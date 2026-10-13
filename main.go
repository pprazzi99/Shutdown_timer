// Command shutdown-timer is a desktop app that turns the computer off after a
// countdown or at a chosen time of day.
package main

import "github.com/pprazzi99/shutdown_timer/internal/ui"

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	ui.Run(version)
}
