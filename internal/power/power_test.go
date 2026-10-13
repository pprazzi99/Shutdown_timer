package power

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestCommandsKnownSystems(t *testing.T) {
	for _, goos := range []string{"windows", "darwin", "linux"} {
		cmds, err := Commands(goos)
		if err != nil || len(cmds) == 0 {
			t.Errorf("%s: got %v, %v", goos, cmds, err)
		}
	}
	if _, err := Commands("plan9"); !errors.Is(err, ErrUnsupported) {
		t.Errorf("plan9: want ErrUnsupported, got %v", err)
	}
}

func TestWindowsCommandForcesShutdown(t *testing.T) {
	cmds, _ := Commands("windows")
	if got := cmds[0].String(); got != "shutdown /s /f /t 0" {
		t.Fatalf("got %q", got)
	}
}

func TestShutdownFallsBackToNextCommand(t *testing.T) {
	var tried []string
	run := func(_ context.Context, c Command) error {
		tried = append(tried, c.Name)
		if c.Name == "systemctl" {
			return errors.New("no systemd")
		}
		return nil
	}
	if err := shutdown(context.Background(), "linux", run); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(tried, []string{"systemctl", "loginctl"}) {
		t.Fatalf("tried %v", tried)
	}
}

func TestShutdownReportsAllFailures(t *testing.T) {
	boom := errors.New("boom")
	err := shutdown(context.Background(), "linux", func(context.Context, Command) error { return boom })
	if !errors.Is(err, boom) {
		t.Fatalf("want wrapped boom, got %v", err)
	}
}

func TestShutdownStopsWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := shutdown(ctx, "linux", func(context.Context, Command) error {
		calls++
		cancel()
		return context.Canceled
	})
	if err == nil || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}
