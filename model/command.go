package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pokanop/nostromo/stringutil"
	"github.com/spf13/cobra"

	"github.com/pokanop/nostromo/keypath"
	"github.com/shivamMg/ppds/tree"
)

// Command is a scope for running one or more commands
type Command struct {
	parent      *Command
	KeyPath     string                   `json:"keyPath"`
	Name        string                   `json:"name"`
	Alias       string                   `json:"alias"`
	AliasOnly   bool                     `json:"aliasOnly"`
	Description string                   `json:"description"`
	Commands    map[string]*Command      `json:"commands"`
	Subs        map[string]*Substitution `json:"subs"`
	Code        *Code                    `json:"code"`
	Mode        Mode                     `json:"mode"`
	Disabled    bool                     `json:"disabled"`
}

func (c *Command) String() string {
	return fmt.Sprintf("[%s] %s -> %s", c.KeyPath, c.Name, c.Alias)
}

// Keys as ordered list of fields for logging
func (c *Command) Keys() []string {
	return []string{"keypath", "alias", "command", "description", "commands", "substitutions", "code", "mode", "aliasOnly", "disabled"}
}

// Fields interface for logging
func (c *Command) Fields() map[string]interface{} {
	return map[string]interface{}{
		"keypath":       c.KeyPath,
		"alias":         c.Alias,
		"command":       c.Name,
		"description":   c.Description,
		"commands":      joinedCommands(c.Commands),
		"substitutions": joinedSubs(c.Subs),
		"code":          c.Code.valid(),
		"mode":          c.Mode.String(),
		"aliasOnly":     c.AliasOnly,
		"disabled":      c.Disabled,
	}
}

// Data method for Node interface to print tree
func (c *Command) Data() interface{} {
	return c.Alias
}

// Children method for Node interface to print tree
func (c *Command) Children() []tree.Node {
	nodes := make([]tree.Node, 0, len(c.Commands))
	for _, v := range c.Commands {
		nodes = append(nodes, v)
	}
	return nodes
}

// Walk the command tree and run supplied func
func (c *Command) Walk(fn func(*Command, *bool)) {
	c.forwardWalk(fn)
}

// CobraCommand returns a cobra.Command for this command
func (c *Command) CobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:       c.Alias,
		Short:     c.Description,
		Long:      c.Description,
		ValidArgs: c.commandList(),
	}
	for _, childCmd := range c.Commands {
		cmd.AddCommand(childCmd.CobraCommand())
	}
	return cmd
}

// newCommand returns a newly initialized command
func newCommand(name, alias, description string, code *Code, aliasOnly bool, mode string) *Command {
	// Default alias to same as command name
	if len(alias) == 0 {
		alias = name
	}

	if code == nil {
		code = &Code{}
	}

	return &Command{
		KeyPath:     alias,
		Name:        name,
		Alias:       alias,
		AliasOnly:   aliasOnly,
		Description: description,
		Commands:    map[string]*Command{},
		Subs:        map[string]*Substitution{},
		Code:        code,
		Mode:        ModeFromString(mode),
		Disabled:    false,
	}
}

func (c *Command) effectiveCommand() string {
	if c.Code.valid() {
		return c.Code.Snippet
	} else if len(c.Name) > 0 {
		switch c.Mode {
		case ConcatenateMode:
			return c.Name
		case IndependentMode, ExclusiveMode:
			return c.Name + ";"
		}
	}
	return ""
}

// addCommand at this scope
func (c *Command) addCommand(cmd *Command) {
	if cmd == nil {
		return
	}

	c.Commands[cmd.Alias] = cmd
	cmd.parent = c
	cmd.KeyPath = fmt.Sprintf("%s.%s", c.KeyPath, cmd.Alias)
}

// removeCommand at this scope
func (c *Command) removeCommand(cmd *Command) {
	if cmd == nil {
		return
	}

	delete(c.Commands, cmd.Alias)
}

// addSubstitution at this scope
func (c *Command) addSubstitution(sub *Substitution) {
	if sub == nil {
		return
	}

	c.Subs[sub.Alias] = sub
}

// removeSubstitution at this scope
func (c *Command) removeSubstitution(sub *Substitution) {
	if sub == nil {
		return
	}

	delete(c.Subs, sub.Alias)
}

// count of the total number of commands including this one
func (c *Command) count() int {
	count := 0
	c.forwardWalk(func(cmd *Command, stop *bool) {
		count++
	})
	return count
}

// find matching command for given key path
func (c *Command) find(keyPath string) *Command {
	if c.Alias == keyPath {
		return c
	}

	// The first key in path should be this command
	keys := keypath.Keys(keyPath)
	if len(keys) < 2 || c.Alias != keys[0] {
		return nil
	}

	cmd := c.Commands[keys[1]]
	if cmd == nil {
		return nil
	}

	return cmd.find(keypath.KeyPath(keys[1:]))
}

// shortestKeyPath valid key path
func (c *Command) shortestKeyPath(keyPath string) string {
	if c.Alias == keyPath {
		return keyPath
	}

	keys := keypath.Keys(keyPath)
	if c.Alias != keys[0] {
		return ""
	}

	cmd := c
	i := 0
	for i = 1; i < len(keys); i++ {
		cmd = cmd.Commands[keys[i]]
		if cmd == nil {
			break
		}
	}

	return keypath.KeyPath(keys[0:i])
}

// executionString to run the command with provided arguments
func (c *Command) executionString(args []string) string {
	var cmd string
	if c.Mode == ExclusiveMode { // Only run this command
		cmd = c.Name
	} else {
		cmd = c.expand()
	}
	var subs []string
	for _, arg := range args {
		subs = append(subs, c.substitute(arg))
	}
	return stringutil.ReplaceShellVars(cmd, subs)
}

func (c *Command) expand() string {
	var cmds []string
	c.reverseWalk(func(cmd *Command, stop *bool) {
		val := cmd.effectiveCommand()
		if len(val) > 0 {
			cmds = append(cmds, val)
		}
	})
	return strings.Join(stringutil.ReversedStrings(cmds), " ")
}

func (c *Command) substitute(arg string) string {
	sub := arg
	c.reverseWalk(func(cmd *Command, stop *bool) {
		s := cmd.Subs[arg]
		if s != nil {
			sub = s.Name
			*stop = true
		}
	})
	return sub
}

func (c *Command) reverseWalk(fn func(*Command, *bool)) {
	if fn == nil {
		return
	}

	stop := false
	cmd := c
	for {
		if cmd == nil {
			break
		}
		fn(cmd, &stop)
		if stop {
			break
		}
		cmd = cmd.parent
	}
}

func (c *Command) forwardWalk(fn func(*Command, *bool)) bool {
	if fn == nil {
		return true
	}

	stop := false
	fn(c, &stop)
	if stop {
		return true
	}

	for _, cmd := range c.Commands {
		if stop := cmd.forwardWalk(fn); stop {
			return true
		}
	}

	return false
}

func (c *Command) link(parent *Command) {
	c.parent = parent
	if c.Code == nil {
		c.Code = &Code{}
	}
	for _, cmd := range c.Commands {
		cmd.link(c)
	}
}

func (c *Command) build(keyPath, command, description string, code *Code, aliasOnly bool, mode string) {
	if len(keyPath) == 0 {
		return
	}

	cmd := c
	keys := keypath.Keys(keyPath)

	// Ensure this command is the first key
	if cmd.Alias != keys[0] {
		return
	}

	// Advance commands to next key and create in between
	var last *Command
	for i := 1; i < len(keys); i++ {
		key := keys[i]
		last = cmd
		cmd = cmd.Commands[key]
		if cmd == nil {
			cmd = newCommand("", key, "", nil, false, mode)
			last.addCommand(cmd)
		}
	}

	// Last key will use actual command
	cmd.Name = command
	cmd.Description = description
	cmd.Code = code
	cmd.AliasOnly = aliasOnly
	cmd.Mode = ModeFromString(mode)
}

func (c *Command) commandList() []string {
	var cmds []string
	for _, cmd := range c.Commands {
		cmds = append(cmds, cmd.Alias)
	}
	sort.Strings(cmds)
	return cmds
}

// checkDisabled returns true if this command or any parent node is disabled, and otherwise false
//
// Returns command if disabled, and otherwise nil
func (c *Command) checkDisabled() (bool, *Command) {
	cmd := c
	for {
		if cmd == nil {
			break
		}
		if cmd.Disabled == true {
			return true, cmd
		}
		cmd = cmd.parent
	}
	return false, nil
}

func joinedSubs(subMap map[string]*Substitution) string {
	subs := []string{}
	for sub := range subMap {
		subs = append(subs, sub)
	}
	return strings.Join(subs, ", ")
}
