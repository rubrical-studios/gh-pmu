package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type createOptions struct {
	title       string
	body        string
	status      string
	priority    string
	microsprint string
	labels      []string
	assignees   []string
	milestone   string
	repo        string
	fromFile    string
	interactive bool
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
	cmd.Flags().StringVarP(&opts.microsprint, "microsprint", "M", "", "Set microsprint field (use 'current' for active microsprint)")
	cmd.Flags().StringArrayVarP(&opts.labels, "label", "l", nil, "Add labels (can be specified multiple times)")
	cmd.Flags().StringArrayVarP(&opts.assignees, "assignee", "a", nil, "Assign users (can be specified multiple times)")
	cmd.Flags().StringVarP(&opts.milestone, "milestone", "m", "", "Set milestone (title or number)")
	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Target repository (owner/repo format)")
	cmd.Flags().StringVarP(&opts.fromFile, "from-file", "f", "", "Create issue from YAML/JSON file")
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "Use interactive mode with prompts")

	return cmd
}

// issueFromFile represents an issue definition in a YAML/JSON file
type issueFromFile struct {
	Title     string   `json:"title" yaml:"title"`
	Body      string   `json:"body" yaml:"body"`
	Labels    []string `json:"labels" yaml:"labels"`
	Assignees []string `json:"assignees" yaml:"assignees"`
	Milestone string   `json:"milestone" yaml:"milestone"`
	Status    string   `json:"status" yaml:"status"`
	Priority  string   `json:"priority" yaml:"priority"`
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

	// Determine repository
	var owner, repo string
	if opts.repo != "" {
		// Use --repo flag
		repoParts := strings.Split(opts.repo, "/")
		if len(repoParts) != 2 {
			return fmt.Errorf("invalid --repo format: expected owner/repo, got %s", opts.repo)
		}
		owner, repo = repoParts[0], repoParts[1]
	} else {
		// Use config
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repository configured")
		}
		repoParts := strings.Split(cfg.Repositories[0], "/")
		if len(repoParts) != 2 {
			return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
		}
		owner, repo = repoParts[0], repoParts[1]
	}

	// Handle --from-file
	if opts.fromFile != "" {
		return runCreateFromFile(cmd, opts, cfg, owner, repo)
	}

	// Handle interactive mode
	if opts.interactive {
		return fmt.Errorf("interactive mode not yet implemented")
	}

	// Handle non-interactive mode
	title := opts.title
	body := opts.body

	if title == "" {
		return fmt.Errorf("--title is required (use --interactive for prompted mode)")
	}

	// Merge labels: config defaults + command line
	labels := append([]string{}, cfg.Defaults.Labels...)
	labels = append(labels, opts.labels...)

	// Create API client
	client := api.NewClient()

	// Create the issue with extended options
	issue, err := client.CreateIssueWithOptions(owner, repo, title, body, labels, opts.assignees, opts.milestone)
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

	// Set microsprint field
	if opts.microsprint != "" {
		microsprintValue := opts.microsprint
		if opts.microsprint == "current" {
			// Resolve "current" to active microsprint name
			microsprintIssues, err := client.GetOpenIssuesByLabel(owner, repo, "microsprint")
			if err != nil {
				return fmt.Errorf("failed to get microsprint issues: %w", err)
			}
			activeTracker := findActiveMicrosprintForCreate(microsprintIssues)
			if activeTracker == nil {
				return fmt.Errorf("no active microsprint found. Run 'gh pmu microsprint start' to create one")
			}
			microsprintValue = strings.TrimPrefix(activeTracker.Title, "Microsprint: ")
		}
		if err := client.SetProjectItemField(project.ID, itemID, "Microsprint", microsprintValue); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set microsprint: %v\n", err)
		}
	}

	// Output the result
	fmt.Printf("Created issue #%d: %s\n", issue.Number, issue.Title)
	fmt.Printf("%s\n", issue.URL)

	return nil
}

func runCreateFromFile(cmd *cobra.Command, opts *createOptions, cfg *config.Config, owner, repo string) error {
	// Read the file
	data, err := os.ReadFile(opts.fromFile)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", opts.fromFile, err)
	}

	// Parse the file (try YAML first, then JSON)
	var issueData issueFromFile
	if strings.HasSuffix(opts.fromFile, ".json") {
		if err := json.Unmarshal(data, &issueData); err != nil {
			return fmt.Errorf("failed to parse JSON file: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, &issueData); err != nil {
			return fmt.Errorf("failed to parse YAML file: %w", err)
		}
	}

	if issueData.Title == "" {
		return fmt.Errorf("title is required in file")
	}

	// Merge with command line options (command line takes precedence)
	title := issueData.Title
	body := issueData.Body
	if opts.body != "" {
		body = opts.body
	}

	labels := append([]string{}, cfg.Defaults.Labels...)
	labels = append(labels, issueData.Labels...)
	labels = append(labels, opts.labels...)

	assignees := append([]string{}, issueData.Assignees...)
	assignees = append(assignees, opts.assignees...)

	milestone := issueData.Milestone
	if opts.milestone != "" {
		milestone = opts.milestone
	}

	status := issueData.Status
	if opts.status != "" {
		status = opts.status
	}

	priority := issueData.Priority
	if opts.priority != "" {
		priority = opts.priority
	}

	// Create API client
	client := api.NewClient()

	// Create the issue
	issue, err := client.CreateIssueWithOptions(owner, repo, title, body, labels, assignees, milestone)
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
	if status != "" {
		statusValue := cfg.ResolveFieldValue("status", status)
		if err := client.SetProjectItemField(project.ID, itemID, "Status", statusValue); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set status: %v\n", err)
		}
	} else if cfg.Defaults.Status != "" {
		statusValue := cfg.ResolveFieldValue("status", cfg.Defaults.Status)
		if err := client.SetProjectItemField(project.ID, itemID, "Status", statusValue); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set default status: %v\n", err)
		}
	}

	if priority != "" {
		priorityValue := cfg.ResolveFieldValue("priority", priority)
		if err := client.SetProjectItemField(project.ID, itemID, "Priority", priorityValue); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set priority: %v\n", err)
		}
	} else if cfg.Defaults.Priority != "" {
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

// findActiveMicrosprintForCreate finds today's active microsprint tracker from a list of issues
// Returns nil if no active microsprint is found for today
func findActiveMicrosprintForCreate(issues []api.Issue) *api.Issue {
	today := time.Now().Format("2006-01-02")
	prefix := "Microsprint: " + today + "-"

	for i := range issues {
		if strings.HasPrefix(issues[i].Title, prefix) {
			return &issues[i]
		}
	}
	return nil
}
