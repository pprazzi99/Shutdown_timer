package schedule

import (
	"errors"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		h, m, s string
		want    time.Duration
		wantErr bool
	}{
		{name: "all fields", h: "1", m: "30", s: "15", want: time.Hour + 30*time.Minute + 15*time.Second},
		{name: "empty fields are zero", m: "5", want: 5 * time.Minute},
		{name: "whitespace", h: " 2 ", want: 2 * time.Hour},
		{name: "upper bounds", h: "98", m: "59", s: "59", want: 98*time.Hour + 59*time.Minute + 59*time.Second},
		{name: "max hours", h: "99", want: MaxDuration},
		{name: "zero", h: "0", m: "0", s: "0", wantErr: true},
		{name: "all empty", wantErr: true},
		{name: "minutes 60", m: "60", wantErr: true},
		{name: "seconds 60", s: "60", wantErr: true},
		{name: "negative", h: "-1", wantErr: true},
		{name: "not a number", h: "abc", wantErr: true},
		{name: "fraction", m: "1.5", wantErr: true},
		{name: "too long", h: "99", s: "1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.h, tt.m, tt.s)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("want ErrInvalid, got %v (value %v)", err, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseClock(t *testing.T) {
	h, m, err := ParseClock(" 23:05 ")
	if err != nil || h != 23 || m != 5 {
		t.Fatalf("got %d:%d, %v", h, m, err)
	}
	for _, bad := range []string{"", "24:00", "12:60", "1230", "noon", "12:30:00"} {
		if _, _, err := ParseClock(bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("ParseClock(%q): want ErrInvalid, got %v", bad, err)
		}
	}
}

func TestNextOccurrence(t *testing.T) {
	loc := time.FixedZone("test", 2*3600)
	now := time.Date(2026, 3, 31, 22, 15, 30, 0, loc)

	tests := []struct {
		name         string
		hour, minute int
		want         time.Time
	}{
		{"later today", 23, 30, time.Date(2026, 3, 31, 23, 30, 0, 0, loc)},
		{"earlier rolls to tomorrow", 7, 0, time.Date(2026, 4, 1, 7, 0, 0, 0, loc)},
		{"same minute rolls to tomorrow", 22, 15, time.Date(2026, 4, 1, 22, 15, 0, 0, loc)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NextOccurrence(now, tt.hour, tt.minute); !got.Equal(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatRemaining(t *testing.T) {
	tests := map[time.Duration]string{
		-time.Second:     "00:00:00",
		0:                "00:00:00",
		time.Millisecond: "00:00:01",
		59 * time.Second: "00:00:59",
		time.Hour + 2*time.Minute + 3*time.Second: "01:02:03",
		99 * time.Hour: "99:00:00",
	}
	for in, want := range tests {
		if got := FormatRemaining(in); got != want {
			t.Errorf("FormatRemaining(%v) = %q, want %q", in, got, want)
		}
	}
}
