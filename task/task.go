package task

import (
	"fmt"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/model"
)

// InitConfig of nostromo config file if not already initialized
func InitConfig() {
	cfg := config.NewConfig("~/.nostromo", model.NewManifest())
	if cfg.Exists() {
		fmt.Println(".nostromo config already exists")
		return
	}

	err := cfg.Save()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("created .nostromo config")
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

	fmt.Println("removed .nostromo config")
}

// ShowConfig for nostromo config file
func ShowConfig() {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	fmt.Println(cfg.Manifest.AsJSON())
}

// AddCommand to the manifest
func AddCommand(keyPath, command string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.AddCommand(keyPath, command)
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

	sh, err := cfg.Manifest.ExecutionString(args)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sh)
}

func checkConfig() *config.Config {
	cfg, err := config.Parse("~/.nostromo")
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
}
