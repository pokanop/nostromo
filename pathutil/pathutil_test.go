package pathutil

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAbs(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tests := []struct {
		name     string
		path     string
		home     string
		expected string
	}{
		{"empty path", "", u.HomeDir, wd},
		{"home not set", "~/foo", "", filepath.Join(wd, "~/foo")},
		{"valid path", "~/foo", u.HomeDir, filepath.Join(u.HomeDir, "foo")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer patchHomeEnv(test.home)()

			if actual := Abs(test.path); actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestExpand(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tests := []struct {
		path     string
		expected string
	}{
		{"/foo", "/foo"},
		{"~/foo", filepath.Join(u.HomeDir, "foo")},
		{"", ""},
		{"~", u.HomeDir},
		{"~foo/foo", "~foo/foo"},
	}

	defer patchHomeEnv(u.HomeDir)()

	for _, test := range tests {
		if actual := Expand(test.path); actual != test.expected {
			t.Errorf("expected: %s actual: %s", test.expected, actual)
		}
	}
}

func TestEnsurePath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		expErr bool
	}{
		{"valid path", "valid_path", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := EnsurePath(filepath.Join(t.TempDir(), test.path))
			if err != nil && !test.expErr {
				t.Errorf("expected no error but got %s", err)
			} else if err == nil && test.expErr {
				t.Errorf("expected error but got none")
			}
		})
	}
}

// patchHomeEnv sets the home directory environment variable(s) for the
// current platform: HOME on unix, USERPROFILE/HOMEDRIVE/HOMEPATH on Windows.
func patchHomeEnv(value string) func() {
	deferFuncs := []func(){}
	for _, key := range homeEnvVars() {
		deferFuncs = append(deferFuncs, patchEnv(key, value))
	}
	return func() {
		for _, f := range deferFuncs {
			f()
		}
	}
}

func homeEnvVars() []string {
	if runtime.GOOS == "windows" {
		return []string{"USERPROFILE", "HOMEDRIVE", "HOMEPATH"}
	}
	return []string{"HOME"}
}

func patchEnv(key, value string) func() {
	bck := os.Getenv(key)
	deferFunc := func() {
		os.Setenv(key, bck)
	}

	if value != "" {
		os.Setenv(key, value)
	} else {
		os.Unsetenv(key)
	}

	return deferFunc
}
