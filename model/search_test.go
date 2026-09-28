package model

import (
	"reflect"
	"testing"

	"github.com/pokanop/nostromo/version"
)

func searchManifest(name string, cmds map[string]string, subs map[string][]string) *Manifest {
	m := NewManifest(name, "", "", version.NewInfo("1.0.0", "", ""))
	for keyPath, command := range cmds {
		_, _ = m.AddCommand(keyPath, command, "", nil, false, "")
	}
	for keyPath, sub := range subs {
		_ = m.AddSubstitution(keyPath, sub[0], sub[1])
	}
	return m
}

func searchSpaceport() *Spaceport {
	core := searchManifest(CoreManifestName, map[string]string{
		"dev.install.goenv": "brew install goenv",
		"dev.install.pyenv": "brew install pyenv",
		"dev.env":           "env",
		"git.status":        "git status",
	}, map[string][]string{
		"dev.install": {"goenv", "ge"},
		"git.status":  {"--short", "s"},
	})
	ship := searchManifest("ship", map[string]string{
		"dev.env":       "printenv",
		"deploy.docker": "docker compose up",
	}, map[string][]string{
		"deploy.docker": {"--detach", "d"},
	})
	return NewSpaceport([]*Manifest{core, ship})
}

func keyPaths(results []*SearchResult) []string {
	paths := []string{}
	for _, r := range results {
		path := r.Manifest.Name + ":" + r.Command.KeyPath
		if r.Sub != nil {
			path += "/" + r.Sub.Alias
		}
		paths = append(paths, path)
	}
	return paths
}

func TestSpaceportFindCommands(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		expected []string
	}{
		{"empty key path", "", []string{}},
		{"missing key path", "dev.install.nope", []string{}},
		{"partial key path", "install.goenv", []string{}},
		{"root command", "dev", []string{"manifest:dev", "ship:dev"}},
		{"nested command", "dev.install.goenv", []string{"manifest:dev.install.goenv"}},
		{"same key path in multiple manifests", "dev.env", []string{"manifest:dev.env", "ship:dev.env"}},
		{"case sensitive", "DEV.ENV", []string{}},
	}

	sp := searchSpaceport()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := keyPaths(sp.FindCommands(test.keyPath))
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, actual)
			}
		})
	}
}

func TestSpaceportSearch(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		expCmds []string
		expSubs []string
	}{
		{"no matches", "missing", []string{}, []string{}},
		{"alias match", "pyenv", []string{"manifest:dev.install.pyenv"}, []string{}},
		{"command name match", "compose", []string{"ship:deploy.docker"}, []string{}},
		{"key path segment match", "install", []string{"manifest:dev.install", "manifest:dev.install.goenv", "manifest:dev.install.pyenv"}, []string{}},
		{"key path with dots match", "install.go", []string{"manifest:dev.install.goenv"}, []string{}},
		{"case insensitive", "GOENV", []string{"manifest:dev.install.goenv"}, []string{"manifest:dev.install/ge"}},
		{"across manifests in sequence order", "env", []string{"manifest:dev.env", "manifest:dev.install.goenv", "manifest:dev.install.pyenv", "ship:dev.env"}, []string{"manifest:dev.install/ge"}},
		{"substitution alias match", "ge", []string{}, []string{"manifest:dev.install/ge"}},
		{"substitution name match", "--", []string{}, []string{"manifest:git.status/s", "ship:deploy.docker/d"}},
	}

	sp := searchSpaceport()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmds, subs := sp.Search(test.query)
			if actual := keyPaths(cmds); !reflect.DeepEqual(actual, test.expCmds) {
				t.Errorf("expected commands %v but got %v", test.expCmds, actual)
			}
			if actual := keyPaths(subs); !reflect.DeepEqual(actual, test.expSubs) {
				t.Errorf("expected substitutions %v but got %v", test.expSubs, actual)
			}
		})
	}
}

func TestManifestSearchSorted(t *testing.T) {
	m := searchManifest("sorted", map[string]string{
		"c.b": "cmd", "c.a": "cmd", "b": "cmd", "a.z.y": "cmd",
	}, map[string][]string{
		"b": {"cmd", "z"},
	})
	_ = m.AddSubstitution("b", "cmd", "a")

	for i := 0; i < 10; i++ {
		cmds, subs := m.Search("")
		if actual := keyPaths(cmds); !reflect.DeepEqual(actual, []string{"sorted:a", "sorted:a.z", "sorted:a.z.y", "sorted:b", "sorted:c", "sorted:c.a", "sorted:c.b"}) {
			t.Fatalf("unexpected command order %v", actual)
		}
		if actual := keyPaths(subs); !reflect.DeepEqual(actual, []string{"sorted:b/a", "sorted:b/z"}) {
			t.Fatalf("unexpected substitution order %v", actual)
		}
	}
}

func TestSubstitutionFields(t *testing.T) {
	s := &Substitution{Name: "original", Alias: "short"}
	if keys := s.Keys(); !reflect.DeepEqual(keys, []string{"alias", "name"}) {
		t.Errorf("unexpected keys %v", keys)
	}
	fields := s.Fields()
	if fields["alias"] != "short" || fields["name"] != "original" {
		t.Errorf("unexpected fields %v", fields)
	}
}
