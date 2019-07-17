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

// NewCommand returns a newly initialized command
func NewCommand(name, alias, comment string, code *Code) *Command {
	// Default alias to same as command name
	if len(alias) == 0 {
		alias = name
	}

	return &Command{
		KeyPath:  name,
		Name:     name,
		Alias:    alias,
		Comment:  comment,
		Commands: map[string]*Command{},
		Subs:     map[string]*Substitution{},
		Code:     code,
	}
}

// AddCommand at this scope
func (c *Command) AddCommand(cmd *Command) {
	if cmd == nil {
		return
	}

	c.Commands[cmd.Alias] = cmd
	cmd.parent = c
	cmd.KeyPath = fmt.Sprintf("%s.%s", c.KeyPath, cmd.Alias)
}

// RemoveCommand at this scope
func (c *Command) RemoveCommand(cmd *Command) {
	if cmd == nil {
		return
	}

	c.Commands[cmd.Alias] = nil
}

// AddSubstitution at this scope
func (c *Command) AddSubstitution(sub *Substitution) {
	if sub == nil {
		return
	}

	c.Subs[sub.Alias] = sub
}

// RemoveSubstitution at this scope
func (c *Command) RemoveSubstitution(sub *Substitution) {
	if sub == nil {
		return
	}

	c.Subs[sub.Alias] = nil
}

// Find matching command for given key path
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

// ShortestKeyPath valid key path
func (c *Command) ShortestKeyPath(keyPath string) string {
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
	return strings.TrimSpace(fmt.Sprintf("sh -c %s %s", cmd, strings.Join(subs, " ")))
}

func (c *Command) expand() string {
	cmds := []string{}
	c.walk(func(cmd *Command, stop *bool) {
		cmds = append(cmds, cmd.Name)
	})
	return strings.Join(reversed(cmds), " ")
}

func (c *Command) substitute(arg string) string {
	sub := arg
	c.walk(func(cmd *Command, stop *bool) {
		s := cmd.Subs[arg]
		if s != nil {
			sub = s.Name
			*stop = true
		}
	})
	return sub
}

func (c *Command) walk(fn func(*Command, *bool)) {
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
