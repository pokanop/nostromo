package shell

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
)

// Supported shells
const (
	Bash       = "bash"
	Zsh        = "zsh"
	Fish       = "fish"
	Powershell = "powershell"
)

var validLanguages = []string{"sh", "ruby", "python", "perl", "js"}

var (
	initFiles = loadStartupFiles()
	prefFiles = preferredStartupFiles(initFiles)
)

// EvalString returns the command as a string to evaluate or an error.
func EvalString(command, language string, verbose bool) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("cannot run empty command")
	}

	command = strings.TrimSuffix(command, "\n")

	cmdStr := buildEvalCmd(command, language)
	if verbose {
		log.Debugf("executing: %s\n", cmdStr)
	}

	return cmdStr, nil
}

// Commit manifest updates to shell initialization files
//
// Loads all shell config files and replaces nostromo aliases
// with manifest's commands.
func Commit(manifest *model.Manifest) error {
	if len(prefFiles) == 0 {
		return fmt.Errorf("could not find preferred init file [%s]", strings.Join(preferredFilenames, ", "))
	}

	for _, f := range initFiles {
		// Apply the manifest
		f.apply(manifest)

		// Write updated file
		if f.canCommit() {
			err := f.commit()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// InitFileLines returns the shell initialization file lines
func InitFileLines() string {
	var s string
	for _, prefFile := range prefFiles {
		s += fmt.Sprintf("|%s|", prefFile.name())
		c, err := prefFile.contentBlock()
		if err == nil {
			s += c
		}
	}
	return s
}

// SupportedLanguages that can be executed
func SupportedLanguages() []string {
	return validLanguages
}

// IsSupportedLanguage returns true if supported snippet language and false otherwise
func IsSupportedLanguage(language string) bool {
	for _, l := range validLanguages {
		if language == l {
			return true
		}
	}
	return false
}

func buildEvalCmd(cmd, language string) string {
	switch language {
	case "ruby":
		return fmt.Sprintf("ruby -e '%s'", cmd)
	case "python":
		return fmt.Sprintf("python -c '%s'", cmd)
	case "perl":
		return fmt.Sprintf("perl -e '%s'", cmd)
	case "js":
		return fmt.Sprintf("node -e '%s'", cmd)
	case "sh":
		fallthrough
	default:
		return cmd
	}
}

// shellWrapperFunc returns the `nostromo` wrapper function for a shell.
//
// `__nostromo_cmd` always invokes the real binary and is what every other
// generated function (and the completion scripts) call. The `nostromo`
// wrapper re-sources the completion scripts after each successful command
// in case something changed.
func shellWrapperFunc(sh string) string {
	switch sh {
	case Fish:
		return "function __nostromo_cmd; command nostromo $argv; end\n" +
			"function nostromo; __nostromo_cmd $argv; and __nostromo_cmd completion fish | source; end"
	case Powershell:
		return "function __nostromo_cmd { & (Get-Command nostromo -CommandType Application | Select-Object -First 1).Source @args }\n" +
			"function nostromo { __nostromo_cmd @args; if ($?) { __nostromo_cmd completion powershell | Out-String | Invoke-Expression } }"
	default:
		return fmt.Sprintf("__nostromo_cmd() { command nostromo \"$@\"; }\nnostromo() { __nostromo_cmd \"$@\" && eval \"$(__nostromo_cmd completion %s)\"; }", sh)
	}
}

// shellAliasFuncs returns the function/alias definitions for a manifest's
// top level commands in the given shell's syntax.
//
// When users run a command, it actually runs `eval` on the result of
// `nostromo eval` with arguments resolved. Commands unavailable on the
// current platform are left out.
func shellAliasFuncs(sh string, m *model.Manifest) string {
	var aliases []string
	for _, c := range m.Commands {
		if !c.IsAvailable() {
			continue
		}
		aliases = append(aliases, shellAliasFunc(sh, c))
	}
	sort.Strings(aliases)
	return fmt.Sprintf("\n%s\n", strings.Join(aliases, "\n"))
}

func shellAliasFunc(sh string, c *model.Command) string {
	switch sh {
	case Fish:
		if c.AliasOnly {
			return fmt.Sprintf("alias %s='%s'", c.Alias, c.Name)
		}
		return fmt.Sprintf("function %s; eval (__nostromo_cmd eval %s $argv | string collect); end", c.Alias, c.Alias)
	case Powershell:
		if c.AliasOnly {
			// Set-Alias cannot carry arguments, so use a function
			return fmt.Sprintf("function %s { %s @args }", c.Alias, c.Name)
		}
		return fmt.Sprintf("function %s { Invoke-Expression (__nostromo_cmd eval %s @args | Out-String) }", c.Alias, c.Alias)
	default:
		if c.AliasOnly {
			return fmt.Sprintf("alias %s='%s'", c.Alias, c.Name)
		}
		return fmt.Sprintf("%s() { eval $(__nostromo_cmd eval %s \"$@\"); }", c.Alias, c.Alias)
	}
}
