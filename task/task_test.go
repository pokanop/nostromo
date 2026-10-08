package task

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/version"
)

func TestShowConfig(t *testing.T) {
	t.Cleanup(func() {
		model.SetVerbose(false)
	})

	tests := []struct {
		name    string
		asJSON  bool
		asYAML  bool
		asTree  bool
		verbose bool
	}{
		{name: "default"},
		{name: "verbose", verbose: true},
		{name: "json", asJSON: true},
		{name: "yaml", asYAML: true},
		{name: "tree", asTree: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.SetVersion(version.NewInfo("", "", ""))

			home := t.TempDir()
			baseDir := t.TempDir()
			t.Setenv("NOSTROMO_HOME", baseDir)
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			if err := os.MkdirAll(filepath.Join(baseDir, "ships"), 0755); err != nil {
				t.Fatal(err)
			}

			cfg, err := config.NewConfig()
			if err != nil {
				t.Fatal(err)
			}
			if err := cfg.Save(); err != nil {
				t.Fatal(err)
			}

			model.SetVerbose(tt.verbose)
			if got := ShowConfig(tt.asJSON, tt.asYAML, tt.asTree); got != 0 {
				t.Errorf("ShowConfig() = %d, want 0", got)
			}
		})
	}
}
