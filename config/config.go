package config

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/pokanop/nostromo/version"
	"github.com/pokanop/nostromo/yamlutil"
	"gopkg.in/yaml.v2"
)

// Path for standard nostromo config
const (
	SystemPrefixDir       = "/usr/local"
	DefaultBaseDir        = "~/.nostromo"
	DefaultConfigFile     = "%s.yaml"
	DefaultManifestsDir   = "ships"
	DefaultBackupsDir     = "cargo"
	DefaultDownloadsDir   = "downloads"
	DefaultCompletionsDir = "completions"
	DefaultManDir         = "man"
)

// URL scheme constants
const (
	FileURLScheme = "file://"
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
	m, legacy, err := parse(coreManifestPath())
	if err != nil {
		return nil, err
	}
	m.Name = model.CoreManifestName
	m.Source = source.String()
	m.Path = path
	manifests := []*model.Manifest{m}

	// Track manifests that must be rewritten in the current format
	stale := map[*model.Manifest]bool{m: legacy}

	// Load synchronized manifests
	docked, legacyDocked := loadManifests()
	manifests = append(manifests, docked...)
	for _, d := range legacyDocked {
		stale[d] = true
	}

	// Load spaceport
	s, err := loadSpaceport()
	if err != nil {
		log.Warning("spaceport not found, creating...")
		s = &model.Spaceport{}
		s.Init()
	}

	s.Import(manifests)
	migrated := s.Migrate()
	s.Link()
	if err := SaveSpaceport(s); err != nil {
		return nil, err
	}

	// Settings were lifted out of the core manifest, rewrite it without them
	if migrated {
		log.Debug("migrating settings into spaceport")
		stale[m] = true
	}

	// Rewrite manifests that were migrated or still used legacy keys
	for _, sm := range manifests {
		if !stale[sm] {
			continue
		}
		log.Debugf("migrating manifest %s to current format\n", sm.Name)
		if err := saveManifest(sm, true, s.Config.BackupCount); err != nil {
			return nil, err
		}
	}

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
	docked, _ := loadManifests()
	manifests = append(manifests, docked...)

	s := model.NewSpaceport(manifests)
	s.Link()

	return &Config{s}, nil
}

// NewManifest creates a new manifest with provided name
func NewManifest(name string) *model.Manifest {
	path := manifestFile(name)
	source := ""
	if u, err := fileURL(path); err != nil {
		log.Warningf("invalid manifest path %s: %s\n", path, err)
	} else {
		source = u.String()
	}
	return model.NewManifest(name, source, path, ver)
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

// ManDir returns the nostromo man dir
func ManDir() string {
	return filepath.Join(pathutil.Abs(BaseDir()), DefaultManDir)
}

// LinkManPages creates a symlink for man pages
func LinkManPages() []error {
	if runtime.GOOS == "windows" {
		log.Debug("skipping man page linking on windows")
		return nil
	}

	mandir := ManDir()
	manpages, err := os.ReadDir(mandir)
	if err != nil {
		return []error{err}
	}

	errors := []error{}
	sysmandir := filepath.Join(SystemPrefixDir, "share", "man", "man1")
	for _, manpage := range manpages {
		name := manpage.Name()
		oldname := filepath.Join(mandir, name)
		newname := filepath.Join(sysmandir, name)

		log.Debugf("adding man page: %s\n", name)
		if err := os.Symlink(oldname, newname); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// UnlinkManPages from system man dir
func UnlinkManPages() []error {
	if runtime.GOOS == "windows" {
		log.Debug("skipping man page unlinking on windows")
		return nil
	}

	sysmandir := filepath.Join(SystemPrefixDir, "share", "man", "man1")
	manpages, err := os.ReadDir(sysmandir)
	if err != nil {
		return []error{err}
	}

	errors := []error{}
	for _, manpage := range manpages {
		name := manpage.Name()
		if strings.HasPrefix(name, "nostromo") {
			path := filepath.Join(sysmandir, name)

			log.Debugf("removing man page: %s\n", name)
			if err := os.Remove(path); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

// Parse nostromo config at path into a `Manifest` object
func Parse(path string) (*model.Manifest, error) {
	m, _, err := parse(path)
	return m, err
}

// parse nostromo config at path into a `Manifest` object, also reporting
// whether the file used legacy lowercase keys and should be saved again
func parse(path string) (*model.Manifest, bool, error) {
	log.Debugf("parsing manifest at %s\n", path)

	f, err := os.Open(pathutil.Abs(path))
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, false, err
	}

	// Initialize manifest with default legacy settings so partial config
	// blocks in older manifests migrate with sensible values
	m := &model.Manifest{
		Config: model.NewConfig(),
	}
	var legacy bool
	ext := filepath.Ext(path)
	if ext == ".yaml" {
		legacy, err = yamlutil.Unmarshal(b, &m)
		if err != nil {
			return nil, false, err
		}
	} else {
		return nil, false, fmt.Errorf("invalid file format: %s", ext)
	}

	// Sanity check parsed manifest
	if len(m.Name) == 0 {
		return nil, false, fmt.Errorf("invalid file content")
	}

	// Normalize file: sources written by older versions (e.g. "file:/path")
	m.Source = normalizeFileSource(m.Source)

	// Only the core manifest carries settings worth migrating into the spaceport
	if !m.IsCore() {
		m.Config = nil
	}

	// Manifest path should match
	m.Path = path

	return m, legacy, nil
}

// SaveSpaceport to nostromo config folder
func SaveSpaceport(s *model.Spaceport) error {
	if s == nil {
		return fmt.Errorf("spaceport is nil")
	}

	log.Debug("saving spaceport")

	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	err = os.WriteFile(spaceportFile(), b, 0644)
	if err != nil {
		return err
	}

	return nil
}

// SaveManifest to nostromo config folder and backup optionally
func (c *Config) SaveManifest(manifest *model.Manifest, backup bool) error {
	return saveManifest(manifest, backup, c.spaceport.Config.BackupCount)
}

// saveManifest to nostromo config folder keeping up to backupCount backups
// if requested
func saveManifest(manifest *model.Manifest, backup bool, backupCount int) error {
	if manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	log.Debugf("saving manifest %s\n", manifest.Name)

	if len(manifest.Path) == 0 {
		return fmt.Errorf("invalid path to save")
	}

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
		if err = backupManifest(manifest, backupCount); err != nil {
			return err
		}
	}

	err = os.WriteFile(pathutil.Abs(manifest.Path), b, 0644)
	if err != nil {
		return err
	}

	return nil
}

// WriteCompletion writes a file to the completions folder
func WriteCompletion(sh, s string) error {
	if len(sh) == 0 || len(s) == 0 {
		return fmt.Errorf("attempt to write 0 length file")
	}

	path := filepath.Join(completionsPath(), fmt.Sprintf("nostromo.%s", sh))
	return os.WriteFile(path, []byte(s), 0644)
}

// Spaceport associated with this config
func (c *Config) Spaceport() *model.Spaceport {
	return c.spaceport
}

// Save nostromo config to file
func (c *Config) Save() error {
	// Update version
	c.spaceport.UpdateVersion(ver)

	// Save spaceport
	if err := SaveSpaceport(c.spaceport); err != nil {
		return err
	}

	// Save core manifest
	if err := c.SaveManifest(c.spaceport.CoreManifest(), true); err != nil {
		return err
	}

	return nil
}

// DeleteManifest from nostromo manfiests folder
func (c *Config) DeleteManifest(name string) error {
	if !c.Exists() {
		return fmt.Errorf("invalid path to remove")
	}

	m := c.spaceport.FindManifest(name)
	if m == nil {
		return fmt.Errorf("no manifest named %s found", name)
	}

	if err := os.Remove(pathutil.Abs(m.Path)); err != nil {
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
	cfg := c.spaceport.Config
	switch key {
	case "verbose":
		return strconv.FormatBool(cfg.Verbose)
	case "aliasesOnly":
		return strconv.FormatBool(cfg.AliasesOnly)
	case "mode":
		return cfg.Mode.String()
	case "backupCount":
		return strconv.FormatInt(int64(cfg.BackupCount), 10)
	case "theme":
		return log.ThemeToString(cfg.Theme)
	}
	return "key not found"
}

// Set setting value for key
func (c *Config) Set(key, value string) error {
	cfg := c.spaceport.Config
	switch key {
	case "verbose":
		verbose, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.Verbose = verbose
		return nil
	case "aliasesOnly":
		aliasesOnly, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.AliasesOnly = aliasesOnly
		return nil
	case "mode":
		if !model.IsModeSupported(value) {
			return fmt.Errorf("invalid mode, supported modes: %s", model.SupportedModes())
		}
		cfg.Mode = model.ModeFromString(value)
		return nil
	case "backupCount":
		count, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return err
		}
		cfg.BackupCount = int(count)
		return nil
	case "theme":
		cfg.Theme = log.ThemeFromString(value)
		return nil
	}
	return fmt.Errorf("key not found")
}

// spaceportFile provides the path for spaceports
func spaceportFile() string {
	return filepath.Join(pathutil.Abs(BaseDir()), fmt.Sprintf(DefaultConfigFile, model.DefaultSpaceportName))
}

// manifestFile joins the manifests path with provided name
func manifestFile(name string) string {
	return filepath.Join(manifestsPath(), fmt.Sprintf(DefaultConfigFile, name))
}

// manifestsPath joins the base directory and the manifest directory
func manifestsPath() string {
	return filepath.Join(pathutil.Abs(BaseDir()), DefaultManifestsDir)
}

// backupsPath joins the base directory and the backups directory
func backupsPath() string {
	return filepath.Join(pathutil.Abs(BaseDir()), DefaultBackupsDir)
}

// completionsPath joins the base directory and the manifest directory
func completionsPath() string {
	return filepath.Join(pathutil.Abs(BaseDir()), DefaultCompletionsDir)
}

func coreManifestFile() string {
	return fmt.Sprintf(DefaultConfigFile, model.CoreManifestName)
}

func coreManifestPath() string {
	return filepath.Join(manifestsPath(), coreManifestFile())
}

func downloadsPath() string {
	return filepath.Join(pathutil.Abs(BaseDir()), DefaultDownloadsDir)
}

// coreManifestURL returns the core manifest URL
func coreManifestURL() (*url.URL, error) {
	return fileURL(coreManifestPath())
}

// fileURL returns a canonical file:// URL for the provided local path.
// On Windows the path gets a leading slash so drive letters produce
// URLs like file:///C:/... which go-getter expects.
func fileURL(path string) (*url.URL, error) {
	p := filepath.ToSlash(pathutil.Abs(path))
	// Make sure the path can be represented inside a URL
	if _, err := url.Parse(p); err != nil {
		return nil, err
	}
	if runtime.GOOS == "windows" && !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return &url.URL{Scheme: "file", Path: p}, nil
}

// normalizeFileSource converts legacy file: source strings stored in
// manifests (e.g. "file:/path" or "file:\\C:\\path") into canonical
// file:// URLs so they continue to resolve correctly.
func normalizeFileSource(source string) string {
	if !strings.HasPrefix(source, "file:") {
		return source
	}

	u, err := url.Parse(source)
	if err != nil {
		return source
	}

	if u.Host != "" && u.Host != "localhost" {
		// File URL on another host (e.g. a network share), leave as-is
		return source
	}

	p := u.Path
	if p == "" {
		p = u.Opaque
	}
	p = strings.ReplaceAll(p, "\\", "/")
	if runtime.GOOS == "windows" && len(p) > 2 && p[0] == '/' && p[2] == ':' {
		p = p[1:]
	}
	if p == "" {
		return source
	}

	n, err := fileURL(p)
	if err != nil {
		return source
	}
	return n.String()
}

// manifestURL verifies target and returns a valid URL or error
//
// For remote URLs, this method makes a HEAD request to confirm the file exists.
func manifestURL(target string) (*url.URL, error) {
	u, err := url.Parse(target)
	if err == nil && u.Scheme == "file" {
		// Check for file path
		p := u.Path
		// file:///C:/... URLs carry a leading slash before the drive letter
		if runtime.GOOS == "windows" && len(p) > 2 && p[0] == '/' && p[2] == ':' {
			p = p[1:]
		}
		p = filepath.Join(u.Host, filepath.FromSlash(p))
		if filepath.IsAbs(p) {
			if _, err = os.Stat(p); err == nil {
				// Local file exists
				return u, nil
			}
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
		return fileURL(target)
	}

	return nil, fmt.Errorf("file not found for target")
}

func loadSpaceport() (*model.Spaceport, error) {
	path := spaceportFile()
	log.Debugf("parsing spaceport at %s\n", path)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	s := &model.Spaceport{}
	ext := filepath.Ext(path)
	if ext == ".yaml" {
		_, err = yamlutil.Unmarshal(b, &s)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid file format: %s", ext)
	}

	// Initialize spaceport
	s.Init()

	return s, nil
}

// loadManifests from the manifests folder, also returning the ones that
// still use legacy keys and should be saved again
func loadManifests() ([]*model.Manifest, []*model.Manifest) {
	manifests := []*model.Manifest{}
	legacy := []*model.Manifest{}
	path := manifestsPath()
	files, err := os.ReadDir(path)
	if err != nil {
		return manifests, legacy
	}

	for _, file := range files {
		// Skip core manifest
		if file.Name() == coreManifestFile() {
			continue
		}

		path := filepath.Join(path, file.Name())
		m, stale, err := parse(path)
		if err != nil {
			log.Warningf("cannot read manifest %s\n", path)
			continue
		}

		// Skip core manifest
		if m.Name != model.CoreManifestName {
			manifests = append(manifests, m)
			if stale {
				legacy = append(legacy, m)
			}
		}
	}

	return manifests, legacy
}

// sanitizeFiles is used for moving config files and fixing up any files during upgrades.
func sanitizeFiles() error {
	if err := pathutil.EnsurePath(manifestsPath()); err != nil {
		return err
	}

	if err := pathutil.EnsurePath(completionsPath()); err != nil {
		return err
	}

	if err := pathutil.EnsurePath(ManDir()); err != nil {
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
