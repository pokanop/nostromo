package shell

import (
	"fmt"
	"os"
	"os/exec"
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

// Run a command on the shell
func Run(command string, verbose bool) error {
	if len(command) == 0 {
		return fmt.Errorf("cannot run empty command")
	}

	command = strings.TrimSuffix(command, "\n")

	name, args := buildExecArgs(command)
	if verbose {
		log.Debugf("executing: %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

// Commit manifest updates to shell initialization files
//
// Loads all shell config files and replaces nostromo aliases
// with manifest's commands.
func Commit(manifest *model.Manifest) error {
	initFiles := loadStartupFiles()
	prefFiles := preferredStartupFiles(initFiles)
	if len(prefFiles) == 0 {
		return fmt.Errorf("could not find preferred init file")
	}

	for _, p := range prefFiles {
		// Forget previous aliases
		p.reset()

		// Since nostromo works by aliasing only the top level commands,
		// iterate the manifest's list and update.
		for _, cmd := range manifest.Commands {
			p.add(cmd.Alias)
		}
	}

	for _, f := range initFiles {
		err := f.commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// InitFileLines returns the shell initialization file lines
func InitFileLines() (string, error) {
	initFiles := loadStartupFiles()
	prefFiles := preferredStartupFiles(initFiles)
	if len(prefFiles) == 0 {
		return "", fmt.Errorf("could not find preferred init file")
	}

	return prefFiles[0].makeAliasBlock(), nil
}

// Which shell is currently running
func Which() Shell {
	sh := os.Getenv("SHELL")
	if strings.Contains(sh, "zsh") {
		return Zsh
	}
	return Bash
}

func buildExecArgs(cmd string) (string, []string) {
	if Which() == Zsh {
		return "zsh", []string{"-ic", cmd}
	}
	return "bash", []string{"-ic", cmd}
}
