package cmd

import (
	"bytes"
	"testing"
)

func TestCreateCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("create command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("create")) {
		t.Error("Expected help output to mention 'create'")
	}
}

func TestCreateCommand_HasTitleFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("title")
	if flag == nil {
		t.Error("Expected --title flag to exist")
	}
}

func TestCreateCommand_HasBodyFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("body")
	if flag == nil {
		t.Error("Expected --body flag to exist")
	}
}

func TestCreateCommand_HasStatusFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("status")
	if flag == nil {
		t.Error("Expected --status flag to exist")
	}
}

func TestCreateCommand_HasPriorityFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("priority")
	if flag == nil {
		t.Error("Expected --priority flag to exist")
	}
}

func TestCreateCommand_HasLabelFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("label")
	if flag == nil {
		t.Error("Expected --label flag to exist")
	}
}

func TestCreateCommand_RequiresTitleInNonInteractiveMode(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--body", "test body"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when title not provided with --body")
	}
}
