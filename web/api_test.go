package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/version"
)

// setup creates a fresh nostromo home with a core manifest containing a few
// commands and one docked manifest, and returns a server for it
func setup(t *testing.T) *Server {
	t.Helper()

	ver := version.NewInfo("1.2.3", "abc", "today")
	config.SetVersion(ver)

	home := t.TempDir()
	t.Setenv("NOSTROMO_HOME", home)
	if err := os.MkdirAll(filepath.Join(home, "ships"), 0777); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	core := cfg.Spaceport().CoreManifest()
	core.Config.BackupCount = 5
	mustAdd(t, core, "docker", "docker", "docker root")
	mustAdd(t, core, "docker.up", "compose up -d", "start services")
	mustAdd(t, core, "docker.down", "compose down", "stop services")
	mustAdd(t, core, "git", "git", "")
	if err := core.AddSubstitution("docker", "docker-compose", "dc"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	docked := config.NewManifest("team")
	docked.Source = "https://example.com/team.yaml"
	mustAdd(t, docked, "deploy", "kubectl apply", "deploy the thing")
	mustAdd(t, docked, "deploy.prod", "kubectl apply -f prod", "")
	if err := config.SaveManifest(docked, false); err != nil {
		t.Fatal(err)
	}

	return NewServer(ver)
}

func mustAdd(t *testing.T, m *model.Manifest, keyPath, name, description string) {
	t.Helper()
	if _, err := m.AddCommand(keyPath, name, description, &model.Code{}, false, ""); err != nil {
		t.Fatalf("add %s: %s", keyPath, err)
	}
}

func request(t *testing.T, s *Server, method, target string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, target, &buf)
	req.Host = "127.0.0.1:8080"
	if method != http.MethodGet {
		req.Header.Set(csrfHeader, "1")
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode %q: %s", rec.Body.String(), err)
	}
}

func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("want status %d, got %d: %s", want, rec.Code, rec.Body.String())
	}
}

func TestMeta(t *testing.T) {
	s := setup(t)
	rec := request(t, s, http.MethodGet, "/api/meta", nil)
	expectStatus(t, rec, http.StatusOK)

	var meta metaResponse
	decode(t, rec, &meta)
	if meta.Version != "1.2.3" {
		t.Errorf("want version 1.2.3, got %q", meta.Version)
	}
	if len(meta.Modes) != 3 {
		t.Errorf("want 3 modes, got %v", meta.Modes)
	}
	if len(meta.Languages) == 0 {
		t.Errorf("want languages, got none")
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("api responses must not be cached")
	}
}

func TestListManifests(t *testing.T) {
	s := setup(t)
	rec := request(t, s, http.MethodGet, "/api/manifests", nil)
	expectStatus(t, rec, http.StatusOK)

	var infos []manifestInfo
	decode(t, rec, &infos)
	if len(infos) != 2 {
		t.Fatalf("want 2 manifests, got %d", len(infos))
	}

	byName := map[string]manifestInfo{}
	for _, info := range infos {
		byName[info.Name] = info
	}

	core, ok := byName[model.CoreManifestName]
	if !ok {
		t.Fatalf("core manifest missing from %v", infos)
	}
	if !core.Core || core.ReadOnly {
		t.Errorf("core manifest should be core and writable: %+v", core)
	}
	if core.CommandCount != 4 || core.RootCount != 2 {
		t.Errorf("want 4 commands and 2 roots, got %d and %d", core.CommandCount, core.RootCount)
	}
	if core.Path == "" {
		t.Errorf("want manifest path")
	}

	team, ok := byName["team"]
	if !ok {
		t.Fatalf("docked manifest missing from %v", infos)
	}
	if team.Core || !team.ReadOnly {
		t.Errorf("docked manifest should be read-only: %+v", team)
	}
	if team.Source != "https://example.com/team.yaml" {
		t.Errorf("want source preserved, got %q", team.Source)
	}
	if team.CommandCount != 2 {
		t.Errorf("want 2 commands, got %d", team.CommandCount)
	}
}

func TestGetManifestTree(t *testing.T) {
	s := setup(t)

	rec := request(t, s, http.MethodGet, "/api/manifest", nil)
	expectStatus(t, rec, http.StatusOK)

	var detail manifestDetail
	decode(t, rec, &detail)
	if detail.Name != model.CoreManifestName {
		t.Errorf("default manifest should be core, got %s", detail.Name)
	}
	if detail.Config == nil || detail.Config.BackupCount != 5 {
		t.Errorf("want core config with backup count 5, got %+v", detail.Config)
	}
	if len(detail.Commands) != 2 {
		t.Fatalf("want 2 root commands, got %d", len(detail.Commands))
	}
	if detail.Commands[0].KeyPath != "docker" || detail.Commands[1].KeyPath != "git" {
		t.Errorf("roots should be sorted, got %s and %s", detail.Commands[0].KeyPath, detail.Commands[1].KeyPath)
	}
	docker := detail.Commands[0]
	if len(docker.Commands) != 2 {
		t.Fatalf("want 2 children, got %d", len(docker.Commands))
	}
	if docker.Commands[0].KeyPath != "docker.down" || docker.Commands[1].KeyPath != "docker.up" {
		t.Errorf("children should be sorted, got %s and %s", docker.Commands[0].KeyPath, docker.Commands[1].KeyPath)
	}
	if len(docker.Subs) != 1 || docker.Subs[0].Alias != "dc" || docker.Subs[0].Name != "docker-compose" {
		t.Errorf("want dc substitution, got %+v", docker.Subs)
	}
	if docker.ReadOnly || docker.Manifest != model.CoreManifestName {
		t.Errorf("core commands should be writable, got %+v", docker)
	}

	rec = request(t, s, http.MethodGet, "/api/manifest?name=team", nil)
	expectStatus(t, rec, http.StatusOK)
	var team manifestDetail
	decode(t, rec, &team)
	if !team.ReadOnly || team.Config != nil {
		t.Errorf("docked manifest should be read-only without config, got %+v", team)
	}
	if len(team.Commands) != 1 || !team.Commands[0].ReadOnly {
		t.Errorf("docked commands should be read-only, got %+v", team.Commands)
	}

	rec = request(t, s, http.MethodGet, "/api/manifest?name=missing", nil)
	expectStatus(t, rec, http.StatusNotFound)
}

func TestGetCommand(t *testing.T) {
	s := setup(t)

	rec := request(t, s, http.MethodGet, "/api/command?keypath=docker.up", nil)
	expectStatus(t, rec, http.StatusOK)

	var node commandNode
	decode(t, rec, &node)
	if node.Alias != "up" || node.Name != "compose up -d" || node.Description != "start services" {
		t.Errorf("unexpected command: %+v", node)
	}
	if node.Mode != "concatenate" {
		t.Errorf("want default mode, got %s", node.Mode)
	}

	rec = request(t, s, http.MethodGet, "/api/command?keypath=deploy.prod", nil)
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &node)
	if !node.ReadOnly || node.Manifest != "team" {
		t.Errorf("docked command should be read-only, got %+v", node)
	}

	rec = request(t, s, http.MethodGet, "/api/command?keypath=nope", nil)
	expectStatus(t, rec, http.StatusNotFound)

	rec = request(t, s, http.MethodGet, "/api/command", nil)
	expectStatus(t, rec, http.StatusBadRequest)

	// Key paths shared between manifests resolve to the core manifest unless
	// a manifest is named explicitly
	rec = request(t, s, http.MethodPost, "/api/command", addCommandRequest{KeyPath: "deploy", Name: "make deploy"})
	expectStatus(t, rec, http.StatusCreated)
	for i := 0; i < 5; i++ {
		rec = request(t, s, http.MethodGet, "/api/command?keypath=deploy", nil)
		expectStatus(t, rec, http.StatusOK)
		decode(t, rec, &node)
		if node.Manifest != "manifest" || node.ReadOnly || node.Name != "make deploy" {
			t.Fatalf("want core command for ambiguous key path, got %+v", node)
		}
	}
	rec = request(t, s, http.MethodGet, "/api/command?keypath=deploy&manifest=team", nil)
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &node)
	if node.Manifest != "team" || !node.ReadOnly || node.Name != "kubectl apply" {
		t.Errorf("want docked command when manifest is named, got %+v", node)
	}
	rec = request(t, s, http.MethodGet, "/api/command?keypath=docker&manifest=team", nil)
	expectStatus(t, rec, http.StatusNotFound)
	rec = request(t, s, http.MethodGet, "/api/command?keypath=deploy&manifest=nope", nil)
	expectStatus(t, rec, http.StatusNotFound)
}

func TestAddCommand(t *testing.T) {
	s := setup(t)

	rec := request(t, s, http.MethodPost, "/api/command", addCommandRequest{
		KeyPath:     "docker.logs",
		Name:        "compose logs -f",
		Description: "tail logs",
		Mode:        "independent",
		Code:        codeInfo{Language: "sh", Snippet: "echo hi"},
	})
	expectStatus(t, rec, http.StatusCreated)

	var node commandNode
	decode(t, rec, &node)
	if node.KeyPath != "docker.logs" || node.Alias != "logs" || node.Mode != "independent" {
		t.Errorf("unexpected command: %+v", node)
	}
	if node.Code.Language != "sh" || node.Code.Snippet != "echo hi" {
		t.Errorf("want code preserved, got %+v", node.Code)
	}

	// Persisted through the normal save path
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Spaceport().CoreManifest().Find("docker.logs") == nil {
		t.Errorf("command was not saved to disk")
	}

	// Saving created a cargo backup
	backups, _ := filepath.Glob(filepath.Join(os.Getenv("NOSTROMO_HOME"), "cargo", "*.yaml"))
	if len(backups) == 0 {
		t.Errorf("want cargo backup to be created on save")
	}

	// Conflicts and validation
	rec = request(t, s, http.MethodPost, "/api/command", addCommandRequest{KeyPath: "docker.logs", Name: "x"})
	expectStatus(t, rec, http.StatusConflict)

	rec = request(t, s, http.MethodPost, "/api/command", addCommandRequest{KeyPath: "", Name: "x"})
	expectStatus(t, rec, http.StatusBadRequest)

	rec = request(t, s, http.MethodPost, "/api/command", addCommandRequest{KeyPath: "empty"})
	expectStatus(t, rec, http.StatusBadRequest)

	rec = request(t, s, http.MethodPost, "/api/command", addCommandRequest{KeyPath: "bad", Name: "x", Mode: "sideways"})
	expectStatus(t, rec, http.StatusBadRequest)

	rec = request(t, s, http.MethodPost, "/api/command", addCommandRequest{KeyPath: "bad", Code: codeInfo{Language: "cobol", Snippet: "x"}})
	expectStatus(t, rec, http.StatusBadRequest)

	rec = request(t, s, http.MethodPost, "/api/command", map[string]string{"unknown": "field"})
	expectStatus(t, rec, http.StatusBadRequest)
}

func TestAddCommandWithVersionlessDockedManifest(t *testing.T) {
	s := setup(t)

	// A hand-written docked manifest without version metadata must not break saves
	handwritten := config.NewManifest("handwritten")
	handwritten.Version = nil
	mustAdd(t, handwritten, "hello", "echo hello", "")
	if err := config.SaveManifest(handwritten, false); err != nil {
		t.Fatal(err)
	}

	rec := request(t, s, http.MethodPost, "/api/command", addCommandRequest{KeyPath: "docker.ps", Name: "compose ps"})
	expectStatus(t, rec, http.StatusCreated)
}

func TestUpdateCommand(t *testing.T) {
	s := setup(t)

	// Partial update keeps other fields
	rec := request(t, s, http.MethodPut, "/api/command?keypath=docker.up", map[string]interface{}{
		"description": "boot everything",
		"disabled":    true,
	})
	expectStatus(t, rec, http.StatusOK)

	var node commandNode
	decode(t, rec, &node)
	if node.Description != "boot everything" || !node.Disabled || node.Name != "compose up -d" {
		t.Errorf("unexpected update: %+v", node)
	}

	// Rename alias updates keypaths
	rec = request(t, s, http.MethodPut, "/api/command?keypath=docker", map[string]interface{}{
		"alias": "dk",
		"code":  codeInfo{Language: "python", Snippet: "print('hi')"},
	})
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &node)
	if node.KeyPath != "dk" || node.Alias != "dk" {
		t.Errorf("want renamed command, got %+v", node)
	}
	if node.Code.Language != "python" {
		t.Errorf("want code updated, got %+v", node.Code)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	core := cfg.Spaceport().CoreManifest()
	if core.Find("docker") != nil {
		t.Errorf("old alias should be gone")
	}
	if c := core.Find("dk.up"); c == nil || c.Description != "boot everything" {
		t.Errorf("children should follow the rename, got %+v", c)
	}

	// Alias collisions and invalid aliases are rejected
	rec = request(t, s, http.MethodPut, "/api/command?keypath=dk", map[string]interface{}{"alias": "git"})
	expectStatus(t, rec, http.StatusBadRequest)
	rec = request(t, s, http.MethodPut, "/api/command?keypath=dk", map[string]interface{}{"alias": "a.b"})
	expectStatus(t, rec, http.StatusBadRequest)
	rec = request(t, s, http.MethodPut, "/api/command?keypath=dk", map[string]interface{}{"mode": "bogus"})
	expectStatus(t, rec, http.StatusBadRequest)
	rec = request(t, s, http.MethodPut, "/api/command?keypath=dk", map[string]interface{}{"code": codeInfo{Language: "sh"}})
	expectStatus(t, rec, http.StatusBadRequest)

	// Empty alias keeps the current one
	rec = request(t, s, http.MethodPut, "/api/command?keypath=dk", map[string]interface{}{"alias": ""})
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &node)
	if node.KeyPath != "dk" {
		t.Errorf("want alias unchanged, got %+v", node)
	}

	// Docked and missing commands cannot be updated
	rec = request(t, s, http.MethodPut, "/api/command?keypath=deploy", map[string]interface{}{"description": "x"})
	expectStatus(t, rec, http.StatusForbidden)
	rec = request(t, s, http.MethodPut, "/api/command?keypath=nope", map[string]interface{}{"description": "x"})
	expectStatus(t, rec, http.StatusNotFound)
}

func TestRemoveCommand(t *testing.T) {
	s := setup(t)

	rec := request(t, s, http.MethodDelete, "/api/command?keypath=docker", nil)
	expectStatus(t, rec, http.StatusOK)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	core := cfg.Spaceport().CoreManifest()
	if core.Find("docker") != nil || core.Find("docker.up") != nil {
		t.Errorf("command and children should be removed")
	}
	if core.Find("git") == nil {
		t.Errorf("unrelated commands should remain")
	}

	rec = request(t, s, http.MethodDelete, "/api/command?keypath=docker", nil)
	expectStatus(t, rec, http.StatusNotFound)

	rec = request(t, s, http.MethodDelete, "/api/command?keypath=deploy", nil)
	expectStatus(t, rec, http.StatusForbidden)

	rec = request(t, s, http.MethodDelete, "/api/command", nil)
	expectStatus(t, rec, http.StatusBadRequest)
}

func TestSubstitutions(t *testing.T) {
	s := setup(t)

	rec := request(t, s, http.MethodPost, "/api/command/sub?keypath=git", subInfo{Name: "origin", Alias: "o"})
	expectStatus(t, rec, http.StatusCreated)

	var node commandNode
	decode(t, rec, &node)
	if len(node.Subs) != 1 || node.Subs[0].Alias != "o" || node.Subs[0].Name != "origin" {
		t.Errorf("want substitution added, got %+v", node.Subs)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Spaceport().CoreManifest().Find("git").Subs["o"]; !ok {
		t.Errorf("substitution was not saved")
	}

	rec = request(t, s, http.MethodPost, "/api/command/sub?keypath=git", subInfo{Name: "", Alias: "x"})
	expectStatus(t, rec, http.StatusBadRequest)

	rec = request(t, s, http.MethodDelete, "/api/command/sub?keypath=git&alias=o", nil)
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &node)
	if len(node.Subs) != 0 {
		t.Errorf("want substitution removed, got %+v", node.Subs)
	}

	rec = request(t, s, http.MethodDelete, "/api/command/sub?keypath=git&alias=o", nil)
	expectStatus(t, rec, http.StatusNotFound)

	rec = request(t, s, http.MethodDelete, "/api/command/sub?keypath=git", nil)
	expectStatus(t, rec, http.StatusBadRequest)

	rec = request(t, s, http.MethodPost, "/api/command/sub?keypath=deploy", subInfo{Name: "a", Alias: "b"})
	expectStatus(t, rec, http.StatusForbidden)
}

func TestSearch(t *testing.T) {
	s := setup(t)

	rec := request(t, s, http.MethodGet, "/api/search?q=compose", nil)
	expectStatus(t, rec, http.StatusOK)

	var res searchResponse
	decode(t, rec, &res)
	if len(res.Commands) != 2 {
		t.Errorf("want docker.up and docker.down, got %+v", res.Commands)
	}
	if len(res.Substitutions) != 1 || res.Substitutions[0].KeyPath != "docker" {
		t.Errorf("want docker-compose substitution match, got %+v", res.Substitutions)
	}

	// Matches across docked manifests are labelled
	rec = request(t, s, http.MethodGet, "/api/search?q=KUBECTL", nil)
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &res)
	if len(res.Commands) != 2 {
		t.Fatalf("want 2 kubectl commands, got %+v", res.Commands)
	}
	for _, c := range res.Commands {
		if c.Manifest != "team" || !c.ReadOnly {
			t.Errorf("want docked search result, got %+v", c)
		}
		if len(c.Commands) != 0 {
			t.Errorf("search results should be flat, got %+v", c.Commands)
		}
	}

	// A command with several matching substitutions is listed once
	rec = request(t, s, http.MethodPost, "/api/command/sub?keypath=docker", subInfo{Alias: "dcp", Name: "docker-compose-plus"})
	expectStatus(t, rec, http.StatusCreated)
	rec = request(t, s, http.MethodGet, "/api/search?q=docker-compose", nil)
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &res)
	if len(res.Substitutions) != 1 || res.Substitutions[0].KeyPath != "docker" {
		t.Errorf("want a single substitution match for docker, got %+v", res.Substitutions)
	}

	rec = request(t, s, http.MethodGet, "/api/search?q=", nil)
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &res)
	if len(res.Commands) != 0 || len(res.Substitutions) != 0 {
		t.Errorf("empty query should return no results")
	}
}

func TestSecurity(t *testing.T) {
	s := setup(t)

	// Mutations without the custom header are rejected
	req := httptest.NewRequest(http.MethodDelete, "/api/command?keypath=git", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	expectStatus(t, rec, http.StatusForbidden)

	// Non loopback hosts are rejected (DNS rebinding)
	req = httptest.NewRequest(http.MethodGet, "/api/manifests", nil)
	req.Host = "evil.example.com"
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	expectStatus(t, rec, http.StatusForbidden)

	// Cross origin requests are rejected
	req = httptest.NewRequest(http.MethodGet, "/api/manifests", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Origin", "http://evil.example.com")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	expectStatus(t, rec, http.StatusForbidden)

	req = httptest.NewRequest(http.MethodGet, "/api/manifests", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	expectStatus(t, rec, http.StatusForbidden)

	// Same origin requests are fine
	req = httptest.NewRequest(http.MethodGet, "/api/manifests", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Origin", "http://localhost:8080")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	expectStatus(t, rec, http.StatusOK)

	// Config was not modified by the rejected delete
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Spaceport().CoreManifest().Find("git") == nil {
		t.Errorf("rejected request must not mutate config")
	}
}

func TestStatic(t *testing.T) {
	s := setup(t)

	rec := request(t, s, http.MethodGet, "/", nil)
	expectStatus(t, rec, http.StatusOK)
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("want html, got %s", ct)
	}
	if !strings.Contains(rec.Body.String(), "<title>nostromo</title>") {
		t.Errorf("index.html not served")
	}
	if rec.Header().Get("Content-Security-Policy") == "" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("security headers missing")
	}

	for _, tt := range []struct{ path, ctype string }{
		{"/app.js", "text/javascript"},
		{"/style.css", "text/css"},
		{"/favicon.svg", "image/svg+xml"},
	} {
		rec = request(t, s, http.MethodGet, tt.path, nil)
		expectStatus(t, rec, http.StatusOK)
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, tt.ctype) {
			t.Errorf("%s: want %s, got %s", tt.path, tt.ctype, ct)
		}
	}

	// No directory listings or traversal
	rec = request(t, s, http.MethodGet, "/static/", nil)
	expectStatus(t, rec, http.StatusNotFound)
	rec = request(t, s, http.MethodGet, "/../server.go", nil)
	if rec.Code == http.StatusOK {
		t.Errorf("path traversal must not serve files")
	}
	rec = request(t, s, http.MethodGet, "/missing.txt", nil)
	expectStatus(t, rec, http.StatusNotFound)
	rec = request(t, s, http.MethodGet, "/api/nope", nil)
	expectStatus(t, rec, http.StatusNotFound)
}
