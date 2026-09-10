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

// Sync adds new manifests from provided sources or updates every docked
// manifest when no sources are given
//
// Manifests linked from the fetched manifests are fetched as well, so
// docking or syncing a manifest also pulls in its dependencies. When
// syncing everything, manifests that are no longer linked or docked are
// removed.
func (c *Config) Sync(force, keep bool, sources []string) ([]*model.Manifest, error) {
	manifests := []*model.Manifest{}
	sources, err := c.syncPrep(sources)
	if err != nil {
		return manifests, err
	}

	defer c.syncCleanup(keep)

	// Docking explicitly given sources, otherwise refresh every root
	dock := len(sources) > 0
	if !dock {
		// Track unique sources
		seen := map[string]bool{}
		for _, m := range c.spaceport.Roots() {
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

	r, err := c.newResolver(force)
	if err != nil {
		return manifests, err
	}

	top := []*model.Manifest{}
	for _, source := range sources {
		// Check if manifest name was provided
		if m := c.spaceport.FindManifest(source); m != nil {
			source = m.Source
		}

		ms, err := r.fetch(source, nil)
		if err != nil {
			return manifests, err
		}
		top = append(top, ms...)
	}

	// Links from the core manifest are always followed
	if err := r.resolveLinks(c.spaceport.CoreManifest(), nil); err != nil {
		return manifests, err
	}

	manifests = r.apply()
	if len(manifests) == 0 {
		return manifests, fmt.Errorf("no manifests found")
	}
	if dock {
		for _, m := range top {
			c.spaceport.Dock(m.Name)
		}
	} else if _, err := c.Prune(); err != nil {
		return manifests, err
	}

	if err := SaveSpaceport(c.spaceport); err != nil {
		return manifests, err
	}

	return manifests, nil
}

// Link fetches a manifest from source, docks it along with its own links
// and records the dependency on the target manifest
//
// The core manifest is the target when no name is given. Linking fails if
// the source resolves to a manifest that already exists from a different
// source or if it would introduce a circular reference back to the target.
func (c *Config) Link(source, target string, force, keep bool) ([]*model.Manifest, error) {
	manifests := []*model.Manifest{}
	t, err := c.linkTarget(target)
	if err != nil {
		return manifests, err
	}

	sources, err := c.syncPrep([]string{source})
	if err != nil {
		return manifests, err
	}
	source = sources[0]

	defer c.syncCleanup(keep)

	r, err := c.newResolver(force)
	if err != nil {
		return manifests, err
	}
	r.strict = true

	linked, err := r.fetch(source, []*model.Manifest{t})
	if err != nil {
		return manifests, err
	}
	if len(linked) == 0 {
		return manifests, fmt.Errorf("no manifests found")
	}

	// Anything fetched that is the target or links back to it is circular
	for _, m := range r.order {
		if m.Name == t.Name || sameSource(m.Source, t.Source) {
			return manifests, fmt.Errorf("circular reference: %s links back to %s", source, t.Name)
		}
		if m.LinksTo(t) {
			return manifests, fmt.Errorf("circular reference: %s links back to %s", m.Name, t.Name)
		}
	}

	manifests = r.apply()

	added := []string{}
	for _, m := range linked {
		if t.AddLink(model.NewLinkedManifest(m)) {
			added = append(added, m.Name)
		} else {
			log.Warningf("%s already linked to %s\n", m.Name, t.Name)
		}
	}
	if len(added) > 0 {
		if err := c.SaveManifest(t, t.IsCore()); err != nil {
			return manifests, err
		}
	}

	if err := SaveSpaceport(c.spaceport); err != nil {
		return manifests, err
	}

	return manifests, nil
}

// Unlink removes the dependency on the named manifest from the target
// manifest and removes any manifests nothing references anymore
//
// Returns the names of the manifests that were removed.
func (c *Config) Unlink(name, target string) ([]string, error) {
	t, err := c.linkTarget(target)
	if err != nil {
		return nil, err
	}

	l := t.RemoveLink(name)
	if l == nil {
		return nil, fmt.Errorf("%s is not linked to %s", name, t.Name)
	}

	if err := c.SaveManifest(t, t.IsCore()); err != nil {
		return nil, err
	}

	removed, err := c.Prune()
	if err != nil {
		return removed, err
	}

	if err := SaveSpaceport(c.spaceport); err != nil {
		return removed, err
	}

	return removed, nil
}

// Prune removes manifests that are neither docked nor linked from a root
//
// Returns the names of the removed manifests.
func (c *Config) Prune() ([]string, error) {
	removed := []string{}
	for _, m := range c.spaceport.Orphans() {
		log.Infof("removing unlinked %s manifest\n", m.Name)
		if err := c.DeleteManifest(m.Name); err != nil && !os.IsNotExist(err) {
			return removed, err
		}
		c.spaceport.RemoveManifest(m.Name)
		removed = append(removed, m.Name)
	}
	return removed, nil
}

// linkTarget returns the named manifest or the core manifest if empty
func (c *Config) linkTarget(name string) (*model.Manifest, error) {
	if len(name) == 0 {
		return c.spaceport.CoreManifest(), nil
	}
	m := c.spaceport.FindManifest(name)
	if m == nil {
		return nil, fmt.Errorf("no manifest named %s found", name)
	}
	return m, nil
}

func (c *Config) syncPrep(sources []string) ([]string, error) {
	// Ensure downloads folder exists
	if err := pathutil.EnsurePath(downloadsPath()); err != nil {
		return nil, err
	}

	s := []string{}
	for _, source := range sources {
		s = append(s, sanitizeSource(source))
	}

	return s, nil
}

// sanitizeSource adjusts github web urls to point to the raw file
func sanitizeSource(source string) string {
	if strings.Contains(source, "github") && strings.Contains(source, "blob") {
		return strings.Replace(source, "blob", "raw", 1)
	}
	return source
}

// sameSource checks if two sources refer to the same location
func sameSource(a, b string) bool {
	return normalizeFileSource(sanitizeSource(a)) == normalizeFileSource(sanitizeSource(b))
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

// syncParse reads every manifest found in a downloaded item
//
// Parsed manifests are renamed if they collide with the core manifest and
// have their path and source pointed at the local store, but nothing is
// added to the spaceport or written to disk.
func (c *Config) syncParse(item *syncItem) ([]*model.Manifest, error) {
	manifests := []*model.Manifest{}
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

		// Docked manifests never carry settings
		m.Config = nil

		manifests = append(manifests, m)
		return nil
	})
	return manifests, err
}

// syncApply adds or updates a fetched manifest in the spaceport and saves
// it when changed, returning whether it was saved
func (c *Config) syncApply(m *model.Manifest, force, linked bool) bool {
	var shouldSave bool
	verb := "adding"
	if linked {
		verb = "linking"
	}
	if c.spaceport.IsUnique(m.Name) {
		// New manifest
		c.spaceport.AddManifest(m)
		shouldSave = true
		log.Infof("%s %s manifest\n", verb, m.Name)
	} else if u := c.spaceport.FindManifest(m.Name); u != nil {
		// Update manifest
		if force || manifestUUID(m) != manifestUUID(u) {
			if !sameSource(m.Source, u.Source) {
				log.Warningf("replacing %s manifest from %s with %s\n", m.Name, u.Source, m.Source)
			}
			c.spaceport.AddManifest(m)
			shouldSave = true
			log.Infof("updating %s manifest\n", m.Name)
		}
	} else {
		// Should not be possible
		panic("manifest not found")
	}

	if shouldSave {
		if err := c.SaveManifest(m, false); err != nil {
			log.Warningf("failed to save manifest %s\n", m.Name)
		}
	}
	return shouldSave
}

func manifestUUID(m *model.Manifest) string {
	if m == nil || m.Version == nil {
		return ""
	}
	return m.Version.UUID
}

// resolver fetches manifests and follows their links before anything is
// applied to the spaceport
type resolver struct {
	c     *Config
	pwd   string
	force bool
	// strict fails on name collisions instead of replacing manifests
	strict bool
	// fetched manifests by source, in fetch order
	sources map[string][]*model.Manifest
	order   []*model.Manifest
	// linked marks manifests pulled in through links rather than directly
	linked map[string]bool
	// updated holds manifests whose link records were rewritten
	updated map[string]*model.Manifest
}

func (c *Config) newResolver(force bool) (*resolver, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &resolver{
		c:       c,
		pwd:     pwd,
		force:   force,
		sources: map[string][]*model.Manifest{},
		linked:  map[string]bool{},
		updated: map[string]*model.Manifest{},
	}, nil
}

// fetch downloads and parses a source then follows its links
//
// The path holds the manifests leading to this fetch and is used to cut
// circular references. Sources are only fetched once.
func (r *resolver) fetch(source string, path []*model.Manifest) ([]*model.Manifest, error) {
	if ms, ok := r.sources[source]; ok {
		return ms, nil
	}
	// Reserve the source before downloading so cycles terminate
	r.sources[source] = nil

	item := newSyncItem(source)
	if err := r.c.syncDownload(r.pwd, item); err != nil {
		return nil, err
	}

	ms, err := r.c.syncParse(item)
	if err != nil {
		return nil, err
	}

	linked := len(path) > 0
	for _, m := range ms {
		if err := r.checkCollision(m, linked); err != nil {
			return nil, err
		}
		if linked {
			r.linked[m.Name] = true
		}
		r.order = append(r.order, m)
	}
	r.sources[source] = ms

	for _, m := range ms {
		if err := r.resolveLinks(m, path); err != nil {
			return nil, err
		}
	}

	return ms, nil
}

// resolveLinks fetches every manifest linked from m
func (r *resolver) resolveLinks(m *model.Manifest, path []*model.Manifest) error {
	if m == nil {
		return nil
	}
	path = append(path, m)

	for _, l := range m.Links {
		source := sanitizeSource(l.Source)
		if len(source) == 0 {
			return fmt.Errorf("invalid link %s in %s manifest: missing source", l.Name, m.Name)
		}

		if p := r.inPath(path, l); p != nil {
			log.Warningf("circular reference: %s -> %s, skipping\n", pathString(path), p.Name)
			continue
		}

		ms, err := r.fetch(source, path)
		if err != nil {
			return fmt.Errorf("unable to fetch %s linked from %s: %s", l.Name, m.Name, err)
		}

		if len(ms) == 0 {
			log.Warningf("no manifests found for %s linked from %s\n", l.Name, m.Name)
		}

		// Keep the link record in step with what the source resolves to
		if len(ms) == 1 && (ms[0].Name != l.Name || manifestUUID(ms[0]) != l.UUID) {
			log.Debugf("updating link %s in %s manifest to %s\n", l.Name, m.Name, ms[0].Name)
			l.Name = ms[0].Name
			l.UUID = manifestUUID(ms[0])
			r.updated[m.Name] = m
		}
	}

	return nil
}

// inPath returns the manifest in path that the link refers to, if any
func (r *resolver) inPath(path []*model.Manifest, l *model.LinkedManifest) *model.Manifest {
	for _, p := range path {
		if l.Name == p.Name || sameSource(l.Source, p.Source) {
			return p
		}
	}
	return nil
}

// checkCollision fails if a manifest with the same name exists from a
// different source
//
// Directly docked manifests are allowed to replace existing ones unless
// strict, linked manifests never are.
func (r *resolver) checkCollision(m *model.Manifest, linked bool) error {
	collision := func(existing string) error {
		return fmt.Errorf("manifest collision: %s already exists from %s, cannot add it from %s", m.Name, existing, m.Source)
	}
	for _, f := range r.order {
		if f.Name == m.Name && !sameSource(f.Source, m.Source) {
			return collision(f.Source)
		}
	}
	if !linked && !r.strict {
		return nil
	}
	if existing := r.c.spaceport.FindManifest(m.Name); existing != nil && !sameSource(existing.Source, m.Source) {
		return collision(existing.Source)
	}
	return nil
}

// apply adds every fetched manifest to the spaceport and saves any
// manifest whose link records changed
func (r *resolver) apply() []*model.Manifest {
	for _, m := range r.order {
		if r.c.syncApply(m, r.force, r.linked[m.Name]) {
			delete(r.updated, m.Name)
		}
	}
	for _, m := range r.updated {
		if err := r.c.SaveManifest(m, m.IsCore()); err != nil {
			log.Warningf("failed to save manifest %s\n", m.Name)
		}
	}
	return r.order
}

func pathString(path []*model.Manifest) string {
	names := []string{}
	for _, m := range path {
		names = append(names, m.Name)
	}
	return strings.Join(names, " -> ")
}
