package log

import "github.com/logrusorgru/aurora"

type theme interface {
	formatLevel(logLevel, string) aurora.Value
	formatStyle(fieldStyle, string) aurora.Value
	formatRegular(string) aurora.Value
	formatHighlight(string) aurora.Value
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
		return aurora.Gray(24-1, text)
	case warningLevel:
		return aurora.Gray(16-1, text).BgGray(8 - 1)
	case errorLevel:
		return aurora.Gray(1-1, text).BgGray(24 - 1)
	default:
		return aurora.Reset(text)
	}
}

func (t *grayscaleTheme) formatStyle(style fieldStyle, text string) aurora.Value {
	switch style {
	case keyFieldStyle, headerFieldStyle:
		return aurora.Gray(24-1, text)
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
	return aurora.Gray(1-1, text)
}
