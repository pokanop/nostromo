package task

import (
	"reflect"
	"testing"
)

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
			if got := sanitizeArgs(tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sanitizeArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}