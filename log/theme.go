package log

import "github.com/logrusorgru/aurora"

type theme interface {
	color(colorLevel, string) aurora.Value
}

type defaultTheme struct{}

func (t *defaultTheme) color(level colorLevel, text string) aurora.Value {
	switch level {
	case debugColorLevel:
		return aurora.Gray(8-1, text).BgGray(16 - 1)
	case infoColorLevel:
		return aurora.Gray(24-1, text)
	case warningColorLevel:
		return aurora.Gray(16-1, text).BgGray(8 - 1)
	case errorColorLevel:
		return aurora.Gray(1-1, text).BgGray(24 - 1)
	default:
		return aurora.White(text)
	}
}

type grayscaleTheme struct{}

func (t *grayscaleTheme) color(level colorLevel, text string) aurora.Value {
	switch level {
	case debugColorLevel:
		return aurora.Gray(8-1, text).BgGray(16 - 1)
	case infoColorLevel:
		return aurora.Gray(24-1, text)
	case warningColorLevel:
		return aurora.Gray(16-1, text).BgGray(8 - 1)
	case errorColorLevel:
		return aurora.Gray(1-1, text).BgGray(24 - 1)
	default:
		return aurora.White(text)
	}
}
