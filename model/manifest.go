package model

import (
	"fmt"

	"github.com/pokanop/nostromo/keypath"
)

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

// AddCommand tree up to key path
func (m *Manifest) AddCommand(keyPath, command string) error {
	if len(keyPath) == 0 {
		return fmt.Errorf("invalid key path")
	}

	// Build commands up to key path otherwise
	key := keypath.Keys(keyPath)[0]
	cmd := m.Commands[key]
	if cmd == nil {
		// Create new command to build our the rest
		cmd = newCommand(key, "", "", nil)
		m.Commands[cmd.Alias] = cmd
	}

	// Modify or build the rest of the key path of commands
	cmd.build(keyPath, command)

	return nil
}

// RemoveCommand tree at key path
func (m *Manifest) RemoveCommand(keyPath string) error {
	cmd := m.Find(keyPath)
	if cmd == nil {
		return fmt.Errorf("command not found")
	}

	parent := cmd.parent
	if parent == nil {
		delete(m.Commands, keyPath)
		return nil
	}

	parent.removeCommand(cmd)

	return nil
}

// AddSubstitution with name and alias at key path
func (m *Manifest) AddSubstitution(keyPath, name, sub string) error {
	cmd := m.Find(keyPath)
	if cmd == nil {
		return fmt.Errorf("command not found")
	}

	s := &Substitution{name, sub}
	cmd.addSubstitution(s)

	return nil
}

// RemoveSubstitution at key path for given alias
func (m *Manifest) RemoveSubstitution(keyPath, alias string) error {
	cmd := m.Find(keyPath)
	if cmd == nil {
		return fmt.Errorf("command not found")
	}

	s := &Substitution{"", alias}
	cmd.removeSubstitution(s)

	return nil
}

// Find command at key path or nil if missing
func (m *Manifest) Find(keyPath string) *Command {
	for _, cmd := range m.Commands {
		if c := cmd.Find(keyPath); c != nil {
			return c
		}
	}
	return nil
}

// count of the total number of commands in this manifest
func (m *Manifest) count() int {
	count := 0
	for _, cmd := range m.Commands {
		count += cmd.count()
	}
	return count
}
