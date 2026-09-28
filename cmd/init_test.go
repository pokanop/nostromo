package cmd

import (
	"os"
	"testing"
)

func TestShowBanner(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if showBanner(false, w) {
		t.Error("banner should not show when stdout is not a terminal")
	}
	if showBanner(true, w) {
		t.Error("banner should not show when disabled")
	}
}

func TestInitFlags(t *testing.T) {
	if err := initCmd.ParseFlags([]string{"--no-banner"}); err != nil {
		t.Fatal(err)
	}
	if !noBanner {
		t.Error("expected --no-banner to set noBanner")
	}
	noBanner = false

	if err := versionCmd.ParseFlags([]string{"--banner"}); err != nil {
		t.Fatal(err)
	}
	if !versionBanner {
		t.Error("expected --banner to set versionBanner")
	}
	versionBanner = false
}
