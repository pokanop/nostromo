package config

import (
	"testing"

	"github.com/pokanop/nostromo/model"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		expErr bool
	}{
		{"invalid path", "", true},
		{"missing path", "/does/not/exist/.nostromo", true},
		{"bad file format", "../testdata/bad.nostromo", true},
		{"good file format", "../testdata/good.nostromo", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Config{}
			m, err := c.Parse(test.path)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr {
				if err != nil {
					t.Errorf("expected no error but got %s", err)
				} else if m == nil {
					t.Errorf("manifest is nil")
				}
			}
		})
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		manifest *model.Manifest
		expErr   bool
	}{
		{"invalid path", "", nil, true},
		{"nil manifest", "path", nil, true},
		{"valid manifest", "/tmp/test.nostromo", fakeManifest(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Config{test.path, test.manifest}
			err := c.Save()
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			}
		})
	}
}

func fakeManifest() *model.Manifest {
	m := model.NewManifest()
	m.AddCommand("one.two.three", "command")
	m.AddSubstitution("one.two", "name", "alias")
	return m
}
