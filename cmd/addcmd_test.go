package cmd

import (
	"reflect"
	"testing"

	"github.com/pokanop/nostromo/model"
	"github.com/spf13/cobra"
)

func TestEnvFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected model.EnvChanges
		wantErr  bool
	}{
		{"none", nil, model.EnvChanges{}, false},
		{"env", []string{"--env", "A=1", "-e", "B=x=y"}, model.EnvChanges{Set: map[string]string{"A": "1", "B": "x=y"}}, false},
		{"unset", []string{"--unset-env", "A"}, model.EnvChanges{Unset: []string{"A"}}, false},
		{"dotenv", []string{"--dotenv", "~/.env", "--dotenv", ".env.local"}, model.EnvChanges{Dotenv: []string{"~/.env", ".env.local"}}, false},
		{"clear dotenv", []string{"--dotenv", ""}, model.EnvChanges{Dotenv: []string{}}, false},
		{"invalid env", []string{"--env", "A"}, model.EnvChanges{}, true},
		{"invalid env key", []string{"--env", "1A=b"}, model.EnvChanges{}, true},
		{"invalid unset key", []string{"--unset-env", "A-B"}, model.EnvChanges{Unset: []string{"A-B"}}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, unsetEnv, dotenvFiles = nil, nil, nil
			cmd := &cobra.Command{Use: "test"}
			addEnvFlags(cmd)
			if err := cmd.ParseFlags(test.args); err != nil {
				t.Fatalf("parse flags: %v", err)
			}
			if err := envValid(); (err != nil) != test.wantErr {
				t.Fatalf("envValid() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.wantErr {
				return
			}
			actual := envFlags(cmd)
			if len(actual.Set) == 0 {
				actual.Set = nil
			}
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected %#v, got %#v", test.expected, actual)
			}
		})
	}
}
