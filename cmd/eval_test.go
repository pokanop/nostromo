package cmd

import (
	"reflect"
	"testing"
)

func TestEvalFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantArgs    []string
		wantVerbose bool
	}{
		{"no flags", []string{"foo", "bar"}, []string{"foo", "bar"}, false},
		{"leading -v", []string{"-v", "a", "b"}, []string{"a", "b"}, true},
		{"leading --verbose", []string{"--verbose", "a", "b"}, []string{"a", "b"}, true},
		{"leading -v=true", []string{"-v=true", "a"}, []string{"a"}, true},
		{"leading --verbose=false", []string{"--verbose=false", "a"}, []string{"a"}, false},
		{"repeated -v", []string{"-v", "--verbose", "a"}, []string{"a"}, true},
		{"trailing -v untouched", []string{"a", "-v"}, []string{"a", "-v"}, false},
		{"user flag untouched", []string{"foo", "--flag"}, []string{"foo", "--flag"}, false},
		{"unknown leading flag untouched", []string{"--flag", "a"}, []string{"--flag", "a"}, false},
		{"empty", []string{}, []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, verbose := evalFlags(tt.args)
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("want args %v, got %v", tt.wantArgs, args)
			}
			if verbose != tt.wantVerbose {
				t.Errorf("want verbose %v, got %v", tt.wantVerbose, verbose)
			}
		})
	}
}
