package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/scooter-indie/gh-pmu/internal/api"
	"github.com/scooter-indie/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

type moveOptions struct {
	status   string
	priority string
}

func newMoveCommand() *cobra.Command {
	opts := &moveOptions{}

	cmd := &cobra.Command{
		Use:   "move <issue-number>",
		Short: "Update project fields for an issue",
		Long: `Update project field values for an issue.

Changes the status, priority, or other project fields for an issue
that is already in the configured project.

Field values are resolved through config aliases, so you can use
shorthand values like "in_progress" which will be mapped to "In Progress".`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(cmd, args, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.status, "status", "s", "", "Set project status field")
	cmd.Flags().StringVarP(&opts.priority, "priority", "p", "", "Set project priority field")

	return cmd
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

	// Parse issue reference
	owner, repo, number, err := parseIssueReference(args[0])
	if err != nil {
		return err
	}

	// If owner/repo not specified, use first repo from config
	if owner == "" || repo == "" {
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repository specified and none configured")
		}
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
		}
		owner = parts[0]
		repo = parts[1]
	}

	// Create API client
	client := api.NewClient()

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

	var itemID string
	for _, item := range items {
		if item.Issue != nil && item.Issue.Number == number {
			itemID = item.ID
			break
		}
	}

	if itemID == "" {
		return fmt.Errorf("issue #%d is not in the project", number)
	}

	// Track changes for confirmation message
	var changes []string

	// Update status if provided
	if opts.status != "" {
		statusValue := cfg.ResolveFieldValue("status", opts.status)
		if err := client.SetProjectItemField(project.ID, itemID, "Status", statusValue); err != nil {
			return fmt.Errorf("failed to set status: %w", err)
		}
		changes = append(changes, fmt.Sprintf("Status → %s", statusValue))
	}

	// Update priority if provided
	if opts.priority != "" {
		priorityValue := cfg.ResolveFieldValue("priority", opts.priority)
		if err := client.SetProjectItemField(project.ID, itemID, "Priority", priorityValue); err != nil {
			return fmt.Errorf("failed to set priority: %w", err)
		}
		changes = append(changes, fmt.Sprintf("Priority → %s", priorityValue))
	}

	// Output confirmation
	fmt.Printf("✓ Updated issue #%d: %s\n", issue.Number, issue.Title)
	for _, change := range changes {
		fmt.Printf("  • %s\n", change)
	}
	fmt.Printf("🔗 %s\n", issue.URL)

	return nil
}
