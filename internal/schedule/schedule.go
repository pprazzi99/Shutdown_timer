// Package schedule converts user input into a shutdown deadline.
package schedule

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalid is wrapped by every validation error returned by this package.
	ErrInvalid = errors.New("invalid input")
	// ErrZeroDuration is returned when the requested countdown is empty.
	ErrZeroDuration = fmt.Errorf("%w: duration must be greater than zero", ErrInvalid)
)

// MaxDuration caps the countdown so a typo cannot schedule a shutdown weeks ahead.
const MaxDuration = 99 * time.Hour

// ParseDuration builds a duration from hours, minutes and seconds fields.
// Empty fields count as zero. Minutes and seconds must be in 0-59.
func ParseDuration(hours, minutes, seconds string) (time.Duration, error) {
	h, err := parseField("hours", hours, 0, int(MaxDuration/time.Hour))
	if err != nil {
		return 0, err
	}
	m, err := parseField("minutes", minutes, 0, 59)
	if err != nil {
		return 0, err
	}
	s, err := parseField("seconds", seconds, 0, 59)
	if err != nil {
		return 0, err
	}

	d := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
	switch {
	case d <= 0:
		return 0, ErrZeroDuration
	case d > MaxDuration:
		return 0, fmt.Errorf("%w: duration must not exceed %d hours", ErrInvalid, int(MaxDuration/time.Hour))
	}
	return d, nil
}

// ParseClock parses a wall-clock time in 24-hour "HH:MM" format.
func ParseClock(s string) (hour, minute int, err error) {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return 0, 0, fmt.Errorf("%w: time must be in HH:MM format (e.g. 23:30)", ErrInvalid)
	}
	return t.Hour(), t.Minute(), nil
}

// NextOccurrence returns the first moment strictly after now at which the local
// clock reads hour:minute. If that time has already passed today, tomorrow is used.
func NextOccurrence(now time.Time, hour, minute int) time.Time {
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !target.After(now) {
		target = time.Date(now.Year(), now.Month(), now.Day()+1, hour, minute, 0, 0, now.Location())
	}
	return target
}

// FormatRemaining renders a duration as HH:MM:SS, rounding up to whole seconds
// so the display never shows 00:00:00 while time is still left.
func FormatRemaining(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int64((d + time.Second - 1) / time.Second)
	return fmt.Sprintf("%02d:%02d:%02d", total/3600, total/60%60, total%60)
}

func parseField(name, value string, minVal, maxVal int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < minVal || n > maxVal {
		return 0, fmt.Errorf("%w: %s must be a whole number between %d and %d", ErrInvalid, name, minVal, maxVal)
	}
	return n, nil
}
