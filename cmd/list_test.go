package cmd

import (
	"bytes"
	"testing"
)

func TestListCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"list", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("list")) {
		t.Error("Expected help output to mention 'list'")
	}
}

func TestListCommand_HasStatusFlag(t *testing.T) {
	cmd := NewRootCommand()
	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("list command not found: %v", err)
	}

	flag := listCmd.Flags().Lookup("status")
	if flag == nil {
		t.Error("Expected --status flag to exist")
	}
}

func TestListCommand_HasPriorityFlag(t *testing.T) {
	cmd := NewRootCommand()
	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("list command not found: %v", err)
	}

	flag := listCmd.Flags().Lookup("priority")
	if flag == nil {
		t.Error("Expected --priority flag to exist")
	}
}

func TestListCommand_HasJSONFlag(t *testing.T) {
	cmd := NewRootCommand()
	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("list command not found: %v", err)
	}

	flag := listCmd.Flags().Lookup("json")
	if flag == nil {
		t.Error("Expected --json flag to exist")
	}
}

func TestListCommand_HasSubIssuesFlag(t *testing.T) {
	cmd := NewRootCommand()
	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("list command not found: %v", err)
	}

	flag := listCmd.Flags().Lookup("has-sub-issues")
	if flag == nil {
		t.Error("Expected --has-sub-issues flag to exist")
	}

	// Verify it's a boolean flag
	if flag.Value.Type() != "bool" {
		t.Errorf("Expected --has-sub-issues to be bool, got %s", flag.Value.Type())
	}
}

func TestListCommand_HasSubIssuesHelpText(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"list", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list help failed: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("has-sub-issues")) {
		t.Error("Expected help to mention --has-sub-issues flag")
	}
}
