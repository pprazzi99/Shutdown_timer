# Shutdown Timer

A small desktop app that turns your computer off after a countdown or at a set
time of day. It runs on Windows, Linux and macOS.

## Features

- **Countdown**: enter hours, minutes and seconds, or pick a preset
  (15 min, 30 min, 1 h, 2 h).
- **At time**: shut down at a wall-clock time such as `23:30`. If that time has
  already passed today, the shutdown happens tomorrow.
- Desktop notifications **5 minutes** and **30 seconds** before shutdown.
- Remaining time, the planned shutdown time and a progress bar while the timer runs.
- **Cancel** at any time. Closing the window also cancels, after you confirm.
- Remembers the last values you used.

## Download

Get the package for your system from the
[Releases](https://github.com/pprazzi99/shutdown_timer/releases) page:

| System  | File                                     |
|---------|------------------------------------------|
| Windows | `shutdown-timer-v2.0-windows-amd64.zip`    |
| Linux   | `shutdown-timer-v2.0-linux-amd64.tar.gz`   |
| macOS   | `shutdown-timer-v2.0-macos-universal.zip` (Apple Silicon and Intel) |

`SHA256SUMS.txt` lists the checksums of every file.

### Platform notes

- **Windows**: shuts down with `shutdown /s /f /t 0`. Open apps are closed
  without asking, so save your work.
- **Linux**: tries `systemctl poweroff`, then `loginctl poweroff`, then
  `shutdown -h now`. On most desktops your user can power off without `sudo`.
- **macOS**: asks System Events to shut down. The first time, macOS asks you to
  let Shutdown Timer control System Events. The app is not notarized, so on the
  first launch right-click it and choose **Open**.

## Building from source

You need Go (see the version in `go.mod`) and a C compiler, because Fyne uses cgo.

On Linux you also need the graphics headers:

```sh
sudo apt-get install libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev
```

Then build:

```sh
go test ./...
go build -o shutdown-timer .
```

On Windows, add `-ldflags "-H windowsgui"` so no console window opens.