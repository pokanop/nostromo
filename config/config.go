package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pokanop/nostromo/model"
)

var validFileTypes = []string{".nostromo"}

// Config manages working with .nostromo configuration files
// The file format is JSON this just provides convenience around converting
// to a manifest
type Config struct {
	path     string
	manifest *model.Manifest
}

// Parse nostromo config at path into a `Manifest` object
func (c *Config) Parse(path string) (*model.Manifest, error) {
	if !isValidFileType(path) {
		return nil, fmt.Errorf("file must be of type [%s]", strings.Join(validFileTypes, ", "))
	}

	f, err := os.Open(path)
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

	c.path = path
	c.manifest = m

	return m, nil
}

// Save nostromo config to file
func (c *Config) Save() error {
	if len(c.path) == 0 {
		return fmt.Errorf("invalid path to save")
	}

	if c.manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	b, err := json.MarshalIndent(c.manifest, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.path, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func isValidFileType(path string) bool {
	ext := filepath.Ext(path)
	for _, validFileType := range validFileTypes {
		if ext == validFileType {
			return true
		}
	}
	return false
}
