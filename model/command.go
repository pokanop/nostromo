package model

import (
	"fmt"
	"strings"

	"github.com/pokanop/nostromo/keypath"
)

// Command is a scope for running one or more commands
type Command struct {
	parent   *Command
	KeyPath  string                   `json:"keyPath"`
	Name     string                   `json:"name"`
	Alias    string                   `json:"alias"`
	Comment  string                   `json:"comment"`
	Commands map[string]*Command      `json:"commands"`
	Subs     map[string]*Substitution `json:"subs"`
	Code     *Code                    `json:"code"`
}

// newCommand returns a newly initialized command
func newCommand(name, alias, comment string, code *Code) *Command {
	// Default alias to same as command name
	if len(alias) == 0 {
		alias = name
	}

	return &Command{
		KeyPath:  alias,
		Name:     name,
		Alias:    alias,
		Comment:  comment,
		Commands: map[string]*Command{},
		Subs:     map[string]*Substitution{},
		Code:     code,
	}
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

// Find matching command for given key path
// First key in path should be this command
func (c *Command) Find(keyPath string) *Command {
	if c.Alias == keyPath {
		return c
	}

	cmd := c.Commands[keypath.Get(keyPath, 1)]
	if cmd != nil {
		return cmd.Find(keypath.DropFirst(keyPath, 1))
	}

	return nil
}

// shortestKeyPath valid key path
func (c *Command) shortestKeyPath(keyPath string) string {
	keys := keypath.Keys(keyPath)
	for i := 0; i < len(keys); i++ {
		kp := keypath.DropLast(keyPath, i)
		cmd := c.Find(kp)
		if cmd != nil {
			return kp
		}
	}
	return ""
}

// ExecutionString to run the command with provided arguments
func (c *Command) ExecutionString(args []string) string {
	cmd := c.expand()
	subs := []string{}
	for _, arg := range args {
		subs = append(subs, c.substitute(arg))
	}

	return strings.TrimSpace(fmt.Sprintf("%s %s", cmd, strings.Join(subs, " ")))
}

func (c *Command) expand() string {
	cmds := []string{}
	c.reverseWalk(func(cmd *Command, stop *bool) {
		if len(cmd.Name) > 0 {
			cmds = append(cmds, cmd.Name)
		}
	})
	return strings.Join(reversed(cmds), " ")
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
	for _, cmd := range c.Commands {
		cmd.link(c)
	}
}

func (c *Command) build(keyPath, command string) {
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
			cmd = newCommand("", key, "", nil)
			last.addCommand(cmd)
		}
	}

	// Last key will use actual command
	cmd.Name = command
}

func (c *Command) String() string {
	return fmt.Sprintf("[%s] %s -> %s", c.KeyPath, c.Name, c.Alias)
}

func reversed(strs []string) []string {
	if strs == nil {
		return nil
	}

	r := []string{}
	for i := len(strs) - 1; i >= 0; i-- {
		r = append(r, strs[i])
	}
	return r
}
