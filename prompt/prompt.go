package prompt

import (
	"bufio"
	"github.com/pokanop/nostromo/log"
	"os"
	"strconv"
	"strings"
)

func stringWithDefault(prompt, def string) string {
	var s string
	log.Boldf(prompt + ": ")
	reader := bufio.NewReader(os.Stdin)
	s, _ = reader.ReadString('\n')
	s = strings.Trim(s, "\n")
	if len(s) == 0 {
		s = def
	}
	return s
}

// String prompt with default.
func String(prompt, def string) string {
	return stringWithDefault(prompt, def)
}

// StringRequired prompt without a default.
func StringRequired(prompt string) (s string) {
	for strings.Trim(s, " ") == "" {
		s = stringWithDefault(prompt, "")
	}
	return s
}

func confirmWithDefault(prompt, def string) bool {
	switch stringWithDefault(prompt, def) {
	case "Yes", "yes", "y", "Y":
		return true
	case "No", "no", "n", "N":
		return false
	}
	return false
}

// Confirm prompts for input that is boolean-ish or defaults.
func Confirm(prompt string, def bool) bool {
	if def {
		return confirmWithDefault(prompt, "y")
	}
	return confirmWithDefault(prompt, "n")
}

// ConfirmRequired continues prompting until the input is boolean-ish.
func ConfirmRequired(prompt string) bool {
	for {
		if result := confirmWithDefault(prompt, ""); result {
			return result
		}
	}
}

// Choose prompts for a single selection from `list` with a default, returning in the index.
func Choose(prompt string, list []string, def int) int {
	log.Regular()
	for i, val := range list {
		log.Regularf("  %d) %s\n", i+1, val)
	}

	log.Regular()
	i := -1

	s := stringWithDefault(prompt, list[def])

	// index
	n, err := strconv.Atoi(s)
	if err == nil {
		if n > 0 && n <= len(list) {
			i = n - 1
			return i
		} else {
			return def
		}
	}

	// value
	i = indexOf(s, list)
	if i != -1 {
		return i
	}
	return def
}

// ChooseRequired continues prompting for a single selection from `list` until valid, returning in the index.
func ChooseRequired(prompt string, list []string) int {
	log.Regular()
	for i, val := range list {
		log.Regularf("  %d) %s\n", i+1, val)
	}

	log.Regular()
	i := -1

	for {
		s := stringWithDefault(prompt, "")

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
