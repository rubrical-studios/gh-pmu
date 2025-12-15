package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// patchClient defines the interface for patch operations
// This allows mocking in tests
type patchClient interface {
	// CreateIssue creates a new issue in the repository
	CreateIssue(owner, repo, title, body string, labels []string) (*api.Issue, error)
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
	// GetIssueByNumber returns an issue by its number
	GetIssueByNumber(owner, repo string, number int) (*api.Issue, error)
	// GetProjectItemID returns the project item ID for an issue
	GetProjectItemID(projectID, issueID string) (string, error)
	// GetProjectItemFieldValue returns the current value of a field on a project item
	GetProjectItemFieldValue(projectID, itemID, fieldID string) (string, error)
	// GetIssuesByPatch returns issues assigned to a specific patch
	GetIssuesByPatch(owner, repo, patchVersion string) ([]api.Issue, error)
	// UpdateIssueBody updates an issue's body
	UpdateIssueBody(issueID, body string) error
	// WriteFile writes content to a file path
	WriteFile(path, content string) error
	// MkdirAll creates a directory and all parents
	MkdirAll(path string) error
	// GitAdd stages files to git
	GitAdd(paths ...string) error
	// CloseIssue closes an issue
	CloseIssue(issueID string) error
	// GitTag creates an annotated git tag
	GitTag(tag, message string) error
	// GetLatestGitTag returns the most recent git tag
	GetLatestGitTag() (string, error)
}

// patchStartOptions holds the options for the patch start command
type patchStartOptions struct {
	version string
}

// patchAddOptions holds the options for the patch add command
type patchAddOptions struct {
	issueNumber int
	force       bool
}

// patchRemoveOptions holds the options for the patch remove command
type patchRemoveOptions struct {
	issueNumber int
}

// patchCurrentOptions holds the options for the patch current command
type patchCurrentOptions struct {
	refresh bool
}

// patchCloseOptions holds the options for the patch close command
type patchCloseOptions struct {
	tag bool
}

// patchListOptions holds the options for the patch list command
type patchListOptions struct {
}

// newPatchCommand creates the patch command group
func newPatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Manage patches for IDPF-LTS patch/hotfix development",
		Long:  `Patch commands for long-term support and maintenance releases.`,
	}

	cmd.AddCommand(newPatchStartCommand())
	cmd.AddCommand(newPatchAddCommand())
	cmd.AddCommand(newPatchRemoveCommand())
	cmd.AddCommand(newPatchCurrentCommand())
	cmd.AddCommand(newPatchCloseCommand())
	cmd.AddCommand(newPatchListCommand())

	return cmd
}

// newPatchStartCommand creates the patch start subcommand
func newPatchStartCommand() *cobra.Command {
	opts := &patchStartOptions{}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new patch",
		Long:  `Creates a tracker issue for a new patch.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Wire up real implementation
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.version, "version", "", "Version number for the patch (required)")
	_ = cmd.MarkFlagRequired("version")

	return cmd
}

// newPatchAddCommand creates the patch add subcommand
func newPatchAddCommand() *cobra.Command {
	opts := &patchAddOptions{}

	cmd := &cobra.Command{
		Use:   "add <issue-number>",
		Short: "Add an issue to the current patch",
		Long: `Assigns an issue to the active patch by setting its Patch field.

Validates label constraints:
- Error if issue has breaking-change label
- Warning if issue does not have bug/fix/hotfix label`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var issueNum int
			if _, err := fmt.Sscanf(args[0], "%d", &issueNum); err != nil {
				return fmt.Errorf("invalid issue number: %s", args[0])
			}
			opts.issueNumber = issueNum

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			cfg, err := config.LoadFromDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			client := api.NewClient()
			return runPatchAddWithDeps(cmd, opts, cfg, client)
		},
	}

	cmd.Flags().BoolVar(&opts.force, "force", false, "Add issue even with label warnings")

	return cmd
}

// newPatchRemoveCommand creates the patch remove subcommand
func newPatchRemoveCommand() *cobra.Command {
	opts := &patchRemoveOptions{}

	cmd := &cobra.Command{
		Use:   "remove <issue-number>",
		Short: "Remove an issue from the current patch",
		Long:  `Clears the Patch field from an issue.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var issueNum int
			if _, err := fmt.Sscanf(args[0], "%d", &issueNum); err != nil {
				return fmt.Errorf("invalid issue number: %s", args[0])
			}
			opts.issueNumber = issueNum

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			cfg, err := config.LoadFromDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			client := api.NewClient()
			return runPatchRemoveWithDeps(cmd, opts, cfg, client)
		},
	}

	return cmd
}

// newPatchCurrentCommand creates the patch current subcommand
func newPatchCurrentCommand() *cobra.Command {
	opts := &patchCurrentOptions{}

	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the active patch",
		Long:  `Displays details about the currently active patch.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			cfg, err := config.LoadFromDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			client := api.NewClient()
			return runPatchCurrentWithDeps(cmd, opts, cfg, client)
		},
	}

	cmd.Flags().BoolVar(&opts.refresh, "refresh", false, "Update tracker issue body with current issue list")

	return cmd
}

// newPatchCloseCommand creates the patch close subcommand
func newPatchCloseCommand() *cobra.Command {
	opts := &patchCloseOptions{}

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close the active patch",
		Long:  `Closes the active patch, generates artifacts, and optionally creates a git tag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			cfg, err := config.LoadFromDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			client := api.NewClient()
			return runPatchCloseWithDeps(cmd, opts, cfg, client)
		},
	}

	cmd.Flags().BoolVar(&opts.tag, "tag", false, "Create a git tag for the patch")

	return cmd
}

// newPatchListCommand creates the patch list subcommand
func newPatchListCommand() *cobra.Command {
	opts := &patchListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all patches",
		Long:  `Displays a table of all patches sorted by version.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			cfg, err := config.LoadFromDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			client := api.NewClient()
			return runPatchListWithDeps(cmd, opts, cfg, client)
		},
	}

	return cmd
}

// runPatchStartWithDeps is the testable entry point for patch start
// It receives all dependencies as parameters for easy mocking in tests
func runPatchStartWithDeps(cmd *cobra.Command, opts *patchStartOptions, cfg *config.Config, client patchClient) error {
	// Validate version format
	if err := validateVersion(opts.version); err != nil {
		return err
	}

	// Validate this is a PATCH increment only
	latestTag, err := client.GetLatestGitTag()
	if err == nil {
		if !isPatchIncrement(latestTag, opts.version) {
			return fmt.Errorf("patch releases must be PATCH increments only (e.g., %s -> X.Y.Z+1)", latestTag)
		}
	}

	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Normalize version (strip v prefix for comparison)
	normalizedVersion := normalizeVersion(opts.version)

	// Check for existing active patch (AC-023-3)
	existingIssues, err := client.GetOpenIssuesByLabel(owner, repo, "patch")
	if err != nil {
		return fmt.Errorf("failed to get existing patches: %w", err)
	}

	// Find any active patch tracker
	activePatch := findActivePatch(existingIssues)
	if activePatch != nil {
		return fmt.Errorf("active patch exists: %s", activePatch.Title)
	}

	// Check for duplicate version in closed patches
	closedIssues, err := client.GetClosedIssuesByLabel(owner, repo, "patch")
	if err != nil {
		return fmt.Errorf("failed to get closed patches: %w", err)
	}

	if isDuplicatePatchVersion(normalizedVersion, closedIssues) {
		return fmt.Errorf("version v%s already patched", normalizedVersion)
	}

	// Generate patch title (AC-023-1)
	title := fmt.Sprintf("Patch: v%s", normalizedVersion)

	// Create tracker issue with patch label (AC-023-2)
	labels := []string{"patch"}
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
	fmt.Fprintf(cmd.OutOrStdout(), "Started patch: %s\n", title)
	fmt.Fprintf(cmd.OutOrStdout(), "Tracker issue: #%d\n", issue.Number)

	return nil
}

// findActivePatch finds any active patch tracker from a list of issues
// Returns nil if no active patch is found
func findActivePatch(issues []api.Issue) *api.Issue {
	for i := range issues {
		if strings.HasPrefix(issues[i].Title, "Patch: v") {
			return &issues[i]
		}
	}
	return nil
}

// isDuplicatePatchVersion checks if a version already exists in the list of closed patches
func isDuplicatePatchVersion(version string, closedIssues []api.Issue) bool {
	targetTitle := fmt.Sprintf("Patch: v%s", version)
	for _, issue := range closedIssues {
		if strings.HasPrefix(issue.Title, targetTitle) {
			return true
		}
	}
	return false
}

// isPatchIncrement checks if newVersion is a valid patch increment from baseVersion
func isPatchIncrement(baseVersion, newVersion string) bool {
	base := strings.TrimPrefix(baseVersion, "v")
	new := strings.TrimPrefix(newVersion, "v")

	baseParts := strings.Split(base, ".")
	newParts := strings.Split(new, ".")

	if len(baseParts) != 3 || len(newParts) != 3 {
		return false
	}

	baseMajor, _ := strconv.Atoi(baseParts[0])
	baseMinor, _ := strconv.Atoi(baseParts[1])
	basePatch, _ := strconv.Atoi(baseParts[2])

	newMajor, _ := strconv.Atoi(newParts[0])
	newMinor, _ := strconv.Atoi(newParts[1])
	newPatch, _ := strconv.Atoi(newParts[2])

	// Major and minor must match, patch must be greater
	return newMajor == baseMajor && newMinor == baseMinor && newPatch > basePatch
}

// runPatchAddWithDeps is the testable entry point for patch add
func runPatchAddWithDeps(cmd *cobra.Command, opts *patchAddOptions, cfg *config.Config, client patchClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open patch issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "patch")
	if err != nil {
		return fmt.Errorf("failed to get patch issues: %w", err)
	}

	// Find active patch tracker
	activePatch := findActivePatch(issues)
	if activePatch == nil {
		return fmt.Errorf("no active patch found")
	}

	// Extract version from title
	patchVersion := extractPatchVersion(activePatch.Title)

	// Get the issue to add
	issue, err := client.GetIssueByNumber(owner, repo, opts.issueNumber)
	if err != nil {
		return fmt.Errorf("failed to get issue #%d: %w", opts.issueNumber, err)
	}

	// LTS Constraints (REQ-024)
	// AC-024-2: Check for breaking-change label first (error)
	if hasLabelName(issue.Labels, "breaking-change") {
		return fmt.Errorf("Breaking changes not allowed in patches")
	}

	// AC-024-1: Warn if not labeled bug/fix/hotfix
	hasBugLabel := hasLabelName(issue.Labels, "bug") || hasLabelName(issue.Labels, "fix") || hasLabelName(issue.Labels, "hotfix")

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

	// Set the Patch text field
	patchField, ok := cfg.Fields["patch"]
	if !ok {
		return fmt.Errorf("patch field not configured")
	}

	err = client.SetProjectItemField(project.ID, itemID, patchField.Field, patchVersion)
	if err != nil {
		return fmt.Errorf("failed to set patch field: %w", err)
	}

	// Show warning if applicable (AC-024-1)
	if !hasBugLabel && !opts.force {
		fmt.Fprintf(cmd.OutOrStdout(), "Warning: Issue #%d is not labeled bug/hotfix\n", opts.issueNumber)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Added #%d to patch %s\n", opts.issueNumber, patchVersion)

	return nil
}

// hasLabelName checks if the issue has a specific label by name
func hasLabelName(labels []api.Label, name string) bool {
	for _, label := range labels {
		if label.Name == name {
			return true
		}
	}
	return false
}

// extractPatchVersion extracts the version from a patch title
func extractPatchVersion(title string) string {
	version := strings.TrimPrefix(title, "Patch: ")
	return version
}

// runPatchRemoveWithDeps is the testable entry point for patch remove
func runPatchRemoveWithDeps(cmd *cobra.Command, opts *patchRemoveOptions, cfg *config.Config, client patchClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open patch issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "patch")
	if err != nil {
		return fmt.Errorf("failed to get patch issues: %w", err)
	}

	// Find active patch tracker
	activePatch := findActivePatch(issues)
	if activePatch == nil {
		return fmt.Errorf("no active patch found")
	}

	// Extract version from title
	patchVersion := extractPatchVersion(activePatch.Title)

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

	// Get patch field config
	patchField, ok := cfg.Fields["patch"]
	if !ok {
		return fmt.Errorf("patch field not configured")
	}

	// Check current field value
	currentValue, err := client.GetProjectItemFieldValue(project.ID, itemID, patchField.Field)
	if err != nil {
		return fmt.Errorf("failed to get current patch field value: %w", err)
	}

	// If not assigned to a patch, warn and return
	if currentValue == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Issue #%d is not assigned to a patch\n", opts.issueNumber)
		return nil
	}

	// Clear the Patch text field
	err = client.SetProjectItemField(project.ID, itemID, patchField.Field, "")
	if err != nil {
		return fmt.Errorf("failed to clear patch field: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Removed #%d from patch %s\n", opts.issueNumber, patchVersion)

	return nil
}

// runPatchCurrentWithDeps is the testable entry point for patch current
func runPatchCurrentWithDeps(cmd *cobra.Command, opts *patchCurrentOptions, cfg *config.Config, client patchClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open patch issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "patch")
	if err != nil {
		return fmt.Errorf("failed to get patch issues: %w", err)
	}

	// Find active patch tracker
	activePatch := findActivePatch(issues)
	if activePatch == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "No active patch\n")
		return nil
	}

	// Extract version from title
	patchVersion := extractPatchVersion(activePatch.Title)

	// Get issues assigned to this patch
	patchIssues, err := client.GetIssuesByPatch(owner, repo, patchVersion)
	if err != nil {
		return fmt.Errorf("failed to get patch issues: %w", err)
	}

	// Display patch details
	fmt.Fprintf(cmd.OutOrStdout(), "Current Patch: %s\n", patchVersion)
	fmt.Fprintf(cmd.OutOrStdout(), "Tracker: #%d\n", activePatch.Number)
	fmt.Fprintf(cmd.OutOrStdout(), "Issues: %d\n", len(patchIssues))

	// If refresh flag is set, update tracker issue body
	if opts.refresh {
		body := generatePatchTrackerBody(patchIssues)
		err = client.UpdateIssueBody(activePatch.ID, body)
		if err != nil {
			return fmt.Errorf("failed to update tracker body: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Tracker body updated\n")
	}

	return nil
}

// generatePatchTrackerBody generates the body content for a patch tracker issue
func generatePatchTrackerBody(issues []api.Issue) string {
	var sb strings.Builder
	sb.WriteString("## Issues in this patch\n\n")
	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
	}
	return sb.String()
}

// runPatchCloseWithDeps is the testable entry point for patch close
func runPatchCloseWithDeps(cmd *cobra.Command, opts *patchCloseOptions, cfg *config.Config, client patchClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open patch issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "patch")
	if err != nil {
		return fmt.Errorf("failed to get patch issues: %w", err)
	}

	// Find active patch tracker
	activePatch := findActivePatch(issues)
	if activePatch == nil {
		return fmt.Errorf("no active patch found")
	}

	// Extract version from title
	patchVersion := extractPatchVersion(activePatch.Title)

	// Get issues assigned to this patch
	patchIssues, err := client.GetIssuesByPatch(owner, repo, patchVersion)
	if err != nil {
		return fmt.Errorf("failed to get patch issues: %w", err)
	}

	// Create artifact directory (configurable via release.artifacts.directory)
	// Patches go under {baseDir}/patch/{version}
	artifactDir := fmt.Sprintf("%s/patch/%s", cfg.GetArtifactDirectory(), patchVersion)
	err = client.MkdirAll(artifactDir)
	if err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}

	var artifactPaths []string

	// Generate and write patch-notes.md
	if cfg.ShouldGenerateReleaseNotes() {
		patchNotesPath := fmt.Sprintf("%s/patch-notes.md", artifactDir)
		patchNotesContent := generatePatchNotesContent(patchVersion, activePatch.Number, patchIssues)
		err = client.WriteFile(patchNotesPath, patchNotesContent)
		if err != nil {
			return fmt.Errorf("failed to write patch-notes.md: %w", err)
		}
		artifactPaths = append(artifactPaths, patchNotesPath)
	}

	// Generate and write changelog.md
	if cfg.ShouldGenerateChangelog() {
		changelogPath := fmt.Sprintf("%s/changelog.md", artifactDir)
		changelogContent := generatePatchChangelogContent(patchVersion, patchIssues)
		err = client.WriteFile(changelogPath, changelogContent)
		if err != nil {
			return fmt.Errorf("failed to write changelog.md: %w", err)
		}
		artifactPaths = append(artifactPaths, changelogPath)
	}

	// Stage artifacts to git
	if len(artifactPaths) > 0 {
		err = client.GitAdd(artifactPaths...)
		if err != nil {
			return fmt.Errorf("failed to stage artifacts: %w", err)
		}
	}

	// Create git tag if requested
	if opts.tag {
		tagMessage := fmt.Sprintf("Patch %s", patchVersion)
		err = client.GitTag(patchVersion, tagMessage)
		if err != nil {
			return fmt.Errorf("failed to create git tag: %w", err)
		}
	}

	// Close the tracker issue
	err = client.CloseIssue(activePatch.ID)
	if err != nil {
		return fmt.Errorf("failed to close tracker issue: %w", err)
	}

	// Output confirmation
	fmt.Fprintf(cmd.OutOrStdout(), "Closed patch: %s\n", patchVersion)
	fmt.Fprintf(cmd.OutOrStdout(), "Artifacts created in: %s\n", artifactDir)
	if opts.tag {
		fmt.Fprintf(cmd.OutOrStdout(), "Tag created: %s\n", patchVersion)
	}

	return nil
}

// generatePatchNotesContent generates the patch notes content
func generatePatchNotesContent(version string, trackerNumber int, issues []api.Issue) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Patch %s\n\n", version))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("**Tracker:** #%d\n\n", trackerNumber))

	// Group issues by label
	bugFixes := []api.Issue{}
	other := []api.Issue{}

	for _, issue := range issues {
		labeled := false
		for _, label := range issue.Labels {
			if label.Name == "bug" {
				bugFixes = append(bugFixes, issue)
				labeled = true
				break
			}
		}
		if !labeled {
			other = append(other, issue)
		}
	}

	if len(bugFixes) > 0 {
		sb.WriteString("## Bug Fixes\n\n")
		for _, issue := range bugFixes {
			sb.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
		}
		sb.WriteString("\n")
	}

	if len(other) > 0 {
		sb.WriteString("## Other Changes\n\n")
		for _, issue := range other {
			sb.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// generatePatchChangelogContent generates the changelog content for a patch
func generatePatchChangelogContent(version string, issues []api.Issue) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", version, time.Now().Format("2006-01-02")))

	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
	}

	return sb.String()
}

// runPatchListWithDeps is the testable entry point for patch list
func runPatchListWithDeps(cmd *cobra.Command, opts *patchListOptions, cfg *config.Config, client patchClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open and closed patch issues
	openIssues, err := client.GetOpenIssuesByLabel(owner, repo, "patch")
	if err != nil {
		return fmt.Errorf("failed to get open patches: %w", err)
	}

	closedIssues, err := client.GetClosedIssuesByLabel(owner, repo, "patch")
	if err != nil {
		return fmt.Errorf("failed to get closed patches: %w", err)
	}

	// Combine and filter for patch trackers
	var patches []patchInfo
	for _, issue := range openIssues {
		if strings.HasPrefix(issue.Title, "Patch: v") {
			patches = append(patches, extractPatchInfo(issue, "Active"))
		}
	}
	for _, issue := range closedIssues {
		if strings.HasPrefix(issue.Title, "Patch: v") {
			patches = append(patches, extractPatchInfo(issue, "Released"))
		}
	}

	if len(patches) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No patches found\n")
		return nil
	}

	// Sort by version descending
	sortPatchesByVersionDesc(patches)

	// Display table
	fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-10s %-10s\n", "VERSION", "TRACKER", "STATUS")
	fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-10s %-10s\n", "-------", "-------", "------")
	for _, p := range patches {
		fmt.Fprintf(cmd.OutOrStdout(), "%-12s #%-9d %-10s\n", p.version, p.trackerNum, p.status)
	}

	return nil
}

// patchInfo holds parsed patch information
type patchInfo struct {
	version    string
	trackerNum int
	status     string
}

// extractPatchInfo extracts patch information from an issue
func extractPatchInfo(issue api.Issue, status string) patchInfo {
	version := extractPatchVersion(issue.Title)
	return patchInfo{
		version:    version,
		trackerNum: issue.Number,
		status:     status,
	}
}

// sortPatchesByVersionDesc sorts patches by version in descending order
func sortPatchesByVersionDesc(patches []patchInfo) {
	for i := 0; i < len(patches)-1; i++ {
		for j := 0; j < len(patches)-i-1; j++ {
			if compareVersions(patches[j].version, patches[j+1].version) < 0 {
				patches[j], patches[j+1] = patches[j+1], patches[j]
			}
		}
	}
}
