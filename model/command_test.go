package model

import (
	"reflect"
	"testing"
)

var depthKeys = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		alias    string
		comment  string
		code     *Code
		expected *Command
	}{
		{"empty alias", "cmd", "", "", nil, &Command{nil, "cmd", "cmd", "cmd", "", map[string]*Command{}, map[string]*Substitution{}, nil}},
		{"valid alias", "cmd", "cmd-alias", "comment", nil, &Command{nil, "cmd", "cmd", "cmd-alias", "comment", map[string]*Command{}, map[string]*Substitution{}, nil}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := NewCommand(test.cmdName, test.alias, test.comment, test.code)
			if !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestAddCommand(t *testing.T) {
	tests := []struct {
		name    string
		command *Command
		add     *Command
	}{
		{"nil command", fakeCommand(1), nil},
		{"invalid command", fakeCommand(1), fakeCommand(2).Commands["two"]},
		{"valid command", fakeCommand(1), fakeCommand(1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.AddCommand(test.add)
			if test.add != nil && test.command.Commands[test.add.Alias] == nil {
				t.Errorf("expected command to be added but was not")
			}
		})
	}
}

func TestRemoveCommand(t *testing.T) {
	tests := []struct {
		name    string
		command *Command
		remove  *Command
	}{
		{"nil command", fakeCommand(1), nil},
		{"invalid command", fakeCommand(1), fakeCommand(2).Commands["two"]},
		{"valid command", fakeCommand(1), fakeCommand(1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.RemoveCommand(test.remove)
			if test.remove != nil && test.command.Commands[test.remove.Alias] != nil {
				t.Errorf("expected command to be removed but was not")
			}
		})
	}
}

func TestAddSubstitution(t *testing.T) {
	tests := []struct {
		name    string
		command *Command
		add     *Substitution
	}{
		{"nil sub", fakeCommand(1), nil},
		{"invalid sub", fakeCommand(1), fakeCommand(2).Subs["one"]},
		{"valid sub", fakeCommand(1), &Substitution{"two", ""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.AddSubstitution(test.add)
			if test.add != nil && test.command.Subs[test.add.Alias] == nil {
				t.Errorf("expected sub to be added but was not")
			}
		})
	}
}

func TestRemoveSubstitution(t *testing.T) {
	tests := []struct {
		name    string
		command *Command
		remove  *Substitution
	}{
		{"nil sub", fakeCommand(1), nil},
		{"invalid sub", fakeCommand(1), fakeCommand(2).Subs["one"]},
		{"valid sub", fakeCommand(1), &Substitution{"two", ""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.RemoveSubstitution(test.remove)
			if test.remove != nil && test.command.Subs[test.remove.Alias] != nil {
				t.Errorf("expected sub to be removed but was not")
			}
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		keyPath  string
		expected *Command
	}{
		{"empty key path", fakeCommand(1), "", nil},
		{"wrong key path", fakeCommand(1), "wrong", nil},
		{"valid key path first level", fakeCommand(1), "one-alias", fakeCommand(1)},
		{"valid key path nth level", fakeCommand(4), "one-alias.two-alias.three-alias", fakeCommand(4).Commands["two-alias"].Commands["three-alias"]},
		{"valid key path last level", fakeCommand(7), "one-alias.two-alias.three-alias.four-alias.five-alias.six-alias.seven-alias", fakeCommand(7).Commands["two-alias"].Commands["three-alias"].Commands["four-alias"].Commands["five-alias"].Commands["six-alias"].Commands["seven-alias"]},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.command.Find(test.keyPath)
			if !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestShortestKeyPath(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		keyPath  string
		expected string
	}{
		{"empty key path", fakeCommand(1), "", ""},
		{"missing key path", fakeCommand(1), "missing", ""},
		{"missing long key path", fakeCommand(1), "this.is.missing", ""},
		{"valid key path first level", fakeCommand(1), "one-alias", "one-alias"},
		{"valid key path nth level", fakeCommand(3), "one-alias.two-alias", "one-alias.two-alias"},
		{"valid key path last level", fakeCommand(4), "one-alias.two-alias.three-alias.four-alias", "one-alias.two-alias.three-alias.four-alias"},
		{"valid key path shortened", fakeCommand(2), "one-alias.two-alias.three-alias.four-alias", "one-alias.two-alias"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.command.ShortestKeyPath(test.keyPath); test.expected != actual {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestExecutionString(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		args     []string
		expected string
	}{
		{"one level nil args", fakeCommand(1), nil, "sh -c one"},
		{"one level empty args", fakeCommand(1), []string{}, "sh -c one"},
		{"one level no dot arg", fakeCommand(1), []string{"arg"}, "sh -c one arg"},
		{"one level dot arg", fakeCommand(1), []string{"arg.1"}, "sh -c one arg.1"},
		{"n level no dot args", fakeCommand(3).Commands["two-alias"].Commands["three-alias"], []string{"arg1", "arg2"}, "sh -c one two three arg1 arg2"},
		{"n level dot args", fakeCommand(4).Commands["two-alias"], []string{"arg.1", "arg2", "arg.3"}, "sh -c one two arg.1 arg2 arg.3"},
		{"n level dot sub args", fakeCommand(4).Commands["two-alias"], []string{"arg.1", "one-sub", "two-sub"}, "sh -c one two arg.1 one two"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.command.ExecutionString(test.args); test.expected != actual {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func fakeCommand(depth int) *Command {
	var firstCmd *Command
	var lastCmd *Command
	var cmd *Command
	for i := 0; i < depth; i++ {
		name := depthKeys[i+1]
		cmd = NewCommand(name, name+"-alias", "", nil)
		cmd.AddSubstitution(&Substitution{name, name + "-sub"})
		if lastCmd != nil {
			lastCmd.AddCommand(cmd)
		} else {
			firstCmd = cmd
		}
		lastCmd = cmd
	}
	return firstCmd
}
