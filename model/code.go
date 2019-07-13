package model

import "fmt"

// Code container for snippet
type Code struct {
	Language string `json:"language"`
	Snippet  string `json:"snippet"`
}

// CommandString for execution translated from code snippet and language
func (c *Code) CommandString() (string, error) {
	switch c.Language {
	case "ruby":
		return fmt.Sprintf("ruby -e '%s'", c.Snippet), nil
	case "python":
		return fmt.Sprintf("python -c '%s'", c.Snippet), nil
	case "shell":
		return fmt.Sprintf("sh -c %s", c.Snippet), nil
	}
	return "", fmt.Errorf("unsupported language for code snippet: %s", c.Language)
}
