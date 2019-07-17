package keypath

import (
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		n        int
		expected string
	}{
		{"empty key path", "", 0, ""},
		{"valid key path clamp negative", "one.two", -1, "one"},
		{"valid key path get first", "one.two", 0, "one"},
		{"valid key path get nth", "one.two.three.four", 2, "three"},
		{"valid key path get last", "one.two.three.four", 3, "four"},
		{"valid key path clamp positive", "one.two.three.four", 8, "four"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := Get(test.keyPath, test.n); actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestDropFirst(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		n        int
		expected string
	}{
		{"empty key path", "", 0, ""},
		{"valid key path clamp negative", "one.two", -1, "one.two"},
		{"valid key path drop zero", "one.two", 0, "one.two"},
		{"valid key path drop many", "one.two.three.four", 2, "three.four"},
		{"valid key path drop all", "one.two.three.four", 4, ""},
		{"valid key path clamp positive", "one.two.three.four", 8, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := DropFirst(test.keyPath, test.n); actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestDropLast(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		n        int
		expected string
	}{
		{"empty key path", "", 0, ""},
		{"valid key path clamp negative", "one.two", -1, "one.two"},
		{"valid key path drop zero", "one.two", 0, "one.two"},
		{"valid key path drop many", "one.two.three.four", 2, "one.two"},
		{"valid key path drop all", "one.two.three.four", 4, ""},
		{"valid key path clamp positive", "one.two.three.four", 8, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := DropLast(test.keyPath, test.n); actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

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
