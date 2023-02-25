package config

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-getter"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
)

type syncItem struct {
	// Identifier is a unique string used for the destination
	identifier string
	// Source is the source URL to download
	source string
	// Destination is the destination directory to download to
	destination string
	// RelativePath from the destination to the manifest
	relativePath string
	// SyncPath is the path to the manifest after sync
	syncPath string
}

func newSyncItem(source string) *syncItem {
	identifier := uuid.NewString()
	destination := path.Join(downloadsPath(), identifier)
	return &syncItem{
		identifier:  identifier,
		source:      source,
		destination: destination,
	}
}

func (i *syncItem) setPath(p string) {
	i.relativePath = strings.TrimPrefix(p, i.destination)
	i.syncPath = path.Join(i.source, i.relativePath)
}

// Sync adds a new manifest from provided sources
func (c *Config) Sync(force, keep bool, sources []string) ([]*model.Manifest, error) {
	manifests := []*model.Manifest{}
	sources, err := c.syncPrep(sources)
	if err != nil {
		return manifests, err
	}

	defer c.syncCleanup(keep)

	// Fallback to all existing manifests if no sources provided
	if len(sources) == 0 {
		// Track unique sources
		seen := map[string]bool{}
		for _, m := range c.spaceport.Manifests() {
			if m.IsCore() {
				continue
			}
			if seen[m.Source] {
				continue
			}
			seen[m.Source] = true
			sources = append(sources, m.Source)
		}
	}

	// Prepare sync items for sources
	items := []*syncItem{}
	for _, source := range sources {
		items = append(items, newSyncItem(source))
	}

	pwd, err := os.Getwd()
	if err != nil {
		return manifests, err
	}

	for _, item := range items {
		// Check if manifest name was provided
		if m := c.spaceport.FindManifest(item.source); m != nil {
			item.source = m.Source
		}

		err := c.syncDownload(pwd, item)
		if err != nil {
			return manifests, err
		}
	}

	manifests, err = c.syncMerge(items, force)
	if err != nil {
		return manifests, err
	}
	if len(manifests) == 0 {
		return manifests, fmt.Errorf("no manifests found")
	}

	if err := SaveSpaceport(c.spaceport); err != nil {
		return manifests, err
	}

	return manifests, nil
}

func (c *Config) syncPrep(sources []string) ([]string, error) {
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

func (c *Config) syncCleanup(keep bool) error {
	// Persist downloads folder if requested
	if keep {
		return nil
	}

	// Remove downloads folder
	return os.RemoveAll(downloadsPath())
}

func (c *Config) syncDownload(pwd string, item *syncItem) error {
	ctx, cancel := context.WithCancel(context.Background())
	client := &getter.Client{
		Ctx:     ctx,
		Src:     item.source,
		Dst:     item.destination,
		Pwd:     pwd,
		Mode:    getter.ClientModeAny,
		Options: []getter.ClientOption{},
	}

	log.Infof("downloading %s...", item.source)

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

func (c *Config) syncMerge(items []*syncItem, force bool) ([]*model.Manifest, error) {
	// Read all files from download folder
	manifests := []*model.Manifest{}

	// For each sync item parse downloaded files
	for _, item := range items {
		// Parse each destination folder
		err := filepath.Walk(item.destination, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Update item paths
			item.setPath(path)

			fname := info.Name()
			m, err := Parse(path)
			if err != nil {
				return nil
			}

			// Check for conflicts
			if m.Name == model.CoreManifestName {
				// Duplicate core, so rename to file name with timestamp
				name := strings.TrimSuffix(fname, filepath.Ext(fname))
				name = fmt.Sprintf("%s-%s", name, time.Now().Format("20060102150405"))
				m.Name = name
			}

			// Update path
			m.Path = filepath.Join(manifestsPath(), fmt.Sprintf(DefaultConfigFile, m.Name))

			// Update source
			m.Source = item.source

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
			return nil
		})
		if err != nil {
			return manifests, err
		}
	}

	return manifests, nil
}
