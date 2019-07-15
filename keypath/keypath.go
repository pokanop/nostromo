package keypath

import "strings"

var subs = map[string]string{
	".": "[#dot#]",
}

var revs = map[string]string{
	"[#dot#]": ".",
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
