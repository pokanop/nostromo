package pathutil

import (
	"os"
	"os/user"
	"path/filepath"
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
			defer patchEnv("HOME", test.home)()

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

	defer patchEnv("HOME", u.HomeDir)()

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
		{"valid path", "/tmp/pathutil_tests/valid_path", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				err := os.RemoveAll("/tmp/pathutil_tests")
				if err != nil {
					t.Fatalf("unable to clean up tmp folder: %s", err)
				}
			}()

			err := EnsurePath(test.path)
			if err != nil && !test.expErr {
				t.Errorf("expected no error but got %s", err)
			} else if err == nil && test.expErr {
				t.Errorf("expected error but got none")
			}
		})
	}
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
