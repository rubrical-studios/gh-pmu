package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

type listOptions struct {
	status       string
	priority     string
	assignee     string
	label        string
	search       string
	limit        int
	hasSubIssues bool
	json         bool
	web          bool
	repo         string
}

func newListCommand() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues from the configured project",
		Long: `List issues from the configured GitHub project with their field values.

By default, displays Title, Status, Priority, and Assignees for each issue.
Use filters to narrow down the results.`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.status, "status", "s", "", "Filter by status (e.g., backlog, in_progress, done)")
	cmd.Flags().StringVarP(&opts.priority, "priority", "p", "", "Filter by priority (e.g., p0, p1, p2)")
	cmd.Flags().StringVarP(&opts.assignee, "assignee", "a", "", "Filter by assignee login")
	cmd.Flags().StringVarP(&opts.label, "label", "l", "", "Filter by label name")
	cmd.Flags().StringVarP(&opts.search, "search", "q", "", "Search in issue title and body")
	cmd.Flags().IntVarP(&opts.limit, "limit", "n", 0, "Limit number of results (0 for no limit)")
	cmd.Flags().BoolVar(&opts.hasSubIssues, "has-sub-issues", false, "Filter to only show parent issues (issues with sub-issues)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output in JSON format")
	cmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open project board in browser")
	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Filter by repository (owner/repo format)")

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	// Load configuration from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, err := config.LoadFromDirectory(cwd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w\nRun 'gh pmu init' to create a configuration file", err)
	}

	// Validate config
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

	// Handle --web flag: open project in browser
	if opts.web {
		return openInBrowser(project.URL)
	}

	// Build filter
	var filter *api.ProjectItemsFilter

	// Determine repository filter (--repo flag takes precedence over config)
	repoFilter := ""
	if opts.repo != "" {
		// Validate repo format
		parts := strings.Split(opts.repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid --repo format: expected owner/repo, got %s", opts.repo)
		}
		repoFilter = opts.repo
	} else if len(cfg.Repositories) > 0 {
		repoFilter = cfg.Repositories[0]
	}

	if repoFilter != "" {
		filter = &api.ProjectItemsFilter{
			Repository: repoFilter,
		}
	}

	// Fetch project items
	items, err := client.GetProjectItems(project.ID, filter)
	if err != nil {
		return fmt.Errorf("failed to get project items: %w", err)
	}

	// Apply status filter
	if opts.status != "" {
		targetStatus := cfg.ResolveFieldValue("status", opts.status)
		items = filterByFieldValue(items, "Status", targetStatus)
	}

	// Apply priority filter
	if opts.priority != "" {
		targetPriority := cfg.ResolveFieldValue("priority", opts.priority)
		items = filterByFieldValue(items, "Priority", targetPriority)
	}

	// Apply assignee filter
	if opts.assignee != "" {
		items = filterByAssignee(items, opts.assignee)
	}

	// Apply label filter
	if opts.label != "" {
		items = filterByLabel(items, opts.label)
	}

	// Apply search filter
	if opts.search != "" {
		items = filterBySearch(items, opts.search)
	}

	// Apply has-sub-issues filter
	if opts.hasSubIssues {
		items = filterByHasSubIssues(client, items)
	}

	// Apply limit
	if opts.limit > 0 && len(items) > opts.limit {
		items = items[:opts.limit]
	}

	// Output
	if opts.json {
		return outputJSON(cmd, items)
	}

	return outputTable(cmd, items)
}

// filterByFieldValue filters items by a specific field value
func filterByFieldValue(items []api.ProjectItem, fieldName, value string) []api.ProjectItem {
	var filtered []api.ProjectItem
	for _, item := range items {
		for _, fv := range item.FieldValues {
			if strings.EqualFold(fv.Field, fieldName) && strings.EqualFold(fv.Value, value) {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return filtered
}

// filterByHasSubIssues filters items to only those with sub-issues
func filterByHasSubIssues(client *api.Client, items []api.ProjectItem) []api.ProjectItem {
	var filtered []api.ProjectItem
	for _, item := range items {
		if item.Issue == nil {
			continue
		}

		// Check if issue has sub-issues
		subIssues, err := client.GetSubIssues(
			item.Issue.Repository.Owner,
			item.Issue.Repository.Name,
			item.Issue.Number,
		)
		if err != nil {
			// Skip issues where we can't fetch sub-issues
			continue
		}

		if len(subIssues) > 0 {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// getFieldValue gets a field value from an item
func getFieldValue(item api.ProjectItem, fieldName string) string {
	for _, fv := range item.FieldValues {
		if strings.EqualFold(fv.Field, fieldName) {
			return fv.Value
		}
	}
	return ""
}

// outputTable outputs items in a table format
func outputTable(cmd *cobra.Command, items []api.ProjectItem) error {
	if len(items) == 0 {
		cmd.Println("No issues found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NUMBER\tTITLE\tSTATUS\tPRIORITY\tASSIGNEES")

	for _, item := range items {
		if item.Issue == nil {
			continue
		}

		// Get field values
		status := getFieldValue(item, "Status")
		priority := getFieldValue(item, "Priority")

		// Format assignees
		var assignees []string
		for _, a := range item.Issue.Assignees {
			assignees = append(assignees, a.Login)
		}
		assigneeStr := strings.Join(assignees, ", ")
		if assigneeStr == "" {
			assigneeStr = "-"
		}

		// Truncate title if too long
		title := item.Issue.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		fmt.Fprintf(w, "#%d\t%s\t%s\t%s\t%s\n",
			item.Issue.Number,
			title,
			status,
			priority,
			assigneeStr,
		)
	}

	w.Flush()
	return nil
}

// JSONOutput represents the JSON output structure
type JSONOutput struct {
	Items []JSONItem `json:"items"`
}

// JSONItem represents an item in JSON output
type JSONItem struct {
	Number      int               `json:"number"`
	Title       string            `json:"title"`
	State       string            `json:"state"`
	URL         string            `json:"url"`
	Repository  string            `json:"repository"`
	Assignees   []string          `json:"assignees"`
	FieldValues map[string]string `json:"fieldValues"`
}

// outputJSON outputs items in JSON format
func outputJSON(cmd *cobra.Command, items []api.ProjectItem) error {
	output := JSONOutput{
		Items: make([]JSONItem, 0, len(items)),
	}

	for _, item := range items {
		if item.Issue == nil {
			continue
		}

		jsonItem := JSONItem{
			Number:      item.Issue.Number,
			Title:       item.Issue.Title,
			State:       item.Issue.State,
			URL:         item.Issue.URL,
			Repository:  fmt.Sprintf("%s/%s", item.Issue.Repository.Owner, item.Issue.Repository.Name),
			Assignees:   make([]string, 0),
			FieldValues: make(map[string]string),
		}

		for _, a := range item.Issue.Assignees {
			jsonItem.Assignees = append(jsonItem.Assignees, a.Login)
		}

		for _, fv := range item.FieldValues {
			jsonItem.FieldValues[fv.Field] = fv.Value
		}

		output.Items = append(output.Items, jsonItem)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// filterByAssignee filters items by assignee login
func filterByAssignee(items []api.ProjectItem, assignee string) []api.ProjectItem {
	var filtered []api.ProjectItem
	for _, item := range items {
		if item.Issue == nil {
			continue
		}
		for _, a := range item.Issue.Assignees {
			if strings.EqualFold(a.Login, assignee) {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return filtered
}

// filterByLabel filters items by label name
func filterByLabel(items []api.ProjectItem, label string) []api.ProjectItem {
	var filtered []api.ProjectItem
	for _, item := range items {
		if item.Issue == nil {
			continue
		}
		for _, l := range item.Issue.Labels {
			if strings.EqualFold(l.Name, label) {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return filtered
}

// filterBySearch filters items by searching in title and body
func filterBySearch(items []api.ProjectItem, search string) []api.ProjectItem {
	var filtered []api.ProjectItem
	searchLower := strings.ToLower(search)
	for _, item := range items {
		if item.Issue == nil {
			continue
		}
		titleLower := strings.ToLower(item.Issue.Title)
		bodyLower := strings.ToLower(item.Issue.Body)
		if strings.Contains(titleLower, searchLower) || strings.Contains(bodyLower, searchLower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// openInBrowser opens the given URL in the default browser
func openInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
