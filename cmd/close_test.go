package cmd

import (
	"testing"
)

func TestNormalizeCloseReason(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		// not_planned variations
		{
			name:     "underscore not_planned",
			input:    "not_planned",
			expected: "not planned",
		},
		{
			name:     "space not planned",
			input:    "not planned",
			expected: "not planned",
		},
		{
			name:     "uppercase NOT_PLANNED",
			input:    "NOT_PLANNED",
			expected: "not planned",
		},
		{
			name:     "mixed case Not_Planned",
			input:    "Not_Planned",
			expected: "not planned",
		},
		{
			name:     "notplanned no separator",
			input:    "notplanned",
			expected: "not planned",
		},

		// completed variations
		{
			name:     "completed",
			input:    "completed",
			expected: "completed",
		},
		{
			name:     "COMPLETED uppercase",
			input:    "COMPLETED",
			expected: "completed",
		},
		{
			name:     "complete shorthand",
			input:    "complete",
			expected: "completed",
		},
		{
			name:     "done alias",
			input:    "done",
			expected: "completed",
		},

		// empty
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "  ",
			expected: "",
		},

		// invalid
		{
			name:      "invalid reason",
			input:     "invalid",
			expectErr: true,
		},
		{
			name:      "wontfix invalid",
			input:     "wontfix",
			expectErr: true,
		},
		{
			name:      "cancelled invalid",
			input:     "cancelled",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeCloseReason(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("normalizeCloseReason(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewCloseCommand(t *testing.T) {
	cmd := newCloseCommand()

	// Verify command basics
	if cmd.Use != "close <issue-number>" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Verify flags exist
	flags := []struct {
		name      string
		shorthand string
	}{
		{"reason", "r"},
		{"comment", "c"},
		{"update-status", ""},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag --%s not found", f.name)
			continue
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag --%s shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}

func TestNewCloseCommand_RequiresArg(t *testing.T) {
	cmd := newCloseCommand()

	// Command requires exactly 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error when no arguments provided")
	}

	err = cmd.Args(cmd, []string{"123"})
	if err != nil {
		t.Errorf("unexpected error with one argument: %v", err)
	}

	err = cmd.Args(cmd, []string{"123", "456"})
	if err == nil {
		t.Error("expected error when too many arguments provided")
	}
}

func TestRunClose_InvalidIssueNumber(t *testing.T) {
	cmd := newCloseCommand()
	opts := &closeOptions{}

	err := runClose(cmd, []string{"not-a-number"}, opts)
	if err == nil {
		t.Error("expected error for non-numeric issue number")
	}
}

func TestRunClose_InvalidReason(t *testing.T) {
	cmd := newCloseCommand()
	opts := &closeOptions{
		reason: "invalid_reason",
	}

	err := runClose(cmd, []string{"123"}, opts)
	if err == nil {
		t.Error("expected error for invalid close reason")
	}
}
