package model

import (
	"testing"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/version"
)

func TestSpaceportMigrate(t *testing.T) {
	legacy := func() *Manifest {
		m := NewManifest(CoreManifestName, "", "", &version.Info{})
		m.Config = &Config{Verbose: true, AliasesOnly: true, Mode: ExclusiveMode, BackupCount: 3}
		return m
	}
	docked := func() *Manifest {
		m := NewManifest("docked", "", "", &version.Info{})
		m.Config = &Config{BackupCount: 99}
		return m
	}
	theme := log.GrayscaleTheme

	tests := []struct {
		name      string
		spaceport *Spaceport
		manifests []*Manifest
		want      *Config
		migrated  bool
	}{
		{"defaults", &Spaceport{}, []*Manifest{NewManifest(CoreManifestName, "", "", &version.Info{})}, NewConfig(), true},
		{"legacy core config", &Spaceport{}, []*Manifest{legacy(), docked()}, &Config{true, true, ExclusiveMode, 3, log.EmojiTheme}, true},
		{"legacy theme", &Spaceport{LegacyTheme: &theme}, []*Manifest{legacy()}, &Config{true, true, ExclusiveMode, 3, log.GrayscaleTheme}, true},
		{"existing config wins", &Spaceport{Config: &Config{BackupCount: 5, Theme: log.DefaultTheme}, LegacyTheme: &theme}, []*Manifest{legacy(), docked()}, &Config{BackupCount: 5, Theme: log.DefaultTheme}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.spaceport
			s.Init()
			s.Import(tt.manifests)
			if got := s.Migrate(); got != tt.migrated {
				t.Errorf("Migrate() = %t, want %t", got, tt.migrated)
			}
			if s.Config == nil || *s.Config != *tt.want {
				t.Errorf("config = %+v, want %+v", s.Config, tt.want)
			}
			if s.LegacyTheme != nil {
				t.Errorf("want legacy theme cleared")
			}
			for _, m := range s.Manifests() {
				if m.Config != nil {
					t.Errorf("want %s manifest config dropped, got %+v", m.Name, m.Config)
				}
			}
		})
	}
}
