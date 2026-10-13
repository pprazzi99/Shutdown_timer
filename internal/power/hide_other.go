//go:build !windows

package power

import "os/exec"

func hideWindow(*exec.Cmd) {}
