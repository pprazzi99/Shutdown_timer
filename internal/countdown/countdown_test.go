package countdown

import (
	"slices"
	"testing"
	"time"
)

var epoch = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func TestTickRemainingAndProgress(t *testing.T) {
	c := New(epoch, epoch.Add(10*time.Second), nil)

	s := c.Tick(epoch.Add(4 * time.Second))
	if s.Remaining != 6*time.Second || s.Done {
		t.Fatalf("got %+v", s)
	}
	if s.Progress < 0.399 || s.Progress > 0.401 {
		t.Fatalf("progress = %v, want 0.4", s.Progress)
	}

	s = c.Tick(epoch.Add(time.Minute))
	if s.Remaining != 0 || !s.Done || s.Progress != 1 {
		t.Fatalf("after deadline got %+v", s)
	}
}

func TestWarningsFireOnce(t *testing.T) {
	c := New(epoch, epoch.Add(10*time.Minute), DefaultWarnings)

	steps := []struct {
		at   time.Duration
		want []time.Duration
	}{
		{at: time.Minute},
		{at: 5 * time.Minute, want: []time.Duration{5 * time.Minute}},
		{at: 5*time.Minute + time.Second},
		{at: 9*time.Minute + 29*time.Second},
		{at: 9*time.Minute + 30*time.Second + 500*time.Millisecond, want: []time.Duration{30 * time.Second}},
		{at: 9*time.Minute + 45*time.Second},
	}
	for _, st := range steps {
		got := c.Tick(epoch.Add(st.at)).Warnings
		if !slices.Equal(got, st.want) {
			t.Fatalf("at %v: warnings %v, want %v", st.at, got, st.want)
		}
	}
}

func TestSkippedTicksStillFireWarnings(t *testing.T) {
	c := New(epoch, epoch.Add(10*time.Minute), DefaultWarnings)
	got := c.Tick(epoch.Add(9*time.Minute + 50*time.Second)).Warnings
	if !slices.Equal(got, DefaultWarnings) {
		t.Fatalf("got %v, want both warnings", got)
	}
}

func TestWarningsLongerThanCountdownAreDropped(t *testing.T) {
	c := New(epoch, epoch.Add(3*time.Minute), DefaultWarnings)
	if got := c.Tick(epoch).Warnings; got != nil {
		t.Fatalf("immediately got %v", got)
	}
	got := c.Tick(epoch.Add(2*time.Minute + 40*time.Second)).Warnings
	if !slices.Equal(got, []time.Duration{30 * time.Second}) {
		t.Fatalf("got %v, want only the 30s warning", got)
	}
}

func TestNoWarningsWhenDone(t *testing.T) {
	c := New(epoch, epoch.Add(10*time.Minute), DefaultWarnings)
	if s := c.Tick(epoch.Add(time.Hour)); !s.Done || s.Warnings != nil {
		t.Fatalf("got %+v", s)
	}
}
