package pathutil

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestExpand(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tests := []struct {
		path     string
		expected string
		env      bool
	}{
		{
			"/foo",
			"/foo",
			false,
		},

		{
			"~/foo",
			filepath.Join(u.HomeDir, "foo"),
			false,
		},

		{
			"",
			"",
			false,
		},

		{
			"~",
			u.HomeDir,
			false,
		},

		{
			"~foo/foo",
			"~foo/foo",
			false,
		},
		{
			"~/foo",
			"~/foo",
			true,
		},
	}

	for _, test := range tests {
		if test.env {
			defer patchEnv("HOME", "")()
		}

		if actual := Expand(test.path); actual != test.expected {
			t.Errorf("expected: %s actual: %s", test.expected, actual)
		}
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
