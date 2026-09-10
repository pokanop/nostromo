package dotenv

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []Entry
		wantErr bool
	}{
		{"empty", "", nil, false},
		{"comments and blanks", "# comment\n\n   \n  # indented\n", nil, false},
		{"bare", "FOO=bar\n", []Entry{{"FOO", "bar", false}}, false},
		{"bare trimmed", "  FOO = bar baz  \n", []Entry{{"FOO", "bar baz", false}}, false},
		{"bare inline comment", "FOO=bar # comment\n", []Entry{{"FOO", "bar", false}}, false},
		{"bare hash in value", "URL=http://x/#anchor\n", []Entry{{"URL", "http://x/#anchor", false}}, false},
		{"empty value", "FOO=\n", []Entry{{"FOO", "", false}}, false},
		{"export prefix", "export FOO=bar\n", []Entry{{"FOO", "bar", false}}, false},
		{"single quoted literal", `FOO='$HOME # not a comment'`, []Entry{{"FOO", "$HOME # not a comment", true}}, false},
		{"single quoted escapes", `FOO='it\'s a \\ path'`, []Entry{{"FOO", `it's a \ path`, true}}, false},
		{"double quoted", `FOO="a \"b\" \\ c\td"`, []Entry{{"FOO", "a \"b\" \\ c\td", false}}, false},
		{"double quoted newline", `FOO="line1\nline2"`, []Entry{{"FOO", "line1\nline2", false}}, false},
		{"double quoted keeps dollar escape", `FOO="cost \$5 of $TOTAL"`, []Entry{{"FOO", `cost \$5 of $TOTAL`, false}}, false},
		{"double quoted trailing comment", `FOO="bar" # comment`, []Entry{{"FOO", "bar", false}}, false},
		{"multiline double quoted", "KEY=\"-----BEGIN-----\nabc\n-----END-----\"\nNEXT=1\n", []Entry{{"KEY", "-----BEGIN-----\nabc\n-----END-----", false}, {"NEXT", "1", false}}, false},
		{"multiline single quoted", "KEY='a\nb'\n", []Entry{{"KEY", "a\nb", true}}, false},
		{"crlf", "A=1\r\nB=2\r\n", []Entry{{"A", "1", false}, {"B", "2", false}}, false},
		{"order preserved", "Z=1\nA=2\nM=3\n", []Entry{{"Z", "1", false}, {"A", "2", false}, {"M", "3", false}}, false},
		{"missing equals", "FOO\n", nil, true},
		{"invalid key", "1FOO=bar\n", nil, true},
		{"invalid key chars", "FOO-BAR=bar\n", nil, true},
		{"unterminated quote", "FOO=\"bar\nBAZ=1\n", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.content))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestValidKey(t *testing.T) {
	for key, want := range map[string]bool{"FOO": true, "_x1": true, "foo_bar": true, "": false, "1A": false, "A-B": false, "A.B": false, "A B": false} {
		if got := ValidKey(key); got != want {
			t.Errorf("ValidKey(%q) = %v, want %v", key, got, want)
		}
	}
}

func TestExpand(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("NOSTROMO_TEST_DIR", "projects")
	tests := []struct {
		path string
		want string
	}{
		{"~/.env", filepath.Join(home, ".env")},
		{"$NOSTROMO_TEST_DIR/.env", "projects/.env"},
		{"${NOSTROMO_TEST_DIR}/.env", "projects/.env"},
		{"~/$NOSTROMO_TEST_DIR/.env", filepath.Join(home, "projects", ".env")},
		{"/abs/.env", "/abs/.env"},
		{".env", ".env"},
	}
	for _, tt := range tests {
		if got := Expand(tt.path); got != tt.want {
			t.Errorf("Expand(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("A=1\nB='two'\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NOSTROMO_TEST_ENV_DIR", dir)

	got, err := Load("$NOSTROMO_TEST_ENV_DIR/.env")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := []Entry{{"A", "1", false}, {"B", "two", true}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load() = %#v, want %#v", got, want)
	}

	if _, err := Load(filepath.Join(dir, "missing.env")); err == nil {
		t.Errorf("Load() expected error for missing file")
	}

	bad := filepath.Join(dir, "bad.env")
	if err := os.WriteFile(bad, []byte("nope\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(bad); err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Errorf("Load() expected parse error with line number, got %v", err)
	}
}
