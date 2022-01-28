package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-getter"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
)

// Sync adds a new manifest from provided sources
func (c *Config) Sync(force bool, sources []string) error {
	if err := syncPrep(); err != nil {
		return err
	}

	defer syncCleanup()

	for _, source := range sources {
		err := syncDownload(source)
		if err != nil {
			return err
		}
	}

	if err := c.syncMerge(force); err != nil {
		return err
	}

	if err := saveSpaceport(c.spaceport); err != nil {
		return err
	}

	return nil
}

func syncPrep() error {
	// Ensure downloads folder exists
	return pathutil.EnsurePath(downloadsPath())
}

func syncCleanup() error {
	// Remove downloads folder
	return os.RemoveAll(downloadsPath())
}

func syncDownload(source string) error {
	ctx, cancel := context.WithCancel(context.Background())
	client := &getter.Client{
		Ctx:     ctx,
		Src:     source,
		Dst:     downloadsPath(),
		Mode:    getter.ClientModeAny,
		Options: []getter.ClientOption{},
	}

	log.Infof("downloading %s...", source)

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	select {
	case sig := <-ch:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		log.Regularf("signal %v", sig)
	case <-ctx.Done():
		wg.Wait()
		log.Regular("success")
	case err := <-errChan:
		wg.Wait()
		log.Regularf("failed, %s", err)
	}

	return nil
}

func (c *Config) syncMerge(force bool) error {
	// Read all files from download folder
	downloadsPath := downloadsPath()
	files, err := ioutil.ReadDir(downloadsPath)
	if err != nil {
		return err
	}

	// Parse each file
	for _, file := range files {
		fname := file.Name()
		path := filepath.Join(downloadsPath, fname)
		m, err := parse(path)
		if err != nil {
			return err
		}

		// Check for conflicts
		if m.Name == model.CoreManifestName {
			// Duplicate core, so rename to file name
			name := strings.TrimSuffix(fname, filepath.Ext(fname))
			m.Name = name
		}

		// Update path
		m.Path = filepath.Join(manifestsPath(), fmt.Sprintf(DefaultConfigFile, m.Name))

		var shouldSave bool
		if c.spaceport.IsUnique(m.Name) {
			// New manifest
			c.spaceport.AddManifest(m)
			shouldSave = true
			log.Infof("adding %s manifest", m.Name)
		} else if u := c.spaceport.FindManifest(m.Name); u != nil {
			// Update manifest
			if force || m.Version.UUID != u.Version.UUID {
				c.spaceport.AddManifest(m)
				shouldSave = true
				log.Infof("updating %s manifest", m.Name)
			}
		} else {
			// Should not be possible
			panic("manifest not found")
		}

		if shouldSave {
			err = saveManifest(m, false)
			if err != nil {
				log.Warningf("failed to save manifest %s", m.Name)
			}
		}
	}

	return nil
}
