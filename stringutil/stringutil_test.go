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
