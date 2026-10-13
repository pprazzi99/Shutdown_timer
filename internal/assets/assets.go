// Package assets embeds static resources used by the GUI.
package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icon.svg
var iconSVG []byte

// Icon is the application icon.
var Icon = fyne.NewStaticResource("icon.svg", iconSVG)
