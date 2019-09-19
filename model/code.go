package model

// Code container for snippet
type Code struct {
	Language string `json:"language"`
	Snippet  string `json:"snippet"`
}

func (c *Code) valid() bool {
	return len(c.Snippet) > 0 && len(c.Language) > 0
}
