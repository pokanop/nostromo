package cmd

import (
	"reflect"
	"testing"

	"github.com/pokanop/nostromo/shell"
)

func TestEvalFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantArgs    []string
		wantVerbose bool
		wantShell   string
		wantErr     bool
	}{
		{"no flags", []string{"foo", "bar"}, []string{"foo", "bar"}, false, shell.Bash, false},
		{"leading -v", []string{"-v", "a", "b"}, []string{"a", "b"}, true, shell.Bash, false},
		{"leading --verbose", []string{"--verbose", "a", "b"}, []string{"a", "b"}, true, shell.Bash, false},
		{"leading -v=true", []string{"-v=true", "a"}, []string{"a"}, true, shell.Bash, false},
		{"leading --verbose=false", []string{"--verbose=false", "a"}, []string{"a"}, false, shell.Bash, false},
		{"repeated -v", []string{"-v", "--verbose", "a"}, []string{"a"}, true, shell.Bash, false},
		{"trailing -v untouched", []string{"a", "-v"}, []string{"a", "-v"}, false, shell.Bash, false},
		{"user flag untouched", []string{"foo", "--flag"}, []string{"foo", "--flag"}, false, shell.Bash, false},
		{"unknown leading flag untouched", []string{"--flag", "a"}, []string{"--flag", "a"}, false, shell.Bash, false},
		{"empty", []string{}, []string{}, false, shell.Bash, false},
		{"--shell fish", []string{"--shell", "fish", "a", "b"}, []string{"a", "b"}, false, shell.Fish, false},
		{"--shell=zsh", []string{"--shell=zsh", "a"}, []string{"a"}, false, shell.Zsh, false},
		{"-s powershell", []string{"-s", "powershell", "a"}, []string{"a"}, false, shell.Powershell, false},
		{"-s=bash", []string{"-s=bash", "a"}, []string{"a"}, false, shell.Bash, false},
		{"shell and verbose", []string{"--shell", "fish", "-v", "a"}, []string{"a"}, true, shell.Fish, false},
		{"verbose and shell", []string{"-v", "--shell=fish", "a"}, []string{"a"}, true, shell.Fish, false},
		{"trailing --shell untouched", []string{"a", "--shell", "fish"}, []string{"a", "--shell", "fish"}, false, shell.Bash, false},
		{"unsupported shell", []string{"--shell", "tcsh", "a"}, nil, false, "", true},
		{"missing shell", []string{"--shell"}, nil, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, opts, err := evalFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("want err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("want args %v, got %v", tt.wantArgs, args)
			}
			if opts.verbose != tt.wantVerbose {
				t.Errorf("want verbose %v, got %v", tt.wantVerbose, opts.verbose)
			}
			if opts.shell != tt.wantShell {
				t.Errorf("want shell %v, got %v", tt.wantShell, opts.shell)
			}
		})
	}
}
