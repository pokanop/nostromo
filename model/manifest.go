package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pokanop/nostromo/keypath"
	"github.com/pokanop/nostromo/log"
)

// Manifest is the main container for nostromo based commands
type Manifest struct {
	Version  string              `json:"version"`
	Config   *Config             `json:"config"`
	Commands map[string]*Command `json:"commands"`
}

// NewManifest returns a newly initialized manifest
func NewManifest() *Manifest {
	return &Manifest{
		Version:  "1.0",
		Config:   &Config{},
		Commands: map[string]*Command{},
	}
}

// Link a newly loaded manifest
//
// This must be run after parsing a manifest to walk the command
// tree and build links.
func (m *Manifest) Link() {
	for _, cmd := range m.Commands {
		cmd.link(nil)
	}
}

// AddCommand tree up to key path
func (m *Manifest) AddCommand(keyPath, command, description string) error {
	if len(keyPath) == 0 {
		return fmt.Errorf("invalid key path")
	}

	// Build commands up to key path otherwise
	key := keypath.Keys(keyPath)[0]
	cmd := m.Commands[key]
	if cmd == nil {
		// Create new command to build our the rest
		cmd = newCommand("", key, "", nil)
		m.Commands[cmd.Alias] = cmd
	}

	// Modify or build the rest of the key path of commands
	cmd.build(keyPath, command, description)

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
func (m *Manifest) AddSubstitution(keyPath, name, alias string) error {
	cmd := m.Find(keyPath)
	if cmd == nil {
		return fmt.Errorf("command not found")
	}

	s := &Substitution{name, alias}
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
		if c := cmd.find(keyPath); c != nil {
			return c
		}
	}
	return nil
}

// AsJSON returns string representation used for storage
func (m *Manifest) AsJSON() string {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// ExecutionString from input if possible or return error
func (m *Manifest) ExecutionString(args []string) (string, error) {
	for _, cmd := range m.Commands {
		keyPath := cmd.shortestKeyPath(keypath.KeyPath(args))
		if len(keyPath) > 0 {
			count := len(keypath.Keys(keyPath))
			if m.Config.Verbose {
				log.Debug("key path:", keyPath)
				if len(args[count:]) > 0 {
					log.Debug("arguments:", args[count:])
				}
			}
			return cmd.find(keyPath).executionString(args[count:]), nil
		}
	}

	if m.Config.Verbose {
		log.Debug("arguments:", args)
	}

	return "", fmt.Errorf("unable to find execution string")
}

// Keys as ordered list of fields for logging
func (m *Manifest) Keys() []string {
	return []string{"version", "commands"}
}

// Fields interface for logging
func (m *Manifest) Fields() map[string]interface{} {

	return map[string]interface{}{
		"version":  m.Version,
		"commands": joinedCommands(m.Commands),
	}
}

// count of the total number of commands in this manifest
func (m *Manifest) count() int {
	count := 0
	for _, cmd := range m.Commands {
		count += cmd.count()
	}
	return count
}

func joinedCommands(cmdMap map[string]*Command) string {
	commands := []string{}
	for cmd := range cmdMap {
		commands = append(commands, cmd)
	}
	return strings.Join(commands, ", ")
}
