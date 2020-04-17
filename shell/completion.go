package shell

import (
	"bytes"
	"github.com/pokanop/nostromo/model"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

// CobraCompleter interface for types that can generate a cobra.Command
type CobraCompleter interface {
	CobraCommand() *cobra.Command
}

// Completion generates shell completion scripts
func Completion(cmd *cobra.Command) (string, error) {
	var buf bytes.Buffer
	var err error
	if Which() == Zsh {
		err = cmd.GenZshCompletion(&buf)
	} else {
		err = cmd.GenBashCompletion(&buf)
	}
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(&buf)
	if err != nil {
		return "", err
	}
	s := string(b)

	// This is required due to a bug in cobra and zsh support:
	// https://github.com/spf13/cobra/pull/887
	if Which() == Zsh {
		s = strings.Replace(s, "#", "", 1)
	}

	return s, nil
}

// ManifestCompletion scripts for a manifest
func ManifestCompletion(m *model.Manifest) ([]string, error) {
	var completions []string
	completions = append(completions, shellWrapperFunc())
	completions = append(completions, shellAliasFuncs(m))
	for _, cmd := range m.Commands {
		// Skip completion scripts for leaf nodes or pure aliases.
		// This allows for it to fallback to the shell's lookups.
		if cmd.AliasOnly || len(cmd.Commands) == 0 {
			continue
		}
		s, err := CommandCompletion(cmd)
		if err != nil {
			return nil, err
		}
		completions = append(completions, s)
	}
	return completions, nil
}

// CommandCompletion script for a command
func CommandCompletion(cmd *model.Command) (string, error) {
	return Completion(cmd.CobraCommand())
}
