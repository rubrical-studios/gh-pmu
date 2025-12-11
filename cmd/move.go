package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

type moveOptions struct {
	status    string
	priority  string
	recursive bool
	depth     int
	dryRun    bool
	yes       bool   // skip confirmation
	repo      string // repository override (owner/repo format)
}

// moveClient defines the interface for API methods used by move functions.
// This allows for easier testing with mock implementations.
type moveClient interface {
	GetIssue(owner, repo string, number int) (*api.Issue, error)
	GetProject(owner string, number int) (*api.Project, error)
	GetProjectItems(projectID string, filter *api.ProjectItemsFilter) ([]api.ProjectItem, error)
	GetSubIssues(owner, repo string, number int) ([]api.SubIssue, error)
	SetProjectItemField(projectID, itemID, fieldName, value string) error
}

func newMoveCommand() *cobra.Command {
	opts := &moveOptions{
		depth: 10, // default max depth
	}

	cmd := &cobra.Command{
		Use:   "move <issue-number>",
		Short: "Update project fields for an issue",
		Long: `Update project field values for an issue.

Changes the status, priority, or other project fields for an issue
that is already in the configured project.

Field values are resolved through config aliases, so you can use
shorthand values like "in_progress" which will be mapped to "In Progress".

Use --recursive to update all sub-issues as well. This will traverse
the issue tree and apply the same changes to all descendants.

Examples:
  # Move a single issue to "In Progress"
  gh pmu move 42 --status in_progress

  # Set both status and priority
  gh pmu move 42 --status done --priority p1

  # Recursively update an epic and all its sub-issues
  gh pmu move 10 --status in_progress --recursive

  # Preview recursive changes without applying (dry-run)
  gh pmu move 10 --status done --recursive --dry-run

  # Recursively update, skip confirmation prompt
  gh pmu move 10 --status backlog --recursive --yes

  # Limit recursion depth (default is 10)
  gh pmu move 10 --status in_progress --recursive --depth 2

  # Specify repository explicitly
  gh pmu move 42 --status done --repo owner/repo`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(cmd, args, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.status, "status", "s", "", "Set project status field")
	cmd.Flags().StringVarP(&opts.priority, "priority", "p", "", "Set project priority field")
	cmd.Flags().BoolVarP(&opts.recursive, "recursive", "r", false, "Apply changes to all sub-issues recursively")
	cmd.Flags().IntVar(&opts.depth, "depth", 10, "Maximum depth for recursive operations")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVarP(&opts.yes, "yes", "y", false, "Skip confirmation prompt for recursive operations")
	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository for the issue (owner/repo format)")

	return cmd
}

// issueInfo holds information about an issue to be updated
type issueInfo struct {
	Owner  string
	Repo   string
	Number int
	Title  string
	ItemID string
	Depth  int
}

func runMove(cmd *cobra.Command, args []string, opts *moveOptions) error {
	// Validate at least one flag is provided
	if opts.status == "" && opts.priority == "" {
		return fmt.Errorf("at least one of --status or --priority is required")
	}

	// Load configuration
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

	// Create API client
	client := api.NewClient()

	return runMoveWithDeps(cmd, args, opts, cfg, client)
}

// runMoveWithDeps is the testable implementation of runMove
func runMoveWithDeps(cmd *cobra.Command, args []string, opts *moveOptions, cfg *config.Config, client moveClient) error {
	// Parse issue reference
	owner, repo, number, err := parseIssueReference(args[0])
	if err != nil {
		return err
	}

	// Determine default repository (--repo flag takes precedence over config)
	defaultOwner, defaultRepo := "", ""
	if opts.repo != "" {
		parts := strings.Split(opts.repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid --repo format: expected owner/repo, got %s", opts.repo)
		}
		defaultOwner, defaultRepo = parts[0], parts[1]
	} else if len(cfg.Repositories) > 0 {
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) == 2 {
			defaultOwner, defaultRepo = parts[0], parts[1]
		}
	}

	// If owner/repo not specified in issue reference, use default
	if owner == "" || repo == "" {
		if defaultOwner == "" || defaultRepo == "" {
			return fmt.Errorf("no repository specified and none configured (use --repo or configure in .gh-pmu.yml)")
		}
		owner = defaultOwner
		repo = defaultRepo
	}

	// Get issue to verify it exists
	issue, err := client.GetIssue(owner, repo, number)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Get project
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Find the project item ID for this issue
	items, err := client.GetProjectItems(project.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to get project items: %w", err)
	}

	// Build a map of issue numbers to item IDs for quick lookup
	itemIDMap := make(map[string]string) // "owner/repo#number" -> itemID
	for _, item := range items {
		if item.Issue != nil {
			key := fmt.Sprintf("%s/%s#%d", item.Issue.Repository.Owner, item.Issue.Repository.Name, item.Issue.Number)
			itemIDMap[key] = item.ID
		}
	}

	rootKey := fmt.Sprintf("%s/%s#%d", owner, repo, number)
	rootItemID, inProject := itemIDMap[rootKey]
	if !inProject {
		return fmt.Errorf("issue #%d is not in the project", number)
	}

	// Collect all issues to update
	issuesToUpdate := []issueInfo{{
		Owner:  owner,
		Repo:   repo,
		Number: number,
		Title:  issue.Title,
		ItemID: rootItemID,
		Depth:  0,
	}}

	// If recursive, collect all sub-issues
	if opts.recursive {
		subIssues, err := collectSubIssuesRecursive(client, owner, repo, number, itemIDMap, 1, opts.depth)
		if err != nil {
			return fmt.Errorf("failed to collect sub-issues: %w", err)
		}
		issuesToUpdate = append(issuesToUpdate, subIssues...)
	}

	// Resolve field values
	statusValue := ""
	priorityValue := ""
	var changeDescriptions []string

	if opts.status != "" {
		statusValue = cfg.ResolveFieldValue("status", opts.status)
		changeDescriptions = append(changeDescriptions, fmt.Sprintf("Status → %s", statusValue))
	}
	if opts.priority != "" {
		priorityValue = cfg.ResolveFieldValue("priority", opts.priority)
		changeDescriptions = append(changeDescriptions, fmt.Sprintf("Priority → %s", priorityValue))
	}

	// Show what will be updated
	if opts.recursive || opts.dryRun {
		if opts.dryRun {
			fmt.Println("Dry run - no changes will be made")
			fmt.Println()
		}

		fmt.Printf("Issues to update (%d):\n", len(issuesToUpdate))
		for _, info := range issuesToUpdate {
			indent := strings.Repeat("  ", info.Depth)
			if info.ItemID != "" {
				fmt.Printf("%s• #%d - %s\n", indent, info.Number, info.Title)
			} else {
				fmt.Printf("%s• #%d - %s (not in project, will skip)\n", indent, info.Number, info.Title)
			}
		}

		fmt.Println("\nChanges to apply:")
		for _, desc := range changeDescriptions {
			fmt.Printf("  • %s\n", desc)
		}

		if opts.dryRun {
			return nil
		}

		// Prompt for confirmation unless --yes is provided
		if !opts.yes {
			fmt.Printf("\nProceed with updating %d issues? [y/N]: ", len(issuesToUpdate))
			var response string
			_, _ = fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}
		fmt.Println()
	}

	// Apply updates
	updatedCount := 0
	skippedCount := 0

	for _, info := range issuesToUpdate {
		if info.ItemID == "" {
			skippedCount++
			continue
		}

		// Update status if provided
		if statusValue != "" {
			if err := client.SetProjectItemField(project.ID, info.ItemID, "Status", statusValue); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to set status for #%d: %v\n", info.Number, err)
				continue
			}
		}

		// Update priority if provided
		if priorityValue != "" {
			if err := client.SetProjectItemField(project.ID, info.ItemID, "Priority", priorityValue); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to set priority for #%d: %v\n", info.Number, err)
				continue
			}
		}

		updatedCount++
		if !opts.recursive {
			// Single issue - show detailed output
			fmt.Printf("✓ Updated issue #%d: %s\n", info.Number, info.Title)
			for _, desc := range changeDescriptions {
				fmt.Printf("  • %s\n", desc)
			}
			fmt.Printf("🔗 https://github.com/%s/%s/issues/%d\n", info.Owner, info.Repo, info.Number)
		}
	}

	// Summary for recursive operations
	if opts.recursive {
		fmt.Printf("✓ Updated %d issues", updatedCount)
		if skippedCount > 0 {
			fmt.Printf(" (%d skipped - not in project)", skippedCount)
		}
		fmt.Println()
	}

	return nil
}

// collectSubIssuesRecursive recursively collects all sub-issues up to maxDepth
func collectSubIssuesRecursive(client moveClient, owner, repo string, number int, itemIDMap map[string]string, currentDepth, maxDepth int) ([]issueInfo, error) {
	if currentDepth > maxDepth {
		return nil, nil
	}

	subIssues, err := client.GetSubIssues(owner, repo, number)
	if err != nil {
		return nil, err
	}

	var result []issueInfo
	for _, sub := range subIssues {
		// Determine the repo for this sub-issue
		subOwner := sub.Repository.Owner
		subRepo := sub.Repository.Name
		if subOwner == "" {
			subOwner = owner
		}
		if subRepo == "" {
			subRepo = repo
		}

		key := fmt.Sprintf("%s/%s#%d", subOwner, subRepo, sub.Number)
		itemID := itemIDMap[key] // may be empty if not in project

		info := issueInfo{
			Owner:  subOwner,
			Repo:   subRepo,
			Number: sub.Number,
			Title:  sub.Title,
			ItemID: itemID,
			Depth:  currentDepth,
		}
		result = append(result, info)

		// Recurse into this sub-issue's children
		children, err := collectSubIssuesRecursive(client, subOwner, subRepo, sub.Number, itemIDMap, currentDepth+1, maxDepth)
		if err != nil {
			// Log warning but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to get sub-issues for #%d: %v\n", sub.Number, err)
			continue
		}
		result = append(result, children...)
	}

	return result, nil
}
