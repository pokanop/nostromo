package model

// Config model for holding nostromo settings
type Config struct {
	Verbose     bool `json:"verbose"`
	AliasesOnly bool `json:"aliasesOnly"`
	Mode        Mode `json:"mode"`
	BackupCount int  `json:"backupCount"`
}

// Create a new config model with default values
func NewConfig() *Config {
	return &Config{
		BackupCount: 10,
	}
}

// Keys as ordered list of fields for logging
func (c *Config) Keys() []string {
	return []string{"verbose", "aliasesOnly", "mode", "backupCount"}
}

// Fields interface for logging
func (c *Config) Fields() map[string]interface{} {
	return map[string]interface{}{
		"verbose":     c.Verbose,
		"aliasesOnly": c.AliasesOnly,
		"mode":        c.Mode.String(),
		"backupCount": c.BackupCount,
	}
}
