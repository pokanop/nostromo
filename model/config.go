package model

import "github.com/pokanop/nostromo/log"

var verbose bool

// Config model for holding nostromo settings
type Config struct {
	Verbose     bool          `json:"verbose" yaml:"verbose"`
	AliasesOnly bool          `json:"aliasesOnly" yaml:"aliasesOnly"`
	Mode        Mode          `json:"mode" yaml:"mode"`
	BackupCount int           `json:"backupCount" yaml:"backupCount"`
	Theme       log.ThemeType `json:"theme" yaml:"theme"`
}

// Create a new config model with default values
func NewConfig() *Config {
	return &Config{
		BackupCount: 10,
		Theme:       log.EmojiTheme,
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
	return []string{"verbose", "aliasesOnly", "mode", "backupCount", "theme"}
}

// Fields interface for logging
func (c *Config) Fields() map[string]interface{} {
	return map[string]interface{}{
		"verbose":     c.Verbose,
		"aliasesOnly": c.AliasesOnly,
		"mode":        c.Mode.String(),
		"backupCount": c.BackupCount,
		"theme":       log.ThemeToString(c.Theme),
	}
}
