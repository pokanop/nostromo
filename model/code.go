package model

// Code container for snippet
type Code struct {
	Language string `json:"language" yaml:"language"`
	Snippet  string `json:"snippet" yaml:"snippet"`
}

func (c *Code) valid() bool {
	return c != nil && len(c.Snippet) > 0 && len(c.Language) > 0
}
