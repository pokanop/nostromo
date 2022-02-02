package model

import "github.com/pokanop/nostromo/version"

// Spaceport type that manages and docks multiple ships' manifests
type Spaceport struct {
	manifests map[string]*Manifest
	Sequence  []string `json:"order"`
}

func NewSpaceport(manifests []*Manifest) *Spaceport {
	s := &Spaceport{map[string]*Manifest{}, []string{}}
	s.Import(manifests)
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
		s.Sequence = append(s.Sequence, m.Name)
	}
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

func (s *Spaceport) FindCommand(name string) *Command {
	for _, m := range s.manifests {
		if cmd := m.Find(name); cmd != nil {
			return cmd
		}
	}
	return nil
}
