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

func (m *Manifest) AddCommand(keyPath, command string) error {
	return nil
}

func (m *Manifest) RemoveCommand(keyPath string) error {
	return nil
}

func (m *Manifest) AddSubstitution(keyPath, name, sub string) error {
	return nil
}

func (m *Manifest) RemoveSubstitution(keyPath, name string) error {
	return nil
}

func (m *Manifest) Find(args []string) *Command {
	return nil
}
