package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

type filterOptions struct {
	status   string
	priority string
	assignee string
	label    string
	json     bool
}

// FilterInput represents the expected JSON input format from gh issue list
type FilterInput struct {
	Number    int     `json:"number"`
	Title     string  `json:"title"`
	State     string  `json:"state"`
	URL       string  `json:"url"`
	Body      string  `json:"body"`
	Labels    []Label `json:"labels"`
	Assignees []User  `json:"assignees"`
}

// Label represents a label in the input JSON
type Label struct {
	Name string `json:"name"`
}

// User represents a user/assignee in the input JSON
type User struct {
	Login string `json:"login"`
}

func newFilterCommand() *cobra.Command {
	opts := &filterOptions{}

	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Filter piped issue JSON by project field values",
		Long: `Filter JSON input from 'gh issue list' by project field values.

This command reads issue JSON from stdin and filters based on project board
field values like status and priority. Issues are looked up in the configured
project to determine their field values.

Example:
  gh issue list --repo owner/repo --json number,title,state | gh pmu filter --status ready
  gh issue list -R owner/repo --json number,title --limit 100 | gh pmu filter --status in_progress --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFilter(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.status, "status", "s", "", "Filter by status (e.g., backlog, ready, in_progress)")
	cmd.Flags().StringVarP(&opts.priority, "priority", "p", "", "Filter by priority (e.g., p0, p1, p2)")
	cmd.Flags().StringVarP(&opts.assignee, "assignee", "a", "", "Filter by assignee login")
	cmd.Flags().StringVarP(&opts.label, "label", "l", "", "Filter by label name")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output in JSON format (default is table)")

	return cmd
}

func runFilter(cmd *cobra.Command, opts *filterOptions) error {
	// Check if stdin has data
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("no input provided - pipe issue JSON from 'gh issue list --json ...'")
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

	// Get project
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Read and parse JSON input from stdin
	var issues []FilterInput
	scanner := bufio.NewScanner(os.Stdin)
	var inputBuilder strings.Builder
	for scanner.Scan() {
		inputBuilder.WriteString(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	inputStr := strings.TrimSpace(inputBuilder.String())
	if inputStr == "" {
		return fmt.Errorf("empty input - pipe issue JSON from 'gh issue list --json ...'")
	}

	if err := json.Unmarshal([]byte(inputStr), &issues); err != nil {
		return fmt.Errorf("failed to parse JSON input: %w\nExpected JSON array from 'gh issue list --json ...'", err)
	}

	// Fetch all project items to get field values
	items, err := client.GetProjectItems(project.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to get project items: %w", err)
	}

	// Build a map of issue number -> project item for quick lookup
	itemsByNumber := make(map[int]api.ProjectItem)
	for _, item := range items {
		if item.Issue != nil {
			itemsByNumber[item.Issue.Number] = item
		}
	}

	// Filter issues based on project field values
	var filtered []FilterInput
	for _, issue := range issues {
		item, exists := itemsByNumber[issue.Number]
		if !exists {
			// Issue not in project, skip
			continue
		}

		// Apply status filter
		if opts.status != "" {
			targetStatus := cfg.ResolveFieldValue("status", opts.status)
			if !hasFieldValue(item, "Status", targetStatus) {
				continue
			}
		}

		// Apply priority filter
		if opts.priority != "" {
			targetPriority := cfg.ResolveFieldValue("priority", opts.priority)
			if !hasFieldValue(item, "Priority", targetPriority) {
				continue
			}
		}

		// Apply assignee filter (from input JSON, not project)
		if opts.assignee != "" {
			if !hasAssignee(issue, opts.assignee) {
				continue
			}
		}

		// Apply label filter (from input JSON, not project)
		if opts.label != "" {
			if !hasLabel(issue, opts.label) {
				continue
			}
		}

		filtered = append(filtered, issue)
	}

	// Output results
	if opts.json {
		return outputFilterJSON(filtered)
	}
	return outputFilterTable(cmd, filtered)
}

// hasFieldValue checks if a project item has a specific field value
func hasFieldValue(item api.ProjectItem, fieldName, value string) bool {
	for _, fv := range item.FieldValues {
		if strings.EqualFold(fv.Field, fieldName) && strings.EqualFold(fv.Value, value) {
			return true
		}
	}
	return false
}

// hasAssignee checks if an issue has a specific assignee
func hasAssignee(issue FilterInput, assignee string) bool {
	for _, a := range issue.Assignees {
		if strings.EqualFold(a.Login, assignee) {
			return true
		}
	}
	return false
}

// hasLabel checks if an issue has a specific label
func hasLabel(issue FilterInput, label string) bool {
	for _, l := range issue.Labels {
		if strings.EqualFold(l.Name, label) {
			return true
		}
	}
	return false
}

// outputFilterJSON outputs filtered issues as JSON
func outputFilterJSON(issues []FilterInput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(issues)
}

// outputFilterTable outputs filtered issues as a table
func outputFilterTable(cmd *cobra.Command, issues []FilterInput) error {
	if len(issues) == 0 {
		cmd.Println("No matching issues found")
		return nil
	}

	// Simple table output
	fmt.Println("NUMBER\tTITLE\tSTATE")
	for _, issue := range issues {
		title := issue.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		fmt.Printf("#%d\t%s\t%s\n", issue.Number, title, issue.State)
	}
	return nil
}
