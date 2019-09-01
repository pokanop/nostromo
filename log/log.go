package log

import (
	"fmt"
	"strings"
)

type options struct {
	theme   theme
	verbose bool
}

var opt *options

type logLevel int

const (
	debugLevel logLevel = iota
	infoLevel
	warningLevel
	errorLevel
)

// Regular log for body style text
func Regular(a ...interface{}) {
	fmt.Println(opt.theme.formatRegular(joined(a...)))
}

// Highlight log as highlighted text
func Highlight(a ...interface{}) {
	fmt.Println(opt.theme.formatHighlight(joined(a...)))
}

// Debug logs a debug message
func Debug(a ...interface{}) {
	fmt.Println(opt.theme.formatLevel(debugLevel, "debug:"), joined(a...))
}

// Debugf logs a debug message
func Debugf(format string, a ...interface{}) {
	fmt.Print(opt.theme.formatLevel(debugLevel, "debug:"), fmt.Sprintf(format, a...))
}

// Info logs an info message
func Info(a ...interface{}) {
	fmt.Println(opt.theme.formatLevel(infoLevel, "info:"), joined(a...))
}

// Infof logs a debug message
func Infof(format string, a ...interface{}) {
	fmt.Print(opt.theme.formatLevel(infoLevel, "info:"), fmt.Sprintf(format, a...))
}

// Warning logs a warning message
func Warning(a ...interface{}) {
	fmt.Println(opt.theme.formatLevel(warningLevel, "warning:"), fmt.Sprint(a...))
}

// Warningf logs a debug message
func Warningf(format string, a ...interface{}) {
	fmt.Print(opt.theme.formatLevel(warningLevel, "warning:"), fmt.Sprintf(format, a...))
}

// Error logs an error message
func Error(a ...interface{}) {
	fmt.Println(opt.theme.formatLevel(errorLevel, "error:"), fmt.Sprint(a...))
}

// Errorf logs a debug message
func Errorf(format string, a ...interface{}) {
	fmt.Print(opt.theme.formatLevel(errorLevel, "error:"), fmt.Sprintf(format, a...))
}

// SetOptions for logger
func SetOptions(verbose bool) {
	opt.verbose = verbose
}

func joined(a ...interface{}) string {
	sargs := []string{}
	for _, arg := range a {
		sargs = append(sargs, fmt.Sprint(arg))
	}
	return strings.Join(sargs, " ")
}

func init() {
	opt = &options{
		theme:   &defaultTheme{},
		verbose: false,
	}
}
