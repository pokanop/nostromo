package stringutil

import (
	"reflect"
	"testing"
)

func TestContainsCaseInsensitive(t *testing.T) {
	type args struct {
		s      string
		substr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty string", args{"", ""}, true},
		{"matching same case", args{"foo", "foo"}, true},
		{"matching diff case", args{"foo", "FoO"}, true},
		{"not matching short", args{"foo bar baz", "qux"}, false},
		{"not matching long", args{"bar", "foo bar baz"}, false},
		{"matching long same case", args{"fOO baR bAz", "bAR"}, true},
		{"matching long diff case", args{"fOO baR bAz", "bar"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsCaseInsensitive(tt.args.s, tt.args.substr); got != tt.want {
				t.Errorf("ContainsCaseInsensitive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"nil args", args{nil}, []string{}},
		{"empty args", args{[]string{}}, []string{}},
		{"single arg", args{[]string{"one"}}, []string{"one"}},
		{"multi arg", args{[]string{"one", "two", "three"}}, []string{"one", "two", "three"}},
		{"multi arg strings", args{[]string{"one foo", "two foo bar", "three"}}, []string{"one", "foo", "two", "foo", "bar", "three"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeArgs(tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SanitizeArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReversed(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		expected []string
	}{
		{"nil strs", nil, nil},
		{"empty strs", []string{}, []string{}},
		{"valid strs", []string{"a", "b", "c"}, []string{"c", "b", "a"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := ReversedStrings(test.strs); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestReplaceShellVars(t *testing.T) {
	type args struct {
		cmd  string
		args []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nil cmd and args", args{"", nil}, ""},
		{"nil args", args{"cmd", nil}, "cmd"},
		{"empty args", args{"cmd", []string{}}, "cmd"},
		{"no shell vars", args{"cmd", []string{"arg1", "arg2"}}, "cmd arg1 arg2"},
		{"shell vars", args{"cmd $1 $2", []string{"arg1", "arg2"}}, "cmd arg1 arg2"},
		{"more shell vars", args{"cmd $1 $2 $3", []string{"arg1", "arg2"}}, "cmd arg1 arg2 $3"},
		{"less shell vars", args{"cmd $1 $2", []string{"arg1", "arg2", "arg3"}}, "cmd arg1 arg2 arg3"},
		{"shell vars no space", args{"cmd $1foo$2bar", []string{"arg1", "arg2", "arg3"}}, "cmd arg1fooarg2bar arg3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceShellVars(tt.args.cmd, tt.args.args); got != tt.want {
				t.Errorf("ReplaceShellVars() = %v, want %v", got, tt.want)
			}
		})
	}
}
