package banner

import (
	"os"

	figure "github.com/common-nighthawk/go-figure"
)

const (
	// Font is the figlet font used to render the banner; fits in 80 columns.
	Font = "slant"

	// Tagline printed below the ascii art.
	Tagline = "CLI for building powerful aliases and tools"
)

// Lines renders text as ascii-art rows.
func Lines(text string) []string {
	return figure.NewFigure(text, Font, false).Slicify()
}

// IsTerminal reports whether f is attached to a character device (TTY).
func IsTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
