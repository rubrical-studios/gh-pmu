package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// microsprintClient defines the interface for microsprint operations
// This allows mocking in tests
type microsprintClient interface {
	// CreateIssue creates a new issue in the repository
	CreateIssue(owner, repo, title, body string, labels []string) (*api.Issue, error)
	// GetAuthenticatedUser returns the current authenticated user
	GetAuthenticatedUser() (string, error)
	// GetOpenIssuesByLabel returns open issues with a specific label
	GetOpenIssuesByLabel(owner, repo, label string) ([]api.Issue, error)
	// GetClosedIssuesByLabel returns closed issues with a specific label
	GetClosedIssuesByLabel(owner, repo, label string) ([]api.Issue, error)
	// AddIssueToProject adds an issue to a project and returns the item ID
	AddIssueToProject(projectID, issueID string) (string, error)
	// SetProjectItemField sets a field value on a project item
	SetProjectItemField(projectID, itemID, fieldID, value string) error
	// GetProject returns project details
	GetProject(owner string, number int) (*api.Project, error)
	// CloseIssue closes an issue
	CloseIssue(issueID string) error
	// GetIssueByNumber returns an issue by its number
	GetIssueByNumber(owner, repo string, number int) (*api.Issue, error)
	// GetProjectItemID returns the project item ID for an issue
	GetProjectItemID(projectID, issueID string) (string, error)
	// UpdateIssueBody updates an issue's body (not used for add, but in interface for completeness)
	UpdateIssueBody(issueID, body string) error
	// GetProjectItemFieldValue returns the current value of a field on a project item
	GetProjectItemFieldValue(projectID, itemID, fieldID string) (string, error)
	// GetIssuesByMicrosprint returns issues assigned to a specific microsprint
	GetIssuesByMicrosprint(owner, repo, microsprintName string) ([]api.Issue, error)
	// WriteFile writes content to a file path
	WriteFile(path, content string) error
	// MkdirAll creates a directory and all parents
	MkdirAll(path string) error
	// GitAdd stages files to git
	GitAdd(paths ...string) error
	// GitCommit creates a git commit with the given message
	GitCommit(message string) error
}

// microsprintStartOptions holds the options for the microsprint start command
type microsprintStartOptions struct {
	name string
}

// microsprintCloseOptions holds the options for the microsprint close command
type microsprintCloseOptions struct {
	skipRetro bool
	commit    bool
}

// microsprintAddOptions holds the options for the microsprint add command
type microsprintAddOptions struct {
	issueNumber int
}

// microsprintRemoveOptions holds the options for the microsprint remove command
type microsprintRemoveOptions struct {
	issueNumber int
}

// microsprintCurrentOptions holds the options for the microsprint current command
type microsprintCurrentOptions struct {
	refresh bool
}

// microsprintListOptions holds the options for the microsprint list command
type microsprintListOptions struct {
	// No options needed for basic list
}

// parseOwnerRepo extracts owner and repo from the first configured repository
func parseOwnerRepo(cfg *config.Config) (string, string, error) {
	if len(cfg.Repositories) == 0 {
		return "", "", fmt.Errorf("no repositories configured")
	}
	parts := strings.SplitN(cfg.Repositories[0], "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: %s", cfg.Repositories[0])
	}
	return parts[0], parts[1], nil
}

// findActiveMicrosprint finds today's active microsprint tracker from a list of issues
// Returns nil if no active microsprint is found for today
func findActiveMicrosprint(issues []api.Issue) *api.Issue {
	today := time.Now().Format("2006-01-02")
	prefix := "Microsprint: " + today + "-"

	for i := range issues {
		if strings.HasPrefix(issues[i].Title, prefix) {
			return &issues[i]
		}
	}
	return nil
}

// countActiveMicrosprints counts how many active microsprints exist for today
func countActiveMicrosprints(issues []api.Issue) int {
	today := time.Now().Format("2006-01-02")
	prefix := "Microsprint: " + today + "-"

	count := 0
	for _, issue := range issues {
		if strings.HasPrefix(issue.Title, prefix) {
			count++
		}
	}
	return count
}

// checkMultipleActiveMicrosprints returns an error if more than one active microsprint exists
func checkMultipleActiveMicrosprints(issues []api.Issue) error {
	if countActiveMicrosprints(issues) > 1 {
		return fmt.Errorf("Multiple active microsprints detected. Run 'gh pmu microsprint resolve' to fix")
	}
	return nil
}

// newMicrosprintCommand creates the microsprint command group
func newMicrosprintCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microsprint",
		Short: "Manage microsprints for AI-assisted development",
		Long:  `Microsprint commands for hour-scale work batches.`,
	}

	cmd.AddCommand(newMicrosprintStartCommand())

	return cmd
}

// newMicrosprintStartCommand creates the microsprint start subcommand
func newMicrosprintStartCommand() *cobra.Command {
	opts := &microsprintStartOptions{}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new microsprint",
		Long:  `Creates a tracker issue for a new microsprint.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Wire up real implementation
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.name, "name", "", "Optional name suffix for the microsprint")

	return cmd
}

// runMicrosprintStartWithDeps is the testable entry point for microsprint start
// It receives all dependencies as parameters for easy mocking in tests
func runMicrosprintStartWithDeps(cmd *cobra.Command, opts *microsprintStartOptions, cfg *config.Config, client microsprintClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Generate microsprint name: YYYY-MM-DD-{suffix}
	today := time.Now().Format("2006-01-02")

	// Get existing microsprints to determine next suffix
	existingIssues, err := client.GetOpenIssuesByLabel(owner, repo, "microsprint")
	if err != nil {
		return fmt.Errorf("failed to get existing microsprints: %w", err)
	}

	suffix := getNextMicrosprintSuffix(today, existingIssues)
	title := fmt.Sprintf("Microsprint: %s-%s", today, suffix)

	// Append custom name if provided
	if opts.name != "" {
		title = fmt.Sprintf("%s-%s", title, opts.name)
	}

	// Create tracker issue with microsprint label
	labels := []string{"microsprint"}
	issue, err := client.CreateIssue(owner, repo, title, "", labels)
	if err != nil {
		return fmt.Errorf("failed to create tracker issue: %w", err)
	}

	// Get project
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Add issue to project
	itemID, err := client.AddIssueToProject(project.ID, issue.ID)
	if err != nil {
		return fmt.Errorf("failed to add issue to project: %w", err)
	}

	// Set status to In Progress
	statusField, ok := cfg.Fields["status"]
	if ok {
		statusValue := statusField.Values["in_progress"]
		if statusValue == "" {
			statusValue = "In progress"
		}
		err = client.SetProjectItemField(project.ID, itemID, statusField.Field, statusValue)
		if err != nil {
			return fmt.Errorf("failed to set status: %w", err)
		}
	}

	// Output confirmation
	fmt.Fprintf(cmd.OutOrStdout(), "Started microsprint: %s\n", title)
	fmt.Fprintf(cmd.OutOrStdout(), "Tracker issue: #%d\n", issue.Number)

	return nil
}

// getNextMicrosprintSuffix determines the next available suffix for today's date
// It examines existing microsprint issues and returns the next letter in sequence
func getNextMicrosprintSuffix(today string, existingIssues []api.Issue) string {
	prefix := "Microsprint: " + today + "-"

	// Find all suffixes used today
	var suffixes []string
	for _, issue := range existingIssues {
		if strings.HasPrefix(issue.Title, prefix) {
			// Extract suffix: everything after the prefix, up to the next dash (if custom name)
			rest := strings.TrimPrefix(issue.Title, prefix)
			// Handle custom names: "a-auth" -> "a"
			if idx := strings.Index(rest, "-"); idx > 0 {
				rest = rest[:idx]
			}
			suffixes = append(suffixes, rest)
		}
	}

	// If no suffixes found, start with "a"
	if len(suffixes) == 0 {
		return "a"
	}

	// Find the highest suffix
	highest := suffixes[0]
	for _, s := range suffixes[1:] {
		if compareSuffixes(s, highest) > 0 {
			highest = s
		}
	}

	// Increment to next suffix
	return incrementSuffix(highest)
}

// compareSuffixes compares two suffixes (a < b < ... < z < aa < ab < ...)
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func compareSuffixes(a, b string) int {
	// Longer suffix is greater (aa > z)
	if len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}
		return 1
	}
	// Same length: compare lexicographically
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// incrementSuffix returns the next suffix in sequence
// a -> b, z -> aa, az -> ba, zz -> aaa
func incrementSuffix(s string) string {
	runes := []rune(s)

	// Work from right to left, incrementing
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] < 'z' {
			runes[i]++
			return string(runes)
		}
		// Carry over: set to 'a' and continue to next position
		runes[i] = 'a'
	}

	// All positions were 'z', need to add another character
	return "a" + string(runes)
}

// runMicrosprintCloseWithDeps is the testable entry point for microsprint close
// It receives all dependencies as parameters for easy mocking in tests
func runMicrosprintCloseWithDeps(cmd *cobra.Command, opts *microsprintCloseOptions, cfg *config.Config, client microsprintClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open microsprint issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "microsprint")
	if err != nil {
		return fmt.Errorf("failed to get microsprint issues: %w", err)
	}

	// Check for multiple active microsprints (REQ-013)
	if err := checkMultipleActiveMicrosprints(issues); err != nil {
		return err
	}

	// Find active tracker issue for today
	activeTracker := findActiveMicrosprint(issues)
	if activeTracker == nil {
		return fmt.Errorf("no active microsprint. Run 'gh pmu microsprint start' first")
	}

	// Close the tracker issue
	err = client.CloseIssue(activeTracker.ID)
	if err != nil {
		return fmt.Errorf("failed to close tracker issue: %w", err)
	}

	return nil
}

// runMicrosprintAddWithDeps is the testable entry point for microsprint add
// It receives all dependencies as parameters for easy mocking in tests
func runMicrosprintAddWithDeps(cmd *cobra.Command, opts *microsprintAddOptions, cfg *config.Config, client microsprintClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open microsprint issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "microsprint")
	if err != nil {
		return fmt.Errorf("failed to get microsprint issues: %w", err)
	}

	// Check for multiple active microsprints (REQ-013)
	if err := checkMultipleActiveMicrosprints(issues); err != nil {
		return err
	}

	// Find active tracker issue for today
	activeTracker := findActiveMicrosprint(issues)
	if activeTracker == nil {
		return fmt.Errorf("no active microsprint. Run 'gh pmu microsprint start' first")
	}

	// Extract microsprint name from title (e.g., "Microsprint: 2025-12-13-a" -> "2025-12-13-a")
	microsprintName := strings.TrimPrefix(activeTracker.Title, "Microsprint: ")

	// Get the issue to add
	issue, err := client.GetIssueByNumber(owner, repo, opts.issueNumber)
	if err != nil {
		return fmt.Errorf("failed to get issue #%d: %w", opts.issueNumber, err)
	}

	// Get project
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get project item ID for the issue
	itemID, err := client.GetProjectItemID(project.ID, issue.ID)
	if err != nil {
		return fmt.Errorf("failed to get project item for issue #%d: %w", opts.issueNumber, err)
	}

	// Set the Microsprint text field
	microsprintField, ok := cfg.Fields["microsprint"]
	if !ok {
		return fmt.Errorf("microsprint field not configured")
	}

	err = client.SetProjectItemField(project.ID, itemID, microsprintField.Field, microsprintName)
	if err != nil {
		return fmt.Errorf("failed to set microsprint field: %w", err)
	}

	// Output confirmation (AC-003-2)
	fmt.Fprintf(cmd.OutOrStdout(), "Added #%d to microsprint %s\n", opts.issueNumber, microsprintName)

	return nil
}

// runMicrosprintRemoveWithDeps is the testable entry point for microsprint remove
// It receives all dependencies as parameters for easy mocking in tests
func runMicrosprintRemoveWithDeps(cmd *cobra.Command, opts *microsprintRemoveOptions, cfg *config.Config, client microsprintClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open microsprint issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "microsprint")
	if err != nil {
		return fmt.Errorf("failed to get microsprint issues: %w", err)
	}

	// Check for multiple active microsprints (REQ-013)
	if err := checkMultipleActiveMicrosprints(issues); err != nil {
		return err
	}

	// Find active tracker issue for today
	activeTracker := findActiveMicrosprint(issues)
	if activeTracker == nil {
		return fmt.Errorf("no active microsprint. Run 'gh pmu microsprint start' first")
	}

	// Extract microsprint name from title
	microsprintName := strings.TrimPrefix(activeTracker.Title, "Microsprint: ")

	// Get the issue to remove
	issue, err := client.GetIssueByNumber(owner, repo, opts.issueNumber)
	if err != nil {
		return fmt.Errorf("failed to get issue #%d: %w", opts.issueNumber, err)
	}

	// Get project
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get project item ID for the issue
	itemID, err := client.GetProjectItemID(project.ID, issue.ID)
	if err != nil {
		return fmt.Errorf("failed to get project item for issue #%d: %w", opts.issueNumber, err)
	}

	// Get microsprint field config
	microsprintField, ok := cfg.Fields["microsprint"]
	if !ok {
		return fmt.Errorf("microsprint field not configured")
	}

	// Check current field value (AC-038-3)
	currentValue, err := client.GetProjectItemFieldValue(project.ID, itemID, microsprintField.Field)
	if err != nil {
		return fmt.Errorf("failed to get current microsprint field value: %w", err)
	}

	// If not assigned to a microsprint, warn and return
	if currentValue == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Issue #%d is not assigned to a microsprint\n", opts.issueNumber)
		return nil
	}

	// Clear the Microsprint text field (AC-038-1)
	err = client.SetProjectItemField(project.ID, itemID, microsprintField.Field, "")
	if err != nil {
		return fmt.Errorf("failed to clear microsprint field: %w", err)
	}

	// Output confirmation (AC-038-2)
	fmt.Fprintf(cmd.OutOrStdout(), "Removed #%d from microsprint %s\n", opts.issueNumber, microsprintName)

	return nil
}

// runMicrosprintCurrentWithDeps is the testable entry point for microsprint current
// It receives all dependencies as parameters for easy mocking in tests
func runMicrosprintCurrentWithDeps(cmd *cobra.Command, opts *microsprintCurrentOptions, cfg *config.Config, client microsprintClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open microsprint issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "microsprint")
	if err != nil {
		return fmt.Errorf("failed to get microsprint issues: %w", err)
	}

	// Check for multiple active microsprints (REQ-013)
	if err := checkMultipleActiveMicrosprints(issues); err != nil {
		return err
	}

	// Find active tracker issue for today
	activeTracker := findActiveMicrosprint(issues)
	if activeTracker == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "No active microsprint\n")
		return nil
	}

	// Extract microsprint name from title
	microsprintName := strings.TrimPrefix(activeTracker.Title, "Microsprint: ")

	// Get issues assigned to this microsprint
	microsprintIssues, err := client.GetIssuesByMicrosprint(owner, repo, microsprintName)
	if err != nil {
		return fmt.Errorf("failed to get microsprint issues: %w", err)
	}

	// Display microsprint details (AC-035-1)
	fmt.Fprintf(cmd.OutOrStdout(), "Microsprint: %s\n", microsprintName)
	fmt.Fprintf(cmd.OutOrStdout(), "Tracker: #%d\n", activeTracker.Number)
	fmt.Fprintf(cmd.OutOrStdout(), "Issues: %d\n", len(microsprintIssues))

	// If refresh flag is set, update tracker issue body (AC-035-3)
	if opts.refresh {
		body := generateTrackerBody(microsprintIssues)
		err = client.UpdateIssueBody(activeTracker.ID, body)
		if err != nil {
			return fmt.Errorf("failed to update tracker issue body: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Tracker issue updated\n")
	}

	return nil
}

// generateTrackerBody creates the body content for a tracker issue
func generateTrackerBody(issues []api.Issue) string {
	if len(issues) == 0 {
		return "No issues assigned to this microsprint."
	}

	var body strings.Builder
	body.WriteString("## Issues\n\n")
	for _, issue := range issues {
		body.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
	}
	return body.String()
}

// runMicrosprintCloseArtifactsWithDeps is the testable entry point for microsprint close with artifacts
// It receives all dependencies as parameters for easy mocking in tests
func runMicrosprintCloseArtifactsWithDeps(cmd *cobra.Command, opts *microsprintCloseOptions, cfg *config.Config, client microsprintClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open microsprint issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "microsprint")
	if err != nil {
		return fmt.Errorf("failed to get microsprint issues: %w", err)
	}

	// Find active tracker issue for today
	activeTracker := findActiveMicrosprint(issues)
	if activeTracker == nil {
		return fmt.Errorf("no active microsprint. Run 'gh pmu microsprint start' first")
	}

	// Extract microsprint name from title
	microsprintName := strings.TrimPrefix(activeTracker.Title, "Microsprint: ")

	// Get issues assigned to this microsprint
	microsprintIssues, err := client.GetIssuesByMicrosprint(owner, repo, microsprintName)
	if err != nil {
		return fmt.Errorf("failed to get microsprint issues: %w", err)
	}

	// Create artifacts directory
	artifactDir := "Microsprints/" + microsprintName
	err = client.MkdirAll(artifactDir)
	if err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	// Generate review.md (AC-004-1)
	reviewPath := artifactDir + "/review.md"
	reviewContent := generateReviewContent(microsprintName, microsprintIssues)
	err = client.WriteFile(reviewPath, reviewContent)
	if err != nil {
		return fmt.Errorf("failed to write review.md: %w", err)
	}

	// Generate retro.md with empty template when --skip-retro (AC-004-3)
	retroPath := artifactDir + "/retro.md"
	retroContent := generateRetroTemplate(microsprintName)
	err = client.WriteFile(retroPath, retroContent)
	if err != nil {
		return fmt.Errorf("failed to write retro.md: %w", err)
	}

	// Stage files to git (AC-004-4)
	err = client.GitAdd(reviewPath, retroPath)
	if err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Commit if --commit flag set (AC-004-5)
	if opts.commit {
		commitMsg := fmt.Sprintf("docs: Close microsprint %s", microsprintName)
		err = client.GitCommit(commitMsg)
		if err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
	}

	// Update tracker issue body with artifact links (AC-004-6)
	trackerBody := generateTrackerCloseBody(microsprintIssues, artifactDir)
	err = client.UpdateIssueBody(activeTracker.ID, trackerBody)
	if err != nil {
		return fmt.Errorf("failed to update tracker issue: %w", err)
	}

	// Close the tracker issue (AC-004-6)
	err = client.CloseIssue(activeTracker.ID)
	if err != nil {
		return fmt.Errorf("failed to close tracker issue: %w", err)
	}

	return nil
}

// generateReviewContent creates the review.md content with issue summary
func generateReviewContent(microsprintName string, issues []api.Issue) string {
	var body strings.Builder
	body.WriteString(fmt.Sprintf("# Microsprint Review: %s\n\n", microsprintName))
	body.WriteString("## Issues\n\n")

	// Handle empty microsprint (REQ-016)
	if len(issues) == 0 {
		body.WriteString("No issues completed in this microsprint.\n")
	} else {
		for _, issue := range issues {
			status := "open"
			if issue.State == "CLOSED" {
				status = "closed"
			}
			body.WriteString(fmt.Sprintf("- #%d %s (%s)\n", issue.Number, issue.Title, status))
		}
	}
	return body.String()
}

// generateRetroTemplate creates an empty retrospective template
func generateRetroTemplate(microsprintName string) string {
	var body strings.Builder
	body.WriteString(fmt.Sprintf("# Microsprint Retrospective: %s\n\n", microsprintName))
	body.WriteString("## What Went Well\n\n")
	body.WriteString("- \n\n")
	body.WriteString("## What Could Be Improved\n\n")
	body.WriteString("- \n\n")
	body.WriteString("## Action Items\n\n")
	body.WriteString("- \n")
	return body.String()
}

// generateTrackerCloseBody creates the tracker issue body for close with artifact links
func generateTrackerCloseBody(issues []api.Issue, artifactDir string) string {
	var body strings.Builder
	body.WriteString("## Issues\n\n")
	for _, issue := range issues {
		body.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
	}
	body.WriteString("\n## Artifacts\n\n")
	body.WriteString(fmt.Sprintf("- [review.md](%s/review.md)\n", artifactDir))
	body.WriteString(fmt.Sprintf("- [retro.md](%s/retro.md)\n", artifactDir))
	return body.String()
}

// runMicrosprintListWithDeps is the testable entry point for microsprint list
// It receives all dependencies as parameters for easy mocking in tests
func runMicrosprintListWithDeps(cmd *cobra.Command, opts *microsprintListOptions, cfg *config.Config, client microsprintClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get both open and closed microsprint issues
	openIssues, err := client.GetOpenIssuesByLabel(owner, repo, "microsprint")
	if err != nil {
		return fmt.Errorf("failed to get open microsprint issues: %w", err)
	}

	closedIssues, err := client.GetClosedIssuesByLabel(owner, repo, "microsprint")
	if err != nil {
		return fmt.Errorf("failed to get closed microsprint issues: %w", err)
	}

	// Filter for valid tracker issues and combine
	var allTrackers []api.Issue
	for _, issue := range openIssues {
		if strings.HasPrefix(issue.Title, "Microsprint: ") {
			allTrackers = append(allTrackers, issue)
		}
	}
	for _, issue := range closedIssues {
		if strings.HasPrefix(issue.Title, "Microsprint: ") {
			allTrackers = append(allTrackers, issue)
		}
	}

	// Handle no microsprints found
	if len(allTrackers) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No microsprints found\n")
		return nil
	}

	// Sort by microsprint name (date) descending
	sortMicrosprintsByDateDesc(allTrackers)

	// Print table header
	fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-10s  %-10s\n", "MICROSPRINT", "TRACKER", "STATUS")
	fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-10s  %-10s\n", "--------------------", "----------", "----------")

	// Print each microsprint
	for _, tracker := range allTrackers {
		name := strings.TrimPrefix(tracker.Title, "Microsprint: ")
		trackerNum := fmt.Sprintf("#%d", tracker.Number)
		status := "Active"
		if tracker.State == "CLOSED" {
			status = "Closed"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-10s  %-10s\n", name, trackerNum, status)
	}

	return nil
}

// sortMicrosprintsByDateDesc sorts microsprint issues by their date in descending order
func sortMicrosprintsByDateDesc(issues []api.Issue) {
	// Simple bubble sort for now (microsprints list is typically small)
	for i := 0; i < len(issues)-1; i++ {
		for j := i + 1; j < len(issues); j++ {
			// Extract date portion from title: "Microsprint: YYYY-MM-DD-x"
			nameI := strings.TrimPrefix(issues[i].Title, "Microsprint: ")
			nameJ := strings.TrimPrefix(issues[j].Title, "Microsprint: ")

			// Compare - larger (more recent) date should come first
			if nameI < nameJ {
				issues[i], issues[j] = issues[j], issues[i]
			}
		}
	}
}
