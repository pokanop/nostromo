package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
)

// Shell type
type Shell int

// Supported shells
const (
	Bash Shell = iota
	Zsh
)

var validLanguages = []string{"sh", "ruby", "python", "perl", "js"}

var (
	initFiles = loadStartupFiles()
	prefFiles = preferredStartupFiles(initFiles)
	prefFile  = currentStartupFile(initFiles)
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
func InitFileLines() (string, error) {
	if prefFile == nil {
		return "", fmt.Errorf("could not find current init file")
	}
	return prefFile.contentBlock()
}

// Which shell is currently running
func Which() Shell {
	sh := os.Getenv("SHELL")
	if strings.Contains(sh, "zsh") {
		return Zsh
	}
	return Bash
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

func shellWrapperFunc() string {
	// Sources completion scripts after each command in case something changes
	sh := "bash"
	if Which() == Zsh {
		sh = "zsh"
	}
	return fmt.Sprintf("__nostromo_cmd() { command ./nostromo \"$@\"; }\nnostromo() { __nostromo_cmd \"$@\" && eval \"$(__nostromo_cmd completion %s)\"; }", sh)
}

func shellAliasFuncs(m *model.Manifest) string {
	var aliases []string
	for _, c := range m.Commands {
		var alias string
		if c.AliasOnly {
			alias = fmt.Sprintf("alias %s='%s'", c.Alias, c.Name)
		} else {
			// This will generate a shell command provided to the completion script
			// generation. When users run a command, it actually runs `eval` on
			// the result of `nostromo eval` with arguments resolved.
			cmd := fmt.Sprintf("__nostromo_cmd eval %s \"$@\"", c.Alias)
			alias = strings.TrimSpace(fmt.Sprintf("%s() { eval $(%s); }", c.Alias, cmd))
		}
		aliases = append(aliases, alias)
	}
	return fmt.Sprintf("\n%s\n", strings.Join(aliases, "\n"))
}
