package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
	"gopkg.in/yaml.v2"
)

// Path for standard nostromo config
const (
	DefaultManifestFile = "manifest.yaml"
	DefaultBaseDir      = "~/.nostromo"
)

// Config manages working with nostromo configuration files
// The file format is YAML this just provides convenience around converting
// to a manifest
type Config struct {
	path     string
	manifest *model.Manifest
}

// NewConfig returns a new nostromo config
func NewConfig(path string, manifest *model.Manifest) *Config {
	return &Config{path, manifest}
}

// GetBaseDir returns the base directory for nostromo files
func GetBaseDir() string {
	customDir := os.Getenv("NOSTROMO_HOME")

	if customDir != "" {
		return customDir
	}

	return DefaultBaseDir
}

// GetConfigPath joins the base directory and the manifest file
func GetConfigPath() string {
	baseDir := GetBaseDir()
	return filepath.Join(baseDir, DefaultManifestFile)
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
	ext := filepath.Ext(path)
	if ext == ".yaml" {
		err = yaml.Unmarshal(b, &m)
	} else {
		return nil, fmt.Errorf("invalid file format: %s", ext)
	}

	if err != nil {
		return nil, err
	}
	m.Link()

	return NewConfig(path, m), nil
}

// Path of the config
func (c *Config) Path() string {
	return c.path
}

// Manifest associated with this config
func (c *Config) Manifest() *model.Manifest {
	return c.manifest
}

// Save nostromo config to file
func (c *Config) Save() error {
	if len(c.path) == 0 {
		return fmt.Errorf("invalid path to save")
	}

	if c.manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	var b []byte
	var err error
	ext := filepath.Ext(c.path)
	if ext == ".yaml" {
		b, err = yaml.Marshal(c.manifest)
	} else {
		return fmt.Errorf("invalid file format: %s", ext)
	}

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
		return strconv.FormatBool(c.manifest.Config.Verbose)
	case "aliasesOnly":
		return strconv.FormatBool(c.manifest.Config.AliasesOnly)
	case "mode":
		return c.manifest.Config.Mode.String()
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
		c.manifest.Config.Verbose = verbose
		return nil
	case "aliasesOnly":
		aliasesOnly, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.manifest.Config.AliasesOnly = aliasesOnly
		return nil
	case "mode":
		if !model.IsModeSupported(value) {
			return fmt.Errorf("invalid mode, supported modes: %s", model.SupportedModes())
		}
		c.manifest.Config.Mode = model.ModeFromString(value)
		return nil
	}
	return fmt.Errorf("key not found")
}
