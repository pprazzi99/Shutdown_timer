// Package power turns the computer off using the operating system's own tools.
package power

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Command is a program invocation that powers the machine off.
type Command struct {
	Name string
	Args []string
}

func (c Command) String() string { return strings.Join(append([]string{c.Name}, c.Args...), " ") }

// ErrUnsupported is returned on operating systems without a known shutdown command.
var ErrUnsupported = errors.New("shutdown is not supported on this operating system")

// Commands lists the shutdown commands for goos, in the order they should be tried.
func Commands(goos string) ([]Command, error) {
	switch goos {
	case "windows":
		// /f closes running applications without asking, as the timer is meant
		// to work unattended.
		return []Command{{Name: "shutdown", Args: []string{"/s", "/f", "/t", "0"}}}, nil
	case "darwin":
		// Asking System Events does not need root, unlike /sbin/shutdown.
		return []Command{{Name: "osascript", Args: []string{"-e", `tell application "System Events" to shut down`}}}, nil
	case "linux", "freebsd", "openbsd", "netbsd":
		return []Command{
			{Name: "systemctl", Args: []string{"poweroff"}},
			{Name: "loginctl", Args: []string{"poweroff"}},
			{Name: "shutdown", Args: []string{"-h", "now"}},
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, goos)
	}
}

// Runner executes a single command.
type Runner func(ctx context.Context, cmd Command) error

// Shutdown powers off the current machine.
func Shutdown(ctx context.Context) error {
	return shutdown(ctx, runtime.GOOS, execRunner)
}

func shutdown(ctx context.Context, goos string, run Runner) error {
	cmds, err := Commands(goos)
	if err != nil {
		return err
	}
	var errs []error
	for _, c := range cmds {
		err := run(ctx, c)
		if err == nil {
			return nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", c, err))
		if ctx.Err() != nil {
			break
		}
	}
	return fmt.Errorf("shutdown failed: %w", errors.Join(errs...))
}

func execRunner(ctx context.Context, c Command) error {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if msg := strings.TrimSpace(string(out)); err != nil && msg != "" {
		return fmt.Errorf("%w: %s", err, msg)
	}
	return err
}
