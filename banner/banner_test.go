package banner

import (
	"os"
	"strings"
	"testing"
)

func TestLines(t *testing.T) {
	lines := Lines("nostromo")
	if len(lines) == 0 {
		t.Fatal("expected banner rows")
	}
	for _, line := range lines {
		if len(line) > 80 {
			t.Errorf("row exceeds 80 columns (%d): %q", len(line), line)
		}
		if strings.TrimRight(line, " ") != line {
			t.Errorf("row has trailing whitespace: %q", line)
		}
	}
	if strings.TrimSpace(strings.Join(lines, "")) == "" {
		t.Error("banner is blank")
	}
}

func TestLinesNonASCII(t *testing.T) {
	// non-strict rendering must not panic on unsupported runes
	if len(Lines("nöstromo")) == 0 {
		t.Error("expected banner rows")
	}
}

func TestIsTerminal(t *testing.T) {
	f, err := os.CreateTemp("", "banner")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if IsTerminal(f) {
		t.Error("regular file should not be a terminal")
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if IsTerminal(w) {
		t.Error("pipe should not be a terminal")
	}
}
