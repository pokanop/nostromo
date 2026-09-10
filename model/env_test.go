package model

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name    string
		pairs   []string
		want    map[string]string
		wantErr bool
	}{
		{"none", nil, map[string]string{}, false},
		{"single", []string{"FOO=bar"}, map[string]string{"FOO": "bar"}, false},
		{"multiple", []string{"FOO=bar", "BAZ=qux"}, map[string]string{"FOO": "bar", "BAZ": "qux"}, false},
		{"empty value", []string{"FOO="}, map[string]string{"FOO": ""}, false},
		{"value with equals", []string{"FOO=a=b"}, map[string]string{"FOO": "a=b"}, false},
		{"last wins", []string{"FOO=1", "FOO=2"}, map[string]string{"FOO": "2"}, false},
		{"missing equals", []string{"FOO"}, nil, true},
		{"invalid key", []string{"FOO BAR=1"}, nil, true},
		{"empty key", []string{"=1"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnv(tt.pairs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyEnv(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		dotenv     []string
		changes    EnvChanges
		wantEnv    map[string]string
		wantDotenv []string
	}{
		{"noop", nil, nil, EnvChanges{}, nil, nil},
		{"set on nil", nil, nil, EnvChanges{Set: map[string]string{"A": "1"}}, map[string]string{"A": "1"}, nil},
		{"set merges", map[string]string{"A": "1"}, nil, EnvChanges{Set: map[string]string{"B": "2"}}, map[string]string{"A": "1", "B": "2"}, nil},
		{"set overrides", map[string]string{"A": "1"}, nil, EnvChanges{Set: map[string]string{"A": "2"}}, map[string]string{"A": "2"}, nil},
		{"unset", map[string]string{"A": "1", "B": "2"}, nil, EnvChanges{Unset: []string{"A"}}, map[string]string{"B": "2"}, nil},
		{"unset last clears", map[string]string{"A": "1"}, nil, EnvChanges{Unset: []string{"A"}}, nil, nil},
		{"unset missing", map[string]string{"A": "1"}, nil, EnvChanges{Unset: []string{"Z"}}, map[string]string{"A": "1"}, nil},
		{"set then unset", nil, nil, EnvChanges{Set: map[string]string{"A": "1"}, Unset: []string{"A"}}, nil, nil},
		{"dotenv nil keeps", nil, []string{".env"}, EnvChanges{}, nil, []string{".env"}},
		{"dotenv replaces", nil, []string{".env"}, EnvChanges{Dotenv: []string{"a.env", "b.env"}}, nil, []string{"a.env", "b.env"}},
		{"dotenv empty clears", nil, []string{".env"}, EnvChanges{Dotenv: []string{}}, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCommand(1)
			c.Env = tt.env
			c.Dotenv = tt.dotenv
			c.ApplyEnv(tt.changes)
			if !reflect.DeepEqual(c.Env, tt.wantEnv) {
				t.Errorf("Env = %v, want %v", c.Env, tt.wantEnv)
			}
			if !reflect.DeepEqual(c.Dotenv, tt.wantDotenv) {
				t.Errorf("Dotenv = %v, want %v", c.Dotenv, tt.wantDotenv)
			}
		})
	}
}

func TestEnvironment(t *testing.T) {
	dir := t.TempDir()
	writeEnv := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	rootEnv := writeEnv("root.env", "FROM_FILE=root\nSHARED=file\nLITERAL='$KEEP'\n")
	leafEnv := writeEnv("leaf.env", "FROM_FILE=leaf\n")
	missing := filepath.Join(dir, "missing.env")
	bad := writeEnv("bad.env", "not valid\n")

	root := "one-alias"
	mid := "one-alias.two-alias"
	leaf := "one-alias.two-alias.three-alias"

	tests := []struct {
		name     string
		env      map[string]map[string]string
		dotenv   map[string][]string
		check    string
		want     []EnvVar
		wantErrs int
	}{
		{"none", nil, nil, leaf, nil, 0},
		{
			"single node sorted",
			map[string]map[string]string{root: {"B": "2", "A": "1"}},
			nil, root,
			[]EnvVar{{"A", "1", false, root}, {"B", "2", false, root}},
			0,
		},
		{
			"inherited",
			map[string]map[string]string{root: {"A": "1"}},
			nil, leaf,
			[]EnvVar{{"A", "1", false, root}},
			0,
		},
		{
			"child overrides parent keeping position",
			map[string]map[string]string{root: {"A": "root", "B": "root"}, leaf: {"A": "leaf"}},
			nil, leaf,
			[]EnvVar{{"A", "leaf", false, leaf}, {"B", "root", false, root}},
			0,
		},
		{
			"child not visible to parent",
			map[string]map[string]string{leaf: {"A": "leaf"}},
			nil, mid,
			nil,
			0,
		},
		{
			"same node references ordered",
			map[string]map[string]string{root: {"A": "$B/x", "B": "${C}", "C": "base"}},
			nil, root,
			[]EnvVar{{"C", "base", false, root}, {"B", "${C}", false, root}, {"A", "$B/x", false, root}},
			0,
		},
		{
			"self reference",
			map[string]map[string]string{root: {"PATH": "$HOME/bin:$PATH"}},
			nil, root,
			[]EnvVar{{"PATH", "$HOME/bin:$PATH", false, root}},
			0,
		},
		{
			"cycle falls back to alphabetical",
			map[string]map[string]string{root: {"A": "$B", "B": "$A"}},
			nil, root,
			[]EnvVar{{"A", "$B", false, root}, {"B", "$A", false, root}},
			0,
		},
		{
			"dotenv loaded",
			nil,
			map[string][]string{root: {rootEnv}}, root,
			[]EnvVar{{"FROM_FILE", "root", false, rootEnv}, {"SHARED", "file", false, rootEnv}, {"LITERAL", "$KEEP", true, rootEnv}},
			0,
		},
		{
			"env overrides dotenv on same node",
			map[string]map[string]string{root: {"SHARED": "env"}},
			map[string][]string{root: {rootEnv}}, root,
			[]EnvVar{{"FROM_FILE", "root", false, rootEnv}, {"SHARED", "env", false, root}, {"LITERAL", "$KEEP", true, rootEnv}},
			0,
		},
		{
			"child dotenv overrides parent env",
			map[string]map[string]string{root: {"FROM_FILE": "env"}},
			map[string][]string{leaf: {leafEnv}}, leaf,
			[]EnvVar{{"FROM_FILE", "leaf", false, leafEnv}},
			0,
		},
		{
			"later dotenv overrides earlier",
			nil,
			map[string][]string{root: {rootEnv, leafEnv}}, root,
			[]EnvVar{{"FROM_FILE", "leaf", false, leafEnv}, {"SHARED", "file", false, rootEnv}, {"LITERAL", "$KEEP", true, rootEnv}},
			0,
		},
		{
			"missing and bad dotenv skipped",
			map[string]map[string]string{leaf: {"A": "1"}},
			map[string][]string{root: {missing}, mid: {bad}, leaf: {leafEnv}}, leaf,
			[]EnvVar{{"FROM_FILE", "leaf", false, leafEnv}, {"A", "1", false, leaf}},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeCommand(3)
			for kp, env := range tt.env {
				c.find(kp).Env = env
			}
			for kp, dotenv := range tt.dotenv {
				c.find(kp).Dotenv = dotenv
			}
			got, errs := c.find(tt.check).Environment()
			if len(errs) != tt.wantErrs {
				t.Errorf("Environment() errs = %v, want %d", errs, tt.wantErrs)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Environment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvironmentErrorsMentionKeyPath(t *testing.T) {
	c := fakeCommand(2)
	c.Dotenv = []string{filepath.Join(t.TempDir(), "missing.env")}
	_, errs := c.find("one-alias.two-alias").Environment()
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "one-alias: dotenv") {
		t.Errorf("Environment() errs = %v, want one error mentioning one-alias", errs)
	}
}

func TestEnvYAMLRoundtrip(t *testing.T) {
	m := fakeManifest(1, 2)
	m.Find("0-one-alias").Env = map[string]string{"A": "1"}
	m.Find("0-one-alias").Dotenv = []string{"~/.env"}
	m.Find("0-one-alias.0-two-alias").Env = map[string]string{"B": "2"}

	out := m.AsYAML()
	for _, want := range []string{"env:\n", "A: \"1\"", "dotenv:\n", "- ~/.env"} {
		if !strings.Contains(out, want) {
			t.Errorf("AsYAML() missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "env: {}") || strings.Contains(out, "dotenv: []") {
		t.Errorf("AsYAML() should omit empty env and dotenv:\n%s", out)
	}

	loaded := &Manifest{}
	if err := yaml.Unmarshal([]byte(out), loaded); err != nil {
		t.Fatal(err)
	}
	loaded.Link()
	c, rest, err := loaded.Resolve([]string{"0-one-alias", "0-two-alias", "arg"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rest, []string{"arg"}) {
		t.Errorf("Resolve() rest = %v, want [arg]", rest)
	}
	if !reflect.DeepEqual(loaded.Find("0-one-alias").Dotenv, []string{"~/.env"}) {
		t.Errorf("Dotenv = %v, want [~/.env]", loaded.Find("0-one-alias").Dotenv)
	}
	loaded.Find("0-one-alias").Dotenv = nil
	got, errs := c.Environment()
	want := []EnvVar{{"A", "1", false, "0-one-alias"}, {"B", "2", false, "0-one-alias.0-two-alias"}}
	if len(errs) != 0 || !reflect.DeepEqual(got, want) {
		t.Errorf("Environment() = %v, %v, want %v", got, errs, want)
	}
}

func TestEnvFields(t *testing.T) {
	c := fakeCommand(1)
	c.Env = map[string]string{"B": "2", "A": "1"}
	c.Dotenv = []string{"~/.env", ".env.local"}
	fields := c.Fields()
	if fields["env"] != "A=1, B=2" {
		t.Errorf("Fields()[env] = %q, want %q", fields["env"], "A=1, B=2")
	}
	if fields["dotenv"] != "~/.env, .env.local" {
		t.Errorf("Fields()[dotenv] = %q, want %q", fields["dotenv"], "~/.env, .env.local")
	}
}
