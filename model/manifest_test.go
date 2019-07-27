package model

import (
	"strconv"
	"testing"

	"github.com/kr/pretty"
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
			err := test.manifest.AddCommand(test.keyPath, test.command)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			} else if count := test.manifest.count(); count != test.expected {
				t.Errorf("expected %d commands but got %d", test.expected, count)
				pretty.Println(test.manifest)
			}
		})
	}
}

func fakeManifest(n, depth int) *Manifest {
	m := NewManifest()
	for i := 0; i < n; i++ {
		c := fakeCommandWithPrefix(depth, strconv.Itoa(i)+"-")
		m.Commands[c.Alias] = c
	}
	return m
}
