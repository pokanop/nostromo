package config

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
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
	DefaultBaseDir      = "~/.nostromo"
	DefaultManifestFile = "%s.yaml"
	DefaultManifestsDir = "ships"
	DefaultBackupsDir   = "cargo"
)

// URL scheme constants
const (
	FileURLScheme  = "file://"
	GitURLScheme   = "git://"
	HTTPURLScheme  = "http://"
	HTTPSURLScheme = "https://"
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
	spaceport *model.Spaceport
}

func LoadConfig() (*Config, error) {
	// Sanitize files
	if err := sanitizeFiles(); err != nil {
		return nil, err
	}

	// Load core manifest
	source, err := coreManifestURL()
	if err != nil {
		return nil, err
	}
	path := coreManifestPath()
	m, err := parse(coreManifestPath())
	if err != nil {
		return nil, err
	}
	m.Name = model.CoreManifestName
	m.Source = source.String()
	m.Path = path
	manifests := []*model.Manifest{m}

	// Load synchronized manifests
	manifests = append(manifests, loadManifests()...)

	s := &model.Spaceport{Manifests: manifests}
	s.Link()

	return &Config{s}, nil
}

// NewConfig returns a new nostromo config
func NewConfig() (*Config, error) {
	// Create core manifest
	m, err := NewCoreManifest()
	if err != nil {
		return nil, err
	}
	manifests := []*model.Manifest{m}

	// Load synchronized manifests
	manifests = append(manifests, loadManifests()...)

	s := &model.Spaceport{Manifests: manifests}
	s.Link()

	return &Config{s}, nil
}

func loadManifests() []*model.Manifest {
	manifests := []*model.Manifest{}
	path := manifestsPath()
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return manifests
	}

	for _, file := range files {
		path := filepath.Join(path, file.Name())
		m, err := parse(path)
		if err != nil {
			log.Warningf("cannot read manifest %s", path)
			continue
		}

		// Skip core manifest
		if m.Name != model.CoreManifestName {
			manifests = append(manifests, m)
		}
	}

	return manifests
}

// NewCoreManifest creates a new core manifest
func NewCoreManifest() (*model.Manifest, error) {
	m, err := coreManifestURL()
	if err != nil {
		return nil, err
	}
	return model.NewManifest(model.CoreManifestName, m.String(), coreManifestPath(), ver), nil
}

// BaseDir returns the base directory for nostromo files
func BaseDir() string {
	customDir := os.Getenv("NOSTROMO_HOME")

	if customDir != "" {
		return customDir
	}

	return DefaultBaseDir
}

// manifestsPath joins the base directory and the manifest directory
func manifestsPath() string {
	return filepath.Join(pathutil.Abs(BaseDir()), DefaultManifestsDir)
}

// backupsPath joins the base directory and the backups directory
func backupsPath() string {
	return filepath.Join(pathutil.Abs(BaseDir()), DefaultBackupsDir)
}

func coreManifestFile() string {
	return fmt.Sprintf(DefaultManifestFile, model.CoreManifestName)
}

func coreManifestPath() string {
	return filepath.Join(manifestsPath(), coreManifestFile())
}

// coreManifestURL returns the core manifest URL
func coreManifestURL() (*url.URL, error) {
	rawURL := filepath.Join(FileURLScheme, coreManifestPath())
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// manifestURL verifies target and returns a valid URL or error
//
// For remote URLs, this method makes a HEAD request to confirm the file exists.
func manifestURL(target string) (*url.URL, error) {
	u, err := url.Parse(target)
	if err == nil && u.Scheme == "file" {
		// Check for file path
		p := filepath.Join(u.Host, u.Path)
		if _, err = os.Stat(p); !os.IsNotExist(err) {
			// Local file exists
			return u, nil
		}
		// file:// scheme was given but local file does not exist
		return nil, fmt.Errorf("file not found for target")
	} else if err == nil && strings.HasPrefix(u.Scheme, "http") {
		// Check for remote path
		resp, err := http.Head(u.String())
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("remote file not found")
		}
		return u, nil
	}

	// Check for local path
	if _, err = os.Stat(target); !os.IsNotExist(err) {
		// Return url with file scheme
		u, err = url.Parse(filepath.Join(FileURLScheme, target))
		if err != nil {
			return nil, err
		}
		return u, nil
	}

	return nil, fmt.Errorf("file not found for target")
}

// parse nostromo config at path into a `Manifest` object
func parse(path string) (*model.Manifest, error) {
	log.Debugf("parsing manifest at %s\n", path)
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
		err = yaml.Unmarshal(b, &m)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid file format: %s", ext)
	}

	// Manifest path should match
	m.Path = path

	return m, nil
}

// Spaceport associated with this config
func (c *Config) Spaceport() *model.Spaceport {
	return c.spaceport
}

func (c *Config) Load(path string) error {
	// Convert path to url
	u, err := manifestURL(path)
	if err != nil {
		return err
	}

	// Handle local manifest
	if u.Scheme == "file" {
		m, err := parse(u.Path)
		if err != nil {
			return err
		}
		c.spaceport.AddManifest(m)
		return nil
	}

	// TODO: Handle remote manifest
	return nil
}

// Save nostromo config to file
func (c *Config) Save() error {
	return c.save(c.spaceport.CoreManifest(), true)
}

func (c *Config) save(manifest *model.Manifest, backup bool) error {
	if manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	if c.spaceport == nil {
		return fmt.Errorf("spaceport is nil")
	}

	if len(manifest.Path) == 0 {
		return fmt.Errorf("invalid path to save")
	}

	// Update version
	c.spaceport.UpdateVersion(ver)

	var b []byte
	var err error
	ext := filepath.Ext(manifest.Path)
	if ext == ".yaml" {
		b, err = yaml.Marshal(manifest)
	} else {
		return fmt.Errorf("invalid file format: %s", ext)
	}

	if err != nil {
		return err
	}

	// Save backup if requested
	if backup {
		if err = c.Backup(manifest); err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(pathutil.Abs(manifest.Path), b, 0644)
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

	if err := os.Remove(pathutil.Abs(c.spaceport.CoreManifest().Path)); err != nil {
		return err
	}

	return nil
}

// Exists checks if nostromo config file exists
func (c *Config) Exists() bool {
	if c.spaceport == nil {
		return false
	}

	m := c.spaceport.CoreManifest()
	if len(m.Path) == 0 {
		return false
	}

	_, err := os.Stat(pathutil.Abs(m.Path))
	return err == nil
}

// Get setting value from config
func (c *Config) Get(key string) string {
	m := c.spaceport.CoreManifest()
	switch key {
	case "verbose":
		return strconv.FormatBool(m.Config.Verbose)
	case "aliasesOnly":
		return strconv.FormatBool(m.Config.AliasesOnly)
	case "mode":
		return m.Config.Mode.String()
	case "backupCount":
		return strconv.FormatInt(int64(m.Config.BackupCount), 10)
	}
	return "key not found"
}

// Set setting value for key
func (c *Config) Set(key, value string) error {
	m := c.spaceport.CoreManifest()
	switch key {
	case "verbose":
		verbose, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		m.Config.Verbose = verbose
		return nil
	case "aliasesOnly":
		aliasesOnly, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		m.Config.AliasesOnly = aliasesOnly
		return nil
	case "mode":
		if !model.IsModeSupported(value) {
			return fmt.Errorf("invalid mode, supported modes: %s", model.SupportedModes())
		}
		m.Config.Mode = model.ModeFromString(value)
		return nil
	case "backupCount":
		count, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return err
		}
		m.Config.BackupCount = int(count)
		return nil
	}
	return fmt.Errorf("key not found")
}

// Backup manifest at config path based on timestamp
func (c *Config) Backup(m *model.Manifest) error {
	// Before saving backup, prune old files
	c.pruneBackups(m)

	// Prevent backups if max count is 0
	if m.Config.BackupCount == 0 {
		return nil
	}

	// Create backups under base dir in a folder
	backupDir, err := ensureBackupDir()
	if err != nil {
		return err
	}

	// Copy existing manifest to backup path
	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UnixNano()/int64(time.Millisecond))
	destinationFile := filepath.Join(backupDir, m.Name+"_"+ts+".yaml")
	sourceFile := pathutil.Abs(m.Path)

	// Check if manifest exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return nil
	}

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

func (c *Config) pruneBackups(m *model.Manifest) {
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

	// Filter file list for matching names
	matches := []fs.FileInfo{}
	for _, file := range files {
		if isMatch, _ := regexp.MatchString(fmt.Sprintf("%s_\\d+\\.yaml", m.Name), file.Name()); isMatch {
			matches = append(matches, file)
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].ModTime().After(matches[j].ModTime())
	})

	// Add one more to backup count since a new backup will be created
	maxCount := m.Config.BackupCount - 1
	if maxCount < 0 {
		maxCount = 0
	}
	if len(matches) > maxCount {
		for _, file := range matches[maxCount:] {
			filename := filepath.Join(backupDir, file.Name())
			if err := os.Remove(filename); err != nil {
				log.Warningf("failed to prune backup file %s: %s\n", filename, err)
			}
		}
	}
}

func ensureBackupDir() (string, error) {
	backupDir := backupsPath()
	if err := pathutil.EnsurePath(backupDir); err != nil {
		return "", err
	}
	return pathutil.Abs(backupDir), nil
}

// sanitizeFiles is used for moving config files and fixing up any files during upgrades.
func sanitizeFiles() error {
	if err := pathutil.EnsurePath(manifestsPath()); err != nil {
		return err
	}

	// The core manifest was previously in the root folder of NOSTROMO_HOME.
	// Check there first and move to new location if needed.
	oldPath := filepath.Join(pathutil.Abs(BaseDir()), model.CoreManifestName+".yaml")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		// File exists, migrate
		log.Warning("migrating core manifest")
		err := os.Rename(oldPath, coreManifestPath())
		if err != nil {
			return err
		}
	}

	// Backups dir name change
	oldPath = filepath.Join(pathutil.Abs(BaseDir()), "backups")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		// Folder exists, migrate
		log.Warning("migrating backups folder")
		err := os.Rename(oldPath, backupsPath())
		if err != nil {
			return err
		}
	}

	return nil
}
