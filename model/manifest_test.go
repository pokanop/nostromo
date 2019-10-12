package model

import (
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/pokanop/nostromo/keypath"
	"github.com/pokanop/nostromo/pathutil"
)

func TestManifestAddCommand(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		command  string
		manifest *Manifest
		expErr   bool
		expected int
	}{
		{"empty key path", "", "command", fakeManifest(1, 1), true, 1},
		{"empty command", "0-one-alias", "", fakeManifest(1, 1), false, 1},
		{"single new command", "missing", "command", fakeManifest(0, 0), false, 1},
		{"single existing command", "0-one-alias", "", fakeManifest(1, 1), false, 1},
		{"multi new command", "0-one-alias.0-two-alias.three-alias", "command", fakeManifest(1, 2), false, 3},
		{"multi existing command", "0-one-alias.0-two-alias.0-three-alias", "command", fakeManifest(1, 3), false, 3},
		{"multi all new commands", "one-alias.two-alias.three-alias", "command", fakeManifest(2, 1), false, 5},
		{"multi many new commands", "one-alias.two-alias.three-alias.four-alias", "command", fakeManifest(3, 4), false, 16},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.manifest.AddCommand(test.keyPath, test.command, "", nil, false)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			} else if count := test.manifest.count(); count != test.expected {
				t.Errorf("expected %d commands but got %d", test.expected, count)
			}
		})
	}
}

func TestManifestRemoveCommand(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		manifest *Manifest
		expErr   bool
		expected int
	}{
		{"empty key path", "", fakeManifest(1, 1), true, 1},
		{"only command", "0-one-alias", fakeManifest(1, 1), false, 0},
		{"missing command", "missing", fakeManifest(3, 4), true, 12},
		{"first command", "1-one-alias", fakeManifest(2, 5), false, 5},
		{"middle command", "1-one-alias.1-two-alias.1-three-alias", fakeManifest(2, 5), false, 7},
		{"last command", "1-one-alias.1-two-alias", fakeManifest(5, 2), false, 9},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.manifest.RemoveCommand(test.keyPath)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			} else if count := test.manifest.count(); count != test.expected {
				t.Errorf("expected %d commands but got %d", test.expected, count)
			}
		})
	}
}

func TestManifestAddSubstitution(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		original string
		alias    string
		manifest *Manifest
		expErr   bool
	}{
		{"empty inputs", "", "", "", fakeManifest(1, 1), true},
		{"empty key path", "", "original", "alias", fakeManifest(1, 1), true},
		{"missing key path", "missing", "original", "alias", fakeManifest(1, 1), true},
		{"valid sub", "0-one-alias", "original", "alias", fakeManifest(1, 1), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.manifest.AddSubstitution(test.keyPath, test.original, test.alias)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			} else if !test.expErr {
				cmd := test.manifest.Find(test.keyPath)
				sub := cmd.Subs[test.alias]
				if sub.Name != test.original || sub.Alias != test.alias {
					t.Errorf("expected substitution is incorrect")
				}
			}
		})
	}
}

func TestManifestRemoveSubstitution(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		alias    string
		manifest *Manifest
		expErr   bool
	}{
		{"empty inputs", "", "", fakeManifest(1, 1), true},
		{"empty key path", "", "alias", fakeManifest(1, 1), true},
		{"missing key path", "missing", "alias", fakeManifest(1, 1), true},
		{"valid sub", "0-one-alias", "alias", fakeManifest(1, 1), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.manifest.RemoveSubstitution(test.keyPath, test.alias)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			} else if !test.expErr {
				cmd := test.manifest.Find(test.keyPath)
				sub := cmd.Subs[test.alias]
				if sub != nil {
					t.Errorf("expected substitution to not exist")
				}
			}
		})
	}
}

func TestLink(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
	}{
		{"valid manifest", fakeManifest(2, 4)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, cmd := range test.manifest.Commands {
				cmd.forwardWalk(func(child *Command, stop *bool) {
					child.parent = nil
				})
			}

			test.manifest.Link()

			for _, cmd := range test.manifest.Commands {
				cmd.forwardWalk(func(child *Command, stop *bool) {
					if child != cmd && child.parent == nil {
						t.Errorf("command parent not linked correctly")
					}
				})
			}
		})
	}
}

func TestAsJSON(t *testing.T) {
	tests := []struct {
		name        string
		manifest    *Manifest
		fixturePath string
	}{
		{"valid manifest", fakeManifest(2, 4), "../testdata/manifest.json"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := os.Open(pathutil.Abs(test.fixturePath))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}

			expected := string(b)
			if actual := test.manifest.AsJSON(); actual != expected {
				t.Errorf("expected: %s, actual: %s", expected, actual)
			}
		})
	}
}

func TestAsYAML(t *testing.T) {
	tests := []struct {
		name        string
		manifest    *Manifest
		fixturePath string
	}{
		{"valid manifest", fakeManifest(2, 4), "../testdata/manifest.yaml"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := os.Open(pathutil.Abs(test.fixturePath))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}

			expected := string(b)
			if actual := test.manifest.AsYAML(); actual != expected {
				t.Errorf("expected: %s, actual: %s", expected, actual)
			}
		})
	}
}

func TestManifestExecutionString(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		keyPath  string
		expErr   bool
		expected string
	}{
		{"empty key path", fakeManifest(1, 1), "", true, ""},
		{"missing key path", fakeManifest(1, 1), "missing.key.path", true, ""},
		{"valid single key path", fakeManifest(1, 1), "0-one-alias", false, "0-one"},
		{"valid nth key path", fakeManifest(1, 4), "0-one-alias.0-two-alias.0-three-alias", false, "0-one 0-two 0-three"},
		{"valid repeat sub key path", fakeSimilarManifest(3, 4), "1-one-alias.two-alias.three-alias", false, "1-one two three"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, actual, err := test.manifest.ExecutionString(keypath.Keys(test.keyPath))
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			} else if !test.expErr && test.expected != actual {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestManifestKeys(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		expected []string
	}{
		{"keys", fakeManifest(1, 1), []string{"version", "commands"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.manifest.Keys(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestManifestFields(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		expected map[string]interface{}
	}{
		{
			"keys",
			fakeManifest(1, 1),
			map[string]interface{}{
				"version":  "1.0",
				"commands": "0-one-alias",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.manifest.Fields(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func fakeManifest(n, depth int) *Manifest {
	m := NewManifest()
	m.Config.Verbose = true
	for i := 0; i < n; i++ {
		c := fakeCommandWithPrefix(depth, strconv.Itoa(i)+"-")
		m.Commands[c.Alias] = c
	}
	return m
}

func fakeSimilarManifest(n, depth int) *Manifest {
	m := NewManifest()
	for i := 0; i < n; i++ {
		c := fakeCommand(depth)
		c.Alias = strconv.Itoa(i) + "-" + c.Alias
		c.KeyPath = strconv.Itoa(i) + "-" + c.KeyPath
		c.Name = strconv.Itoa(i) + "-" + c.Name
		m.Commands[c.Alias] = c
	}
	return m
}
