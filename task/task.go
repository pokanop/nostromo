package task

import (
	"fmt"
	"strings"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/pokanop/nostromo/shell"
)

// InitConfig of nostromo config file if not already initialized
func InitConfig() {
	cfg := config.NewConfig("~/.nostromo/config", model.NewManifest())
	if cfg.Exists() {
		fmt.Println("nostromo config already exists")
		return
	}

	err := pathutil.EnsurePath("~/.nostromo")
	if err != nil {
		fmt.Println(err)
		return
	}

	saveConfig(cfg)
}

// DestroyConfig deletes nostromo config file
func DestroyConfig() {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Delete()
	if err != nil {
		fmt.Println(err)
		return
	}
}

// ShowConfig for nostromo config file
func ShowConfig() {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	fmt.Println(cfg.Manifest.AsJSON())

	lines, err := shell.InitFileLines()
	if err != nil {
		return
	}
	fmt.Println(lines)
}

// AddCommand to the manifest
func AddCommand(keyPath, command, description string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.AddCommand(keyPath, command, description)
	if err != nil {
		fmt.Println(err)
		return
	}

	saveConfig(cfg)
}

// RemoveCommand from the manifest
func RemoveCommand(keyPath string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.RemoveCommand(keyPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	saveConfig(cfg)
}

// AddSubstitution to the manifest
func AddSubstitution(keyPath, name, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.AddSubstitution(keyPath, name, alias)
	if err != nil {
		fmt.Println(err)
	}

	saveConfig(cfg)
}

// RemoveSubstitution from the manifest
func RemoveSubstitution(keyPath, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.RemoveSubstitution(keyPath, alias)
	if err != nil {
		fmt.Println(err)
	}

	saveConfig(cfg)
}

// Run a command from the manifest
func Run(args []string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	cmd, err := cfg.Manifest.ExecutionString(sanitizeArgs(args))
	if err != nil {
		fmt.Println(err)
		return
	}

	err = shell.Run(cmd)
	if err != nil {
		fmt.Println(err)
	}
}

func checkConfig() *config.Config {
	cfg, err := config.Parse("~/.nostromo/config")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return cfg
}

func saveConfig(cfg *config.Config) {
	err := cfg.Save()
	if err != nil {
		fmt.Println(err)
	}

	err = shell.Commit(cfg.Manifest)
	if err != nil {
		fmt.Println(err)
	}
}

func sanitizeArgs(args []string) []string {
	sanitizedArgs := []string{}
	for _, arg := range args {
		if len(arg) > 0 {
			sanitizedArgs = append(sanitizedArgs, strings.TrimSpace(arg))
		}
	}
	return sanitizedArgs
}
