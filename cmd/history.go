package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// historyOptions holds command flags
type historyOptions struct {
	since  string // Date filter (e.g., "2024-01-01", "1 week ago")
	limit  int    // Max commits (default 50)
	output bool   // Write to History/ directory
	force  bool   // Override safety limits
	json   bool   // JSON output format
}

// CommitInfo represents parsed information from a git commit
type CommitInfo struct {
	Hash       string           `json:"hash"`
	Author     string           `json:"author"`
	Date       time.Time        `json:"date"`
	Subject    string           `json:"subject"`
	ChangeType string           `json:"change_type"`
	References []IssueReference `json:"references,omitempty"`
}

// IssueReference represents a parsed issue/PR reference
type IssueReference struct {
	Number int    `json:"number"`
	Owner  string `json:"owner,omitempty"`
	Repo   string `json:"repo,omitempty"`
	Type   string `json:"type"` // fixes, closes, related
	URL    string `json:"url"`
}

// Lipgloss styles
var (
	historyHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(0, 1)

	hashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	authorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	changeTypeStyles = map[string]lipgloss.Style{
		"Fix":      lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		"Add":      lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		"Update":   lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true),
		"Remove":   lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		"Refactor": lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		"Docs":     lipgloss.NewStyle().Foreground(lipgloss.Color("141")),
		"Test":     lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
		"Chore":    lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		"Change":   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	}

	issueRefStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))

	summaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
)

func newHistoryCommand() *cobra.Command {
	opts := &historyOptions{
		limit: 50,
	}

	cmd := &cobra.Command{
		Use:   "history <path> [path...]",
		Short: "Show git commit history with issue references",
		Long: `Show git commit history for file(s) or directories with issue/PR references.

Parses commit messages for issue references (#123, fixes #456) and creates
links to GitHub issues. Infers change type from commit prefixes (Fix:, Add:, etc.).

Safety protections:
  - Refuses to run from repository root without explicit path
  - Limited to 25 files by default (use --force to override)

Examples:
  gh pmu history cmd/move.go                    # History for single file
  gh pmu history cmd/                           # History for directory
  gh pmu history internal/api/ --since "1 week" # Recent changes only
  gh pmu history cmd/ --output                  # Write to History/ directory
  gh pmu history . --force                      # Override safety limits
  gh pmu history cmd/ --json                    # JSON output
  gh pmu history cmd/ --limit 100               # More commits`,
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistory(cmd, args, opts)
		},
	}

	cmd.Flags().StringVar(&opts.since, "since", "", "Show commits since date (e.g., '2024-01-01', '1 week ago')")
	cmd.Flags().IntVar(&opts.limit, "limit", 50, "Maximum number of commits to show")
	cmd.Flags().BoolVar(&opts.output, "output", false, "Write output to History/ directory as markdown")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Override safety limits (root directory, file count)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output in JSON format")

	return cmd
}

func runHistory(cmd *cobra.Command, args []string, opts *historyOptions) error {
	// Load config for repo info
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, err := config.LoadFromDirectory(cwd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w\nRun 'gh pmu init' to create a configuration file", err)
	}

	// Default to current directory if no args
	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	// Safety validation
	if err := validateHistorySafety(paths, opts); err != nil {
		return err
	}

	// Get commit history
	commits, err := getCommitHistory(paths, opts.since, opts.limit)
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	if len(commits) == 0 {
		fmt.Println("No commits found for the specified path(s).")
		return nil
	}

	// Get repo info for issue URLs
	repoOwner, repoName := parseRepoFromConfig(cfg)

	// Parse references and infer change types
	for i := range commits {
		commits[i].ChangeType = inferChangeType(commits[i].Subject)
		commits[i].References = parseCommitReferences(commits[i].Subject, repoOwner, repoName)
	}

	// Generate target path string for display
	targetPath := strings.Join(paths, ", ")

	// Output based on format
	if opts.json {
		return outputHistoryJSON(commits)
	}

	if opts.output {
		return outputMarkdown(commits, targetPath, repoOwner, repoName)
	}

	// Default: styled screen output
	renderHistoryScreen(commits, targetPath)
	return nil
}

// validateHistorySafety checks safety constraints
func validateHistorySafety(paths []string, opts *historyOptions) error {
	if opts.force {
		return nil
	}

	// Check for repository root
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil // Can't determine, let git handle it
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	for _, path := range paths {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(cwd, path)
		}
		absPath = filepath.Clean(absPath)

		if absPath == repoRoot || path == "." {
			// Check if we're at repo root
			if cwd == repoRoot {
				return fmt.Errorf("refusing to run at repository root\n" +
					"Specify a subdirectory or file, or use --force to override")
			}
		}
	}

	// Check file count
	totalFiles := 0
	for _, path := range paths {
		count, err := countFilesInPath(path)
		if err != nil {
			continue
		}
		totalFiles += count
	}

	if totalFiles > 25 {
		return fmt.Errorf("path contains %d files (limit is 25)\n"+
			"Use --force to override this limit", totalFiles)
	}

	return nil
}

// getRepoRoot returns the git repository root
func getRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// countFilesInPath counts tracked files in a path
func countFilesInPath(path string) (int, error) {
	cmd := exec.Command("git", "ls-files", path)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}
	return len(lines), nil
}

// getCommitHistory executes git log and parses the output
func getCommitHistory(paths []string, since string, limit int) ([]CommitInfo, error) {
	args := []string{
		"log",
		"--format=%h|%an|%aI|%s",
		fmt.Sprintf("--max-count=%d", limit),
	}

	if since != "" {
		args = append(args, fmt.Sprintf("--since=%s", since))
	}

	args = append(args, "--")
	args = append(args, paths...)

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}

	var commits []CommitInfo
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}

		date, _ := time.Parse(time.RFC3339, parts[2])
		commits = append(commits, CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    date,
			Subject: parts[3],
		})
	}

	return commits, nil
}

// inferChangeType determines the change type from commit subject prefix
func inferChangeType(subject string) string {
	prefixMap := map[string]string{
		"fix":      "Fix",
		"bug":      "Fix",
		"add":      "Add",
		"feat":     "Add",
		"feature":  "Add",
		"update":   "Update",
		"enhance":  "Update",
		"remove":   "Remove",
		"delete":   "Remove",
		"refactor": "Refactor",
		"docs":     "Docs",
		"doc":      "Docs",
		"test":     "Test",
		"chore":    "Chore",
		"build":    "Chore",
		"ci":       "Chore",
	}

	lowerSubject := strings.ToLower(subject)
	for prefix, changeType := range prefixMap {
		if strings.HasPrefix(lowerSubject, prefix+":") ||
			strings.HasPrefix(lowerSubject, prefix+"(") ||
			strings.HasPrefix(lowerSubject, prefix+" ") {
			return changeType
		}
	}
	return "Change"
}

// parseCommitReferences extracts all issue/PR references from commit message
func parseCommitReferences(subject, defaultOwner, defaultRepo string) []IssueReference {
	var refs []IssueReference
	seen := make(map[int]bool)

	// Pattern: fixes #123, closes #456, resolves #789
	actionPattern := regexp.MustCompile(`(?i)(?:fix(?:es)?|close[sd]?|resolve[sd]?)\s+#(\d+)`)
	actionMatches := actionPattern.FindAllStringSubmatch(subject, -1)
	for _, match := range actionMatches {
		num, _ := strconv.Atoi(match[1])
		if !seen[num] {
			seen[num] = true
			refs = append(refs, IssueReference{
				Number: num,
				Owner:  defaultOwner,
				Repo:   defaultRepo,
				Type:   strings.ToLower(strings.TrimSuffix(strings.TrimSuffix(match[0][:strings.Index(match[0], "#")-1], "s"), "d")),
				URL:    fmt.Sprintf("https://github.com/%s/%s/issues/%d", defaultOwner, defaultRepo, num),
			})
		}
	}

	// Pattern: owner/repo#123
	crossRepoPattern := regexp.MustCompile(`(\w[\w-]*)/(\w[\w-]*)#(\d+)`)
	crossMatches := crossRepoPattern.FindAllStringSubmatch(subject, -1)
	for _, match := range crossMatches {
		num, _ := strconv.Atoi(match[3])
		if !seen[num] {
			seen[num] = true
			refs = append(refs, IssueReference{
				Number: num,
				Owner:  match[1],
				Repo:   match[2],
				Type:   "related",
				URL:    fmt.Sprintf("https://github.com/%s/%s/issues/%d", match[1], match[2], num),
			})
		}
	}

	// Pattern: simple #123 (not already captured)
	simplePattern := regexp.MustCompile(`(?:^|[^/\w])#(\d+)`)
	simpleMatches := simplePattern.FindAllStringSubmatch(subject, -1)
	for _, match := range simpleMatches {
		num, _ := strconv.Atoi(match[1])
		if !seen[num] {
			seen[num] = true
			refs = append(refs, IssueReference{
				Number: num,
				Owner:  defaultOwner,
				Repo:   defaultRepo,
				Type:   "related",
				URL:    fmt.Sprintf("https://github.com/%s/%s/issues/%d", defaultOwner, defaultRepo, num),
			})
		}
	}

	return refs
}

// parseRepoFromConfig extracts owner and repo from config
func parseRepoFromConfig(cfg *config.Config) (string, string) {
	if len(cfg.Repositories) > 0 {
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return "", ""
}

// renderHistoryScreen outputs styled history to terminal
func renderHistoryScreen(commits []CommitInfo, targetPath string) {
	// Header
	header := historyHeaderStyle.Render(fmt.Sprintf(" %s ", targetPath))
	commitCount := fmt.Sprintf("%d commits", len(commits))

	fmt.Printf("\n%s  %s\n\n", header, summaryStyle.Render(commitCount))

	// Count change types for summary
	typeCounts := make(map[string]int)

	// Commit entries
	for _, commit := range commits {
		typeCounts[commit.ChangeType]++

		// Hash and date
		hashStr := hashStyle.Render(commit.Hash)
		dateStr := dateStyle.Render(commit.Date.Format("2006-01-02"))
		authorStr := authorStyle.Render(truncate(commit.Author, 12))

		// Change type badge
		typeStyle, ok := changeTypeStyles[commit.ChangeType]
		if !ok {
			typeStyle = changeTypeStyles["Change"]
		}
		typeStr := typeStyle.Render(fmt.Sprintf("[%s]", commit.ChangeType))

		// Subject (truncated)
		subjectStr := truncate(commit.Subject, 50)

		// Issue references
		var refStrs []string
		for _, ref := range commit.References {
			refStrs = append(refStrs, issueRefStyle.Render(fmt.Sprintf("#%d", ref.Number)))
		}
		refStr := ""
		if len(refStrs) > 0 {
			refStr = " " + strings.Join(refStrs, " ")
		}

		fmt.Printf("  %s  %s  %-12s  %s %s%s\n",
			hashStr, dateStr, authorStr, typeStr, subjectStr, refStr)
	}

	// Summary line
	fmt.Println()
	var summaryParts []string
	typeOrder := []string{"Fix", "Add", "Update", "Remove", "Refactor", "Docs", "Test", "Chore", "Change"}
	for _, t := range typeOrder {
		if count, ok := typeCounts[t]; ok && count > 0 {
			style := changeTypeStyles[t]
			summaryParts = append(summaryParts, style.Render(fmt.Sprintf("%s: %d", t, count)))
		}
	}
	if len(summaryParts) > 0 {
		fmt.Printf("  %s\n\n", strings.Join(summaryParts, " | "))
	}
}

// outputHistoryJSON outputs commits as JSON
func outputHistoryJSON(commits []CommitInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(commits)
}

// outputMarkdown writes history to a markdown file in History/ directory
func outputMarkdown(commits []CommitInfo, targetPath, repoOwner, repoName string) error {
	// Create History/ directory
	historyDir := "History"
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create History directory: %w", err)
	}

	// Generate filename from path
	filename := strings.ReplaceAll(targetPath, "/", "-")
	filename = strings.ReplaceAll(filename, "\\", "-")
	filename = strings.ReplaceAll(filename, ".", "-")
	filename = strings.ReplaceAll(filename, ", ", "_")
	filename = strings.Trim(filename, "-")
	if filename == "" {
		filename = "history"
	}
	filename = filename + ".md"
	fullPath := filepath.Join(historyDir, filename)

	// Generate markdown content
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# History: %s\n\n", targetPath))
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString("---\n\n")

	// Summary counts
	typeCounts := make(map[string]int)
	for _, commit := range commits {
		typeCounts[commit.ChangeType]++
	}

	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("**Total Commits:** %d\n\n", len(commits)))
	typeOrder := []string{"Fix", "Add", "Update", "Remove", "Refactor", "Docs", "Test", "Chore", "Change"}
	for _, t := range typeOrder {
		if count, ok := typeCounts[t]; ok && count > 0 {
			b.WriteString(fmt.Sprintf("- **%s:** %d\n", t, count))
		}
	}
	b.WriteString("\n---\n\n")

	// Commits table
	b.WriteString("## Commits\n\n")
	b.WriteString("| Commit | Date | Author | Type | Message | Issues |\n")
	b.WriteString("|--------|------|--------|------|---------|--------|\n")

	for _, commit := range commits {
		// Issue references as links
		var issueLinks []string
		for _, ref := range commit.References {
			issueLinks = append(issueLinks, fmt.Sprintf("[#%d](%s)", ref.Number, ref.URL))
		}
		issuesStr := "-"
		if len(issueLinks) > 0 {
			issuesStr = strings.Join(issueLinks, ", ")
		}

		// Escape pipe characters in subject
		subject := strings.ReplaceAll(commit.Subject, "|", "\\|")

		b.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s | %s |\n",
			commit.Hash,
			commit.Date.Format("2006-01-02"),
			commit.Author,
			commit.ChangeType,
			subject,
			issuesStr,
		))
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	fmt.Printf("✓ History written to %s\n", fullPath)
	return nil
}

// truncate shortens a string to maxLen, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
