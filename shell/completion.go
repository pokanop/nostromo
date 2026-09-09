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

	return redirectCompletionRequest(sh, cmd.Name(), buf.String())
}

const rootCommandName = "nostromo"

// completionRequest describes how a cobra generated completion script asks
// the program for completions, and what to replace it with.
//
// Cobra scripts invoke the typed program name to get completions. For
// nostromo itself that would call the `nostromo` wrapper function, which
// re-sources completions after every call (and recurses in fish), so the
// real binary is invoked instead. For user commands, which are shell
// functions that `eval` the command, completions are requested from
// `nostromo __complete run <name>` where the `run` cobra command hosts them.
type completionRequest struct {
	find    string
	root    string
	command string // `%[1]s` is the user command's name
}

var completionRequests = map[string]completionRequest{
	Bash: {
		find:    `requestComp="${words[0]} __complete ${args[*]}"`,
		root:    `requestComp="command nostromo __complete ${args[*]}"`,
		command: `requestComp="__nostromo_cmd __complete run %[1]s ${args[*]}"`,
	},
	Zsh: {
		find:    `requestComp="${words[1]} __complete ${words[2,-1]}"`,
		root:    `requestComp="command nostromo __complete ${words[2,-1]}"`,
		command: `requestComp="__nostromo_cmd __complete run %[1]s ${words[2,-1]}"`,
	},
	Fish: {
		find:    `$args[1] __complete $args[2..-1] $lastArg"`,
		root:    `command nostromo __complete $args[2..-1] $lastArg"`,
		command: `__nostromo_cmd __complete run %[1]s $args[2..-1] $lastArg"`,
	},
	Powershell: {
		find:    `$RequestComp="$Program __complete $Arguments"`,
		root:    `$RequestComp="& (Get-Command nostromo -CommandType Application | Select-Object -First 1).Source __complete $Arguments"`,
		command: `$RequestComp="__nostromo_cmd __complete run %[1]s $Arguments"`,
	},
}

// redirectCompletionRequest rewrites how a generated completion script
// requests completions from the program, see completionRequest.
func redirectCompletionRequest(sh, name, script string) (string, error) {
	r, ok := completionRequests[sh]
	if !ok {
		return "", fmt.Errorf("unsupported shell: %s", sh)
	}
	if !strings.Contains(script, r.find) {
		return "", fmt.Errorf("unable to redirect %s completion request for %s", sh, name)
	}
	replace := fmt.Sprintf(r.command, name)
	if name == rootCommandName {
		replace = r.root
	}
	return strings.ReplaceAll(script, r.find, replace), nil
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
	completions = append(completions, shellAliasFuncs(sh, m))
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
