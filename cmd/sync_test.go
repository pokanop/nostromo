package cmd

import "testing"

func TestSyncAcceptsManifestNames(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{}},
		{"one manifest", []string{"docker"}},
		{"many manifests", []string{"docker", "edit", "tools"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := syncCmd.ValidateArgs(tt.args); err != nil {
				t.Errorf("sync %v: unexpected error: %v", tt.args, err)
			}
		})
	}
}
