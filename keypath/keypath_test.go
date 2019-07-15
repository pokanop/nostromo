package keypath

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{"nil args", nil, nil},
		{"empty args", []string{}, []string{}},
		{"no dots no spaces", []string{"arg1", "arg2", "arg3"}, []string{"arg1", "arg2", "arg3"}},
		{"no dots spaces", []string{"arg1", "arg2-1 arg2-2", "arg3"}, []string{"arg1", "arg2-1 arg2-2", "arg3"}},
		{"dots no spaces", []string{"arg1", "arg2-1.arg2-2.arg2-3", "arg3"}, []string{"arg1", "arg2-1[#dot#]arg2-2[#dot#]arg2-3", "arg3"}},
		{"dots and spaces", []string{"arg1", "arg2-1.arg2-2. arg2-3", "arg3"}, []string{"arg1", "arg2-1[#dot#]arg2-2[#dot#] arg2-3", "arg3"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := Encode(test.args); reflect.DeepEqual(test.expected, actual) == false {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{"nil args", nil, nil},
		{"empty args", []string{}, []string{}},
		{"no dots no spaces", []string{"arg1", "arg2", "arg3"}, []string{"arg1", "arg2", "arg3"}},
		{"no dots spaces", []string{"arg1", "arg2-1 arg2-2", "arg3"}, []string{"arg1", "arg2-1 arg2-2", "arg3"}},
		{"dots no spaces", []string{"arg1", "arg2-1[#dot#]arg2-2[#dot#]arg2-3", "arg3"}, []string{"arg1", "arg2-1.arg2-2.arg2-3", "arg3"}},
		{"dots and spaces", []string{"arg1", "arg2-1[#dot#]arg2-2[#dot#] arg2-3", "arg3"}, []string{"arg1", "arg2-1.arg2-2. arg2-3", "arg3"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := Decode(test.args); reflect.DeepEqual(test.expected, actual) == false {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}
