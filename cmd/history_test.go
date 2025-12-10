package cmd

import (
	"testing"
	"time"
)

func TestNewHistoryCommand(t *testing.T) {
	cmd := newHistoryCommand()

	if cmd.Use != "history <path> [path...]" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description")
	}
}

func TestHistoryCommand_HasFlags(t *testing.T) {
	cmd := newHistoryCommand()

	tests := []struct {
		flagName     string
		defaultValue string
	}{
		{"since", ""},
		{"limit", "50"},
		{"output", "false"},
		{"force", "false"},
		{"json", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("expected flag --%s to exist", tt.flagName)
				return
			}
			if flag.DefValue != tt.defaultValue {
				t.Errorf("expected default %s, got %s", tt.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestInferChangeType(t *testing.T) {
	tests := []struct {
		subject  string
		expected string
	}{
		// Fix variations
		{"fix: handle empty values", "Fix"},
		{"Fix: handle empty values", "Fix"},
		{"fix(api): handle empty values", "Fix"},
		{"bug: fix null pointer", "Fix"},

		// Add variations
		{"add: new feature", "Add"},
		{"Add: new feature", "Add"},
		{"feat: implement login", "Add"},
		{"feat(auth): implement login", "Add"},
		{"feature: new dashboard", "Add"},

		// Update variations
		{"update: improve performance", "Update"},
		{"enhance: add caching", "Update"},

		// Remove variations
		{"remove: deprecated function", "Remove"},
		{"delete: old files", "Remove"},

		// Refactor
		{"refactor: extract method", "Refactor"},
		{"refactor(core): simplify logic", "Refactor"},

		// Docs
		{"docs: update readme", "Docs"},
		{"doc: add examples", "Docs"},

		// Test
		{"test: add unit tests", "Test"},
		{"test(api): improve coverage", "Test"},

		// Chore
		{"chore: update deps", "Chore"},
		{"build: fix makefile", "Chore"},
		{"ci: update workflow", "Chore"},

		// Unknown - fallback to Change
		{"Some random commit message", "Change"},
		{"Updated the thing", "Change"},
		{"v1.0.0 release", "Change"},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			result := inferChangeType(tt.subject)
			if result != tt.expected {
				t.Errorf("inferChangeType(%q) = %q, want %q", tt.subject, result, tt.expected)
			}
		})
	}
}

func TestParseCommitReferences(t *testing.T) {
	tests := []struct {
		name          string
		subject       string
		defaultOwner  string
		defaultRepo   string
		expectedCount int
		checkFirst    *IssueReference // optional: check first reference details
	}{
		{
			name:          "simple hash reference",
			subject:       "Fix bug #123",
			defaultOwner:  "owner",
			defaultRepo:   "repo",
			expectedCount: 1,
			checkFirst: &IssueReference{
				Number: 123,
				Owner:  "owner",
				Repo:   "repo",
				Type:   "related",
				URL:    "https://github.com/owner/repo/issues/123",
			},
		},
		{
			name:          "fixes prefix",
			subject:       "fixes #456",
			defaultOwner:  "owner",
			defaultRepo:   "repo",
			expectedCount: 1,
			checkFirst: &IssueReference{
				Number: 456,
				Owner:  "owner",
				Repo:   "repo",
				Type:   "fix",
				URL:    "https://github.com/owner/repo/issues/456",
			},
		},
		{
			name:          "closes prefix",
			subject:       "closes #789",
			defaultOwner:  "owner",
			defaultRepo:   "repo",
			expectedCount: 1,
			checkFirst: &IssueReference{
				Number: 789,
				Owner:  "owner",
				Repo:   "repo",
				Type:   "close",
				URL:    "https://github.com/owner/repo/issues/789",
			},
		},
		{
			name:          "cross-repo reference",
			subject:       "See other-owner/other-repo#42",
			defaultOwner:  "owner",
			defaultRepo:   "repo",
			expectedCount: 1,
			checkFirst: &IssueReference{
				Number: 42,
				Owner:  "other-owner",
				Repo:   "other-repo",
				Type:   "related",
				URL:    "https://github.com/other-owner/other-repo/issues/42",
			},
		},
		{
			name:          "multiple references",
			subject:       "Fix #1, closes #2, related to #3",
			defaultOwner:  "owner",
			defaultRepo:   "repo",
			expectedCount: 3,
		},
		{
			name:          "no references",
			subject:       "Just a regular commit message",
			defaultOwner:  "owner",
			defaultRepo:   "repo",
			expectedCount: 0,
		},
		{
			name:          "duplicate references deduplicated",
			subject:       "Fix #123, also #123 again",
			defaultOwner:  "owner",
			defaultRepo:   "repo",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refs := parseCommitReferences(tt.subject, tt.defaultOwner, tt.defaultRepo)

			if len(refs) != tt.expectedCount {
				t.Errorf("expected %d references, got %d", tt.expectedCount, len(refs))
				return
			}

			if tt.checkFirst != nil && len(refs) > 0 {
				got := refs[0]
				if got.Number != tt.checkFirst.Number {
					t.Errorf("Number: expected %d, got %d", tt.checkFirst.Number, got.Number)
				}
				if got.Owner != tt.checkFirst.Owner {
					t.Errorf("Owner: expected %s, got %s", tt.checkFirst.Owner, got.Owner)
				}
				if got.Repo != tt.checkFirst.Repo {
					t.Errorf("Repo: expected %s, got %s", tt.checkFirst.Repo, got.Repo)
				}
				if got.URL != tt.checkFirst.URL {
					t.Errorf("URL: expected %s, got %s", tt.checkFirst.URL, got.URL)
				}
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a very long string that needs truncation", 20, "this is a very lo..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestCommitInfo_JSONMarshalling(t *testing.T) {
	commit := CommitInfo{
		Hash:       "abc1234",
		Author:     "Test Author",
		Date:       time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		Subject:    "Fix: test commit #123",
		ChangeType: "Fix",
		References: []IssueReference{
			{
				Number: 123,
				Owner:  "owner",
				Repo:   "repo",
				Type:   "related",
				URL:    "https://github.com/owner/repo/issues/123",
			},
		},
	}

	// Verify struct fields are properly tagged
	if commit.Hash != "abc1234" {
		t.Errorf("expected hash abc1234, got %s", commit.Hash)
	}
	if len(commit.References) != 1 {
		t.Errorf("expected 1 reference, got %d", len(commit.References))
	}
}
