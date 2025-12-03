package cmd

import (
	"bytes"
	"testing"
)

func TestSubCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("sub command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("sub")) {
		t.Error("Expected help output to mention 'sub'")
	}
}

func TestSubAddCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "add", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("sub add command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("add")) {
		t.Error("Expected help output to mention 'add'")
	}
}

func TestSubAddCommand_RequiresTwoArgs(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "add", "123"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when only one argument provided")
	}
}

func TestSubAddCommand_RequiresParentAndChild(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "add"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no arguments provided")
	}
}

func TestSubCreateCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "create", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("sub create command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("create")) {
		t.Error("Expected help output to mention 'create'")
	}
}

func TestSubCreateCommand_HasParentFlag(t *testing.T) {
	cmd := NewRootCommand()
	subCmd, _, err := cmd.Find([]string{"sub", "create"})
	if err != nil {
		t.Fatalf("sub create command not found: %v", err)
	}

	flag := subCmd.Flags().Lookup("parent")
	if flag == nil {
		t.Error("Expected --parent flag to exist")
	}
}

func TestSubCreateCommand_HasTitleFlag(t *testing.T) {
	cmd := NewRootCommand()
	subCmd, _, err := cmd.Find([]string{"sub", "create"})
	if err != nil {
		t.Fatalf("sub create command not found: %v", err)
	}

	flag := subCmd.Flags().Lookup("title")
	if flag == nil {
		t.Error("Expected --title flag to exist")
	}
}

func TestSubCreateCommand_RequiresParentFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "create", "--title", "Test"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when --parent not provided")
	}
}

func TestSubCreateCommand_RequiresTitleFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "create", "--parent", "123"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when --title not provided")
	}
}

func TestSubListCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "list", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("sub list command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("list")) {
		t.Error("Expected help output to mention 'list'")
	}
}

func TestSubListCommand_RequiresParentArg(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "list"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when parent issue not provided")
	}
}

func TestSubListCommand_HasJSONFlag(t *testing.T) {
	cmd := NewRootCommand()
	subCmd, _, err := cmd.Find([]string{"sub", "list"})
	if err != nil {
		t.Fatalf("sub list command not found: %v", err)
	}

	flag := subCmd.Flags().Lookup("json")
	if flag == nil {
		t.Error("Expected --json flag to exist")
	}
}

func TestSubRemoveCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "remove", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("sub remove command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("remove")) {
		t.Error("Expected help output to mention 'remove'")
	}
}

func TestSubRemoveCommand_RequiresTwoArgs(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "remove", "123"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when only one argument provided")
	}
}

func TestSubRemoveCommand_RequiresParentAndChild(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"sub", "remove"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no arguments provided")
	}
}
