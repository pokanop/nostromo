package stringutil

import (
	"fmt"
	"strings"
)

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

// ReversedStrings returns a slice of reversed strings.
func ReversedStrings(strs []string) []string {
	if strs == nil {
		return nil
	}

	r := []string{}
	for i := len(strs) - 1; i >= 0; i-- {
		r = append(r, strs[i])
	}
	return r
}

// ReplaceShellVars swaps command args like $1 and returns the result.
func ReplaceShellVars(cmd string, args []string) string {
	// Deal with $1 - $N for now, not sure if we need to deal with or how
	// to handle $#, $@, $*, $!, $$, and $?
	count := 0
	for _, arg := range args {
		shellVar := fmt.Sprintf("$%d", count+1)
		if !strings.Contains(cmd, shellVar) {
			break
		}
		count += 1
		cmd = strings.ReplaceAll(cmd, shellVar, arg)
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s", cmd, strings.Join(args[count:], " ")))
}
