package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	return cmd.Run()
}
