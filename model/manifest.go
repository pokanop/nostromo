package model

// Manifest is the main container for nostromo based commands
type Manifest struct {
	Version  string              `json:"version"`
	Commands map[string]*Command `json:"commands"`
}

// NewManifest returns a newly initialized manifest
func NewManifest() *Manifest {
	return &Manifest{
		Version:  "1.0",
		Commands: map[string]*Command{},
	}
}
