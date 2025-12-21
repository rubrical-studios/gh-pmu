package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// semverRegex matches valid semver versions with optional v prefix
var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

// releaseClient defines the interface for release operations
// This allows mocking in tests
type releaseClient interface {
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
	// GetIssuesByRelease returns issues assigned to a specific release
	GetIssuesByRelease(owner, repo, releaseVersion string) ([]api.Issue, error)
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
	// ReopenIssue reopens a closed issue
	ReopenIssue(issueID string) error
	// GitTag creates an annotated git tag
	GitTag(tag, message string) error
	// GitCheckoutNewBranch creates and checks out a new git branch
	GitCheckoutNewBranch(branch string) error
}

// releaseStartOptions holds the options for the release start command
type releaseStartOptions struct {
	branch string
}

// releaseAddOptions holds the options for the release add command
type releaseAddOptions struct {
	issueNumber int
}

// releaseRemoveOptions holds the options for the release remove command
type releaseRemoveOptions struct {
	issueNumber int
}

// releaseCurrentOptions holds the options for the release current command
type releaseCurrentOptions struct {
	refresh bool
}

// releaseCloseOptions holds the options for the release close command
type releaseCloseOptions struct {
	tag         bool
	yes         bool
	releaseName string
}

// releaseListOptions holds the options for the release list command
type releaseListOptions struct {
}

// newReleaseCommand creates the release command group
func newReleaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Manage releases for IDPF-Structured development",
		Long:  `Release commands for version-based deployment workflows.`,
	}

	cmd.AddCommand(newReleaseStartCommand())
	cmd.AddCommand(newReleaseAddCommand())
	cmd.AddCommand(newReleaseRemoveCommand())
	cmd.AddCommand(newReleaseCurrentCommand())
	cmd.AddCommand(newReleaseCloseCommand())
	cmd.AddCommand(newReleaseReopenCommand())
	cmd.AddCommand(newReleaseListCommand())

	return cmd
}

// newReleaseStartCommand creates the release start subcommand
func newReleaseStartCommand() *cobra.Command {
	opts := &releaseStartOptions{}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new release",
		Long: `Creates a tracker issue for a new release and creates the git branch.

The --branch flag is required and specifies the branch name to create.
The branch name is used literally for the tracker title, Release field,
and artifact directory.

Examples:
  gh pmu release start --branch release/v2.0.0
  gh pmu release start --branch patch/v1.9.1
  gh pmu release start --branch hotfix-auth-bypass`,
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
			return runReleaseStartWithDeps(cmd, opts, cfg, client)
		},
	}

	cmd.Flags().StringVar(&opts.branch, "branch", "", "Branch name for the release (required)")
	_ = cmd.MarkFlagRequired("branch")

	return cmd
}

// newReleaseAddCommand creates the release add subcommand
func newReleaseAddCommand() *cobra.Command {
	opts := &releaseAddOptions{}

	cmd := &cobra.Command{
		Use:   "add <issue-number>",
		Short: "Add an issue to the current release",
		Long:  `Assigns an issue to the active release by setting its Release field.`,
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
			return runReleaseAddWithDeps(cmd, opts, cfg, client)
		},
	}

	return cmd
}

// newReleaseRemoveCommand creates the release remove subcommand
func newReleaseRemoveCommand() *cobra.Command {
	opts := &releaseRemoveOptions{}

	cmd := &cobra.Command{
		Use:   "remove <issue-number>",
		Short: "Remove an issue from the current release",
		Long:  `Clears the Release field from an issue.`,
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
			return runReleaseRemoveWithDeps(cmd, opts, cfg, client)
		},
	}

	return cmd
}

// newReleaseCurrentCommand creates the release current subcommand
func newReleaseCurrentCommand() *cobra.Command {
	opts := &releaseCurrentOptions{}

	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the active release",
		Long:  `Displays details about the currently active release.`,
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
			return runReleaseCurrentWithDeps(cmd, opts, cfg, client)
		},
	}

	cmd.Flags().BoolVar(&opts.refresh, "refresh", false, "Update tracker issue body with current issue list")

	return cmd
}

// newReleaseCloseCommand creates the release close subcommand
func newReleaseCloseCommand() *cobra.Command {
	opts := &releaseCloseOptions{}

	cmd := &cobra.Command{
		Use:   "close <release-name>",
		Short: "Close a release",
		Long: `Closes a release, generates artifacts, and optionally creates a git tag.

The release name must be specified explicitly (e.g., release/v2.0.0).
Incomplete issues will be moved to backlog with Release and Microsprint fields cleared.

Examples:
  gh pmu release close release/v2.0.0
  gh pmu release close patch/v1.9.1 --tag
  gh pmu release close release/v2.0.0 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.releaseName = args[0]

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			cfg, err := config.LoadFromDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			client := api.NewClient()
			return runReleaseCloseWithDeps(cmd, opts, cfg, client)
		},
	}

	cmd.Flags().BoolVar(&opts.tag, "tag", false, "Create a git tag for the release")
	cmd.Flags().BoolVarP(&opts.yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// newReleaseListCommand creates the release list subcommand
func newReleaseListCommand() *cobra.Command {
	opts := &releaseListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all releases",
		Long:  `Displays a table of all releases sorted by version.`,
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
			return runReleaseListWithDeps(cmd, opts, cfg, client)
		},
	}

	return cmd
}

// runReleaseStartWithDeps is the testable entry point for release start
// It receives all dependencies as parameters for easy mocking in tests
func runReleaseStartWithDeps(cmd *cobra.Command, opts *releaseStartOptions, cfg *config.Config, client releaseClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Check for existing active release
	existingIssues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
	if err != nil {
		return fmt.Errorf("failed to get existing releases: %w", err)
	}

	// Find any active release tracker
	activeRelease := findActiveRelease(existingIssues)
	if activeRelease != nil {
		return fmt.Errorf("active release exists: %s", activeRelease.Title)
	}

	// Create the git branch
	err = client.GitCheckoutNewBranch(opts.branch)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Use branch name for tracker title and Release field
	title := fmt.Sprintf("Release: %s", opts.branch)

	// Create tracker issue with release label
	labels := []string{"release"}
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

	// Parse version and track from branch name for config storage
	version, track := parseReleaseTitle(title)

	// Add to active releases in config
	cfg.AddActiveRelease(config.ActiveRelease{
		Version:      version,
		TrackerIssue: issue.Number,
		Track:        track,
	})

	// Save config
	cwd, _ := os.Getwd()
	configPath, err := config.FindConfigFile(cwd)
	if err == nil {
		if err := cfg.Save(configPath); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update config: %v\n", err)
		}
	}

	// Output confirmation
	fmt.Fprintf(cmd.OutOrStdout(), "Created branch: %s\n", opts.branch)
	fmt.Fprintf(cmd.OutOrStdout(), "Started release: %s\n", title)
	fmt.Fprintf(cmd.OutOrStdout(), "Tracker issue: #%d\n", issue.Number)

	return nil
}

// findActiveRelease finds any active release tracker from a list of issues
// Returns nil if no active release is found
func findActiveRelease(issues []api.Issue) *api.Issue {
	for i := range issues {
		if strings.HasPrefix(issues[i].Title, "Release: ") {
			return &issues[i]
		}
	}
	return nil
}

// validateVersion validates that a version string is valid semver format
// Accepts X.Y.Z or vX.Y.Z format
func validateVersion(version string) error {
	if !semverRegex.MatchString(version) {
		return fmt.Errorf("Invalid version format. Use semver: X.Y.Z")
	}
	return nil
}

// runReleaseAddWithDeps is the testable entry point for release add
// It receives all dependencies as parameters for easy mocking in tests
func runReleaseAddWithDeps(cmd *cobra.Command, opts *releaseAddOptions, cfg *config.Config, client releaseClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open release issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
	if err != nil {
		return fmt.Errorf("failed to get release issues: %w", err)
	}

	// Find active release tracker
	activeRelease := findActiveRelease(issues)
	if activeRelease == nil {
		return fmt.Errorf("no active release found")
	}

	// Extract version from title (e.g., "Release: v1.2.0" or "Release: v1.2.0 (Phoenix)" -> "v1.2.0")
	releaseVersion := extractReleaseVersion(activeRelease.Title)

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

	// Set the Release text field
	releaseField, ok := cfg.Fields["release"]
	if !ok {
		return fmt.Errorf("release field not configured")
	}

	err = client.SetProjectItemField(project.ID, itemID, releaseField.Field, releaseVersion)
	if err != nil {
		return fmt.Errorf("failed to set release field: %w", err)
	}

	// Output confirmation (AC-019-2)
	fmt.Fprintf(cmd.OutOrStdout(), "Added #%d to release %s\n", opts.issueNumber, releaseVersion)

	return nil
}

// extractReleaseVersion extracts the version from a release title
// e.g., "Release: v1.2.0" -> "v1.2.0", "Release: v1.2.0 (Phoenix)" -> "v1.2.0"
func extractReleaseVersion(title string) string {
	// Remove "Release: " prefix
	version := strings.TrimPrefix(title, "Release: ")
	// If there's a codename in parentheses, remove it
	if idx := strings.Index(version, " ("); idx > 0 {
		version = version[:idx]
	}
	return version
}

// runReleaseRemoveWithDeps is the testable entry point for release remove
// It receives all dependencies as parameters for easy mocking in tests
func runReleaseRemoveWithDeps(cmd *cobra.Command, opts *releaseRemoveOptions, cfg *config.Config, client releaseClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open release issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
	if err != nil {
		return fmt.Errorf("failed to get release issues: %w", err)
	}

	// Find active release tracker
	activeRelease := findActiveRelease(issues)
	if activeRelease == nil {
		return fmt.Errorf("no active release found")
	}

	// Extract version from title
	releaseVersion := extractReleaseVersion(activeRelease.Title)

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

	// Get release field config
	releaseField, ok := cfg.Fields["release"]
	if !ok {
		return fmt.Errorf("release field not configured")
	}

	// Check current field value (AC-039-3)
	currentValue, err := client.GetProjectItemFieldValue(project.ID, itemID, releaseField.Field)
	if err != nil {
		return fmt.Errorf("failed to get current release field value: %w", err)
	}

	// If not assigned to a release, warn and return
	if currentValue == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Issue #%d is not assigned to a release\n", opts.issueNumber)
		return nil
	}

	// Clear the Release text field (AC-039-1)
	err = client.SetProjectItemField(project.ID, itemID, releaseField.Field, "")
	if err != nil {
		return fmt.Errorf("failed to clear release field: %w", err)
	}

	// Output confirmation (AC-039-2)
	fmt.Fprintf(cmd.OutOrStdout(), "Removed #%d from release %s\n", opts.issueNumber, releaseVersion)

	return nil
}

// runReleaseCurrentWithDeps is the testable entry point for release current
// It receives all dependencies as parameters for easy mocking in tests
func runReleaseCurrentWithDeps(cmd *cobra.Command, opts *releaseCurrentOptions, cfg *config.Config, client releaseClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open release issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
	if err != nil {
		return fmt.Errorf("failed to get release issues: %w", err)
	}

	// Find active release tracker
	activeRelease := findActiveRelease(issues)
	if activeRelease == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "No active release\n")
		return nil
	}

	// Extract version from title
	releaseVersion := extractReleaseVersion(activeRelease.Title)

	// Get issues assigned to this release
	releaseIssues, err := client.GetIssuesByRelease(owner, repo, releaseVersion)
	if err != nil {
		return fmt.Errorf("failed to get release issues: %w", err)
	}

	// Display release details (AC-036-1)
	fmt.Fprintf(cmd.OutOrStdout(), "Current Release: %s\n", releaseVersion)
	fmt.Fprintf(cmd.OutOrStdout(), "Tracker: #%d\n", activeRelease.Number)
	fmt.Fprintf(cmd.OutOrStdout(), "Issues: %d\n", len(releaseIssues))

	// If refresh flag is set, update tracker issue body (AC-036-3)
	if opts.refresh {
		body := generateReleaseTrackerBody(releaseIssues)
		err = client.UpdateIssueBody(activeRelease.ID, body)
		if err != nil {
			return fmt.Errorf("failed to update tracker body: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Tracker body updated\n")
	}

	return nil
}

// generateReleaseTrackerBody generates the body content for a release tracker issue
func generateReleaseTrackerBody(issues []api.Issue) string {
	var sb strings.Builder
	sb.WriteString("## Issues in this release\n\n")
	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
	}
	return sb.String()
}

// runReleaseCloseWithDeps is the testable entry point for release close
// It receives all dependencies as parameters for easy mocking in tests
func runReleaseCloseWithDeps(cmd *cobra.Command, opts *releaseCloseOptions, cfg *config.Config, client releaseClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open release issues
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
	if err != nil {
		return fmt.Errorf("failed to get release issues: %w", err)
	}

	// Find the specified release by name
	var targetRelease *api.Issue
	expectedTitle := fmt.Sprintf("Release: %s", opts.releaseName)
	for i := range issues {
		if issues[i].Title == expectedTitle || strings.HasPrefix(issues[i].Title, expectedTitle+" (") {
			targetRelease = &issues[i]
			break
		}
	}
	if targetRelease == nil {
		return fmt.Errorf("release not found: %s", opts.releaseName)
	}

	// Extract version and codename from title
	releaseVersion := extractReleaseVersion(targetRelease.Title)
	codename := extractReleaseCodename(targetRelease.Title)

	// Get issues assigned to this release
	releaseIssues, err := client.GetIssuesByRelease(owner, repo, releaseVersion)
	if err != nil {
		return fmt.Errorf("failed to get release issues: %w", err)
	}

	// Count done vs incomplete issues
	var doneIssues, incompleteIssues []api.Issue
	for _, issue := range releaseIssues {
		if issue.State == "CLOSED" || issue.State == "closed" {
			doneIssues = append(doneIssues, issue)
		} else {
			incompleteIssues = append(incompleteIssues, issue)
		}
	}

	// Show release summary
	fmt.Fprintf(cmd.OutOrStdout(), "Closing release: %s\n", opts.releaseName)
	fmt.Fprintf(cmd.OutOrStdout(), "  Tracker issue: #%d\n", targetRelease.Number)
	fmt.Fprintf(cmd.OutOrStdout(), "  Issues in release: %d (%d done, %d incomplete)\n",
		len(releaseIssues), len(doneIssues), len(incompleteIssues))
	fmt.Fprintln(cmd.OutOrStdout())

	// Warn about incomplete issues and confirm
	if len(incompleteIssues) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "⚠️  %d issue(s) are not done. They will be moved to backlog.\n", len(incompleteIssues))

		if !opts.yes {
			fmt.Fprint(cmd.OutOrStdout(), "Proceed? (y/n): ")
			var response string
			_, _ = fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}
		}
		fmt.Fprintln(cmd.OutOrStdout())

		// Move incomplete issues to backlog and clear Release/Microsprint fields
		fmt.Fprintln(cmd.OutOrStdout(), "Moving incomplete issues to backlog...")
		project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}

		for _, issue := range incompleteIssues {
			// Get project item ID
			itemID, err := client.GetProjectItemID(project.ID, issue.ID)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "  Warning: could not find project item for #%d: %v\n", issue.Number, err)
				continue
			}

			// Clear Release field
			if releaseField, ok := cfg.Fields["release"]; ok {
				_ = client.SetProjectItemField(project.ID, itemID, releaseField.Field, "")
			}

			// Clear Microsprint field
			if microsprintField, ok := cfg.Fields["microsprint"]; ok {
				_ = client.SetProjectItemField(project.ID, itemID, microsprintField.Field, "")
			}

			// Set status to backlog
			if statusField, ok := cfg.Fields["status"]; ok {
				backlogValue := statusField.Values["backlog"]
				if backlogValue == "" {
					backlogValue = "Backlog"
				}
				_ = client.SetProjectItemField(project.ID, itemID, statusField.Field, backlogValue)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "  #%d - %s\n", issue.Number, issue.Title)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	} else if !opts.yes {
		// Confirm even without incomplete issues
		fmt.Fprint(cmd.OutOrStdout(), "Proceed? (y/n): ")
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Create artifact directory (configurable via release.artifacts.directory)
	artifactDir := fmt.Sprintf("%s/%s", cfg.GetArtifactDirectory(), releaseVersion)
	err = client.MkdirAll(artifactDir)
	if err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}

	var artifactPaths []string

	// Generate and write release-notes.md
	if cfg.ShouldGenerateReleaseNotes() {
		releaseNotesPath := fmt.Sprintf("%s/release-notes.md", artifactDir)
		releaseNotesContent := generateReleaseNotesContent(releaseVersion, codename, targetRelease.Number, doneIssues)
		err = client.WriteFile(releaseNotesPath, releaseNotesContent)
		if err != nil {
			return fmt.Errorf("failed to write release-notes.md: %w", err)
		}
		artifactPaths = append(artifactPaths, releaseNotesPath)
	}

	// Generate and write changelog.md
	if cfg.ShouldGenerateChangelog() {
		changelogPath := fmt.Sprintf("%s/changelog.md", artifactDir)
		changelogContent := generateChangelogContent(releaseVersion, doneIssues)
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
		tagMessage := fmt.Sprintf("Release %s", releaseVersion)
		err = client.GitTag(releaseVersion, tagMessage)
		if err != nil {
			return fmt.Errorf("failed to create git tag: %w", err)
		}
	}

	// Close the tracker issue
	err = client.CloseIssue(targetRelease.ID)
	if err != nil {
		return fmt.Errorf("failed to close tracker issue: %w", err)
	}

	// Remove from active releases in config
	cfg.RemoveActiveRelease(targetRelease.Number)

	// Save config
	cwd, _ := os.Getwd()
	configPath, err := config.FindConfigFile(cwd)
	if err == nil {
		if err := cfg.Save(configPath); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update config: %v\n", err)
		}
	}

	// Output confirmation
	fmt.Fprintf(cmd.OutOrStdout(), "✓ Release closed: %s\n", releaseVersion)
	fmt.Fprintf(cmd.OutOrStdout(), "✓ Artifacts created in: %s\n", artifactDir)
	if len(incompleteIssues) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ %d issue(s) moved to backlog (Release and Microsprint cleared)\n", len(incompleteIssues))
	}
	if opts.tag {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Tag created: %s\n", releaseVersion)
	}

	return nil
}

// newReleaseReopenCommand creates the release reopen subcommand
func newReleaseReopenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reopen <release-name>",
		Short: "Reopen a closed release",
		Long: `Reopens a previously closed release tracker issue.

Use this to continue work on a release after it has been closed.
The release name must be specified explicitly.

Examples:
  gh pmu release reopen release/v2.0.0
  gh pmu release reopen patch/v1.9.1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseName := args[0]

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			cfg, err := config.LoadFromDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w\nRun 'gh pmu init' to create a configuration file", err)
			}

			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			client := api.NewClient()
			return runReleaseReopenWithDeps(cmd, releaseName, cfg, client)
		},
	}

	return cmd
}

func runReleaseReopenWithDeps(cmd *cobra.Command, releaseName string, cfg *config.Config, client releaseClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get closed release issues
	issues, err := client.GetClosedIssuesByLabel(owner, repo, "release")
	if err != nil {
		return fmt.Errorf("failed to get closed release issues: %w", err)
	}

	// Find the specified release by name
	var targetRelease *api.Issue
	expectedTitle := fmt.Sprintf("Release: %s", releaseName)
	for i := range issues {
		if issues[i].Title == expectedTitle || strings.HasPrefix(issues[i].Title, expectedTitle+" (") {
			targetRelease = &issues[i]
			break
		}
	}

	if targetRelease == nil {
		return fmt.Errorf("closed release not found: %s", releaseName)
	}

	// Reopen the tracker issue
	err = client.ReopenIssue(targetRelease.ID)
	if err != nil {
		return fmt.Errorf("failed to reopen tracker issue: %w", err)
	}

	// Parse version and track for config storage
	version, track := parseReleaseTitle(targetRelease.Title)

	// Add to active releases in config
	cfg.AddActiveRelease(config.ActiveRelease{
		Version:      version,
		TrackerIssue: targetRelease.Number,
		Track:        track,
	})

	// Save config
	cwd, _ := os.Getwd()
	configPath, err := config.FindConfigFile(cwd)
	if err == nil {
		if err := cfg.Save(configPath); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update config: %v\n", err)
		}
	}

	releaseVersion := extractReleaseVersion(targetRelease.Title)
	fmt.Fprintf(cmd.OutOrStdout(), "Reopened release %s (tracker #%d)\n", releaseVersion, targetRelease.Number)

	return nil
}

// extractReleaseCodename extracts the codename from a release title
// e.g., "Release: v1.2.0 (Phoenix)" -> "Phoenix", "Release: v1.2.0" -> ""
func extractReleaseCodename(title string) string {
	start := strings.Index(title, "(")
	end := strings.Index(title, ")")
	if start > 0 && end > start {
		return title[start+1 : end]
	}
	return ""
}

// generateReleaseNotesContent generates the release notes content
func generateReleaseNotesContent(version, codename string, trackerNumber int, issues []api.Issue) string {
	var sb strings.Builder

	// Header with version and codename
	if codename != "" {
		sb.WriteString(fmt.Sprintf("# Release %s (%s)\n\n", version, codename))
	} else {
		sb.WriteString(fmt.Sprintf("# Release %s\n\n", version))
	}

	// Date
	sb.WriteString(fmt.Sprintf("**Date:** %s\n\n", time.Now().Format("2006-01-02")))

	// Tracker issue reference
	sb.WriteString(fmt.Sprintf("**Tracker:** #%d\n\n", trackerNumber))

	// Group issues by label
	enhancements := []api.Issue{}
	bugFixes := []api.Issue{}
	other := []api.Issue{}

	for _, issue := range issues {
		labeled := false
		for _, label := range issue.Labels {
			if label.Name == "enhancement" {
				enhancements = append(enhancements, issue)
				labeled = true
				break
			} else if label.Name == "bug" {
				bugFixes = append(bugFixes, issue)
				labeled = true
				break
			}
		}
		if !labeled {
			other = append(other, issue)
		}
	}

	if len(enhancements) > 0 {
		sb.WriteString("## Features\n\n")
		for _, issue := range enhancements {
			sb.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
		}
		sb.WriteString("\n")
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

// generateChangelogContent generates the changelog content
func generateChangelogContent(version string, issues []api.Issue) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", version, time.Now().Format("2006-01-02")))

	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("- #%d %s\n", issue.Number, issue.Title))
	}

	return sb.String()
}

// runReleaseListWithDeps is the testable entry point for release list
// It receives all dependencies as parameters for easy mocking in tests
func runReleaseListWithDeps(cmd *cobra.Command, opts *releaseListOptions, cfg *config.Config, client releaseClient) error {
	owner, repo, err := parseOwnerRepo(cfg)
	if err != nil {
		return err
	}

	// Get open and closed release issues
	openIssues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
	if err != nil {
		return fmt.Errorf("failed to get open releases: %w", err)
	}

	closedIssues, err := client.GetClosedIssuesByLabel(owner, repo, "release")
	if err != nil {
		return fmt.Errorf("failed to get closed releases: %w", err)
	}

	// Combine and filter for release trackers
	var releases []releaseInfo
	for _, issue := range openIssues {
		if strings.HasPrefix(issue.Title, "Release: ") {
			releases = append(releases, extractReleaseInfo(issue, "Active"))
		}
	}
	for _, issue := range closedIssues {
		if strings.HasPrefix(issue.Title, "Release: ") {
			releases = append(releases, extractReleaseInfo(issue, "Released"))
		}
	}

	if len(releases) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No releases found\n")
		return nil
	}

	// Sort by version descending (AC-022-2)
	sortReleasesByVersionDesc(releases)

	// Display table (AC-022-1)
	fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-15s %-10s %-10s\n", "VERSION", "CODENAME", "TRACKER", "STATUS")
	fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-15s %-10s %-10s\n", "-------", "--------", "-------", "------")
	for _, r := range releases {
		codenameDisplay := r.codename
		if codenameDisplay == "" {
			codenameDisplay = "-"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-15s #%-9d %-10s\n", r.version, codenameDisplay, r.trackerNum, r.status)
	}

	return nil
}

// releaseInfo holds parsed release information
type releaseInfo struct {
	version    string
	codename   string
	trackerNum int
	status     string
}

// extractReleaseInfo extracts release information from an issue
func extractReleaseInfo(issue api.Issue, status string) releaseInfo {
	version := extractReleaseVersion(issue.Title)
	codename := extractReleaseCodename(issue.Title)
	return releaseInfo{
		version:    version,
		codename:   codename,
		trackerNum: issue.Number,
		status:     status,
	}
}

// sortReleasesByVersionDesc sorts releases by version in descending order
func sortReleasesByVersionDesc(releases []releaseInfo) {
	// Simple bubble sort for version comparison
	for i := 0; i < len(releases)-1; i++ {
		for j := 0; j < len(releases)-i-1; j++ {
			if compareVersions(releases[j].version, releases[j+1].version) < 0 {
				releases[j], releases[j+1] = releases[j+1], releases[j]
			}
		}
	}
}

// compareVersions compares two semver versions
// Returns: positive if v1 > v2, negative if v1 < v2, zero if equal
func compareVersions(v1, v2 string) int {
	// Strip 'v' prefix
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < 3; i++ {
		var n1, n2 int
		if i < len(parts1) {
			_, _ = fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			_, _ = fmt.Sscanf(parts2[i], "%d", &n2)
		}
		if n1 != n2 {
			return n1 - n2
		}
	}
	return 0
}

// nextVersions contains calculated next version options
type nextVersions struct {
	patch string
	minor string
	major string
}

// calculateNextVersions computes the next patch, minor, and major versions
func calculateNextVersions(currentVersion string) (*nextVersions, error) {
	// Strip 'v' prefix for parsing
	version := strings.TrimPrefix(currentVersion, "v")
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", currentVersion)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return &nextVersions{
		patch: fmt.Sprintf("v%d.%d.%d", major, minor, patch+1),
		minor: fmt.Sprintf("v%d.%d.0", major, minor+1),
		major: fmt.Sprintf("v%d.0.0", major+1),
	}, nil
}

// releaseActiveEntry represents an active release for config storage
type releaseActiveEntry struct {
	Version      string `yaml:"version"`
	TrackerIssue int    `yaml:"tracker_issue"`
	Started      string `yaml:"started"`
	Track        string `yaml:"track"`
}

// parseReleaseTitle parses a release title into version and track
// Examples:
//
//	"Release: v1.2.0" -> version="1.2.0", track="stable"
//	"Release: v1.2.0 (Phoenix)" -> version="1.2.0", track="stable"
//	"Release: patch/1.1.1" -> version="1.1.1", track="patch"
//	"Release: beta/2.0.0" -> version="2.0.0", track="beta"
func parseReleaseTitle(title string) (version, track string) {
	// Remove "Release: " prefix
	remainder := strings.TrimPrefix(title, "Release: ")

	// Remove codename suffix if present (e.g., " (Phoenix)")
	if idx := strings.Index(remainder, " ("); idx != -1 {
		remainder = remainder[:idx]
	}

	// Check for track prefix (e.g., "patch/", "beta/")
	if strings.Contains(remainder, "/") {
		parts := strings.SplitN(remainder, "/", 2)
		track = parts[0]
		version = strings.TrimPrefix(parts[1], "v")
	} else {
		// Default track is "stable", version starts with v
		track = "stable"
		version = strings.TrimPrefix(remainder, "v")
	}

	return version, track
}

// SyncActiveReleases queries open release issues and returns active release entries
func SyncActiveReleases(client releaseClient, owner, repo string) ([]releaseActiveEntry, error) {
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
	if err != nil {
		return nil, fmt.Errorf("failed to get release issues: %w", err)
	}

	var entries []releaseActiveEntry
	for _, issue := range issues {
		if !strings.HasPrefix(issue.Title, "Release: ") {
			continue
		}

		version, track := parseReleaseTitle(issue.Title)
		entries = append(entries, releaseActiveEntry{
			Version:      version,
			TrackerIssue: issue.Number,
			Started:      "",
			Track:        track,
		})
	}

	return entries, nil
}
