// Package dotenv parses .env files into ordered entries so their values can
// be exported in any supported shell.
package dotenv

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/pokanop/nostromo/pathutil"
)

// Entry is a single KEY=VALUE line of a .env file
type Entry struct {
	Key   string
	Value string
	// Literal is true for single quoted values which must not be expanded
	// by the shell, double quoted and bare values are expanded.
	Literal bool
}

var keyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidKey reports whether key can be used as an environment variable name
func ValidKey(key string) bool {
	return keyPattern.MatchString(key)
}

// Expand a .env path resolving a leading ~ and $VAR or ${VAR} references
func Expand(path string) string {
	return pathutil.Expand(os.ExpandEnv(path))
}

// Load parses the .env file at path after expanding it, see Expand
func Load(path string) ([]Entry, error) {
	f, err := os.Open(Expand(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return entries, nil
}

// Parse .env content into entries in file order.
//
// Supported syntax:
//
//	# comments and blank lines
//	KEY=value            bare values are trimmed, trailing " # comment" removed
//	export KEY=value     optional export prefix
//	KEY='value'          literal, no escapes except \' and \\
//	KEY="value"          \n \r \t \" \\ escapes, other escapes are kept
//
// Quoted values may span multiple lines.
func Parse(r io.Reader) ([]Entry, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	var entries []Entry
	for i := 0; i < len(lines); i++ {
		lineNum := i + 1
		line := strings.TrimSpace(lines[i])
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("line %d: expected KEY=VALUE", lineNum)
		}
		key = strings.TrimSpace(key)
		if !ValidKey(key) {
			return nil, fmt.Errorf("line %d: invalid key %q", lineNum, key)
		}

		value = strings.TrimLeft(value, " \t")
		entry := Entry{Key: key}
		if len(value) > 0 && (value[0] == '"' || value[0] == '\'') {
			quote := value[0]
			rest := value[1:]
			end := closingQuote(rest, quote)
			for end < 0 && i+1 < len(lines) {
				i++
				rest += "\n" + lines[i]
				end = closingQuote(rest, quote)
			}
			if end < 0 {
				return nil, fmt.Errorf("line %d: unterminated quoted value for %s", lineNum, key)
			}
			if quote == '\'' {
				entry.Value = unescapeSingle(rest[:end])
				entry.Literal = true
			} else {
				entry.Value = unescapeDouble(rest[:end])
			}
		} else {
			entry.Value = stripComment(value)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// closingQuote returns the index of the first unescaped quote or -1
func closingQuote(s string, quote byte) int {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case quote:
			return i
		}
	}
	return -1
}

func unescapeSingle(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && (s[i+1] == '\'' || s[i+1] == '\\') {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func unescapeDouble(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case '"', '\\':
			b.WriteByte(s[i])
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// stripComment removes a trailing " # comment" from a bare value
func stripComment(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '#' && (i == 0 || s[i-1] == ' ' || s[i-1] == '\t') {
			s = s[:i]
			break
		}
	}
	return strings.TrimSpace(s)
}
