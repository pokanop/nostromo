package log

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"strings"
)

type options struct {
	theme   theme
	verbose bool
	echo    bool
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
	if opt.echo {
		echo(a...)
		return
	}
	fmt.Println(opt.theme.formatRegular(joined(a...)))
}

// Regularf log for body style text
func Regularf(format string, a ...interface{}) {
	if opt.echo {
		echof(format, a...)
		return
	}
	fmt.Print(opt.theme.formatRegular(fmt.Sprintf(format, a...)))
}

// Highlight log as highlighted text
func Highlight(a ...interface{}) {
	if opt.echo {
		echo(a...)
		return
	}
	fmt.Println(opt.theme.formatHighlight(joined(a...)))
}

// Highlightf log as highlighted text
func Highlightf(format string, a ...interface{}) {
	if opt.echo {
		echof(format, a...)
		return
	}
	fmt.Print(opt.theme.formatHighlight(fmt.Sprintf(format, a...)))
}

func Bold(a ...interface{}) {
	if opt.echo {
		echo(a...)
		return
	}
	fmt.Println(aurora.Bold(joined(a...)))
}

func Boldf(format string, a ...interface{}) {
	if opt.echo {
		echof(format, a...)
		return
	}
	fmt.Print(aurora.Bold(fmt.Sprintf(format, a...)))
}

// Debug logs a debug message
func Debug(a ...interface{}) {
	if !opt.verbose {
		return
	}
	if opt.echo {
		echo(a...)
		return
	}
	fmt.Println(opt.theme.formatLevel(debugLevel, "debug:"), joined(a...))
}

// Debugf logs a debug message
func Debugf(format string, a ...interface{}) {
	if !opt.verbose {
		return
	}
	if opt.echo {
		echof(format, a...)
		return
	}
	fmt.Print(opt.theme.formatLevel(debugLevel, "debug: "), fmt.Sprintf(format, a...))
}

// Info logs an info message
func Info(a ...interface{}) {
	if opt.echo {
		echo(a...)
		return
	}
	fmt.Println(opt.theme.formatLevel(infoLevel, "info:"), joined(a...))
}

// Infof logs a debug message
func Infof(format string, a ...interface{}) {
	if opt.echo {
		echof(format, a...)
		return
	}
	fmt.Print(opt.theme.formatLevel(infoLevel, "info: "), fmt.Sprintf(format, a...))
}

// Warning logs a warning message
func Warning(a ...interface{}) {
	if opt.echo {
		echo(a...)
		return
	}
	fmt.Println(opt.theme.formatLevel(warningLevel, "warning:"), joined(a...))
}

// Warningf logs a debug message
func Warningf(format string, a ...interface{}) {
	if opt.echo {
		echof(format, a...)
		return
	}
	fmt.Print(opt.theme.formatLevel(warningLevel, "warning: "), fmt.Sprintf(format, a...))
}

// Error logs an error message
func Error(a ...interface{}) {
	if opt.echo {
		echo(a...)
		return
	}
	fmt.Println(opt.theme.formatLevel(errorLevel, "error:"), joined(a...))
}

// Errorf logs a debug message
func Errorf(format string, a ...interface{}) {
	if opt.echo {
		echof(format, a...)
		return
	}
	fmt.Print(opt.theme.formatLevel(errorLevel, "error: "), fmt.Sprintf(format, a...))
}

// Print is effectively a pass-through to fmt.Print
func Print(a ...interface{}) {
	fmt.Print(a...)
}

func echo(a ...interface{}) {
	fmt.Printf("echo \"%s\";", joined(a...))
}

func echof(format string, a ...interface{}) {
	fmt.Printf("echo \"%s\";", fmt.Sprintf(format, a...))
}

// SetVerbose for logger
func SetVerbose(verbose bool) {
	opt.verbose = verbose
}

// SetEcho mode for logger
func SetEcho(echo bool) {
	opt.echo = echo
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
