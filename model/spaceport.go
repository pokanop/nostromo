package model

import (
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/version"
)

// Spaceport type that manages and docks multiple ships' manifests
//
// The spaceport also holds the global nostromo settings that apply to
// every docked manifest.
type Spaceport struct {
	manifests map[string]*Manifest
	Sequence  []string `json:"sequence" yaml:"sequence"`
	// Docked holds manifests the user docked explicitly, as opposed to ones
	// pulled in through links, so cleanup knows what is safe to remove
	Docked []string `json:"docked" yaml:"docked"`
	Config *Config  `json:"config" yaml:"config"`
	// LegacyTheme is the pre-config top level theme setting, only populated
	// when loading an older spaceport and lifted into Config by Migrate
	LegacyTheme *log.ThemeType `json:"-" yaml:"theme,omitempty"`
}

func NewSpaceport(manifests []*Manifest) *Spaceport {
	s := &Spaceport{}
	s.Init()
	s.Import(manifests)
	s.Migrate()
	return s
}

func (s *Spaceport) Init() {
	// Ensure fields are created
	if s.manifests == nil {
		s.manifests = map[string]*Manifest{}
	}
	if s.Sequence == nil {
		s.Sequence = []string{}
	}
	if s.Docked == nil {
		s.Docked = []string{}
	}
}

// Migrate legacy settings into the spaceport config
//
// Settings used to live on the core manifest and the theme at the top level
// of the spaceport. If the spaceport has no config yet, lift those values
// into a new config. Manifest config blocks are always dropped so they are
// no longer persisted.
//
// Returns true if a config was created and should be persisted.
func (s *Spaceport) Migrate() bool {
	migrated := false
	if s.Config == nil {
		s.Config = NewConfig()
		if m := s.CoreManifest(); m != nil && m.Config != nil {
			s.Config.Verbose = m.Config.Verbose
			s.Config.AliasesOnly = m.Config.AliasesOnly
			s.Config.Mode = m.Config.Mode
			s.Config.BackupCount = m.Config.BackupCount
		}
		if s.LegacyTheme != nil {
			s.Config.Theme = *s.LegacyTheme
		}
		migrated = true
	}
	s.LegacyTheme = nil

	for _, m := range s.manifests {
		if m != nil {
			m.Config = nil
		}
	}

	return migrated
}

func (s *Spaceport) Manifests() []*Manifest {
	// Use order to return manifests
	manifests := []*Manifest{}
	for _, name := range s.Sequence {
		m := s.manifests[name]
		manifests = append(manifests, m)
	}
	return manifests
}

func (s *Spaceport) Commands() []*Command {
	cmds := []*Command{}
	for _, m := range s.Manifests() {
		for _, c := range m.Commands {
			cmds = append(cmds, c)
		}
	}
	return cmds
}

func (s *Spaceport) Import(manifests []*Manifest) {
	s.Sequence = []string{}
	for _, m := range manifests {
		s.AddManifest(m)
	}
	s.reconcile()
}

func (s *Spaceport) Link() {
	for _, m := range s.manifests {
		m.Link()
	}
}

func (s *Spaceport) CoreManifest() *Manifest {
	return s.manifests[CoreManifestName]
}

func (s *Spaceport) AddManifest(m *Manifest) {
	s.manifests[m.Name] = m
	if !contains(s.Sequence, m.Name) {
		s.Sequence = append(s.Sequence, m.Name)
	}
}

func (s *Spaceport) RemoveManifest(name string) bool {
	delete(s.manifests, name)
	s.Undock(name)
	index := -1
	for i, n := range s.Sequence {
		if n == name {
			index = i
		}
	}
	if index != -1 {
		s.Sequence = append(s.Sequence[:index], s.Sequence[index+1:]...)
		return true
	}
	return false
}

func (s *Spaceport) FindManifest(name string) *Manifest {
	return s.manifests[name]
}

// IsUnique checks if a manifest name collision exists
func (s *Spaceport) IsUnique(name string) bool {
	return s.manifests[name] == nil
}

func (s *Spaceport) UpdateVersion(ver *version.Info) {
	for _, m := range s.manifests {
		m.Version.Update(ver)
	}
}

func (s *Spaceport) FindCommand(name string) (*Command, *Manifest) {
	for _, m := range s.manifests {
		if cmd := m.Find(name); cmd != nil {
			return cmd, m
		}
	}
	return nil, nil
}
