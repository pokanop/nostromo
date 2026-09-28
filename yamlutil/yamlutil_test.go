package yamlutil

import (
	"reflect"
	"testing"
)

type leaf struct {
	Language string `yaml:"language"`
	Snippet  string `yaml:"snippet"`
}

type node struct {
	KeyPath   string           `yaml:"keyPath"`
	AliasOnly bool             `yaml:"aliasOnly"`
	Children  map[string]*node `yaml:"children"`
	Tags      []string         `yaml:"tags,omitempty"`
	Code      *leaf            `yaml:"code"`
	Untagged  int
	Skipped   string `yaml:"-"`
	hidden    string
}

type link struct {
	SemVer string `yaml:"semVer"`
}

type inlined struct {
	BackupCount int `yaml:"backupCount"`
}

type root struct {
	Name     string           `yaml:"name"`
	Commands map[string]*node `yaml:"commands"`
	Links    []*link          `yaml:"links"`
	Extra    map[string]int   `yaml:"extra"`
	Settings inlined          `yaml:",inline"`
}

const legacyDoc = `
name: root
backupcount: 3
commands:
  one:
    keypath: one
    aliasonly: true
    untagged: 7
    skipped: nope
    hidden: nope
    code:
      language: sh
      snippet: echo hi
    children:
      two:
        keypath: one.two
        aliasonly: false
        tags:
        - a
        - b
        children: {}
links:
- semver: 1.2.3
extra:
  MixedCase: 1
unknownkey: whatever
`

const canonicalDoc = `
name: root
backupCount: 3
commands:
  one:
    keyPath: one
    aliasOnly: true
    untagged: 7
    code:
      language: sh
      snippet: echo hi
    children:
      two:
        keyPath: one.two
        aliasOnly: false
        tags:
        - a
        - b
        children: {}
links:
- semVer: 1.2.3
extra:
  MixedCase: 1
`

func expectedRoot() *root {
	return &root{
		Name: "root",
		Commands: map[string]*node{
			"one": {
				KeyPath:   "one",
				AliasOnly: true,
				Untagged:  7,
				Code:      &leaf{"sh", "echo hi"},
				Children: map[string]*node{
					"two": {
						KeyPath:  "one.two",
						Tags:     []string{"a", "b"},
						Children: map[string]*node{},
					},
				},
			},
		},
		Links:    []*link{{"1.2.3"}},
		Extra:    map[string]int{"MixedCase": 1},
		Settings: inlined{3},
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name         string
		doc          string
		wantMigrated bool
		wantErr      bool
	}{
		{"legacy keys", legacyDoc, true, false},
		{"canonical keys", canonicalDoc, false, false},
		{"invalid yaml", "name: [", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &root{}
			migrated, err := Unmarshal([]byte(tt.doc), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if migrated != tt.wantMigrated {
				t.Errorf("Unmarshal() migrated = %v, want %v", migrated, tt.wantMigrated)
			}
			if want := expectedRoot(); !reflect.DeepEqual(got, want) {
				t.Errorf("Unmarshal() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestUnmarshalLegacyMatchesCanonical(t *testing.T) {
	legacy := &root{}
	if _, err := Unmarshal([]byte(legacyDoc), legacy); err != nil {
		t.Fatal(err)
	}
	canonical := &root{}
	if _, err := Unmarshal([]byte(canonicalDoc), canonical); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(legacy, canonical) {
		t.Errorf("legacy = %+v, canonical = %+v", legacy, canonical)
	}
}

func TestUnmarshalPrefersCanonicalDuplicate(t *testing.T) {
	doc := "keypath: old\nkeyPath: new\nALIASONLY: true\n"
	got := &node{}
	migrated, err := Unmarshal([]byte(doc), got)
	if err != nil {
		t.Fatal(err)
	}
	if !migrated {
		t.Errorf("want migrated")
	}
	if got.KeyPath != "new" {
		t.Errorf("KeyPath = %q, want %q", got.KeyPath, "new")
	}
	if !got.AliasOnly {
		t.Errorf("want AliasOnly set from upper case key")
	}
}

func TestUnmarshalIgnoresMismatchedShapes(t *testing.T) {
	// Struct fields holding scalars or sequences where mappings are expected
	// must not panic and are left for the decoder to reject
	doc := "commands: just a string\nlinks:\n- 42\nextra: [1, 2]\n"
	got := &root{}
	if _, err := Unmarshal([]byte(doc), got); err == nil {
		t.Errorf("want decode error for mismatched shapes")
	}
}

func TestUnmarshalEmptyAndNil(t *testing.T) {
	got := &root{}
	migrated, err := Unmarshal([]byte(""), got)
	if err != nil || migrated {
		t.Errorf("empty doc: migrated = %v, err = %v", migrated, err)
	}
	if _, err := Unmarshal([]byte("name: x"), nil); err == nil {
		t.Errorf("want error for nil target")
	}
}
