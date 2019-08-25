package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// Abs tries to return absolute path if possible or original value
func Abs(path string) string {
	abspath, err := filepath.Abs(Expand(path))
	if err != nil {
		return path
	}
	return abspath
}

// HomeDir returns the home directory for the executing user.
func HomeDir() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	return "", fmt.Errorf("missing env var HOME")
}

// Expand expands the path to include the home directory if the path
// is prefixed with `~`. If it isn't prefixed with `~`, the path is
// returned as-is.
func Expand(path string) string {
	if len(path) == 0 || path[0] != '~' || (len(path) > 1 && path[1] != '/' && path[1] != '\\') {
		return path
	}

	dir, err := HomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(dir, path[1:])
}

// EnsurePath creates all directories up to and including path
func EnsurePath(path string) error {
	return os.MkdirAll(Abs(path), 0755)
}
