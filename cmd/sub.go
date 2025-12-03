package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/scooter-indie/gh-pmu/internal/api"
	"github.com/scooter-indie/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

func newSubCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sub",
		Short: "Manage sub-issues",
		Long: `Manage sub-issue relationships between issues.

Sub-issues allow you to create parent-child hierarchies between issues,
useful for breaking down epics into smaller tasks.`,
	}

	cmd.AddCommand(newSubAddCommand())
	cmd.AddCommand(newSubCreateCommand())
	cmd.AddCommand(newSubListCommand())
	cmd.AddCommand(newSubRemoveCommand())

	return cmd
}

func newSubAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <parent-issue> <child-issue>",
		Short: "Link an issue as a sub-issue of another",
		Long: `Link an existing issue as a sub-issue of a parent issue.

Both issues must already exist. The child issue will appear as a
sub-issue under the parent issue in GitHub's UI.

Examples:
  gh pmu sub add 10 15        # Link issue #15 as sub-issue of #10
  gh pmu sub add #10 #15      # Same, with # prefix
  gh pmu sub add owner/repo#10 owner/repo#15  # Full references`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubAdd(cmd, args)
		},
	}

	return cmd
}

func runSubAdd(cmd *cobra.Command, args []string) error {
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

	// Parse parent issue reference
	parentOwner, parentRepo, parentNumber, err := parseIssueReference(args[0])
	if err != nil {
		return fmt.Errorf("invalid parent issue: %w", err)
	}

	// Parse child issue reference
	childOwner, childRepo, childNumber, err := parseIssueReference(args[1])
	if err != nil {
		return fmt.Errorf("invalid child issue: %w", err)
	}

	// Default to configured repo if not specified
	if parentOwner == "" || parentRepo == "" {
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repository specified and none configured")
		}
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
		}
		parentOwner = parts[0]
		parentRepo = parts[1]
	}

	if childOwner == "" || childRepo == "" {
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repository specified and none configured")
		}
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
		}
		childOwner = parts[0]
		childRepo = parts[1]
	}

	// Create API client
	client := api.NewClient()

	// Validate parent issue exists
	parentIssue, err := client.GetIssue(parentOwner, parentRepo, parentNumber)
	if err != nil {
		return fmt.Errorf("failed to get parent issue #%d: %w", parentNumber, err)
	}

	// Validate child issue exists
	childIssue, err := client.GetIssue(childOwner, childRepo, childNumber)
	if err != nil {
		return fmt.Errorf("failed to get child issue #%d: %w", childNumber, err)
	}

	// Add sub-issue link
	err = client.AddSubIssue(parentIssue.ID, childIssue.ID)
	if err != nil {
		// Check if already linked (GitHub returns "duplicate" or "only have one parent" messages)
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "only have one parent") {
			return fmt.Errorf("issue #%d is already a sub-issue (issues can only have one parent)", childNumber)
		}
		return fmt.Errorf("failed to add sub-issue link: %w", err)
	}

	// Output confirmation
	fmt.Printf("✓ Linked issue #%d as sub-issue of #%d\n", childNumber, parentNumber)
	fmt.Printf("  Parent: %s\n", parentIssue.Title)
	fmt.Printf("  Child:  %s\n", childIssue.Title)

	return nil
}

type subCreateOptions struct {
	parent         string
	title          string
	body           string
	inheritLabels  bool
	inheritAssign  bool
	inheritMilestone bool
}

func newSubCreateCommand() *cobra.Command {
	opts := &subCreateOptions{
		inheritLabels:    true,
		inheritAssign:    false,
		inheritMilestone: true,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue as a sub-issue",
		Long: `Create a new issue and automatically link it as a sub-issue of a parent.

By default, the new issue inherits labels and milestone from the parent.
Use flags to control inheritance behavior.

Examples:
  gh pmu sub create --parent 10 --title "Implement feature X"
  gh pmu sub create --parent #10 --title "Task" --body "Description"
  gh pmu sub create -p 10 -t "Task" --no-inherit-labels`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubCreate(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.parent, "parent", "p", "", "Parent issue number (required)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Issue title (required)")
	cmd.Flags().StringVarP(&opts.body, "body", "b", "", "Issue body")
	cmd.Flags().BoolVar(&opts.inheritLabels, "inherit-labels", true, "Inherit labels from parent")
	cmd.Flags().BoolVar(&opts.inheritAssign, "inherit-assignees", false, "Inherit assignees from parent")
	cmd.Flags().BoolVar(&opts.inheritMilestone, "inherit-milestone", true, "Inherit milestone from parent")

	cmd.MarkFlagRequired("parent")
	cmd.MarkFlagRequired("title")

	return cmd
}

func runSubCreate(cmd *cobra.Command, opts *subCreateOptions) error {
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

	// Parse parent issue reference
	parentOwner, parentRepo, parentNumber, err := parseIssueReference(opts.parent)
	if err != nil {
		return fmt.Errorf("invalid parent issue: %w", err)
	}

	// Default to configured repo if not specified
	if parentOwner == "" || parentRepo == "" {
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repository specified and none configured")
		}
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
		}
		parentOwner = parts[0]
		parentRepo = parts[1]
	}

	// Create API client
	client := api.NewClient()

	// Get parent issue to validate and optionally inherit from
	parentIssue, err := client.GetIssue(parentOwner, parentRepo, parentNumber)
	if err != nil {
		return fmt.Errorf("failed to get parent issue #%d: %w", parentNumber, err)
	}

	// Build labels list
	var labels []string
	if opts.inheritLabels && len(parentIssue.Labels) > 0 {
		for _, l := range parentIssue.Labels {
			labels = append(labels, l.Name)
		}
	}

	// Create the new issue
	newIssue, err := client.CreateIssue(parentOwner, parentRepo, opts.title, opts.body, labels)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	// Link as sub-issue
	err = client.AddSubIssue(parentIssue.ID, newIssue.ID)
	if err != nil {
		// Issue was created but linking failed - inform user
		fmt.Fprintf(os.Stderr, "Warning: Issue created but failed to link as sub-issue: %v\n", err)
		fmt.Printf("Created issue #%d: %s\n", newIssue.Number, newIssue.Title)
		fmt.Printf("%s\n", newIssue.URL)
		return nil
	}

	// Output confirmation
	fmt.Printf("✓ Created sub-issue #%d under parent #%d\n", newIssue.Number, parentNumber)
	fmt.Printf("  Title:  %s\n", newIssue.Title)
	fmt.Printf("  Parent: %s\n", parentIssue.Title)
	if len(labels) > 0 {
		fmt.Printf("  Labels: %s (inherited)\n", strings.Join(labels, ", "))
	}
	fmt.Printf("🔗 %s\n", newIssue.URL)

	return nil
}

type subListOptions struct {
	json bool
}

func newSubListCommand() *cobra.Command {
	opts := &subListOptions{}

	cmd := &cobra.Command{
		Use:   "list <parent-issue>",
		Short: "List sub-issues of a parent issue",
		Long: `List all sub-issues of a parent issue.

Displays the title, state, and assignee for each sub-issue,
along with a completion count.

Examples:
  gh pmu sub list 10        # List sub-issues of issue #10
  gh pmu sub list #10       # Same, with # prefix
  gh pmu sub list 10 --json # Output as JSON`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubList(cmd, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.json, "json", false, "Output in JSON format")

	return cmd
}

func runSubList(cmd *cobra.Command, args []string, opts *subListOptions) error {
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

	// Parse parent issue reference
	parentOwner, parentRepo, parentNumber, err := parseIssueReference(args[0])
	if err != nil {
		return fmt.Errorf("invalid parent issue: %w", err)
	}

	// Default to configured repo if not specified
	if parentOwner == "" || parentRepo == "" {
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repository specified and none configured")
		}
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
		}
		parentOwner = parts[0]
		parentRepo = parts[1]
	}

	// Create API client
	client := api.NewClient()

	// Get parent issue to validate it exists
	parentIssue, err := client.GetIssue(parentOwner, parentRepo, parentNumber)
	if err != nil {
		return fmt.Errorf("failed to get parent issue #%d: %w", parentNumber, err)
	}

	// Get sub-issues
	subIssues, err := client.GetSubIssues(parentOwner, parentRepo, parentNumber)
	if err != nil {
		return fmt.Errorf("failed to get sub-issues: %w", err)
	}

	// Output
	if opts.json {
		return outputSubListJSON(subIssues, parentIssue)
	}

	return outputSubListTable(subIssues, parentIssue)
}

// SubListJSONOutput represents the JSON output for sub list command
type SubListJSONOutput struct {
	Parent    SubListParent    `json:"parent"`
	SubIssues []SubListItem    `json:"subIssues"`
	Summary   SubListSummary   `json:"summary"`
}

type SubListParent struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type SubListItem struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

type SubListSummary struct {
	Total    int `json:"total"`
	Open     int `json:"open"`
	Closed   int `json:"closed"`
}

func outputSubListJSON(subIssues []api.SubIssue, parent *api.Issue) error {
	output := SubListJSONOutput{
		Parent: SubListParent{
			Number: parent.Number,
			Title:  parent.Title,
		},
		SubIssues: make([]SubListItem, 0, len(subIssues)),
		Summary: SubListSummary{
			Total: len(subIssues),
		},
	}

	for _, sub := range subIssues {
		output.SubIssues = append(output.SubIssues, SubListItem{
			Number: sub.Number,
			Title:  sub.Title,
			State:  sub.State,
			URL:    sub.URL,
		})

		if sub.State == "CLOSED" {
			output.Summary.Closed++
		} else {
			output.Summary.Open++
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputSubListTable(subIssues []api.SubIssue, parent *api.Issue) error {
	fmt.Printf("Sub-issues of #%d: %s\n\n", parent.Number, parent.Title)

	if len(subIssues) == 0 {
		fmt.Println("No sub-issues found.")
		return nil
	}

	closedCount := 0
	for _, sub := range subIssues {
		state := "[ ]"
		if sub.State == "CLOSED" {
			state = "[x]"
			closedCount++
		}
		fmt.Printf("  %s #%d - %s\n", state, sub.Number, sub.Title)
	}

	fmt.Printf("\nProgress: %d/%d complete\n", closedCount, len(subIssues))

	return nil
}

func newSubRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <parent-issue> <child-issue>",
		Short: "Remove a sub-issue link from a parent issue",
		Long: `Remove the sub-issue relationship between a parent and child issue.

This does NOT delete the child issue, only removes the parent-child link.
The child issue will become a standalone issue again.

Examples:
  gh pmu sub remove 10 15        # Unlink issue #15 from parent #10
  gh pmu sub remove #10 #15      # Same, with # prefix
  gh pmu sub remove owner/repo#10 owner/repo#15  # Full references`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubRemove(cmd, args)
		},
	}

	return cmd
}

func runSubRemove(cmd *cobra.Command, args []string) error {
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

	// Parse parent issue reference
	parentOwner, parentRepo, parentNumber, err := parseIssueReference(args[0])
	if err != nil {
		return fmt.Errorf("invalid parent issue: %w", err)
	}

	// Parse child issue reference
	childOwner, childRepo, childNumber, err := parseIssueReference(args[1])
	if err != nil {
		return fmt.Errorf("invalid child issue: %w", err)
	}

	// Default to configured repo if not specified
	if parentOwner == "" || parentRepo == "" {
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repository specified and none configured")
		}
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
		}
		parentOwner = parts[0]
		parentRepo = parts[1]
	}

	if childOwner == "" || childRepo == "" {
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repository specified and none configured")
		}
		parts := strings.Split(cfg.Repositories[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format in config: %s", cfg.Repositories[0])
		}
		childOwner = parts[0]
		childRepo = parts[1]
	}

	// Create API client
	client := api.NewClient()

	// Validate parent issue exists
	parentIssue, err := client.GetIssue(parentOwner, parentRepo, parentNumber)
	if err != nil {
		return fmt.Errorf("failed to get parent issue #%d: %w", parentNumber, err)
	}

	// Validate child issue exists
	childIssue, err := client.GetIssue(childOwner, childRepo, childNumber)
	if err != nil {
		return fmt.Errorf("failed to get child issue #%d: %w", childNumber, err)
	}

	// Remove sub-issue link
	err = client.RemoveSubIssue(parentIssue.ID, childIssue.ID)
	if err != nil {
		// Check if not linked
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "not a sub-issue") || strings.Contains(errMsg, "not found") {
			return fmt.Errorf("issue #%d is not a sub-issue of #%d", childNumber, parentNumber)
		}
		return fmt.Errorf("failed to remove sub-issue link: %w", err)
	}

	// Output confirmation
	fmt.Printf("✓ Removed sub-issue link: #%d is no longer a sub-issue of #%d\n", childNumber, parentNumber)
	fmt.Printf("  Former parent: %s\n", parentIssue.Title)
	fmt.Printf("  Unlinked:      %s\n", childIssue.Title)

	return nil
}
