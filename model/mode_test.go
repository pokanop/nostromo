package model

import (
	"testing"
)

func TestModeString(t *testing.T) {
	tests := []struct {
		name string
		m    Mode
		want string
	}{
		{"concatenate", ConcatenateMode, "concatenate"},
		{"independent", IndependentMode, "independent"},
		{"exclusive", ExclusiveMode, "exclusive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsModeSupported(t *testing.T) {
	type args struct {
		mode string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"concatenate", args{"concatenate"}, true},
		{"independent", args{"independent"}, true},
		{"exclusive", args{"exclusive"}, true},
		{"not supported", args{"not supported"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModeSupported(tt.args.mode); got != tt.want {
				t.Errorf("IsModeSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupportedModes(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"supported modes", []string{"concatenate", "independent", "exclusive"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SupportedModes()
			for _, mode := range tt.want {
				found := false
				for _, exp := range got {
					if mode == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("SupportedModes() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestModeFromString(t *testing.T) {
	type args struct {
		mode string
	}
	tests := []struct {
		name string
		args args
		want Mode
	}{
		{"concatenate", args{"concatenate"}, ConcatenateMode},
		{"independent", args{"independent"}, IndependentMode},
		{"exclusive", args{"exclusive"}, ExclusiveMode},
		{"not supported", args{"not supported"}, ConcatenateMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ModeFromString(tt.args.mode); got != tt.want {
				t.Errorf("ModeFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}