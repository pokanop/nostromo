package stringutil

import "strings"

// SanitizeArgs expands all args by spaces and returns a slice.
func SanitizeArgs(args []string) []string {
	sanitizedArgs := []string{}
	if args == nil {
		return sanitizedArgs
	}

	for _, arg := range strings.Split(strings.Join(args, " "), " ") {
		if len(arg) > 0 {
			sanitizedArgs = append(sanitizedArgs, strings.TrimSpace(arg))
		}
	}
	return sanitizedArgs
}

// ContainsCaseInsensitive checks if a string is a substring regardless of case.
func ContainsCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
