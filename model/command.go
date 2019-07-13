package model

// Command is a scope for running one or more commands
type Command struct {
	Name     string
	Alias    string
	Comment  string
	Commands map[string]*Command
	Subs     map[string]*Substitution
	Code     *Code
}

// NewCommand returns a newly initialized command
func NewCommand(name, alias, comment string, code *Code) *Command {
	return &Command{
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
	c.Commands[cmd.Name] = cmd
}

// RemoveCommand at this scope
func (c *Command) RemoveCommand(cmd *Command) {
	c.Commands[cmd.Name] = nil
}

// AddSub at this scope
func (c *Command) AddSub(sub *Substitution) {
	c.Subs[sub.Name] = sub
}

// RemoveSub at this scope
func (c *Command) RemoveSub(sub *Substitution) {
	c.Subs[sub.Name] = nil
}
