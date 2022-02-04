package model

var verbose bool

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

// SetVerbose global flag
func SetVerbose(v bool) {
	verbose = v
}

// IsVerbose check with override and config
func (c *Config) IsVerbose() bool {
	return verbose || c.Verbose
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
