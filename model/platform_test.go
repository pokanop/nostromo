package model

import (
	"reflect"
	"runtime"
	"testing"
)

func TestCurrentPlatform(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string
	}{
		{"runtime", "", runtime.GOOS + "/" + runtime.GOARCH},
		{"override os", "windows", "windows"},
		{"override os/arch", "linux/arm64", "linux/arm64"},
		{"override normalized", "  Darwin ", "darwin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(PlatformEnv, tt.env)
			if got := CurrentPlatform(); got != tt.want {
				t.Errorf("CurrentPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePlatforms(t *testing.T) {
	tests := []struct {
		name    string
		list    []string
		want    []string
		wantErr bool
	}{
		{"nil", nil, []string{}, false},
		{"empty", []string{""}, []string{}, false},
		{"single", []string{"linux"}, []string{"linux"}, false},
		{"comma separated", []string{"linux,darwin"}, []string{"linux", "darwin"}, false},
		{"multiple items", []string{"linux", "darwin"}, []string{"linux", "darwin"}, false},
		{"normalized and deduped", []string{" Linux ", "linux", "DARWIN"}, []string{"linux", "darwin"}, false},
		{"os/arch", []string{"linux/arm64", "windows/amd64"}, []string{"linux/arm64", "windows/amd64"}, false},
		{"typo os", []string{"linxu"}, nil, true},
		{"typo arch", []string{"linux/arm65"}, nil, true},
		{"invalid in list", []string{"linux,macos"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlatforms(tt.list)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParsePlatforms() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePlatforms() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePlatform(t *testing.T) {
	tests := []struct {
		platform string
		wantErr  bool
	}{
		{"linux", false},
		{"darwin", false},
		{"windows", false},
		{"linux/arm64", false},
		{"darwin/amd64", false},
		{"macos", true},
		{"osx", true},
		{"win", true},
		{"linux/x86", true},
		{"linux/", true},
		{"/amd64", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			if err := ValidatePlatform(tt.platform); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlatform() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlatformsInclude(t *testing.T) {
	tests := []struct {
		name      string
		platforms []string
		platform  string
		want      bool
	}{
		{"empty list", nil, "linux/amd64", true},
		{"os match", []string{"linux"}, "linux/amd64", true},
		{"os mismatch", []string{"linux"}, "darwin/amd64", false},
		{"one of many", []string{"windows", "darwin"}, "darwin/arm64", true},
		{"os/arch match", []string{"linux/arm64"}, "linux/arm64", true},
		{"os/arch arch mismatch", []string{"linux/arm64"}, "linux/amd64", false},
		{"os/arch os mismatch", []string{"linux/arm64"}, "darwin/arm64", false},
		{"os/arch spec with os only platform", []string{"linux/arm64"}, "linux", true},
		{"os spec with os only platform", []string{"linux"}, "linux", true},
		{"os spec with os only platform mismatch", []string{"linux"}, "windows", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := platformsInclude(tt.platforms, tt.platform); got != tt.want {
				t.Errorf("platformsInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}
