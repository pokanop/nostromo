package model

import "fmt"

// Language for various code snippets
type Language int

const (
	// UnsupportedLanguage for unknown snippets
	UnsupportedLanguage Language = iota
	// RubyLanguage for ruby based snippets
	RubyLanguage
	// PythonLanguage for ruby based snippets
	PythonLanguage
	// ShellLanguage for ruby based snippets
	ShellLanguage
)

var languageMap = map[string]Language{
	"ruby":   RubyLanguage,
	"python": PythonLanguage,
	"shell":  ShellLanguage,
}

// Code container for snippet
type Code struct {
	Language Language
	Source   string
}

// NewCode returns a newly initialized code snippet
func NewCode(language string, source string) *Code {
	if l, ok := languageMap[language]; ok {
		return &Code{l, source}
	}
	return &Code{UnsupportedLanguage, source}
}

// CommandString for execution translated from code snippet and language
func (c *Code) CommandString() string {
	switch c.Language {
	case RubyLanguage:
		return fmt.Sprintf("ruby -e '%s'", c.Source)
	case PythonLanguage:
		return fmt.Sprintf("python -c '%s'", c.Source)
	case ShellLanguage:
		return fmt.Sprintf("sh -c %s", c.Source)
	}
	return ""
}
