package cmd

import "testing"

func TestDocsCommand(t *testing.T) {
	c, _, err := rootCmd.Find([]string{"docs"})
	if err != nil {
		t.Fatalf("docs command not registered: %v", err)
	}
	if !c.Hidden {
		t.Error("docs command should be hidden")
	}
	if err := c.ValidateArgs([]string{}); err != nil {
		t.Errorf("docs with no args: %v", err)
	}
	if err := c.ValidateArgs([]string{"out"}); err != nil {
		t.Errorf("docs with one arg: %v", err)
	}
	if err := c.ValidateArgs([]string{"a", "b"}); err == nil {
		t.Error("docs should reject more than one arg")
	}
}
