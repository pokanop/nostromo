package model

import (
	"reflect"
	"testing"

	"github.com/pokanop/nostromo/log"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig()
	if c.BackupCount != 10 || c.Theme != log.EmojiTheme {
		t.Errorf("unexpected default values for config")
	}
}

func TestConfigKeys(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{"keys", fakeConfig(), []string{"verbose", "aliasesOnly", "mode", "backupCount", "theme"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.config.Keys(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestConfigFields(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected map[string]interface{}
	}{
		{
			"keys",
			fakeConfig(),
			map[string]interface{}{
				"verbose":     true,
				"aliasesOnly": false,
				"mode":        "concatenate",
				"backupCount": 10,
				"theme":       "emoji",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.config.Fields(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func fakeConfig() *Config {
	c := NewConfig()
	c.Verbose = true
	return c
}
