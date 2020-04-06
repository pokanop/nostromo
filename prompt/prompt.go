package prompt

import (
	"bufio"
	"github.com/pokanop/nostromo/log"
	"os"
	"strconv"
	"strings"
)

// String prompt.
func String(prompt string, args ...interface{}) string {
	var s string
	log.Boldf(prompt+": ", args...)
	reader := bufio.NewReader(os.Stdin)
	s, _ = reader.ReadString('\n')
	return strings.Trim(s, "\n")
}

// String prompt (required).
func StringRequired(prompt string, args ...interface{}) (s string) {
	for strings.Trim(s, " ") == "" {
		s = String(prompt, args...)
	}
	return s
}

// Confirm continues prompting until the input is boolean-ish.
func Confirm(prompt string, args ...interface{}) bool {
	for {
		switch String(prompt, args...) {
		case "Yes", "yes", "y", "Y":
			return true
		case "No", "no", "n", "N":
			return false
		}
	}
}

// Choose prompts for a single selection from `list`, returning in the index.
func Choose(prompt string, list []string) int {
	log.Regular()
	for i, val := range list {
		log.Regularf("  %d) %s\n", i+1, val)
	}

	log.Regular()
	i := -1

	for {
		s := String(prompt)

		// index
		n, err := strconv.Atoi(s)
		if err == nil {
			if n > 0 && n <= len(list) {
				i = n - 1
				break
			} else {
				continue
			}
		}

		// value
		i = indexOf(s, list)
		if i != -1 {
			break
		}
	}

	return i
}

// index of `s` in `list`.
func indexOf(s string, list []string) int {
	for i, val := range list {
		if val == s {
			return i
		}
	}
	return -1
}
