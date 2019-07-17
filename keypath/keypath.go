package keypath

import (
	"math"
	"strings"
)

const delimiter = "."

var subs = map[string]string{
	".": "[#dot#]",
}

var revs = map[string]string{
	"[#dot#]": ".",
}

func KeyPath(args []string) string {
	return strings.Join(args, delimiter)
}

func Keys(keyPath string) []string {
	return strings.Split(keyPath, delimiter)
}

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

func DropFirst(keyPath string, n int) string {
	keys := Keys(keyPath)
	l := len(keys)
	n = clamp(n, 0, l)
	if l > n {
		return KeyPath(keys[n:])
	}
	return ""
}

func DropLast(keyPath string, n int) string {
	keys := Keys(keyPath)
	l := len(keys)
	n = clamp(n, 0, l)
	if l > n {
		return KeyPath(keys[:l-n])
	}
	return ""
}

func Encode(args []string) []string {
	return swap(args, subs)
}

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
