package config

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/version"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		manifests []*model.Manifest
		wantErr   bool
	}{
		{"no core manifest", fakeManifests(), false},
		{"only core manifest", fakeManifests("/tmp/nostromo/manifest.yaml"), false},
		{"multiple manifests", fakeManifests("/tmp/nostromo/manifest.yaml", "/tmp/nostromo/ships/manifest2.yaml", "/tmp/nostromo/ships/manifest3.yaml"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NOSTROMO_HOME", "/tmp/nostromo")
			defer os.Unsetenv("NOSTROMO_HOME")
			os.MkdirAll("/tmp/nostromo/ships", 0777)
			defer os.RemoveAll("/tmp/nostromo")

			// Create temporary spaceport file
			s := model.NewSpaceport(tt.manifests)
			saveSpaceport(s)

			// Copy manifest to target locations
			src, _ := os.Open("../testdata/manifest.yaml")
			defer src.Close()
			for _, manifest := range tt.manifests {
				dest, _ := os.Create(manifest.Path)
				defer dest.Close()
				io.Copy(dest, src)
				dest.Sync()
			}

			c, err := LoadConfig()
			if tt.wantErr && err == nil {
				t.Errorf("want error but got none")
			}

			if len(tt.manifests) > 0 {
				if len(c.spaceport.Manifests()) != len(tt.manifests) {
					t.Errorf("want %d manifests, got %d", len(tt.manifests), len(c.spaceport.Manifests()))
				}

				if len(c.spaceport.CoreManifest().Commands) == 0 {
					t.Errorf("want core manifest with some commands, got %d", len(c.spaceport.CoreManifest().Commands))
				}
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name      string
		manifests []*model.Manifest
		wantErr   bool
	}{
		{"no manifest", fakeManifests(), false},
		{"single manifest", fakeManifests("/tmp/nostromo/manifest.yaml"), false},
		{"multiple manifests", fakeManifests("/tmp/nostromo/manifest.yaml", "/tmp/nostromo/ships/manifest2.yaml", "/tmp/nostromo/ships/manifest3.yaml"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NOSTROMO_HOME", "/tmp/nostromo")
			defer os.Unsetenv("NOSTROMO_HOME")

			// Copy manifest to target locations
			os.MkdirAll("/tmp/nostromo/ships", 0777)
			defer os.RemoveAll("/tmp/nostromo")
			src, _ := os.Open("../testdata/manifest.yaml")
			defer src.Close()
			for _, manifest := range tt.manifests {
				dest, _ := os.Create(manifest.Path)
				defer dest.Close()
				io.Copy(dest, src)
				dest.Sync()
			}

			c, err := NewConfig()
			if tt.wantErr && err == nil {
				t.Errorf("want error but got none")
			}

			if len(tt.manifests) > 0 && len(c.spaceport.Manifests()) != len(tt.manifests) {
				t.Errorf("want %d manifests, got %d", len(tt.manifests), len(c.spaceport.Manifests()))
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
			m, err := parse(test.path)
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
		{"no perms", fakeConfig("/tmp/no-perms/.nostromo"), true},
		{"bad extension", fakeConfig("/tmp/bad.ext"), true},
		{"yaml file format", fakeConfig("/tmp/manifest.yaml"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var m *model.Manifest
			if test.config != nil {
				m = test.config.spaceport.CoreManifest()
			}
			err := saveManifest(m, false)
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
		{"valid path", fakeConfig("/tmp/test.yaml"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.config != nil && !test.expErr {
				var m *model.Manifest
				if test.config != nil {
					m = test.config.spaceport.CoreManifest()
				}
				err := saveManifest(m, false)
				if err != nil {
					t.Errorf("unable to save temporary manifest: %s", err)
				}
			}
			err := test.config.Delete()
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := fakeConfig("")
			m := c.spaceport.CoreManifest()
			m.Config.Verbose = true
			m.Config.AliasesOnly = true
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
		{"keys", fakeConfig(""), []string{"verbose", "aliasesOnly", "mode", "backupCount"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.config.spaceport.CoreManifest().Config.Keys(); !reflect.DeepEqual(actual, test.expected) {
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
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.config.spaceport.CoreManifest().Config.Fields(); !reflect.DeepEqual(actual, test.expected) {
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
		{"valid path", "/tmp", 1, false},
		{"missing manifest", "/tmp", 1, false},
		{"no backups", "/tmp", 0, false},
		{"some backups", "/tmp", 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NOSTROMO_HOME", tt.baseDir)

			m, err := parse("../testdata/manifest.yaml")
			if err != nil {
				t.Errorf("failed to parse manifest: %s", err)
			}

			manifests := []*model.Manifest{m}
			c := &Config{model.NewSpaceport(manifests)}

			c.spaceport.CoreManifest().Config.BackupCount = tt.backupCount
			err = backupManifest(m)
			if err != nil {
				if tt.expErr == true {
					return
				}
				t.Errorf("failed to backup: %s", err)
			}

			for i := 0; i < 9; i++ {
				time.Sleep(10 * time.Millisecond)
				backupManifest(m)
			}

			backupDir, _ := ensureBackupDir()
			files, err := ioutil.ReadDir(backupDir)
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
		{"valid home", "/tmp", "file:///tmp/ships/manifest.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NOSTROMO_HOME", tt.home)

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
	tests := []struct {
		name    string
		target  string
		want    string
		wantErr bool
	}{
		{"empty target", "", "", true},
		{"invalid target", "not a url", "", true},
		{"valid target", "/tmp", "file:///tmp", false},
		{"invalid file target", "file:///does/not/exist", "", true},
		{"valid file target", "file:///tmp", "file:///tmp", false},
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
	m.AddCommand("one.two.three", "command", "", &model.Code{}, false, "concatenate")
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

func init() {
	SetVersion(version.NewInfo("", "", ""))
}
