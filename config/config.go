package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/pokanop/nostromo/version"
	"gopkg.in/yaml.v2"
)

// Path for standard nostromo config
const (
	DefaultManifestFile = "manifest.yaml"
	DefaultBaseDir      = "~/.nostromo"
	DefaultBackupsDir   = "backups"
)

var ver *version.Info

// SetVersion should be called before any task to ensure manifest is updated
func SetVersion(v *version.Info) {
	ver = v
}

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

	// Initialize manifest with some defaults
	m := &model.Manifest{
		Config: &model.Config{
			BackupCount: 10,
		},
	}
	ext := filepath.Ext(path)
	if ext == ".yaml" {
		// Attempt to parse legacy manifest versions first
		m = parseV0(b)
		if m == nil {
			err = yaml.Unmarshal(b, &m)
		}
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
	return c.save(true)
}

func (c *Config) save(backup bool) error {
	if len(c.path) == 0 {
		return fmt.Errorf("invalid path to save")
	}

	if c.manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	// Update version
	c.manifest.Version.Update(ver)

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

	// Save backup if requested
	if backup {
		if err = c.Backup(); err != nil {
			return err
		}
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
	case "backupCount":
		return strconv.FormatInt(int64(c.manifest.Config.BackupCount), 10)
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
	case "backupCount":
		count, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return err
		}
		c.manifest.Config.BackupCount = int(count)
		return nil
	}
	return fmt.Errorf("key not found")
}

// Backup manifest at config path based on timestamp
func (c *Config) Backup() error {
	// Before saving backup, prune old files
	c.pruneBackups()

	// Prevent backups if max count is 0
	if c.manifest.Config.BackupCount == 0 {
		return nil
	}

	// Create backups under base dir in a folder
	backupDir, err := ensureBackupDir()
	if err != nil {
		return err
	}

	// Copy existing manifest to backup path
	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UnixNano()/int64(time.Millisecond))
	basename := strings.TrimSuffix(DefaultManifestFile, filepath.Ext(DefaultManifestFile))
	destinationFile := filepath.Join(backupDir, basename+"_"+ts+".yaml")
	sourceFile := pathutil.Abs(GetConfigPath())

	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) pruneBackups() {
	backupDir, err := ensureBackupDir()
	if err != nil {
		return
	}

	// Read all files, sort by timestamp, and drop items > max count
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		log.Warningf("unable to read backup dir: %s\n", err)
		return
	}

	sort.SliceStable(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	// Add one more to backup count since a new backup will be created
	maxCount := c.manifest.Config.BackupCount - 1
	if maxCount < 0 {
		maxCount = 0
	}
	if len(files) > maxCount {
		for _, file := range files[maxCount:] {
			filename := filepath.Join(backupDir, file.Name())
			if err := os.Remove(filename); err != nil {
				log.Warningf("failed to prune backup file %s: %s\n", filename, err)
			}
		}
	}
}

func ensureBackupDir() (string, error) {
	backupDir := filepath.Join(GetBaseDir(), DefaultBackupsDir)
	if err := pathutil.EnsurePath(backupDir); err != nil {
		return "", err
	}
	return pathutil.Abs(backupDir), nil
}

func parseV0(data []byte) *model.Manifest {
	var prev *model.ManifestV0
	if err := yaml.Unmarshal(data, &prev); err != nil {
		return nil
	}

	// Create new manifest with current version and migrate data
	m := model.NewManifest(ver)
	m.Config = prev.Config
	m.Commands = prev.Commands
	return m
}
