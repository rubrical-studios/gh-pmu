package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/rubrical-studios/gh-pmu/internal/ui"
	"github.com/spf13/cobra"
)

// listClient defines the interface for API methods used by list functions.
// This allows for easier testing with mock implementations.
type listClient interface {
	GetProject(owner string, number int) (*api.Project, error)
	GetProjectItems(projectID string, filter *api.ProjectItemsFilter) ([]api.ProjectItem, error)
	GetOpenIssuesByLabel(owner, repo, label string) ([]api.Issue, error)
	GetSubIssueCounts(owner, repo string, numbers []int) (map[int]int, error)
	SearchRepositoryIssues(owner, repo string, filters api.SearchFilters, limit int) ([]api.Issue, error)
	GetProjectFieldsForIssues(projectID string, issueIDs []string) (map[string][]api.FieldValue, error)
}

type listOptions struct {
	status       string
	priority     string
	assignee     string
	label        string
	search       string
	release      string
	noRelease    bool
	microsprint  string
	state        string
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
	cmd.Flags().StringVarP(&opts.release, "release", "r", "", "Filter by release (e.g., v1.0.0, current)")
	cmd.Flags().BoolVar(&opts.noRelease, "no-release", false, "Filter to issues without a release assignment")
	cmd.Flags().StringVarP(&opts.microsprint, "microsprint", "m", "", "Filter by microsprint (e.g., 2025-12-20-a, current)")
	cmd.Flags().StringVar(&opts.state, "state", "", "Filter by issue state (open or closed)")
	cmd.MarkFlagsMutuallyExclusive("release", "no-release")
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

	return runListWithDeps(cmd, opts, cfg, client)
}

// runListWithDeps is the testable implementation of runList
func runListWithDeps(cmd *cobra.Command, opts *listOptions, cfg *config.Config, client listClient) error {
	// Validate state filter
	if opts.state != "" {
		stateLower := strings.ToLower(opts.state)
		if stateLower != "open" && stateLower != "closed" && stateLower != "all" {
			return fmt.Errorf("invalid --state value: expected 'open', 'closed', or 'all', got %q", opts.state)
		}
	}

	// Get project
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Handle --web flag: open project in browser
	if opts.web {
		return ui.OpenInBrowser(project.URL)
	}

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

	// Determine query strategy based on state filter
	// "all" state uses GetProjectItems (full fetch with optional limit)
	// "open", "closed", or default uses SearchRepositoryIssues (optimized)
	stateLower := strings.ToLower(opts.state)
	useSearchAPI := repoFilter != "" && stateLower != "all"

	var items []api.ProjectItem

	if useSearchAPI {
		// Open-first strategy: Use GitHub Search API for filtered queries
		repoParts := strings.Split(repoFilter, "/")
		searchFilters := api.SearchFilters{
			State:    stateLower, // defaults to "open" if empty
			Assignee: opts.assignee,
			Search:   opts.search,
		}
		if opts.label != "" {
			searchFilters.Labels = []string{opts.label}
		}

		issues, err := client.SearchRepositoryIssues(repoParts[0], repoParts[1], searchFilters, opts.limit)
		if err != nil {
			return fmt.Errorf("failed to search issues: %w", err)
		}

		// Convert issues to project items and enrich with project field values
		items, err = enrichIssuesWithProjectFields(client, project.ID, issues)
		if err != nil {
			return fmt.Errorf("failed to enrich issues with project fields: %w", err)
		}
	} else {
		// Full fetch strategy: Use GetProjectItems for --state all or no repo filter
		var filter *api.ProjectItemsFilter
		if repoFilter != "" || opts.limit > 0 {
			filter = &api.ProjectItemsFilter{
				Repository: repoFilter,
				Limit:      opts.limit,
			}
		}

		items, err = client.GetProjectItems(project.ID, filter)
		if err != nil {
			return fmt.Errorf("failed to get project items: %w", err)
		}
	}

	// Apply status filter (always client-side since it's a project field)
	if opts.status != "" {
		targetStatus := cfg.ResolveFieldValue("status", opts.status)
		items = filterByFieldValue(items, "Status", targetStatus)
	}

	// Apply priority filter (always client-side since it's a project field)
	if opts.priority != "" {
		targetPriority := cfg.ResolveFieldValue("priority", opts.priority)
		items = filterByFieldValue(items, "Priority", targetPriority)
	}

	// Apply assignee filter (skip if search API was used - already filtered server-side)
	if opts.assignee != "" && !useSearchAPI {
		items = filterByAssignee(items, opts.assignee)
	}

	// Apply label filter (skip if search API was used - already filtered server-side)
	if opts.label != "" && !useSearchAPI {
		items = filterByLabel(items, opts.label)
	}

	// Apply search filter (skip if search API was used - already filtered server-side)
	if opts.search != "" && !useSearchAPI {
		items = filterBySearch(items, opts.search)
	}

	// Apply release filter
	if opts.release != "" {
		targetRelease := opts.release
		if opts.release == "current" && repoFilter != "" {
			// Resolve "current" to active release
			parts := strings.Split(repoFilter, "/")
			if len(parts) == 2 {
				releaseIssues, err := client.GetOpenIssuesByLabel(parts[0], parts[1], "branch")
				if err == nil {
					for _, issue := range releaseIssues {
						if strings.HasPrefix(issue.Title, "Release: ") {
							targetRelease = strings.TrimPrefix(issue.Title, "Release: ")
							if idx := strings.Index(targetRelease, " ("); idx > 0 {
								targetRelease = targetRelease[:idx]
							}
							break
						}
					}
				}
			}
		}
		items = filterByFieldValue(items, "Release", targetRelease)
	}

	// Apply no-release filter (issues without release assignment)
	if opts.noRelease {
		items = filterByEmptyField(items, "Release")
	}

	// Apply microsprint filter
	if opts.microsprint != "" {
		targetMicrosprint := opts.microsprint
		if opts.microsprint == "current" && repoFilter != "" {
			// Resolve "current" to active microsprint
			parts := strings.Split(repoFilter, "/")
			if len(parts) == 2 {
				microsprintIssues, err := client.GetOpenIssuesByLabel(parts[0], parts[1], "microsprint")
				if err == nil {
					for _, issue := range microsprintIssues {
						if strings.HasPrefix(issue.Title, "Microsprint: ") {
							targetMicrosprint = strings.TrimPrefix(issue.Title, "Microsprint: ")
							break
						}
					}
				}
			}
		}
		items = filterByFieldValue(items, "Microsprint", targetMicrosprint)
	}

	// Apply state filter (skip if search API was used - already filtered server-side)
	// For "all" state with GetProjectItems, no filtering needed
	if opts.state != "" && !useSearchAPI && stateLower != "all" {
		items = filterByState(items, opts.state)
	}

	// Apply has-sub-issues filter
	var hasSubIssuesFilterFailed bool
	if opts.hasSubIssues {
		items, hasSubIssuesFilterFailed = filterByHasSubIssues(cmd, client, items)
	}

	// Apply limit
	if opts.limit > 0 && len(items) > opts.limit {
		items = items[:opts.limit]
	}

	// Output
	var outputErr error
	if opts.json {
		outputErr = outputJSON(cmd, items)
	} else {
		outputErr = outputTable(cmd, items)
	}

	// Print note if sub-issue filter had failures
	if hasSubIssuesFilterFailed && outputErr == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "\nNote: some items included without sub-issue verification")
	}

	return outputErr
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

// filterByEmptyField filters items where a specific field is empty or not set
func filterByEmptyField(items []api.ProjectItem, fieldName string) []api.ProjectItem {
	var filtered []api.ProjectItem
	for _, item := range items {
		hasValue := false
		for _, fv := range item.FieldValues {
			if strings.EqualFold(fv.Field, fieldName) && fv.Value != "" {
				hasValue = true
				break
			}
		}
		if !hasValue {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// filterByHasSubIssues filters items to only those with sub-issues.
// Uses a batch query to fetch sub-issue counts efficiently.
// Returns the filtered items and a boolean indicating if any repo queries failed.
func filterByHasSubIssues(cmd *cobra.Command, client listClient, items []api.ProjectItem) ([]api.ProjectItem, bool) {
	if len(items) == 0 {
		return nil, false
	}

	// Group items by repository for batch queries
	type repoKey struct {
		owner string
		repo  string
	}
	repoItems := make(map[repoKey][]api.ProjectItem)
	repoNumbers := make(map[repoKey][]int)

	for _, item := range items {
		if item.Issue == nil {
			continue
		}
		key := repoKey{
			owner: item.Issue.Repository.Owner,
			repo:  item.Issue.Repository.Name,
		}
		repoItems[key] = append(repoItems[key], item)
		repoNumbers[key] = append(repoNumbers[key], item.Issue.Number)
	}

	// Fetch sub-issue counts for each repository (one query per repo)
	subIssueCounts := make(map[repoKey]map[int]int)
	var failedRepos []repoKey
	for key, numbers := range repoNumbers {
		counts, err := client.GetSubIssueCounts(key.owner, key.repo, numbers)
		if err != nil {
			// Warn about the failure but don't exclude items
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not fetch sub-issue counts for %s/%s: %v\n", key.owner, key.repo, err)
			failedRepos = append(failedRepos, key)
			continue
		}
		subIssueCounts[key] = counts
	}

	// Filter items based on sub-issue counts
	var filtered []api.ProjectItem
	for key, keyItems := range repoItems {
		counts, ok := subIssueCounts[key]
		if !ok {
			// Include items from failed repos (prefer false positives over silent exclusion)
			filtered = append(filtered, keyItems...)
			continue
		}
		for _, item := range keyItems {
			if count, exists := counts[item.Issue.Number]; exists && count > 0 {
				filtered = append(filtered, item)
			}
		}
	}

	return filtered, len(failedRepos) > 0
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

// filterByState filters items by GitHub issue state (open or closed)
func filterByState(items []api.ProjectItem, state string) []api.ProjectItem {
	var filtered []api.ProjectItem
	stateLower := strings.ToLower(state)
	for _, item := range items {
		if item.Issue == nil {
			continue
		}
		// Issue.State is OPEN or CLOSED (uppercase from GitHub API)
		if strings.EqualFold(item.Issue.State, stateLower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// enrichIssuesWithProjectFields converts issues to ProjectItems and enriches them with project field values.
// This is used when fetching via SearchRepositoryIssues to add project-specific fields (Status, Priority, etc.).
func enrichIssuesWithProjectFields(client listClient, projectID string, issues []api.Issue) ([]api.ProjectItem, error) {
	if len(issues) == 0 {
		return nil, nil
	}

	// Collect issue IDs for batch enrichment
	issueIDs := make([]string, 0, len(issues))
	issueByID := make(map[string]*api.Issue)
	for i := range issues {
		issueIDs = append(issueIDs, issues[i].ID)
		issueByID[issues[i].ID] = &issues[i]
	}

	// Fetch project field values for all issues
	fieldValuesByID, err := client.GetProjectFieldsForIssues(projectID, issueIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get project fields: %w", err)
	}

	// Convert to ProjectItems
	items := make([]api.ProjectItem, 0, len(issues))
	for _, issue := range issues {
		item := api.ProjectItem{
			ID:    issue.ID,
			Issue: &issue,
		}

		// Add field values if available
		if fieldValues, ok := fieldValuesByID[issue.ID]; ok {
			item.FieldValues = fieldValues
		}

		items = append(items, item)
	}

	return items, nil
}
