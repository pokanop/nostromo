package shell

import (
	"reflect"
	"testing"

	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/version"
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

func TestShellWrapperFunc(t *testing.T) {
	if shellWrapperFunc() != `__nostromo_cmd() { command nostromo "$@"; }
nostromo() { __nostromo_cmd "$@" && eval "$(__nostromo_cmd completion)"; }` {
		t.Errorf("shell wrapper func not as expected")
	}
}

// func TestShellAliasFuncs(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		manifest *model.Manifest
// 		expected string
// 	}{
// 		{"no alias", fakeManifest(false), "\none() { eval $(__nostromo_cmd eval one \"$*\"); }\ntwo() { eval $(__nostromo_cmd eval two \"$*\"); }\n"},
// 		{"alias only", fakeManifest(true), ""},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			actual := shellAliasFuncs(tt.manifest)
// 			fmt.Printf("<start>%s<end>", actual)
// 			if tt.expected != actual {
// 				t.Errorf("shell alias funcs incorrect expected: %s, actual: %s", tt.expected, actual)
// 			}
// 		})
// 	}
// }

func fakeManifest(aliasOnly bool) *model.Manifest {
	m := model.NewManifest(&version.Info{})
	m.Config.AliasesOnly = aliasOnly
	m.AddCommand("one.two.three", "command", "", &model.Code{}, false, "concatenate")
	m.AddSubstitution("one.two", "name", "alias")
	m.AddCommand("two", "command", "", &model.Code{}, false, "concatenate")
	return m
}
