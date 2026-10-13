package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func newTestWindow(t *testing.T) *timerWindow {
	t.Helper()
	a := test.NewTempApp(t)
	w := a.NewWindow("test")
	tw := newTimerWindow(a, w)
	tw.show()
	t.Cleanup(func() {
		if tw.cancel != nil {
			tw.cancel()
		}
	})
	return tw
}

func TestPresetFillsDurationFields(t *testing.T) {
	tw := newTestWindow(t)
	tw.modes.Select(tw.clockTab)

	tw.setDuration(2*time.Hour + 5*time.Minute + 7*time.Second)

	if tw.hours.Text != "2" || tw.minutes.Text != "5" || tw.seconds.Text != "7" {
		t.Fatalf("got %s:%s:%s", tw.hours.Text, tw.minutes.Text, tw.seconds.Text)
	}
	if tw.modes.Selected() != tw.durationTab {
		t.Fatal("preset should switch to the countdown tab")
	}
}

func TestStartAndCancel(t *testing.T) {
	tw := newTestWindow(t)
	tw.setDuration(time.Hour)

	tw.start()
	if tw.cancel == nil || tw.win.Content() != tw.running {
		t.Fatal("countdown did not start")
	}
	if got := tw.remaining.Text; got != "01:00:00" {
		t.Fatalf("remaining = %q", got)
	}

	test.Tap(tw.cancelBtn)
	if tw.cancel != nil || tw.win.Content() != tw.setup {
		t.Fatal("cancel did not return to the setup view")
	}
}

func TestInvalidInputDoesNotStart(t *testing.T) {
	tw := newTestWindow(t)
	tw.hours.SetText("0")
	tw.minutes.SetText("75")
	tw.seconds.SetText("0")

	tw.start()
	if tw.cancel != nil || tw.win.Content() != tw.setup {
		t.Fatal("invalid duration started the countdown")
	}

	tw.modes.Select(tw.clockTab)
	tw.clock.SetText("25:00")
	tw.start()
	if tw.cancel != nil {
		t.Fatal("invalid clock time started the countdown")
	}
}

func TestClockModeDeadline(t *testing.T) {
	tw := newTestWindow(t)
	tw.modes.Select(tw.clockTab)
	tw.clock.SetText("07:30")

	now := time.Date(2026, 5, 1, 8, 0, 0, 0, time.Local)
	got, err := tw.deadline(now)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 5, 2, 7, 30, 0, 0, time.Local); !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFormatTarget(t *testing.T) {
	now := time.Date(2026, 5, 1, 22, 0, 0, 0, time.UTC)
	if got := formatTarget(now, now.Add(time.Hour)); got != "23:00:00" {
		t.Errorf("same day: %q", got)
	}
	if got := formatTarget(now, now.Add(3*time.Hour)); got != "01:00:00 (Sat 2 May)" {
		t.Errorf("next day: %q", got)
	}
}
