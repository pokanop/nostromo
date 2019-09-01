package task

import (
	"strings"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/pokanop/nostromo/shell"
)

// InitConfig of nostromo config file if not already initialized
func InitConfig() {
	cfg := checkConfig()
	if cfg == nil {
		cfg = config.NewConfig(config.ConfigPath, model.NewManifest())
		err := pathutil.EnsurePath("~/.nostromo")
		if err != nil {
			log.Error(err)
			return
		}
	} else {
		log.Highlight("nostromo config exists, updating")
	}

	err := saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// DestroyConfig deletes nostromo config file
func DestroyConfig() {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Delete()
	if err != nil {
		log.Error(err)
		return
	}
}

// ShowConfig for nostromo config file
func ShowConfig() {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	log.Highlight(cfg.Manifest.AsJSON())

	lines, err := shell.InitFileLines()
	if err != nil {
		return
	}
	log.Highlight(lines)
}

// SetConfig updates properties for nostromo settings
func SetConfig(key, value string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Set(key, value)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// GetConfig reads properties from nostromo settings
func GetConfig(key string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	log.Highlight(cfg.Get(key))
}

// AddCommand to the manifest
func AddCommand(keyPath, command, description string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.AddCommand(keyPath, command, description)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}

	log.FieldsCompact(cfg.Manifest.Find(keyPath))
}

// RemoveCommand from the manifest
func RemoveCommand(keyPath string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.RemoveCommand(keyPath)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// AddSubstitution to the manifest
func AddSubstitution(keyPath, name, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.AddSubstitution(keyPath, name, alias)
	if err != nil {
		log.Error(err)
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}

	log.FieldsCompact(cfg.Manifest.Find(keyPath))
}

// RemoveSubstitution from the manifest
func RemoveSubstitution(keyPath, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.RemoveSubstitution(keyPath, alias)
	if err != nil {
		log.Error(err)
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// Run a command from the manifest
func Run(args []string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	cmd, err := cfg.Manifest.ExecutionString(sanitizeArgs(args))
	if err != nil {
		log.Error(err)
		return
	}

	if cfg.Manifest.Config.Verbose {
		log.Debug("executing:", cmd)
	}

	err = shell.Run(cmd)
	if err != nil {
		log.Error(err)
	}
}

func checkConfig() *config.Config {
	cfg, err := config.Parse(config.ConfigPath)
	if err != nil {
		log.Error(err)
		return nil
	}
	return cfg
}

func saveConfig(cfg *config.Config) error {
	err := cfg.Save()
	if err != nil {
		return err
	}

	err = shell.Commit(cfg.Manifest)
	if err != nil {
		return err
	}

	return nil
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
