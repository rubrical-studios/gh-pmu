package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/scooter-indie/gh-pmu/internal/api"
	"github.com/scooter-indie/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

type viewOptions struct {
	json bool
}

func newViewCommand() *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <issue-number>",
		Short: "View an issue with project metadata",
		Long: `View an issue with all its project field values.

Displays issue details including title, body, state, labels, assignees,
and all project-specific fields like Status and Priority.

Also shows sub-issues if any exist, and parent issue if this is a sub-issue.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.json, "json", false, "Output in JSON format")

	return cmd
}

func runView(cmd *cobra.Command, args []string, opts *viewOptions) error {
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

	// Fetch issue
	issue, err := client.GetIssue(owner, repo, number)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Fetch project items to get field values for this issue
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	items, err := client.GetProjectItems(project.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to get project items: %w", err)
	}

	// Find this issue in project items to get field values
	var fieldValues []api.FieldValue
	for _, item := range items {
		if item.Issue != nil && item.Issue.Number == number {
			fieldValues = item.FieldValues
			break
		}
	}

	// Fetch sub-issues (if any)
	subIssues, err := client.GetSubIssues(owner, repo, number)
	if err != nil {
		// Non-fatal - issue might not have sub-issues or API might not support it
		subIssues = nil
	}

	// Fetch parent issue (if this is a sub-issue)
	parentIssue, err := client.GetParentIssue(owner, repo, number)
	if err != nil {
		// Non-fatal - issue might not be a sub-issue
		parentIssue = nil
	}

	// Output
	if opts.json {
		return outputViewJSON(cmd, issue, fieldValues, subIssues, parentIssue)
	}

	return outputViewTable(cmd, issue, fieldValues, subIssues, parentIssue)
}

// ViewJSONOutput represents the JSON output for view command
type ViewJSONOutput struct {
	Number       int                `json:"number"`
	Title        string             `json:"title"`
	State        string             `json:"state"`
	Body         string             `json:"body"`
	URL          string             `json:"url"`
	Author       string             `json:"author"`
	Assignees    []string           `json:"assignees"`
	Labels       []string           `json:"labels"`
	Milestone    string             `json:"milestone,omitempty"`
	FieldValues  map[string]string  `json:"fieldValues"`
	SubIssues    []SubIssueJSON     `json:"subIssues,omitempty"`
	SubProgress  *SubProgressJSON   `json:"subProgress,omitempty"`
	ParentIssue  *ParentIssueJSON   `json:"parentIssue,omitempty"`
}

// SubProgressJSON represents sub-issue progress in JSON output
type SubProgressJSON struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	Percentage int `json:"percentage"`
}

// SubIssueJSON represents a sub-issue in JSON output
type SubIssueJSON struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

// ParentIssueJSON represents the parent issue in JSON output
type ParentIssueJSON struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

func outputViewJSON(cmd *cobra.Command, issue *api.Issue, fieldValues []api.FieldValue, subIssues []api.SubIssue, parentIssue *api.Issue) error {
	output := ViewJSONOutput{
		Number:      issue.Number,
		Title:       issue.Title,
		State:       issue.State,
		Body:        issue.Body,
		URL:         issue.URL,
		Author:      issue.Author.Login,
		Assignees:   make([]string, 0),
		Labels:      make([]string, 0),
		FieldValues: make(map[string]string),
	}

	for _, a := range issue.Assignees {
		output.Assignees = append(output.Assignees, a.Login)
	}

	for _, l := range issue.Labels {
		output.Labels = append(output.Labels, l.Name)
	}

	if issue.Milestone != nil {
		output.Milestone = issue.Milestone.Title
	}

	for _, fv := range fieldValues {
		output.FieldValues[fv.Field] = fv.Value
	}

	if len(subIssues) > 0 {
		output.SubIssues = make([]SubIssueJSON, 0, len(subIssues))
		closedCount := 0
		for _, sub := range subIssues {
			output.SubIssues = append(output.SubIssues, SubIssueJSON{
				Number: sub.Number,
				Title:  sub.Title,
				State:  sub.State,
				URL:    sub.URL,
			})
			if sub.State == "CLOSED" {
				closedCount++
			}
		}

		// Add progress info
		total := len(subIssues)
		percentage := 0
		if total > 0 {
			percentage = (closedCount * 100) / total
		}
		output.SubProgress = &SubProgressJSON{
			Total:      total,
			Completed:  closedCount,
			Percentage: percentage,
		}
	}

	if parentIssue != nil {
		output.ParentIssue = &ParentIssueJSON{
			Number: parentIssue.Number,
			Title:  parentIssue.Title,
			URL:    parentIssue.URL,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputViewTable(cmd *cobra.Command, issue *api.Issue, fieldValues []api.FieldValue, subIssues []api.SubIssue, parentIssue *api.Issue) error {
	// Title and state
	fmt.Printf("%s #%d\n", issue.Title, issue.Number)
	fmt.Printf("State: %s\n", issue.State)
	fmt.Printf("URL: %s\n", issue.URL)
	fmt.Println()

	// Author
	fmt.Printf("Author: @%s\n", issue.Author.Login)

	// Assignees
	if len(issue.Assignees) > 0 {
		var assignees []string
		for _, a := range issue.Assignees {
			assignees = append(assignees, "@"+a.Login)
		}
		fmt.Printf("Assignees: %s\n", strings.Join(assignees, ", "))
	}

	// Labels
	if len(issue.Labels) > 0 {
		var labels []string
		for _, l := range issue.Labels {
			labels = append(labels, l.Name)
		}
		fmt.Printf("Labels: %s\n", strings.Join(labels, ", "))
	}

	// Milestone
	if issue.Milestone != nil {
		fmt.Printf("Milestone: %s\n", issue.Milestone.Title)
	}

	// Project field values
	if len(fieldValues) > 0 {
		fmt.Println()
		fmt.Println("Project Fields:")
		for _, fv := range fieldValues {
			fmt.Printf("  %s: %s\n", fv.Field, fv.Value)
		}
	}

	// Parent issue
	if parentIssue != nil {
		fmt.Println()
		fmt.Printf("Parent Issue: #%d - %s\n", parentIssue.Number, parentIssue.Title)
	}

	// Sub-issues with progress bar
	if len(subIssues) > 0 {
		fmt.Println()
		fmt.Println("Sub-Issues:")
		closedCount := 0
		for _, sub := range subIssues {
			state := "[ ]"
			if sub.State == "CLOSED" {
				state = "[x]"
				closedCount++
			}
			// Show repo info if cross-repo
			if sub.Repository.Owner != "" && sub.Repository.Name != "" {
				parentRepo := issue.Repository.Owner + "/" + issue.Repository.Name
				subRepo := sub.Repository.Owner + "/" + sub.Repository.Name
				if subRepo != parentRepo {
					fmt.Printf("  %s %s#%d - %s\n", state, subRepo, sub.Number, sub.Title)
					continue
				}
			}
			fmt.Printf("  %s #%d - %s\n", state, sub.Number, sub.Title)
		}

		// Progress bar and percentage
		total := len(subIssues)
		percentage := 0
		if total > 0 {
			percentage = (closedCount * 100) / total
		}
		progressBar := renderProgressBar(closedCount, total, 20)
		fmt.Printf("\n%s %d of %d sub-issues complete (%d%%)\n", progressBar, closedCount, total, percentage)
	}

	// Body
	if issue.Body != "" {
		fmt.Println()
		fmt.Println("---")
		fmt.Println(issue.Body)
	}

	return nil
}

// renderProgressBar creates a visual progress bar
// Example: [████████░░░░░░░░░░░░] for 40% complete
func renderProgressBar(completed, total, width int) string {
	if total == 0 {
		return "[" + strings.Repeat("░", width) + "]"
	}

	filled := (completed * width) / total
	if filled > width {
		filled = width
	}

	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

// parseIssueNumber parses a string into an issue number
// Accepts formats: "123" or "#123"
func parseIssueNumber(s string) (int, error) {
	// Strip leading # if present
	s = strings.TrimPrefix(s, "#")

	num, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid issue number: %s", s)
	}

	if num <= 0 {
		return 0, fmt.Errorf("issue number must be positive: %d", num)
	}

	return num, nil
}

// parseIssueReference parses an issue reference string
// Accepts formats: "123", "#123", or "owner/repo#123"
// Returns owner, repo, number (owner/repo may be empty if not specified)
func parseIssueReference(s string) (owner, repo string, number int, err error) {
	// Check for owner/repo#number format
	if idx := strings.Index(s, "#"); idx > 0 {
		// Has # with something before it - could be owner/repo#number
		repoRef := s[:idx]
		numStr := s[idx+1:]

		if slashIdx := strings.Index(repoRef, "/"); slashIdx > 0 {
			owner = repoRef[:slashIdx]
			repo = repoRef[slashIdx+1:]

			number, err = parseIssueNumber(numStr)
			if err != nil {
				return "", "", 0, err
			}
			return owner, repo, number, nil
		}
	}

	// Try parsing as simple number or #number
	number, err = parseIssueNumber(s)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid issue reference: %s", s)
	}

	return "", "", number, nil
}
