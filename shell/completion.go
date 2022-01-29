package shell

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pokanop/nostromo/model"
	"github.com/spf13/cobra"
)

// CobraCompleter interface for types that can generate a cobra.Command
type CobraCompleter interface {
	CobraCommand() *cobra.Command
}

// Completion generates shell completion scripts
func Completion(sh string, cmd *cobra.Command) (string, error) {
	var buf bytes.Buffer
	var err error
	switch sh {
	case "bash":
		err = cmd.GenBashCompletionV2(&buf, true)
	case "zsh":
		zshHead := fmt.Sprintf("#compdef %[1]s\ncompdef _%[1]s %[1]s\n", cmd.Name())
		buf.Write([]byte(zshHead))
		err = cmd.GenZshCompletion(&buf)
	case "fish":
		err = cmd.GenFishCompletion(&buf, true)
	case "powershell":
		err = cmd.GenPowerShellCompletionWithDesc(&buf)
	}
	if err != nil {
		return "", err
	}

	s := buf.String()
	if cmd.Name() != "nostromo" {
		s = strings.ReplaceAll(s, "${words[1]} __complete ${words[2,-1]}", "nostromo __complete run ${words[1]} ${words[2,-1]}")
	}

	return s, nil
}

// SpaceportCompletion scripts for all manifests
func SpaceportCompletion(sh string, s *model.Spaceport) ([]string, error) {
	var completions []string
	completions = append(completions, shellWrapperFunc(sh))
	for _, m := range s.Manifests() {
		mc, err := ManifestCompletion(sh, m)
		if err != nil {
			return nil, err
		}
		completions = append(completions, mc...)
	}
	return completions, nil
}

// ManifestCompletion scripts for a manifest
func ManifestCompletion(sh string, m *model.Manifest) ([]string, error) {
	var completions []string
	completions = append(completions, shellAliasFuncs(m))
	for _, cmd := range m.Commands {
		// Skip completion scripts for leaf nodes or pure aliases.
		// This allows for it to fallback to the shell's lookups.
		if cmd.AliasOnly || len(cmd.Commands) == 0 {
			continue
		}
		s, err := CommandCompletion(sh, cmd)
		if err != nil {
			return nil, err
		}
		completions = append(completions, s)
	}
	return completions, nil
}

// CommandCompletion script for a command
func CommandCompletion(sh string, cmd *model.Command) (string, error) {
	return Completion(sh, cmd.CobraCommand())
}
