package log

import "fmt"

var (
	currentTheme theme
)

// colorLevel for representing onscreen text
type colorLevel int

const (
	debugColorLevel colorLevel = iota
	infoColorLevel
	warningColorLevel
	errorColorLevel
)

// Debug logs a debug message
func Debug(text string) {
	fmt.Println(currentTheme.color(debugColorLevel, text))
}

// Debugf logs a debug message
func Debugf(format string, args ...interface{}) {
	fmt.Print(currentTheme.color(debugColorLevel, fmt.Sprintf(format, args...)))
}

// Info logs an info message
func Info(text string) {
	fmt.Println(currentTheme.color(infoColorLevel, text))
}

// Infof logs a debug message
func Infof(format string, args ...interface{}) {
	fmt.Print(currentTheme.color(infoColorLevel, fmt.Sprintf(format, args...)))
}

// Warning logs a warning message
func Warning(text string) {
	fmt.Println(currentTheme.color(warningColorLevel, text))
}

// Warningf logs a debug message
func Warningf(format string, args ...interface{}) {
	fmt.Print(currentTheme.color(warningColorLevel, fmt.Sprintf(format, args...)))
}

// Error logs an error message
func Error(text string) {
	fmt.Println(currentTheme.color(errorColorLevel, text))
}

// Errorf logs a debug message
func Errorf(format string, args ...interface{}) {
	fmt.Print(currentTheme.color(errorColorLevel, fmt.Sprintf(format, args...)))
}

func init() {
	currentTheme = &grayscaleTheme{}
}
