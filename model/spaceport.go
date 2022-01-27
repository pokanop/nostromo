package model

import "github.com/pokanop/nostromo/version"

// Spaceport type that manages and docks multiple ships' manifests
type Spaceport struct {
	Manifests []*Manifest
}

func (s *Spaceport) UpdateVersion(ver *version.Info) {
	for _, m := range s.Manifests {
		m.Version.Update(ver)
	}
}

func (s *Spaceport) CoreManifest() *Manifest {
	for _, m := range s.Manifests {
		if m.Name == CoreManifestName {
			return m
		}
	}
	return nil
}

func (s *Spaceport) Link() {
	for _, m := range s.Manifests {
		m.Link()
	}
}

func (s *Spaceport) AddManifest(m *Manifest) {
	s.Manifests = append(s.Manifests, m)
}
