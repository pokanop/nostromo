package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/pokanop/nostromo/model"
)

// manifestServer serves manifests over http so link fetches go through
// go-getter like a real remote source
type manifestServer struct {
	*httptest.Server
	mu    sync.Mutex
	files map[string]string
}

func newManifestServer(t *testing.T) *manifestServer {
	s := &manifestServer{files: map[string]string{}}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		body, ok := s.files[strings.TrimPrefix(r.URL.Path, "/")]
		s.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *manifestServer) url(file string) string {
	return s.URL + "/" + file
}

// serve a manifest with the given uuid that links the given files
func (s *manifestServer) serve(name, uuid string, links ...string) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "name: %s\nversion:\n  uuid: %s\ncommands:\n  %s:\n    keypath: %s\n    name: %s\n    alias: %s\n    code:\n      language: sh\n      snippet: echo %s\n", name, uuid, name, name, name, name, name)
	if len(links) > 0 {
		fmt.Fprint(b, "links:\n")
		for _, l := range links {
			fmt.Fprintf(b, "- name: %s\n  source: %s\n", strings.TrimSuffix(l, ".yaml"), s.url(l))
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[name+".yaml"] = b.String()
	return s.url(name + ".yaml")
}

// newTestConfig creates a saved nostromo config in a temp home
func newTestConfig(t *testing.T) *Config {
	baseDir := t.TempDir()
	t.Setenv("NOSTROMO_HOME", baseDir)
	if err := os.MkdirAll(manifestsPath(), 0755); err != nil {
		t.Fatal(err)
	}
	cfg, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func manifestNames(manifests []*model.Manifest) []string {
	names := []string{}
	for _, m := range manifests {
		names = append(names, m.Name)
	}
	return names
}

func assertManifests(t *testing.T, cfg *Config, want ...string) {
	t.Helper()
	got := manifestNames(cfg.Spaceport().Manifests())
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("manifests = %v, want %v", got, want)
	}
	for _, name := range want {
		if _, err := os.Stat(manifestFile(name)); err != nil {
			t.Errorf("manifest file for %s missing: %s", name, err)
		}
	}
	files, _ := filepath.Glob(filepath.Join(manifestsPath(), "*.yaml"))
	if len(files) != len(want) {
		t.Errorf("manifest files = %v, want %d", files, len(want))
	}
}

func assertDocked(t *testing.T, cfg *Config, want ...string) {
	t.Helper()
	got := cfg.Spaceport().Docked
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("docked = %v, want %v", got, want)
	}
}

func TestLink(t *testing.T) {
	srv := newManifestServer(t)
	srv.serve("util", "util-1")
	tools := srv.serve("tools", "tools-1", "util.yaml")
	cfg := newTestConfig(t)

	manifests, err := cfg.Link(tools, "", false, false)
	if err != nil {
		t.Fatalf("link failed: %s", err)
	}
	if got := manifestNames(manifests); strings.Join(got, ",") != "tools,util" {
		t.Errorf("linked = %v, want [tools util]", got)
	}
	assertManifests(t, cfg, model.CoreManifestName, "tools", "util")
	assertDocked(t, cfg)

	core := cfg.Spaceport().CoreManifest()
	if len(core.Links) != 1 || core.Links[0].Name != "tools" || core.Links[0].UUID != "tools-1" || core.Links[0].Source != tools {
		t.Errorf("unexpected core links: %+v", core.Links)
	}

	// Linking again is a no-op
	if _, err := cfg.Link(tools, "", false, false); err != nil {
		t.Fatalf("relink failed: %s", err)
	}
	if len(core.Links) != 1 {
		t.Errorf("want 1 link, got %d", len(core.Links))
	}

	// Everything must survive a reload, including link provenance
	cfg, err = LoadConfig()
	if err != nil {
		t.Fatalf("reload failed: %s", err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "tools", "util")
	assertDocked(t, cfg)
	if got := cfg.Spaceport().CoreManifest().Links; len(got) != 1 || got[0].Name != "tools" {
		t.Errorf("links not persisted: %+v", got)
	}
	if cmd, _ := cfg.Spaceport().FindCommand("util"); cmd == nil {
		t.Errorf("want linked commands available")
	}
}

func TestLinkToManifest(t *testing.T) {
	srv := newManifestServer(t)
	srv.serve("util", "util-1")
	tools := srv.serve("tools", "tools-1")
	cfg := newTestConfig(t)

	if _, err := cfg.Link(tools, "", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.Link(srv.url("util.yaml"), "tools", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.Link(srv.url("util.yaml"), "missing", false, false); err == nil || !strings.Contains(err.Error(), "no manifest named missing") {
		t.Errorf("want missing target error, got %v", err)
	}

	tm := cfg.Spaceport().FindManifest("tools")
	if len(tm.Links) != 1 || tm.Links[0].Name != "util" {
		t.Errorf("unexpected tools links: %+v", tm.Links)
	}
	assertManifests(t, cfg, model.CoreManifestName, "tools", "util")

	// Link must be persisted on the target manifest
	m, err := Parse(manifestFile("tools"))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Links) != 1 || m.Links[0].Source != srv.url("util.yaml") {
		t.Errorf("links not saved: %+v", m.Links)
	}
}

func TestLinkNotFound(t *testing.T) {
	srv := newManifestServer(t)
	cfg := newTestConfig(t)

	if _, err := cfg.Link(srv.url("missing.yaml"), "", false, false); err == nil {
		t.Errorf("want error for missing source")
	}
	assertManifests(t, cfg, model.CoreManifestName)
	if len(cfg.Spaceport().CoreManifest().Links) != 0 {
		t.Errorf("want no links recorded")
	}
}

func TestLinkCollision(t *testing.T) {
	srv := newManifestServer(t)
	other := newManifestServer(t)
	util := srv.serve("util", "util-1")
	otherUtil := other.serve("util", "util-2")
	tools := other.serve("tools", "tools-1", "util.yaml")
	cfg := newTestConfig(t)

	if _, err := cfg.Link(util, "", false, false); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		source string
	}{
		{"direct", otherUtil},
		{"transitive", tools},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cfg.Link(tt.source, "", false, false)
			if err == nil || !strings.Contains(err.Error(), "manifest collision: util already exists from "+util) {
				t.Fatalf("want collision error, got %v", err)
			}
			assertManifests(t, cfg, model.CoreManifestName, "util")
			if got := cfg.Spaceport().FindManifest("util").Version.UUID; got != "util-1" {
				t.Errorf("util replaced, uuid = %s", got)
			}
			if len(cfg.Spaceport().CoreManifest().Links) != 1 {
				t.Errorf("want collision to leave links untouched")
			}
		})
	}
}

func TestSyncCollision(t *testing.T) {
	srv := newManifestServer(t)
	other := newManifestServer(t)
	util := srv.serve("util", "util-1")
	other.serve("util", "util-2")
	tools := other.serve("tools", "tools-1", "util.yaml")
	cfg := newTestConfig(t)

	if _, err := cfg.Sync(false, false, []string{util}); err != nil {
		t.Fatal(err)
	}
	// A linked manifest may never replace one from a different source
	if _, err := cfg.Sync(false, false, []string{tools}); err == nil || !strings.Contains(err.Error(), "manifest collision") {
		t.Fatalf("want collision error, got %v", err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "util")
	assertDocked(t, cfg, "util")
}

func TestLinkCircular(t *testing.T) {
	srv := newManifestServer(t)
	a := srv.serve("a", "a-1", "b.yaml")
	b := srv.serve("b", "b-1", "a.yaml")
	cfg := newTestConfig(t)

	// Docking a cycle terminates and keeps both manifests
	if _, err := cfg.Sync(false, false, []string{a}); err != nil {
		t.Fatalf("dock failed: %s", err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "a", "b")
	assertDocked(t, cfg, "a")

	tests := []struct {
		name   string
		source string
		target string
	}{
		{"links back to target", b, "a"},
		{"is the target", a, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cfg.Link(tt.source, tt.target, false, false)
			if err == nil || !strings.Contains(err.Error(), "circular reference") {
				t.Fatalf("want circular error, got %v", err)
			}
		})
	}
	if am := cfg.Spaceport().FindManifest("a"); len(am.Links) != 1 {
		t.Errorf("want a links untouched, got %+v", am.Links)
	}

	// Syncing a cycle terminates as well
	if _, err := cfg.Sync(false, false, nil); err != nil {
		t.Fatalf("sync failed: %s", err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "a", "b")
}

func TestUnlink(t *testing.T) {
	srv := newManifestServer(t)
	srv.serve("util", "util-1")
	tools := srv.serve("tools", "tools-1", "util.yaml")
	cfg := newTestConfig(t)

	if _, err := cfg.Link(tools, "", false, false); err != nil {
		t.Fatal(err)
	}

	removed, err := cfg.Unlink("tools", "")
	if err != nil {
		t.Fatalf("unlink failed: %s", err)
	}
	if strings.Join(removed, ",") != "tools,util" {
		t.Errorf("removed = %v, want [tools util]", removed)
	}
	assertManifests(t, cfg, model.CoreManifestName)
	if len(cfg.Spaceport().CoreManifest().Links) != 0 {
		t.Errorf("want link removed")
	}

	if _, err := cfg.Unlink("tools", ""); err == nil || !strings.Contains(err.Error(), "not linked") {
		t.Errorf("want not linked error, got %v", err)
	}

	cfg, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	assertManifests(t, cfg, model.CoreManifestName)
}

func TestUnlinkKeepsReferenced(t *testing.T) {
	srv := newManifestServer(t)
	util := srv.serve("util", "util-1")
	tools := srv.serve("tools", "tools-1", "util.yaml")
	extra := srv.serve("extra", "extra-1", "util.yaml")
	cfg := newTestConfig(t)

	// util is docked directly and linked from tools and extra
	if _, err := cfg.Sync(false, false, []string{util}); err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.Link(tools, "", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.Link(extra, "", false, false); err != nil {
		t.Fatal(err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "util", "tools", "extra")
	assertDocked(t, cfg, "util")

	// Unlinking tools removes only tools, util is still docked and linked
	removed, err := cfg.Unlink("tools", "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(removed, ",") != "tools" {
		t.Errorf("removed = %v, want [tools]", removed)
	}
	assertManifests(t, cfg, model.CoreManifestName, "util", "extra")

	// Unlinking extra by source leaves util since it was docked directly
	if _, err := cfg.Unlink(extra, ""); err != nil {
		t.Fatal(err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "util")
	assertDocked(t, cfg, "util")
}

func TestSyncFollowsLinks(t *testing.T) {
	srv := newManifestServer(t)
	srv.serve("util", "util-1")
	tools := srv.serve("tools", "tools-1", "util.yaml")
	cfg := newTestConfig(t)

	if _, err := cfg.Sync(false, false, []string{tools}); err != nil {
		t.Fatal(err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "tools", "util")
	assertDocked(t, cfg, "tools")

	// Upstream drops the link, so syncing removes the orphaned util
	srv.serve("tools", "tools-2")
	if _, err := cfg.Sync(false, false, nil); err != nil {
		t.Fatal(err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "tools")
	if got := cfg.Spaceport().FindManifest("tools").Version.UUID; got != "tools-2" {
		t.Errorf("tools not updated, uuid = %s", got)
	}

	// Upstream adds a new link, so syncing fetches it
	srv.serve("more", "more-1")
	srv.serve("tools", "tools-3", "more.yaml")
	if _, err := cfg.Sync(false, false, nil); err != nil {
		t.Fatal(err)
	}
	assertManifests(t, cfg, model.CoreManifestName, "tools", "more")
	assertDocked(t, cfg, "tools")
}

func TestSyncCoreLinks(t *testing.T) {
	srv := newManifestServer(t)
	tools := srv.serve("tools", "tools-1")
	cfg := newTestConfig(t)

	if _, err := cfg.Link(tools, "", false, false); err != nil {
		t.Fatal(err)
	}

	// Linked manifests are refreshed on sync and the link record follows
	srv.serve("tools", "tools-2")
	if _, err := cfg.Sync(false, false, nil); err != nil {
		t.Fatal(err)
	}
	if got := cfg.Spaceport().FindManifest("tools").Version.UUID; got != "tools-2" {
		t.Errorf("tools not updated, uuid = %s", got)
	}
	if got := cfg.Spaceport().CoreManifest().Links[0].UUID; got != "tools-2" {
		t.Errorf("link uuid = %s, want tools-2", got)
	}
	m, err := Parse(coreManifestPath())
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Links[0].UUID; got != "tools-2" {
		t.Errorf("saved link uuid = %s, want tools-2", got)
	}
}

func TestPrune(t *testing.T) {
	srv := newManifestServer(t)
	srv.serve("util", "util-1")
	tools := srv.serve("tools", "tools-1", "util.yaml")
	cfg := newTestConfig(t)

	if _, err := cfg.Sync(false, false, []string{tools}); err != nil {
		t.Fatal(err)
	}

	// Undocking tools leaves util orphaned until pruned
	if err := cfg.DeleteManifest("tools"); err != nil {
		t.Fatal(err)
	}
	cfg.Spaceport().RemoveManifest("tools")
	removed, err := cfg.Prune()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(removed, ",") != "util" {
		t.Errorf("removed = %v, want [util]", removed)
	}
	assertManifests(t, cfg, model.CoreManifestName)
	assertDocked(t, cfg)
}

func TestSameSource(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"https://example.com/a.yaml", "https://example.com/a.yaml", true},
		{"https://github.com/o/r/blob/main/a.yaml", "https://github.com/o/r/raw/main/a.yaml", true},
		{"file:/path/a.yaml", "file:///path/a.yaml", true},
		{"https://example.com/a.yaml", "https://example.com/b.yaml", false},
	}
	for _, tt := range tests {
		if got := sameSource(tt.a, tt.b); got != tt.want {
			t.Errorf("sameSource(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
