package model

// Config model for holding nostromo settings
type Config struct {
	Verbose     bool `json:"verbose"`
	AliasesOnly bool `json:"aliasesOnly"`
}

// Keys as ordered list of fields for logging
func (c *Config) Keys() []string {
	return []string{"verbose", "aliasesOnly"}
}

// Fields interface for logging
func (c *Config) Fields() map[string]interface{} {
	return map[string]interface{}{
		"verbose":     c.Verbose,
		"aliasesOnly": c.AliasesOnly,
	}
}
