package cmd

import (
	"bytes"
	"testing"
)

func TestViewCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"view", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("view command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("view")) {
		t.Error("Expected help output to mention 'view'")
	}
}

func TestViewCommand_RequiresIssueNumber(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"view"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when issue number not provided")
	}
}

func TestViewCommand_HasJSONFlag(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	flag := viewCmd.Flags().Lookup("json")
	if flag == nil {
		t.Error("Expected --json flag to exist")
	}
}

func TestViewCommand_AcceptsIssueNumber(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	// Verify the command accepts exactly 1 argument
	if viewCmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestViewCommand_ParsesIssueNumber(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{"valid number", "123", false},
		{"with hash", "#123", false},
		{"invalid string", "abc", true},
		{"negative number", "-1", true},
		{"zero", "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseIssueNumber(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIssueNumber(%q) error = %v, wantErr %v", tt.arg, err, tt.wantErr)
			}
		})
	}
}

func TestViewCommand_ParsesIssueReference(t *testing.T) {
	tests := []struct {
		name       string
		arg        string
		wantOwner  string
		wantRepo   string
		wantNumber int
		wantErr    bool
	}{
		{"number only", "123", "", "", 123, false},
		{"with hash", "#123", "", "", 123, false},
		{"full reference", "owner/repo#123", "owner", "repo", 123, false},
		{"invalid", "invalid", "", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, number, err := parseIssueReference(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIssueReference(%q) error = %v, wantErr %v", tt.arg, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("parseIssueReference(%q) owner = %v, want %v", tt.arg, owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("parseIssueReference(%q) repo = %v, want %v", tt.arg, repo, tt.wantRepo)
				}
				if number != tt.wantNumber {
					t.Errorf("parseIssueReference(%q) number = %v, want %v", tt.arg, number, tt.wantNumber)
				}
			}
		})
	}
}
