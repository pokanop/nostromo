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
	cfg, err := config.Parse("~/.nostromo")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cfg.Delete()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("removed .nostromo config")
}

// ShowConfig for nostromo config file
func ShowConfig() {
	cfg, err := config.Parse("~/.nostromo")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.Manifest.AsJSON())
}
