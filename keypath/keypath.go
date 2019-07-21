package keypath

import (
	"math"
	"strings"
)

// Delimiter used for key path operations
const Delimiter = "."

var subs = map[string]string{
	".": "[#dot#]",
}

var revs = map[string]string{
	"[#dot#]": ".",
}

// KeyPath for given argument list combined by `Delimiter`
//
// Returns `"foo.bar.baz"` for `["foo", "bar", "baz"]`
func KeyPath(args []string) string {
	return strings.Join(args, Delimiter)
}

// Keys for given key path separated by `Delimiter`
//
// Returns `["foo", "bar", "baz"]` for `"foo.bar.baz"`
func Keys(keyPath string) []string {
	return strings.Split(keyPath, Delimiter)
}

// Get the nth key in the key path
func Get(keyPath string, n int) string {
	keys := Keys(keyPath)
	l := len(keys)
	m := l
	if m > 0 {
		m--
	}
	n = clamp(n, 0, m)
	if l > n {
		return keys[n]
	}
	return ""
}

// DropFirst n keys in the key path
func DropFirst(keyPath string, n int) string {
	keys := Keys(keyPath)
	l := len(keys)
	n = clamp(n, 0, l)
	if l > n {
		return KeyPath(keys[n:])
	}
	return ""
}

// DropLast n keys in the key path
func DropLast(keyPath string, n int) string {
	keys := Keys(keyPath)
	l := len(keys)
	n = clamp(n, 0, l)
	if l > n {
		return KeyPath(keys[:l-n])
	}
	return ""
}

// Encode strings to operate safely as key paths
func Encode(args []string) []string {
	return swap(args, subs)
}

// Decode strings to operate outside of key paths
func Decode(args []string) []string {
	return swap(args, revs)
}

func swap(args []string, lookups map[string]string) []string {
	if args == nil {
		return nil
	}

	mods := make([]string, len(args))
	for i, a := range args {
		for k, v := range lookups {
			a = strings.Replace(a, k, v, -1)
		}
		mods[i] = a
	}
	return mods
}

func clamp(val, min, max int) int {
	return int(math.Max(math.Min(float64(val), float64(max)), float64(min)))
}
