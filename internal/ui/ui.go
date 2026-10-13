// Package ui implements the Fyne desktop interface.
package ui

import (
	"context"
	"fmt"
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/pprazzi99/shutdown_timer/internal/assets"
	"github.com/pprazzi99/shutdown_timer/internal/countdown"
	"github.com/pprazzi99/shutdown_timer/internal/power"
	"github.com/pprazzi99/shutdown_timer/internal/schedule"
)

const (
	appID = "io.github.pprazzi99.shutdowntimer"

	// tickInterval is shorter than a second so the display never visibly lags.
	tickInterval    = 200 * time.Millisecond
	shutdownTimeout = 30 * time.Second

	prefHours   = "hours"
	prefMinutes = "minutes"
	prefSeconds = "seconds"
	prefClock   = "clock"
	prefMode    = "mode"
)

var presets = []struct {
	label string
	d     time.Duration
}{
	{"15 min", 15 * time.Minute},
	{"30 min", 30 * time.Minute},
	{"1 h", time.Hour},
	{"2 h", 2 * time.Hour},
}

// Run opens the main window and blocks until the app exits.
func Run(version string) {
	a := app.NewWithID(appID)
	a.SetIcon(assets.Icon)

	w := a.NewWindow("Shutdown Timer " + version)
	newTimerWindow(a, w).show()
	w.Resize(fyne.NewSize(400, 0))
	w.CenterOnScreen()
	w.ShowAndRun()
}

type timerWindow struct {
	app fyne.App
	win fyne.Window

	hours, minutes, seconds, clock *widget.Entry
	modes                          *container.AppTabs
	durationTab, clockTab          *container.TabItem
	setup                          fyne.CanvasObject

	remaining *canvas.Text
	target    *widget.Label
	progress  *widget.ProgressBar
	cancelBtn *widget.Button
	running   fyne.CanvasObject

	// cancel stops the active countdown; nil when idle. Only touched on the UI goroutine.
	cancel context.CancelFunc
}

func newTimerWindow(a fyne.App, w fyne.Window) *timerWindow {
	t := &timerWindow{app: a, win: w}
	t.setup = t.buildSetup()
	t.running = t.buildRunning()
	t.loadPrefs()
	w.SetCloseIntercept(t.confirmClose)
	return t
}

func (t *timerWindow) show() { t.win.SetContent(t.setup) }

func (t *timerWindow) buildSetup() fyne.CanvasObject {
	t.hours = numberEntry("0")
	t.minutes = numberEntry("0")
	t.seconds = numberEntry("0")

	fields := container.NewGridWithColumns(3,
		labeled("Hours", t.hours),
		labeled("Minutes", t.minutes),
		labeled("Seconds", t.seconds),
	)

	presetRow := container.NewGridWithColumns(len(presets))
	for _, p := range presets {
		presetRow.Add(widget.NewButton(p.label, func() { t.setDuration(p.d) }))
	}

	t.clock = widget.NewEntry()
	t.clock.SetPlaceHolder("HH:MM, e.g. 23:30")
	t.clock.OnSubmitted = func(string) { t.start() }
	hint := widget.NewLabel("If this time has already passed today, the computer shuts down tomorrow.")
	hint.Wrapping = fyne.TextWrapWord
	hint.Importance = widget.LowImportance

	t.durationTab = container.NewTabItemWithIcon("Countdown", theme.HistoryIcon(),
		container.NewPadded(container.NewVBox(fields, widget.NewSeparator(), presetRow)))
	t.clockTab = container.NewTabItemWithIcon("At time", theme.CalendarIcon(),
		container.NewPadded(container.NewVBox(labeled("Shut down at", t.clock), hint)))
	t.modes = container.NewAppTabs(t.durationTab, t.clockTab)

	start := widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), t.start)
	start.Importance = widget.HighImportance

	return container.NewPadded(container.NewBorder(nil, start, nil, nil, t.modes))
}

func (t *timerWindow) buildRunning() fyne.CanvasObject {
	t.remaining = canvas.NewText("00:00:00", theme.Color(theme.ColorNameForeground))
	t.remaining.TextSize = 56
	t.remaining.TextStyle = fyne.TextStyle{Monospace: true}
	t.remaining.Alignment = fyne.TextAlignCenter

	t.target = widget.NewLabel("")
	t.target.Alignment = fyne.TextAlignCenter

	t.progress = widget.NewProgressBar()
	t.progress.TextFormatter = func() string { return "" }

	t.cancelBtn = widget.NewButtonWithIcon("Cancel shutdown", theme.CancelIcon(), t.stop)
	t.cancelBtn.Importance = widget.DangerImportance

	title := widget.NewLabelWithStyle("Time until shutdown", fyne.TextAlignCenter, fyne.TextStyle{})
	return container.NewPadded(container.NewVBox(title, t.remaining, t.target, t.progress, layoutSpacer(), t.cancelBtn))
}

func (t *timerWindow) setDuration(d time.Duration) {
	t.hours.SetText(strconv.Itoa(int(d / time.Hour)))
	t.minutes.SetText(strconv.Itoa(int(d / time.Minute % 60)))
	t.seconds.SetText(strconv.Itoa(int(d / time.Second % 60)))
	t.modes.Select(t.durationTab)
}

// deadline turns the form contents into the moment the computer should shut down.
func (t *timerWindow) deadline(now time.Time) (time.Time, error) {
	if t.modes.Selected() == t.clockTab {
		h, m, err := schedule.ParseClock(t.clock.Text)
		if err != nil {
			return time.Time{}, err
		}
		return schedule.NextOccurrence(now, h, m), nil
	}
	d, err := schedule.ParseDuration(t.hours.Text, t.minutes.Text, t.seconds.Text)
	if err != nil {
		return time.Time{}, err
	}
	return now.Add(d), nil
}

func (t *timerWindow) start() {
	if t.cancel != nil {
		return
	}
	now := time.Now()
	deadline, err := t.deadline(now)
	if err != nil {
		dialog.ShowError(err, t.win)
		return
	}
	t.savePrefs()

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	cd := countdown.New(now, deadline, countdown.DefaultWarnings)

	t.target.SetText("Shutdown at " + formatTarget(now, deadline))
	t.cancelBtn.Enable()
	t.render(cd.Tick(now))
	t.win.SetContent(t.running)

	go t.run(ctx, cd)
}

func (t *timerWindow) stop() {
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
	t.show()
}

// run drives the countdown off the UI goroutine and powers off when it ends.
func (t *timerWindow) run(ctx context.Context, cd *countdown.Countdown) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s := cd.Tick(now)
			for _, w := range s.Warnings {
				t.app.SendNotification(fyne.NewNotification(
					"Shutdown in "+formatWarning(w),
					"Save your work. Open Shutdown Timer to cancel."))
			}
			fyne.Do(func() { t.render(s) })
			if s.Done {
				t.powerOff(ctx)
				return
			}
		}
	}
}

func (t *timerWindow) powerOff(ctx context.Context) {
	fyne.Do(func() {
		t.target.SetText("Shutting down…")
		t.cancelBtn.Disable()
	})

	runCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()
	// A failure caused by the user cancelling at the last moment is not worth reporting.
	if err := power.Shutdown(runCtx); err != nil && ctx.Err() == nil {
		fyne.Do(func() {
			t.stop()
			dialog.ShowError(err, t.win)
		})
	}
}

func (t *timerWindow) render(s countdown.State) {
	t.remaining.Text = schedule.FormatRemaining(s.Remaining)
	t.remaining.Color = theme.Color(theme.ColorNameForeground)
	if s.Remaining <= time.Minute {
		t.remaining.Color = theme.Color(theme.ColorNameError)
	}
	t.remaining.Refresh()
	t.progress.SetValue(s.Progress)
}

func (t *timerWindow) confirmClose() {
	if t.cancel == nil {
		t.win.Close()
		return
	}
	dialog.ShowConfirm("Cancel shutdown?",
		"Closing the window cancels the scheduled shutdown.", func(ok bool) {
			if ok {
				t.stop()
				t.win.Close()
			}
		}, t.win)
}

func (t *timerWindow) loadPrefs() {
	p := t.app.Preferences()
	t.hours.SetText(p.StringWithFallback(prefHours, "0"))
	t.minutes.SetText(p.StringWithFallback(prefMinutes, "30"))
	t.seconds.SetText(p.StringWithFallback(prefSeconds, "0"))
	t.clock.SetText(p.StringWithFallback(prefClock, "23:00"))
	if p.Int(prefMode) == 1 {
		t.modes.Select(t.clockTab)
	}
}

func (t *timerWindow) savePrefs() {
	p := t.app.Preferences()
	p.SetString(prefHours, t.hours.Text)
	p.SetString(prefMinutes, t.minutes.Text)
	p.SetString(prefSeconds, t.seconds.Text)
	p.SetString(prefClock, t.clock.Text)
	mode := 0
	if t.modes.Selected() == t.clockTab {
		mode = 1
	}
	p.SetInt(prefMode, mode)
}

func numberEntry(placeholder string) *widget.Entry {
	e := widget.NewEntry()
	e.SetPlaceHolder(placeholder)
	return e
}

func labeled(text string, obj fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), obj)
}

func layoutSpacer() fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(0, theme.Padding()*2))
	return r
}

// formatTarget shows the deadline's clock time, adding the day when it is not today.
func formatTarget(now, deadline time.Time) string {
	clock := deadline.Format("15:04:05")
	y1, m1, d1 := now.Date()
	y2, m2, d2 := deadline.Date()
	if y1 == y2 && m1 == m2 && d1 == d2 {
		return clock
	}
	return clock + deadline.Format(" (Mon 2 Jan)")
}

func formatWarning(d time.Duration) string {
	if d >= time.Minute {
		return fmt.Sprintf("%d minutes", int(d/time.Minute))
	}
	return fmt.Sprintf("%d seconds", int(d/time.Second))
}
