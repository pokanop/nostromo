package log

import (
	"github.com/logrusorgru/aurora/v3"
)

type ThemeType int

const (
	DefaultTheme ThemeType = iota
	GrayscaleTheme
	EmojiTheme
)

type theme interface {
	formatLevel(logLevel, string) aurora.Value
	formatStyle(fieldStyle, string) aurora.Value
	formatRegular(string) aurora.Value
	formatHighlight(string) aurora.Value
}

// ThemeToString conversion from ThemeType to string
func ThemeToString(theme ThemeType) string {
	switch theme {
	case DefaultTheme:
		return "default"
	case GrayscaleTheme:
		return "grayscale"
	case EmojiTheme:
		return "emoji"
	}
	return "unknown"
}

// ThemeToString conversion from string to ThemeType
func ThemeFromString(theme string) ThemeType {
	switch theme {
	case "default":
		return DefaultTheme
	case "grayscale":
		return GrayscaleTheme
	case "emoji":
		return EmojiTheme
	}
	return DefaultTheme
}

type defaultTheme struct{}

func (t *defaultTheme) formatLevel(level logLevel, text string) aurora.Value {
	switch level {
	case debugLevel:
		return aurora.Green(text)
	case infoLevel:
		return aurora.Blue(text)
	case warningLevel:
		return aurora.Yellow(text)
	case errorLevel:
		return aurora.Red(text)
	default:
		return aurora.Reset(text)
	}
}

func (t *defaultTheme) formatStyle(style fieldStyle, text string) aurora.Value {
	switch style {
	case keyFieldStyle, headerFieldStyle:
		return aurora.Blue(text)
	case valueFieldStyle, cellFieldStyle:
		fallthrough
	default:
		return aurora.Reset(text)
	}
}

func (t *defaultTheme) formatRegular(text string) aurora.Value {
	return aurora.Reset(text)
}

func (t *defaultTheme) formatHighlight(text string) aurora.Value {
	return aurora.Blue(text)
}

type grayscaleTheme struct{}

func (t *grayscaleTheme) formatLevel(level logLevel, text string) aurora.Value {
	switch level {
	case debugLevel:
		return aurora.Gray(8-1, text).BgGray(16 - 1)
	case infoLevel:
		return aurora.Gray(1-1, text).BgGray(24 - 1)
	case warningLevel:
		return aurora.Gray(20-1, text).BgGray(4 - 1)
	case errorLevel:
		return aurora.Gray(24-1, text).BgGray(1 - 1)
	default:
		return aurora.Reset(text)
	}
}

func (t *grayscaleTheme) formatStyle(style fieldStyle, text string) aurora.Value {
	switch style {
	case keyFieldStyle, headerFieldStyle:
		return aurora.Gray(20-1, text).BgGray(4 - 1)
	case valueFieldStyle, cellFieldStyle:
		fallthrough
	default:
		return aurora.Reset(text)
	}
}

func (t *grayscaleTheme) formatRegular(text string) aurora.Value {
	return aurora.Reset(text)
}

func (t *grayscaleTheme) formatHighlight(text string) aurora.Value {
	return aurora.Gray(20-1, text).BgGray(4 - 1)
}

type emojiTheme struct{}

func (t *emojiTheme) formatLevel(level logLevel, text string) aurora.Value {
	switch level {
	case debugLevel:
		return aurora.Reset("🔍")
	case infoLevel:
		return aurora.Reset("💡")
	case warningLevel:
		return aurora.Reset("😮")
	case errorLevel:
		return aurora.Reset("🧨")
	default:
		return aurora.Reset("🤔")
	}
}

func (t *emojiTheme) formatStyle(style fieldStyle, text string) aurora.Value {
	switch style {
	case keyFieldStyle, headerFieldStyle:
		return aurora.Blue(text)
	case valueFieldStyle, cellFieldStyle:
		fallthrough
	default:
		return aurora.Reset(text)
	}
}

func (t *emojiTheme) formatRegular(text string) aurora.Value {
	return aurora.Reset(text)
}

func (t *emojiTheme) formatHighlight(text string) aurora.Value {
	return aurora.Blue("🚀 " + text)
}
