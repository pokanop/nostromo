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

var validLanguages = []string{"ruby", "python", "perl", "js", "sh"}

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
	initFiles := loadStartupFiles()
	prefFiles := preferredStartupFiles(initFiles)
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
	initFiles := loadStartupFiles()
	prefFile := currentStartupFile(initFiles)
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

// ValidLanguages that can be executed
func ValidLanguages() []string {
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
