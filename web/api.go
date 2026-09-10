package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/shell"
	"github.com/pokanop/nostromo/task"
)

const maxBodyBytes = 1 << 20

type errorResponse struct {
	Error string `json:"error"`
}

type metaResponse struct {
	Version   string   `json:"version"`
	Modes     []string `json:"modes"`
	Languages []string `json:"languages"`
}

type manifestInfo struct {
	Name         string `json:"name"`
	Source       string `json:"source"`
	Path         string `json:"path"`
	Version      string `json:"version"`
	Core         bool   `json:"core"`
	ReadOnly     bool   `json:"readOnly"`
	CommandCount int    `json:"commandCount"`
	RootCount    int    `json:"rootCount"`
}

type configInfo struct {
	Verbose     bool   `json:"verbose"`
	AliasesOnly bool   `json:"aliasesOnly"`
	Mode        string `json:"mode"`
	BackupCount int    `json:"backupCount"`
}

type manifestDetail struct {
	manifestInfo
	Config   *configInfo    `json:"config,omitempty"`
	Commands []*commandNode `json:"commands"`
}

type codeInfo struct {
	Language string `json:"language"`
	Snippet  string `json:"snippet"`
}

type subInfo struct {
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type commandNode struct {
	KeyPath     string         `json:"keyPath"`
	Name        string         `json:"name"`
	Alias       string         `json:"alias"`
	Description string         `json:"description"`
	Mode        string         `json:"mode"`
	AliasOnly   bool           `json:"aliasOnly"`
	Disabled    bool           `json:"disabled"`
	Code        codeInfo       `json:"code"`
	Subs        []subInfo      `json:"subs"`
	Commands    []*commandNode `json:"commands"`
	Manifest    string         `json:"manifest"`
	ReadOnly    bool           `json:"readOnly"`
}

type addCommandRequest struct {
	KeyPath     string   `json:"keyPath"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Mode        string   `json:"mode"`
	AliasOnly   bool     `json:"aliasOnly"`
	Code        codeInfo `json:"code"`
}

type updateCommandRequest struct {
	Alias       string   `json:"alias"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Mode        string   `json:"mode"`
	AliasOnly   bool     `json:"aliasOnly"`
	Disabled    bool     `json:"disabled"`
	Code        codeInfo `json:"code"`
}

type searchResponse struct {
	Query         string         `json:"query"`
	Commands      []*commandNode `json:"commands"`
	Substitutions []*commandNode `json:"substitutions"`
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	semver := ""
	if s.ver != nil {
		semver = s.ver.SemVer
	}
	writeJSON(w, http.StatusOK, metaResponse{
		Version:   semver,
		Modes:     model.SupportedModes(),
		Languages: shell.SupportedLanguages(),
	})
}

func (s *Server) handleManifests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := task.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	infos := []manifestInfo{}
	for _, m := range cfg.Spaceport().Manifests() {
		infos = append(infos, newManifestInfo(m))
	}
	writeJSON(w, http.StatusOK, infos)
}

func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := task.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		name = model.CoreManifestName
	}
	m := cfg.Spaceport().FindManifest(name)
	if m == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no manifest named %s found", name))
		return
	}

	writeJSON(w, http.StatusOK, newManifestDetail(m))
}

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := task.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	keyPath := strings.TrimSpace(r.URL.Query().Get("keypath"))

	switch r.Method {
	case http.MethodGet:
		if keyPath == "" {
			writeError(w, http.StatusBadRequest, "keypath is required")
			return
		}
		cmd, m := findCommand(cfg, strings.TrimSpace(r.URL.Query().Get("manifest")), keyPath)
		if cmd == nil {
			writeError(w, http.StatusNotFound, "command not found")
			return
		}
		writeJSON(w, http.StatusOK, newCommandNode(cmd, m))

	case http.MethodPost:
		var req addCommandRequest
		if err := readJSON(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.KeyPath = strings.TrimSpace(req.KeyPath)
		if req.KeyPath == "" {
			writeError(w, http.StatusBadRequest, "keyPath is required")
			return
		}
		if req.Name == "" && req.Code.Snippet == "" {
			writeError(w, http.StatusBadRequest, "must provide command or code snippet")
			return
		}
		m := cfg.Spaceport().CoreManifest()
		if m.Find(req.KeyPath) != nil {
			writeError(w, http.StatusConflict, fmt.Sprintf("command %s already exists", req.KeyPath))
			return
		}
		cmd, err := task.AddCommandToConfig(cfg, req.KeyPath, req.Name, req.Description, req.Code.Snippet, req.Code.Language, req.AliasOnly, req.Mode, false)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, newCommandNode(cmd, m))

	case http.MethodPut:
		m, status, err := writableCommand(cfg, keyPath)
		if err != nil {
			writeError(w, status, err.Error())
			return
		}
		existing := m.Find(keyPath)
		req := updateCommandRequest{
			Alias:       existing.Alias,
			Name:        existing.Name,
			Description: existing.Description,
			Mode:        existing.Mode.String(),
			AliasOnly:   existing.AliasOnly,
			Disabled:    existing.Disabled,
			Code:        newCodeInfo(existing.Code),
		}
		if err := readJSON(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		cmd, err := task.UpdateCommandInConfig(cfg, keyPath, task.CommandUpdate{
			Alias:       strings.TrimSpace(req.Alias),
			Name:        req.Name,
			Description: req.Description,
			Mode:        req.Mode,
			AliasOnly:   req.AliasOnly,
			Disabled:    req.Disabled,
			Language:    req.Code.Language,
			Snippet:     req.Code.Snippet,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, newCommandNode(cmd, m))

	case http.MethodDelete:
		if _, status, err := writableCommand(cfg, keyPath); err != nil {
			writeError(w, status, err.Error())
			return
		}
		if err := task.RemoveCommandFromConfig(cfg, keyPath); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"removed": keyPath})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSubstitution(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := task.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	keyPath := strings.TrimSpace(r.URL.Query().Get("keypath"))
	m, status, err := writableCommand(cfg, keyPath)
	if err != nil {
		writeError(w, status, err.Error())
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req subInfo
		if err := readJSON(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		cmd, err := task.AddSubstitutionToConfig(cfg, keyPath, req.Name, req.Alias)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, newCommandNode(cmd, m))

	case http.MethodDelete:
		alias := r.URL.Query().Get("alias")
		if alias == "" {
			writeError(w, http.StatusBadRequest, "alias is required")
			return
		}
		if _, ok := m.Find(keyPath).Subs[alias]; !ok {
			writeError(w, http.StatusNotFound, "substitution not found")
			return
		}
		if err := task.RemoveSubstitutionFromConfig(cfg, keyPath, alias); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, newCommandNode(m.Find(keyPath), m))

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := task.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	resp := searchResponse{Query: q, Commands: []*commandNode{}, Substitutions: []*commandNode{}}
	if q == "" {
		writeJSON(w, http.StatusOK, resp)
		return
	}

	owners := commandOwners(cfg)
	cmds, subs := task.FindMatches(cfg, q)
	for _, c := range cmds {
		resp.Commands = append(resp.Commands, newFlatCommandNode(c, owners[c]))
	}
	for _, c := range subs {
		resp.Substitutions = append(resp.Substitutions, newFlatCommandNode(c, owners[c]))
	}
	sortNodes(resp.Commands)
	sortNodes(resp.Substitutions)

	writeJSON(w, http.StatusOK, resp)
}

// findCommand looks up a key path in the named manifest, or when no name is
// given, in the core manifest first and then docked manifests in order so the
// result is deterministic when manifests share key paths
func findCommand(cfg *config.Config, manifest, keyPath string) (*model.Command, *model.Manifest) {
	sp := cfg.Spaceport()
	if manifest != "" {
		m := sp.FindManifest(manifest)
		if m == nil {
			return nil, nil
		}
		return m.Find(keyPath), m
	}
	if cmd := sp.CoreManifest().Find(keyPath); cmd != nil {
		return cmd, sp.CoreManifest()
	}
	for _, m := range sp.Manifests() {
		if cmd := m.Find(keyPath); cmd != nil {
			return cmd, m
		}
	}
	return nil, nil
}

// writableCommand finds the manifest for a core command that can be edited
// or returns a status code and error describing why it cannot be
func writableCommand(cfg *config.Config, keyPath string) (*model.Manifest, int, error) {
	if keyPath == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("keypath is required")
	}
	core := cfg.Spaceport().CoreManifest()
	if core.Find(keyPath) != nil {
		return core, http.StatusOK, nil
	}
	if cmd, m := findCommand(cfg, "", keyPath); cmd != nil {
		return nil, http.StatusForbidden, fmt.Errorf("%s belongs to docked manifest %s which is read-only, use the CLI to detach or edit it", keyPath, m.Name)
	}
	return nil, http.StatusNotFound, fmt.Errorf("command not found")
}

func newManifestInfo(m *model.Manifest) manifestInfo {
	info := manifestInfo{
		Name:      m.Name,
		Source:    m.Source,
		Path:      m.Path,
		Core:      m.IsCore(),
		ReadOnly:  !m.IsCore(),
		RootCount: len(m.Commands),
	}
	if m.Version != nil {
		info.Version = m.Version.SemVer
	}
	for _, cmd := range m.Commands {
		cmd.Walk(func(c *model.Command, stop *bool) {
			info.CommandCount++
		})
	}
	return info
}

func newManifestDetail(m *model.Manifest) manifestDetail {
	d := manifestDetail{
		manifestInfo: newManifestInfo(m),
		Commands:     []*commandNode{},
	}
	if m.IsCore() && m.Config != nil {
		d.Config = &configInfo{
			Verbose:     m.Config.Verbose,
			AliasesOnly: m.Config.AliasesOnly,
			Mode:        m.Config.Mode.String(),
			BackupCount: m.Config.BackupCount,
		}
	}
	for _, cmd := range m.Commands {
		d.Commands = append(d.Commands, newCommandNode(cmd, m))
	}
	sortNodes(d.Commands)
	return d
}

func newCommandNode(c *model.Command, m *model.Manifest) *commandNode {
	n := newFlatCommandNode(c, m)
	for _, child := range c.Commands {
		n.Commands = append(n.Commands, newCommandNode(child, m))
	}
	sortNodes(n.Commands)
	return n
}

func newFlatCommandNode(c *model.Command, m *model.Manifest) *commandNode {
	n := &commandNode{
		KeyPath:     c.KeyPath,
		Name:        c.Name,
		Alias:       c.Alias,
		Description: c.Description,
		Mode:        c.Mode.String(),
		AliasOnly:   c.AliasOnly,
		Disabled:    c.Disabled,
		Code:        newCodeInfo(c.Code),
		Subs:        []subInfo{},
		Commands:    []*commandNode{},
	}
	if m != nil {
		n.Manifest = m.Name
		n.ReadOnly = !m.IsCore()
	}
	for _, sub := range c.Subs {
		if sub == nil {
			continue
		}
		n.Subs = append(n.Subs, subInfo{Name: sub.Name, Alias: sub.Alias})
	}
	sort.Slice(n.Subs, func(i, j int) bool { return n.Subs[i].Alias < n.Subs[j].Alias })
	return n
}

func newCodeInfo(c *model.Code) codeInfo {
	if c == nil {
		return codeInfo{}
	}
	return codeInfo{Language: c.Language, Snippet: c.Snippet}
}

func commandOwners(cfg *config.Config) map[*model.Command]*model.Manifest {
	owners := map[*model.Command]*model.Manifest{}
	for _, m := range cfg.Spaceport().Manifests() {
		for _, cmd := range m.Commands {
			manifest := m
			cmd.Walk(func(c *model.Command, stop *bool) {
				owners[c] = manifest
			})
		}
	}
	return owners
}

func sortNodes(nodes []*commandNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Manifest != nodes[j].Manifest {
			return nodes[i].Manifest < nodes[j].Manifest
		}
		return nodes[i].KeyPath < nodes[j].KeyPath
	})
}

func readJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid request body: %s", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
