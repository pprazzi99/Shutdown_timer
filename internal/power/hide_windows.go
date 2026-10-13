package power

import (
	"os/exec"
	"syscall"
)

// hideWindow stops a console window from flashing up when a GUI app runs a command.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
