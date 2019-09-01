package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
)

var cfg *Config

// ConfigPath for standard nostromo config
const ConfigPath = "~/.nostromo/manifest"

// Config manages working with nostromo configuration files
// The file format is JSON this just provides convenience around converting
// to a manifest
type Config struct {
	path     string
	Manifest *model.Manifest
}

// NewConfig returns a new nostromo config
func NewConfig(path string, manifest *model.Manifest) *Config {
	return &Config{path, manifest}
}

// Parse nostromo config at path into a `Manifest` object
func Parse(path string) (*Config, error) {
	f, err := os.Open(pathutil.Abs(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var m *model.Manifest
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	m.Link()

	return NewConfig(path, m), nil
}

// Save nostromo config to file
func (c *Config) Save() error {
	if len(c.path) == 0 {
		return fmt.Errorf("invalid path to save")
	}

	if c.Manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	b, err := json.MarshalIndent(c.Manifest, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(pathutil.Abs(c.path), b, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Delete nostromo config file
func (c *Config) Delete() error {
	if !c.Exists() {
		return fmt.Errorf("invalid path to remove")
	}

	if err := os.Remove(pathutil.Abs(c.path)); err != nil {
		return err
	}

	return nil
}

// Exists checks if nostromo config file exists
func (c *Config) Exists() bool {
	if len(c.path) == 0 {
		return false
	}

	_, err := os.Stat(pathutil.Abs(c.path))
	return err == nil
}

// Get setting value from config
func (c *Config) Get(key string) string {
	switch key {
	case "verbose":
		return strconv.FormatBool(c.Manifest.Config.Verbose)
	}
	return "key not found"
}

// Set setting value for key
func (c *Config) Set(key, value string) error {
	switch key {
	case "verbose":
		verbose, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.Manifest.Config.Verbose = verbose
		return nil
	}
	return fmt.Errorf("key not found")
}
