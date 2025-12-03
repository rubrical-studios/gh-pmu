package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/scooter-indie/gh-pmu/internal/api"
	"github.com/scooter-indie/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

type createOptions struct {
	title    string
	body     string
	status   string
	priority string
	labels   []string
}

func newCreateCommand() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an issue with project metadata",
		Long: `Create a new issue and add it to the configured project.

When --title is provided, creates the issue non-interactively.
Otherwise, opens an editor for composing the issue.

The issue is automatically added to the configured project and
any specified field values (status, priority) are set.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Issue title (required for non-interactive mode)")
	cmd.Flags().StringVarP(&opts.body, "body", "b", "", "Issue body")
	cmd.Flags().StringVarP(&opts.status, "status", "s", "", "Set project status field (e.g., backlog, in_progress)")
	cmd.Flags().StringVarP(&opts.priority, "priority", "p", "", "Set project priority field (e.g., p0, p1, p2)")
	cmd.Flags().StringArrayVarP(&opts.labels, "label", "l", nil, "Add labels (can be specified multiple times)")

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
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

	// Get repository
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repository configured")
	}

	repoParts := strings.Split(cfg.Repositories[0], "/")
	if len(repoParts) != 2 {
		return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
	}
	owner, repo := repoParts[0], repoParts[1]

	// Handle interactive vs non-interactive mode
	title := opts.title
	body := opts.body

	if title == "" {
		// Interactive mode - open editor
		// For now, require --title flag
		return fmt.Errorf("--title is required (interactive mode not yet implemented)")
	}

	// Merge labels: config defaults + command line
	labels := append([]string{}, cfg.Defaults.Labels...)
	labels = append(labels, opts.labels...)

	// Create API client
	client := api.NewClient()

	// Create the issue
	issue, err := client.CreateIssue(owner, repo, title, body, labels)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	// Add issue to project
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	itemID, err := client.AddIssueToProject(project.ID, issue.ID)
	if err != nil {
		return fmt.Errorf("failed to add issue to project: %w", err)
	}

	// Set project field values
	if opts.status != "" {
		statusValue := cfg.ResolveFieldValue("status", opts.status)
		if err := client.SetProjectItemField(project.ID, itemID, "Status", statusValue); err != nil {
			// Non-fatal - warn but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to set status: %v\n", err)
		}
	} else if cfg.Defaults.Status != "" {
		// Apply default status from config
		statusValue := cfg.ResolveFieldValue("status", cfg.Defaults.Status)
		if err := client.SetProjectItemField(project.ID, itemID, "Status", statusValue); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set default status: %v\n", err)
		}
	}

	if opts.priority != "" {
		priorityValue := cfg.ResolveFieldValue("priority", opts.priority)
		if err := client.SetProjectItemField(project.ID, itemID, "Priority", priorityValue); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set priority: %v\n", err)
		}
	} else if cfg.Defaults.Priority != "" {
		// Apply default priority from config
		priorityValue := cfg.ResolveFieldValue("priority", cfg.Defaults.Priority)
		if err := client.SetProjectItemField(project.ID, itemID, "Priority", priorityValue); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set default priority: %v\n", err)
		}
	}

	// Output the result
	fmt.Printf("Created issue #%d: %s\n", issue.Number, issue.Title)
	fmt.Printf("%s\n", issue.URL)

	return nil
}
