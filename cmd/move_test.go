package cmd

import (
	"bytes"
	"testing"
)

func TestMoveCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"move", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("move command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("move")) {
		t.Error("Expected help output to mention 'move'")
	}
}

func TestMoveCommand_RequiresIssueNumber(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"move"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when issue number not provided")
	}
}

func TestMoveCommand_HasStatusFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("status")
	if flag == nil {
		t.Error("Expected --status flag to exist")
	}
}

func TestMoveCommand_HasPriorityFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("priority")
	if flag == nil {
		t.Error("Expected --priority flag to exist")
	}
}

func TestMoveCommand_RequiresAtLeastOneFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"move", "123"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no field flags provided")
	}
}
