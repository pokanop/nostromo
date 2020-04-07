package shell

import (
	"reflect"
	"testing"
)

func TestValidLanguages(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"valid languages", []string{"sh", "ruby", "python", "perl", "js"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SupportedLanguages(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidLanguages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSupportedLanguage(t *testing.T) {
	type args struct {
		language string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"supported", args{"python"}, true},
		{"not supported", args{"jython"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedLanguage(tt.args.language); got != tt.want {
				t.Errorf("IsSupportedLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalString(t *testing.T) {
	type args struct {
		command  string
		language string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty command", args{"", ""}, "", true},
		{"command with newline", args{"echo foo\n", ""}, "echo foo", false},
		{"no lang", args{"echo foo", ""}, "echo foo", false},
		{"node", args{"console.log(\"hello world\")", "js"}, "node -e 'console.log(\"hello world\")'", false},
		{"ruby", args{"puts \"hello world\"", "ruby"}, "ruby -e 'puts \"hello world\"'", false},
		{"python", args{"print()", "python"}, "python -c 'print()'", false},
		{"perl", args{"print \"hello world\";", "perl"}, "perl -e 'print \"hello world\";'", false},
		{"sh", args{"echo foo", "sh"}, "echo foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalString(tt.args.command, tt.args.language, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvalString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EvalString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
