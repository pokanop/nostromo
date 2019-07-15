package model

import (
	"reflect"
	"testing"
)

var depthKeys = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}

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
			if test.add != nil && test.command.Commands[test.add.Name] == nil {
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
			if test.remove != nil && test.command.Commands[test.remove.Name] != nil {
				t.Errorf("expected command to be removed but was not")
			}
		})
	}
}

func TestAddSub(t *testing.T) {
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
			test.command.AddSub(test.add)
			if test.add != nil && test.command.Subs[test.add.Name] == nil {
				t.Errorf("expected sub to be added but was not")
			}
		})
	}
}

func TestRemoveSub(t *testing.T) {
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
			test.command.RemoveSub(test.remove)
			if test.remove != nil && test.command.Subs[test.remove.Name] != nil {
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
		{"valid key path first level", fakeCommand(1), "one", fakeCommand(1)},
		{"valid key path nth level", fakeCommand(4), "one.two.three", fakeCommand(4).Commands["two"].Commands["three"]},
		{"valid key path last level", fakeCommand(7), "one.two.three.four.five.six.seven", fakeCommand(7).Commands["two"].Commands["three"].Commands["four"].Commands["five"].Commands["six"].Commands["seven"]},
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
		{"valid key path first level", fakeCommand(1), "one", "one"},
		{"valid key path nth level", fakeCommand(3), "one.two", "one.two"},
		{"valid key path last level", fakeCommand(4), "one.two.three.four", "one.two.three.four"},
		{"valid key path shortened", fakeCommand(2), "one.two.three.four", "one.two"},
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
		{"n level no dot args", fakeCommand(3).Commands["two"].Commands["three"], []string{"arg1", "arg2"}, "sh -c one two three arg1 arg2"},
		{"n level dot args", fakeCommand(4).Commands["two"], []string{"arg.1", "arg2", "arg.3"}, "sh -c one two arg.1 arg2 arg.3"},
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
		cmd.AddSub(&Substitution{name, name + "-sub"})
		if lastCmd != nil {
			lastCmd.AddCommand(cmd)
		} else {
			firstCmd = cmd
		}
		lastCmd = cmd
	}
	return firstCmd
}
