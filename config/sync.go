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
func (c *Config) Sync(force, keep bool, sources []string) ([]*model.Manifest, error) {
	manifests := []*model.Manifest{}
	sources, err := syncPrep(sources)
	if err != nil {
		return manifests, err
	}

	defer syncCleanup(keep)

	// Fallback to all existing manifests if no sources provided
	if len(sources) == 0 {
		for _, m := range c.spaceport.Manifests() {
			if m.IsCore() {
				continue
			}
			sources = append(sources, m.Source)
		}
	}

	for _, source := range sources {
		// Check if manifest name was provided
		if m := c.spaceport.FindManifest(source); m != nil {
			source = m.Source
		}

		err := syncDownload(source)
		if err != nil {
			return manifests, err
		}
	}

	manifests, err = c.syncMerge(sources, force)
	if err != nil {
		return manifests, err
	}

	if err := SaveSpaceport(c.spaceport); err != nil {
		return manifests, err
	}

	return manifests, nil
}

func syncPrep(sources []string) ([]string, error) {
	// Ensure downloads folder exists
	if err := pathutil.EnsurePath(downloadsPath()); err != nil {
		return nil, err
	}

	// Sanitize github web urls
	s := []string{}
	for _, source := range sources {
		if strings.Contains(source, "github") && strings.Contains(source, "blob") {
			// Adjust github link to point to raw file
			source = strings.Replace(source, "blob", "raw", 1)
			s = append(s, source)
		} else {
			// Fallback to same url
			s = append(s, source)
		}
	}

	return s, nil
}

func syncCleanup(keep bool) error {
	// Persist downloads folder if requested
	if keep {
		return nil
	}

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
		log.Regularf("signal %v\n", sig)
	case <-ctx.Done():
		wg.Wait()
		log.Regular("success")
	case err := <-errChan:
		wg.Wait()
		log.Regularf("failed, %s\n", err)
		return err
	}

	return nil
}

func (c *Config) syncMerge(sources []string, force bool) ([]*model.Manifest, error) {
	// Read all files from download folder
	manifests := []*model.Manifest{}
	downloadsPath := downloadsPath()
	files, err := ioutil.ReadDir(downloadsPath)
	if err != nil {
		return manifests, err
	}

	// Parse each file
	for _, file := range files {
		fname := file.Name()
		path := filepath.Join(downloadsPath, fname)
		m, err := Parse(path)
		if err != nil {
			return manifests, err
		}

		// Check for conflicts
		if m.Name == model.CoreManifestName {
			// Duplicate core, so rename to file name
			name := strings.TrimSuffix(fname, filepath.Ext(fname))
			m.Name = name
		}

		// Update path
		m.Path = filepath.Join(manifestsPath(), fmt.Sprintf(DefaultConfigFile, m.Name))

		// Update source
		for _, source := range sources {
			base := strings.TrimSuffix(filepath.Base(source), filepath.Ext(m.Path))
			if base == m.Name {
				m.Source = source
			}
		}

		var shouldSave bool
		if c.spaceport.IsUnique(m.Name) {
			// New manifest
			c.spaceport.AddManifest(m)
			shouldSave = true
			log.Infof("adding %s manifest\n", m.Name)
		} else if u := c.spaceport.FindManifest(m.Name); u != nil {
			// Update manifest
			if force || m.Version.UUID != u.Version.UUID {
				c.spaceport.AddManifest(m)
				shouldSave = true
				log.Infof("updating %s manifest\n", m.Name)
			}
		} else {
			// Should not be possible
			panic("manifest not found")
		}

		if shouldSave {
			err = SaveManifest(m, false)
			if err != nil {
				log.Warningf("failed to save manifest %s\n", m.Name)
			}
		}

		manifests = append(manifests, m)
	}

	return manifests, nil
}
