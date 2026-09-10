package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/version"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		manifests []string
		wantErr   bool
	}{
		{"no core manifest", []string{}, false},
		{"only core manifest", []string{"manifest.yaml"}, false},
		{"multiple manifests", []string{"manifest.yaml", filepath.Join("ships", "manifest2.yaml"), filepath.Join("ships", "manifest3.yaml")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			t.Setenv("NOSTROMO_HOME", baseDir)
			os.MkdirAll(filepath.Join(baseDir, "ships"), 0777)

			paths := absPaths(baseDir, tt.manifests)
			manifests := fakeManifests(paths...)

			// Create temporary spaceport file
			s := model.NewSpaceport(manifests)
			SaveSpaceport(s)

			// Copy manifest to target locations
			src, err := os.ReadFile("../testdata/manifest.yaml")
			if err != nil {
				t.Fatalf("unable to read test manifest: %s", err)
			}
			for i, manifest := range manifests {
				data := src
				// Non core manifests need a unique name to load
				if i > 0 {
					name := strings.TrimSuffix(filepath.Base(manifest.Path), filepath.Ext(manifest.Path))
					data = []byte(strings.Replace(string(src), "name: manifest", "name: "+name, 1))
				}
				if err := os.WriteFile(manifest.Path, data, 0644); err != nil {
					t.Fatalf("unable to write manifest %s: %s", manifest.Path, err)
				}
			}

			c, err := LoadConfig()
			if tt.wantErr && err == nil {
				t.Errorf("want error but got none")
			}

			if len(manifests) > 0 {
				if len(c.spaceport.Manifests()) != len(manifests) {
					t.Errorf("want %d manifests, got %d", len(manifests), len(c.spaceport.Manifests()))
				}

				if len(c.spaceport.CoreManifest().Commands) == 0 {
					t.Errorf("want core manifest with some commands, got %d", len(c.spaceport.CoreManifest().Commands))
				}
			}
		})
	}
}

func TestLoadConfigMigratesLegacySettings(t *testing.T) {
	tests := []struct {
		name            string
		spaceport       string
		wantVerbose     bool
		wantAliasesOnly bool
		wantMode        model.Mode
		wantBackupCount int
		wantTheme       log.ThemeType
	}{
		{"no spaceport", "", true, true, model.IndependentMode, 3, log.EmojiTheme},
		{"legacy spaceport theme", "sequence:\n- manifest\ntheme: 1\n", true, true, model.IndependentMode, 3, log.GrayscaleTheme},
		{"spaceport already migrated", "sequence:\n- manifest\nconfig:\n  verbose: false\n  aliasesonly: false\n  mode: 2\n  backupcount: 7\n  theme: 1\n", false, false, model.ExclusiveMode, 7, log.GrayscaleTheme},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			t.Setenv("NOSTROMO_HOME", baseDir)
			os.MkdirAll(filepath.Join(baseDir, "ships"), 0777)

			// Core manifest with a legacy config block, plus a docked manifest
			// whose config block must be ignored
			src, err := os.ReadFile("../testdata/legacy_manifest.yaml")
			if err != nil {
				t.Fatalf("unable to read legacy manifest: %s", err)
			}
			core := strings.Replace(string(src), "aliasesonly: false", "aliasesonly: true", 1)
			core = strings.Replace(core, "mode: 0", "mode: 1", 1)
			core = strings.Replace(core, "backupcount: 10", "backupcount: 3", 1)
			if err := os.WriteFile(coreManifestPath(), []byte(core), 0644); err != nil {
				t.Fatalf("unable to write core manifest: %s", err)
			}
			docked := strings.Replace(string(src), "name: manifest", "name: docked", 1)
			docked = strings.Replace(docked, "backupcount: 10", "backupcount: 99", 1)
			if err := os.WriteFile(manifestFile("docked"), []byte(docked), 0644); err != nil {
				t.Fatalf("unable to write docked manifest: %s", err)
			}
			if len(tt.spaceport) > 0 {
				if err := os.WriteFile(spaceportFile(), []byte(tt.spaceport), 0644); err != nil {
					t.Fatalf("unable to write spaceport: %s", err)
				}
			}

			c, err := LoadConfig()
			if err != nil {
				t.Fatalf("unable to load config: %s", err)
			}

			check := func(s *model.Spaceport) {
				cfg := s.Config
				if cfg == nil {
					t.Fatalf("want spaceport config, got nil")
				}
				if cfg.Verbose != tt.wantVerbose || cfg.AliasesOnly != tt.wantAliasesOnly || cfg.Mode != tt.wantMode || cfg.BackupCount != tt.wantBackupCount || cfg.Theme != tt.wantTheme {
					t.Errorf("unexpected spaceport config: %+v", cfg)
				}
				for _, m := range s.Manifests() {
					if m != nil && m.Config != nil {
						t.Errorf("want %s manifest config dropped, got %+v", m.Name, m.Config)
					}
				}
			}
			check(c.spaceport)

			// Settings must persist in the spaceport
			if s, err := loadSpaceport(); err != nil {
				t.Fatalf("unable to reload spaceport: %s", err)
			} else {
				check(s)
			}

			// Migrating rewrites the core manifest without its config block,
			// otherwise stale blocks are dropped on the next save
			checkCore := func() {
				saved, err := os.ReadFile(coreManifestPath())
				if err != nil {
					t.Fatalf("unable to read core manifest: %s", err)
				}
				if strings.Contains(string(saved), "config:") {
					t.Errorf("want config block removed from core manifest, got:\n%s", saved)
				}
			}
			if !strings.Contains(tt.spaceport, "config:") {
				checkCore()
			}
			if err := c.Save(); err != nil {
				t.Fatalf("unable to save config: %s", err)
			}
			checkCore()

			// A second load must not migrate again
			c, err = LoadConfig()
			if err != nil {
				t.Fatalf("unable to reload config: %s", err)
			}
			check(c.spaceport)
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name      string
		manifests []string
		wantErr   bool
	}{
		{"no manifest", []string{}, false},
		{"single manifest", []string{"manifest.yaml"}, false},
		{"multiple manifests", []string{"manifest.yaml", filepath.Join("ships", "manifest2.yaml"), filepath.Join("ships", "manifest3.yaml")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			t.Setenv("NOSTROMO_HOME", baseDir)

			// Copy manifest to target locations
			os.MkdirAll(filepath.Join(baseDir, "ships"), 0777)
			src, err := os.ReadFile("../testdata/manifest.yaml")
			if err != nil {
				t.Fatalf("unable to read test manifest: %s", err)
			}
			manifests := fakeManifests(absPaths(baseDir, tt.manifests)...)
			for i, manifest := range manifests {
				data := src
				// Non core manifests need a unique name to load
				if i > 0 {
					name := strings.TrimSuffix(filepath.Base(manifest.Path), filepath.Ext(manifest.Path))
					data = []byte(strings.Replace(string(src), "name: manifest", "name: "+name, 1))
				}
				if err := os.WriteFile(manifest.Path, data, 0644); err != nil {
					t.Fatalf("unable to write manifest %s: %s", manifest.Path, err)
				}
			}

			c, err := NewConfig()
			if tt.wantErr && err == nil {
				t.Errorf("want error but got none")
			}

			if len(manifests) > 0 && len(c.spaceport.Manifests()) != len(manifests) {
				t.Errorf("want %d manifests, got %d", len(manifests), len(c.spaceport.Manifests()))
			}

			if len(c.spaceport.CoreManifest().Commands) != 0 {
				t.Errorf("want core manifest with 0 commands, got %d", len(c.spaceport.CoreManifest().Commands))
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		expErr bool
	}{
		{"invalid path", "", true},
		{"missing path", "/does/not/exist/.nostromo", true},
		{"bad file contents", "../testdata/bad.yaml", true},
		{"bad extension", "../testdata/bad.ext", true},
		{"yaml file format", "../testdata/manifest.yaml", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m, err := Parse(test.path)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr {
				if err != nil {
					t.Errorf("expected no error but got %s", err)
				} else if m == nil {
					t.Errorf("manifest is nil")
				}
				if m.Path != test.path {
					t.Errorf("expected path %s but got %s", test.path, m.Path)
				}
			}
		})
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		expErr bool
	}{
		{"invalid path", fakeConfig("/does/not/exist"), true},
		{"nil manifest", nil, true},
		{"no perms", fakeConfig(filepath.Join(t.TempDir(), "no-perms", ".nostromo")), true},
		{"bad extension", fakeConfig(filepath.Join(t.TempDir(), "bad.ext")), true},
		{"yaml file format", fakeConfig(filepath.Join(t.TempDir(), "manifest.yaml")), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var m *model.Manifest
			if test.config != nil {
				m = test.config.spaceport.CoreManifest()
			}
			err := saveManifest(m, false, 10)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		expErr bool
	}{
		{"invalid path", fakeConfig("/does/not/exist/test.yaml"), true},
		{"valid path", fakeConfig(filepath.Join(t.TempDir(), "test.yaml")), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.config != nil && !test.expErr {
				var m *model.Manifest
				if test.config != nil {
					m = test.config.spaceport.CoreManifest()
				}
				err := test.config.SaveManifest(m, false)
				if err != nil {
					t.Errorf("unable to save temporary manifest: %s", err)
				}
			}
			err := test.config.DeleteManifest(model.CoreManifestName)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"invalid path", "", false},
		{"missing path", "/does/not/exist", false},
		{"valid path", "../testdata/manifest.yaml", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := fakeConfig(test.path)
			if actual := c.Exists(); actual != test.expected {
				t.Errorf("expected: %t, actual: %t", test.expected, actual)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"no key", "", "key not found"},
		{"missing key", "missing", "key not found"},
		{"verbose", "verbose", "true"},
		{"aliasesOnly", "aliasesOnly", "true"},
		{"mode", "mode", "concatenate"},
		{"backupCount", "backupCount", "10"},
		{"theme", "theme", "emoji"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := fakeConfig("")
			c.spaceport.Config.Verbose = true
			c.spaceport.Config.AliasesOnly = true
			if actual := c.Get(test.key); actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expErr   bool
		expected string
	}{
		{"no key", "", "", true, ""},
		{"missing key", "missing", "", true, ""},
		{"verbose empty", "verbose", "", true, ""},
		{"verbose true", "verbose", "true", false, "true"},
		{"verbose false", "verbose", "false", false, "false"},
		{"aliasesOnly empty", "aliasesOnly", "", true, ""},
		{"aliasesOnly true", "aliasesOnly", "true", false, "true"},
		{"aliasesOnly false", "aliasesOnly", "false", false, "false"},
		{"mode concatenate", "mode", "concatenate", false, "concatenate"},
		{"mode independent", "mode", "independent", false, "independent"},
		{"mode exclusive", "mode", "exclusive", false, "exclusive"},
		{"mode invalid", "mode", "invalid", true, ""},
		{"backupCount empty", "backupCount", "", true, ""},
		{"backupCount 5", "backupCount", "5", false, "5"},
		{"backupCount 100", "backupCount", "100", false, "100"},
		{"theme grayscale", "theme", "grayscale", false, "grayscale"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := fakeConfig("")
			err := c.Set(test.key, test.value)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			} else if !test.expErr {
				if actual := c.Get(test.key); actual != test.expected {
					t.Errorf("expected: %s, actual: %s", test.expected, actual)
				}
			}
		})
	}
}

func TestKeys(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{"keys", fakeConfig(""), []string{"verbose", "aliasesOnly", "mode", "backupCount", "theme"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.config.spaceport.Config.Keys(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected map[string]interface{}
	}{
		{
			"keys",
			fakeConfig(""),
			map[string]interface{}{
				"verbose":     false,
				"aliasesOnly": false,
				"mode":        model.ConcatenateMode.String(),
				"backupCount": 10,
				"theme":       "emoji",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.config.spaceport.Config.Fields(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestBaseDir(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string
	}{
		{"default", "", "~/.nostromo"},
		{"override", "~/.config", "~/.config"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.env) > 0 {
				os.Setenv("NOSTROMO_HOME", tt.env)
			}

			if got := BaseDir(); got != tt.want {
				t.Errorf("BaseDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackup(t *testing.T) {
	tests := []struct {
		name        string
		baseDir     string
		backupCount int
		expErr      bool
	}{
		{"invalid path", "/does/not/exist", 1, true},
		{"valid path", "", 1, false},
		{"missing manifest", "", 1, false},
		{"no backups", "", 0, false},
		{"some backups", "", 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := tt.baseDir
			if baseDir == "" {
				baseDir = t.TempDir()
			}
			t.Setenv("NOSTROMO_HOME", baseDir)

			m, err := Parse("../testdata/manifest.yaml")
			if err != nil {
				t.Errorf("failed to Parse manifest: %s", err)
			}

			err = backupManifest(m, tt.backupCount)
			if err != nil {
				if tt.expErr == true {
					return
				}
				t.Errorf("failed to backup: %s", err)
			}

			for i := 0; i < 9; i++ {
				time.Sleep(10 * time.Millisecond)
				backupManifest(m, tt.backupCount)
			}

			backupDir, _ := ensureBackupDir()
			files, err := os.ReadDir(backupDir)
			if err != nil {
				t.Errorf("could not read backup dir: %s", err)
			}
			if len(files) != tt.backupCount {
				t.Errorf("expected %d backup files but got %d", tt.backupCount, len(files))
			}
		})
	}
}

func TestNewCoreManifest(t *testing.T) {
	m, err := NewCoreManifest()
	if err != nil {
		panic(err)
	}
	if m == nil {
		t.Errorf("want not nil, got nil")
	}
}

func TestGetCoreManifestURL(t *testing.T) {
	tests := []struct {
		name    string
		home    string
		want    string
		wantErr bool
	}{
		{"invalid home", "http://test.com/Segment%%2815197306101420000%29.ts", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NOSTROMO_HOME", tt.home)

			u, err := coreManifestURL()
			if tt.wantErr == true && err == nil {
				t.Errorf("want error, got none")
			} else if tt.wantErr {
				return
			}
			if u == nil {
				t.Errorf("want not nil, got nil")
			}
			got := u.String()
			if got != tt.want {
				t.Errorf("want %s, got %s", tt.want, got)
			}
		})
	}
}

func TestManifestURL(t *testing.T) {
	tmpDir := t.TempDir()
	tmpURL, err := fileURL(tmpDir)
	if err != nil {
		t.Fatalf("unable to build temp dir url: %s", err)
	}

	tests := []struct {
		name    string
		target  string
		want    string
		wantErr bool
	}{
		{"empty target", "", "", true},
		{"invalid target", "not a url", "", true},
		{"valid local path", tmpDir, tmpURL.String(), false},
		{"invalid file target", "file:///does/not/exist", "", true},
		{"valid file target", tmpURL.String(), tmpURL.String(), false},
		{"invalid remote target", "https://does/not/exist", "", true},
		{"valid remote target", "https://jsonplaceholder.typicode.com/users", "https://jsonplaceholder.typicode.com/users", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := manifestURL(tt.target)
			if tt.wantErr && err == nil {
				t.Errorf("want error, got none")
			} else if tt.wantErr {
				return
			}

			if u.String() != tt.want {
				t.Errorf("want %s, got %s", tt.want, u.String())
			}
		})
	}
}

func TestNormalizeFileSource(t *testing.T) {
	tmpDir := t.TempDir()
	tmpURL, err := fileURL(tmpDir)
	if err != nil {
		t.Fatalf("unable to build temp dir url: %s", err)
	}
	slashed := filepath.ToSlash(tmpDir)

	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"empty", "", ""},
		{"non file source", "https://foo.com/x.yaml", "https://foo.com/x.yaml"},
		{"legacy single slash", "file:" + slashed, tmpURL.String()},
		{"canonical", "file:///" + strings.TrimPrefix(slashed, "/"), tmpURL.String()},
		{"file host", "file://host/share/x.yaml", "file://host/share/x.yaml"},
	}
	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name   string
			source string
			want   string
		}{"legacy backslashes", "file:" + tmpDir, tmpURL.String()})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeFileSource(tt.source); got != tt.want {
				t.Errorf("want %s, got %s", tt.want, got)
			}
		})
	}
}

func fakeConfig(path string) *Config {
	manifests := []*model.Manifest{fakeManifest(path)}
	c := &Config{model.NewSpaceport(manifests)}
	return c
}

func fakeManifest(path string) *model.Manifest {
	m, err := NewCoreManifest()
	if err != nil {
		panic(err)
	}
	m.Path = path
	m.AddCommand("one.two.three", "command", "", &model.Code{}, false, "concatenate", nil)
	m.AddSubstitution("one.two", "name", "alias")
	return m
}

func fakeManifests(path ...string) []*model.Manifest {
	manifests := []*model.Manifest{}
	for _, path := range path {
		manifests = append(manifests, fakeManifest(path))
	}
	return manifests
}

func absPaths(baseDir string, paths []string) []string {
	abs := []string{}
	for _, path := range paths {
		abs = append(abs, filepath.Join(baseDir, path))
	}
	return abs
}

func init() {
	SetVersion(version.NewInfo("", "", ""))
}
