package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pokanop/nostromo/model"
)

// Run a command on the shell
func Run(command string) error {
	if len(command) == 0 {
		return fmt.Errorf("cannot run empty command")
	}

	fmt.Println(command)

	command = strings.TrimSuffix(command, "\n")

	cmd := exec.Command("sh", "-c", command)
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
	p := preferredStartupFile(initFiles)
	if p == nil {
		return fmt.Errorf("could not find preferred init file")
	}

	// Forget previous aliases
	p.reset()

	// Since nostromo works by aliasing only the top level commands,
	// iterate the manifest's list and update.
	for _, cmd := range manifest.Commands {
		p.add(cmd.Alias)
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
	p := preferredStartupFile(initFiles)
	if p == nil {
		return "", fmt.Errorf("could not find preferred init file")
	}

	return p.makeAliasBlock(), nil
}
