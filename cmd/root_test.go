package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommandHelp(t *testing.T) {
	// Test that root command executes and shows help
	cmd := NewRootCommand()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("Expected help output, got empty string")
	}

	// Verify it contains expected content
	if !bytes.Contains([]byte(output), []byte("gh pmu")) {
		t.Errorf("Expected output to contain 'gh pmu', got: %s", output)
	}
}

func TestRootCommandVersion(t *testing.T) {
	cmd := NewRootCommand()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	// Cobra uses first word of Use field for version output
	if !bytes.Contains([]byte(output), []byte("version")) {
		t.Errorf("Expected version output to contain 'version', got: %s", output)
	}
}
