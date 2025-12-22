package api

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strings"

	graphql "github.com/cli/shurcooL-graphql"
)

// safeGraphQLInt safely converts an int to graphql.Int with bounds checking.
// Returns an error if the value exceeds int32 range.
func safeGraphQLInt(n int) (graphql.Int, error) {
	if n > math.MaxInt32 || n < math.MinInt32 {
		return 0, fmt.Errorf("integer value %d exceeds int32 range", n)
	}
	return graphql.Int(n), nil
}

// GetProject fetches a project by owner and number
func (c *Client) GetProject(owner string, number int) (*Project, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// First try as user project
	project, err := c.getUserProject(owner, number)
	if err == nil {
		return project, nil
	}

	// If that fails, try as organization project
	project, err = c.getOrgProject(owner, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s/%d: %w", owner, number, err)
	}

	return project, nil
}

func (c *Client) getUserProject(owner string, number int) (*Project, error) {
	var query struct {
		User struct {
			ProjectV2 struct {
				ID     string
				Number int
				Title  string
				URL    string `graphql:"url"`
				Closed bool
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}

	gqlNumber, err := safeGraphQLInt(number)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"number": gqlNumber,
	}

	err = c.gql.Query("GetUserProject", &query, variables)
	if err != nil {
		return nil, err
	}

	return &Project{
		ID:     query.User.ProjectV2.ID,
		Number: query.User.ProjectV2.Number,
		Title:  query.User.ProjectV2.Title,
		URL:    query.User.ProjectV2.URL,
		Closed: query.User.ProjectV2.Closed,
		Owner: ProjectOwner{
			Type:  "User",
			Login: owner,
		},
	}, nil
}

func (c *Client) getOrgProject(owner string, number int) (*Project, error) {
	var query struct {
		Organization struct {
			ProjectV2 struct {
				ID     string
				Number int
				Title  string
				URL    string `graphql:"url"`
				Closed bool
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}

	gqlNumber, err := safeGraphQLInt(number)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"number": gqlNumber,
	}

	err = c.gql.Query("GetOrgProject", &query, variables)
	if err != nil {
		return nil, err
	}

	return &Project{
		ID:     query.Organization.ProjectV2.ID,
		Number: query.Organization.ProjectV2.Number,
		Title:  query.Organization.ProjectV2.Title,
		URL:    query.Organization.ProjectV2.URL,
		Closed: query.Organization.ProjectV2.Closed,
		Owner: ProjectOwner{
			Type:  "Organization",
			Login: owner,
		},
	}, nil
}

// GetProjectFields fetches all fields for a project
func (c *Client) GetProjectFields(projectID string) ([]ProjectField, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Node struct {
			ProjectV2 struct {
				Fields struct {
					Nodes []struct {
						TypeName string `graphql:"__typename"`
						// Common fields
						ProjectV2Field struct {
							ID       string
							Name     string
							DataType string
						} `graphql:"... on ProjectV2Field"`
						// Single select fields have options
						ProjectV2SingleSelectField struct {
							ID       string
							Name     string
							DataType string
							Options  []struct {
								ID   string
								Name string
							}
						} `graphql:"... on ProjectV2SingleSelectField"`
					}
				} `graphql:"fields(first: 50)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
	}

	err := c.gql.Query("GetProjectFields", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get project fields: %w", err)
	}

	var fields []ProjectField
	for _, node := range query.Node.ProjectV2.Fields.Nodes {
		field := ProjectField{}

		switch node.TypeName {
		case "ProjectV2SingleSelectField":
			field.ID = node.ProjectV2SingleSelectField.ID
			field.Name = node.ProjectV2SingleSelectField.Name
			field.DataType = node.ProjectV2SingleSelectField.DataType
			for _, opt := range node.ProjectV2SingleSelectField.Options {
				field.Options = append(field.Options, FieldOption{
					ID:   opt.ID,
					Name: opt.Name,
				})
			}
		case "ProjectV2Field":
			field.ID = node.ProjectV2Field.ID
			field.Name = node.ProjectV2Field.Name
			field.DataType = node.ProjectV2Field.DataType
		default:
			// Skip iteration/other field types for now
			continue
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// GetIssue fetches an issue by repository and number
func (c *Client) GetIssue(owner, repo string, number int) (*Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Repository struct {
			Issue struct {
				ID     string
				Number int
				Title  string
				Body   string
				State  string
				URL    string `graphql:"url"`
				Author struct {
					Login string
				}
				Assignees struct {
					Nodes []struct {
						Login string
					}
				} `graphql:"assignees(first: 10)"`
				Labels struct {
					Nodes []struct {
						Name  string
						Color string
					}
				} `graphql:"labels(first: 20)"`
				Milestone struct {
					Title string
				}
			} `graphql:"issue(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	gqlNumber, err := safeGraphQLInt(number)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"repo":   graphql.String(repo),
		"number": gqlNumber,
	}

	err = c.gql.Query("GetIssue", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %s/%s#%d: %w", owner, repo, number, err)
	}

	issue := &Issue{
		ID:     query.Repository.Issue.ID,
		Number: query.Repository.Issue.Number,
		Title:  query.Repository.Issue.Title,
		Body:   query.Repository.Issue.Body,
		State:  query.Repository.Issue.State,
		URL:    query.Repository.Issue.URL,
		Repository: Repository{
			Owner: owner,
			Name:  repo,
		},
		Author: Actor{Login: query.Repository.Issue.Author.Login},
	}

	for _, a := range query.Repository.Issue.Assignees.Nodes {
		issue.Assignees = append(issue.Assignees, Actor{Login: a.Login})
	}

	for _, l := range query.Repository.Issue.Labels.Nodes {
		issue.Labels = append(issue.Labels, Label{Name: l.Name, Color: l.Color})
	}

	if query.Repository.Issue.Milestone.Title != "" {
		issue.Milestone = &Milestone{Title: query.Repository.Issue.Milestone.Title}
	}

	return issue, nil
}

// GetIssueWithProjectFields fetches an issue and its project field values in a single query.
// This is more efficient than calling GetIssue + GetProjectItems when you only need one issue.
func (c *Client) GetIssueWithProjectFields(owner, repo string, number int) (*Issue, []FieldValue, error) {
	if c.gql == nil {
		return nil, nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Repository struct {
			Issue struct {
				ID     string
				Number int
				Title  string
				Body   string
				State  string
				URL    string `graphql:"url"`
				Author struct {
					Login string
				}
				Assignees struct {
					Nodes []struct {
						Login string
					}
				} `graphql:"assignees(first: 10)"`
				Labels struct {
					Nodes []struct {
						Name  string
						Color string
					}
				} `graphql:"labels(first: 20)"`
				Milestone struct {
					Title string
				}
				ProjectItems struct {
					Nodes []struct {
						FieldValues struct {
							Nodes []struct {
								TypeName string `graphql:"__typename"`
								// Single select field value
								ProjectV2ItemFieldSingleSelectValue struct {
									Name  string
									Field struct {
										ProjectV2SingleSelectField struct {
											Name string
										} `graphql:"... on ProjectV2SingleSelectField"`
									}
								} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
								// Text field value
								ProjectV2ItemFieldTextValue struct {
									Text  string
									Field struct {
										ProjectV2Field struct {
											Name string
										} `graphql:"... on ProjectV2Field"`
									}
								} `graphql:"... on ProjectV2ItemFieldTextValue"`
							}
						} `graphql:"fieldValues(first: 20)"`
					}
				} `graphql:"projectItems(first: 10)"`
			} `graphql:"issue(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	gqlNumber, err := safeGraphQLInt(number)
	if err != nil {
		return nil, nil, err
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"repo":   graphql.String(repo),
		"number": gqlNumber,
	}

	err = c.gql.Query("GetIssueWithProjectFields", &query, variables)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get issue %s/%s#%d: %w", owner, repo, number, err)
	}

	issue := &Issue{
		ID:     query.Repository.Issue.ID,
		Number: query.Repository.Issue.Number,
		Title:  query.Repository.Issue.Title,
		Body:   query.Repository.Issue.Body,
		State:  query.Repository.Issue.State,
		URL:    query.Repository.Issue.URL,
		Repository: Repository{
			Owner: owner,
			Name:  repo,
		},
		Author: Actor{Login: query.Repository.Issue.Author.Login},
	}

	for _, a := range query.Repository.Issue.Assignees.Nodes {
		issue.Assignees = append(issue.Assignees, Actor{Login: a.Login})
	}

	for _, l := range query.Repository.Issue.Labels.Nodes {
		issue.Labels = append(issue.Labels, Label{Name: l.Name, Color: l.Color})
	}

	if query.Repository.Issue.Milestone.Title != "" {
		issue.Milestone = &Milestone{Title: query.Repository.Issue.Milestone.Title}
	}

	// Extract field values from project items
	var fieldValues []FieldValue
	for _, projectItem := range query.Repository.Issue.ProjectItems.Nodes {
		for _, fv := range projectItem.FieldValues.Nodes {
			switch fv.TypeName {
			case "ProjectV2ItemFieldSingleSelectValue":
				if fv.ProjectV2ItemFieldSingleSelectValue.Name != "" {
					fieldValues = append(fieldValues, FieldValue{
						Field: fv.ProjectV2ItemFieldSingleSelectValue.Field.ProjectV2SingleSelectField.Name,
						Value: fv.ProjectV2ItemFieldSingleSelectValue.Name,
					})
				}
			case "ProjectV2ItemFieldTextValue":
				if fv.ProjectV2ItemFieldTextValue.Text != "" {
					fieldValues = append(fieldValues, FieldValue{
						Field: fv.ProjectV2ItemFieldTextValue.Field.ProjectV2Field.Name,
						Value: fv.ProjectV2ItemFieldTextValue.Text,
					})
				}
			}
		}
	}

	return issue, fieldValues, nil
}

// GetProjectItemIDForIssue looks up the project item ID for a specific issue in a project.
// This is more efficient than fetching all project items when you only need one.
func (c *Client) GetProjectItemIDForIssue(projectID, owner, repo string, number int) (string, error) {
	if c.gql == nil {
		return "", fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Repository struct {
			Issue struct {
				ProjectItems struct {
					Nodes []struct {
						ID      string
						Project struct {
							ID string
						}
					}
				} `graphql:"projectItems(first: 20)"`
			} `graphql:"issue(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	gqlNumber, err := safeGraphQLInt(number)
	if err != nil {
		return "", err
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"repo":   graphql.String(repo),
		"number": gqlNumber,
	}

	err = c.gql.Query("GetProjectItemIDForIssue", &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get project item for issue %s/%s#%d: %w", owner, repo, number, err)
	}

	// Find the project item in the target project
	for _, item := range query.Repository.Issue.ProjectItems.Nodes {
		if item.Project.ID == projectID {
			return item.ID, nil
		}
	}

	return "", fmt.Errorf("issue #%d is not in the project", number)
}

// ProjectItemsFilter allows filtering project items
type ProjectItemsFilter struct {
	Repository string // Filter by repository (owner/repo format)
}

// GetProjectItems fetches all items from a project with their field values.
// Uses cursor-based pagination to retrieve all items regardless of project size.
func (c *Client) GetProjectItems(projectID string, filter *ProjectItemsFilter) ([]ProjectItem, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var allItems []ProjectItem
	var cursor *string

	for {
		items, pageInfo, err := c.getProjectItemsPage(projectID, cursor)
		if err != nil {
			return nil, err
		}

		// Filter and process items from this page
		for _, item := range items {
			// Apply repository filter if specified
			if filter != nil && filter.Repository != "" {
				if item.Issue != nil && item.Issue.Repository.Owner != "" {
					repoName := item.Issue.Repository.Owner + "/" + item.Issue.Repository.Name
					if repoName != filter.Repository {
						continue
					}
				}
			}
			allItems = append(allItems, item)
		}

		// Check if there are more pages
		if !pageInfo.HasNextPage {
			break
		}
		cursor = &pageInfo.EndCursor
	}

	return allItems, nil
}

// pageInfo holds pagination information from GraphQL responses
type pageInfo struct {
	HasNextPage bool
	EndCursor   string
}

// getProjectItemsPage fetches a single page of project items
func (c *Client) getProjectItemsPage(projectID string, cursor *string) ([]ProjectItem, pageInfo, error) {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Items struct {
					Nodes []struct {
						ID      string
						Content struct {
							TypeName string `graphql:"__typename"`
							Issue    struct {
								ID         string
								Number     int
								Title      string
								Body       string
								State      string
								URL        string `graphql:"url"`
								Repository struct {
									NameWithOwner string
								}
								Assignees struct {
									Nodes []struct {
										Login string
									}
								} `graphql:"assignees(first: 10)"`
								Labels struct {
									Nodes []struct {
										Name string
									}
								} `graphql:"labels(first: 20)"`
							} `graphql:"... on Issue"`
						}
						FieldValues struct {
							Nodes []struct {
								TypeName string `graphql:"__typename"`
								// Single select field value
								ProjectV2ItemFieldSingleSelectValue struct {
									Name  string
									Field struct {
										ProjectV2SingleSelectField struct {
											Name string
										} `graphql:"... on ProjectV2SingleSelectField"`
									}
								} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
								// Text field value
								ProjectV2ItemFieldTextValue struct {
									Text  string
									Field struct {
										ProjectV2Field struct {
											Name string
										} `graphql:"... on ProjectV2Field"`
									}
								} `graphql:"... on ProjectV2ItemFieldTextValue"`
							}
						} `graphql:"fieldValues(first: 20)"`
					}
					PageInfo struct {
						HasNextPage bool
						EndCursor   string
					}
				} `graphql:"items(first: 100, after: $cursor)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
		"cursor":    (*graphql.String)(nil),
	}
	if cursor != nil {
		variables["cursor"] = graphql.String(*cursor)
	}

	err := c.gql.Query("GetProjectItems", &query, variables)
	if err != nil {
		return nil, pageInfo{}, fmt.Errorf("failed to get project items: %w", err)
	}

	var items []ProjectItem
	for _, node := range query.Node.ProjectV2.Items.Nodes {
		// Skip non-issue items (like draft issues or PRs)
		if node.Content.TypeName != "Issue" {
			continue
		}

		item := ProjectItem{
			ID: node.ID,
			Issue: &Issue{
				ID:     node.Content.Issue.ID,
				Number: node.Content.Issue.Number,
				Title:  node.Content.Issue.Title,
				Body:   node.Content.Issue.Body,
				State:  node.Content.Issue.State,
				URL:    node.Content.Issue.URL,
			},
		}

		// Parse repository
		if node.Content.Issue.Repository.NameWithOwner != "" {
			parts := splitRepoName(node.Content.Issue.Repository.NameWithOwner)
			if len(parts) == 2 {
				item.Issue.Repository = Repository{
					Owner: parts[0],
					Name:  parts[1],
				}
			}
		}

		// Parse assignees
		for _, a := range node.Content.Issue.Assignees.Nodes {
			item.Issue.Assignees = append(item.Issue.Assignees, Actor{Login: a.Login})
		}

		// Parse labels
		for _, l := range node.Content.Issue.Labels.Nodes {
			item.Issue.Labels = append(item.Issue.Labels, Label{Name: l.Name})
		}

		// Parse field values
		for _, fv := range node.FieldValues.Nodes {
			switch fv.TypeName {
			case "ProjectV2ItemFieldSingleSelectValue":
				if fv.ProjectV2ItemFieldSingleSelectValue.Name != "" {
					item.FieldValues = append(item.FieldValues, FieldValue{
						Field: fv.ProjectV2ItemFieldSingleSelectValue.Field.ProjectV2SingleSelectField.Name,
						Value: fv.ProjectV2ItemFieldSingleSelectValue.Name,
					})
				}
			case "ProjectV2ItemFieldTextValue":
				if fv.ProjectV2ItemFieldTextValue.Text != "" {
					item.FieldValues = append(item.FieldValues, FieldValue{
						Field: fv.ProjectV2ItemFieldTextValue.Field.ProjectV2Field.Name,
						Value: fv.ProjectV2ItemFieldTextValue.Text,
					})
				}
			}
		}

		items = append(items, item)
	}

	return items, pageInfo{
		HasNextPage: query.Node.ProjectV2.Items.PageInfo.HasNextPage,
		EndCursor:   query.Node.ProjectV2.Items.PageInfo.EndCursor,
	}, nil
}

// splitRepoName splits "owner/repo" into parts
func splitRepoName(nameWithOwner string) []string {
	for i, c := range nameWithOwner {
		if c == '/' {
			return []string{nameWithOwner[:i], nameWithOwner[i+1:]}
		}
	}
	return nil
}

// GetSubIssues fetches all sub-issues for a given issue
func (c *Client) GetSubIssues(owner, repo string, number int) ([]SubIssue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Repository struct {
			Issue struct {
				SubIssues struct {
					Nodes []struct {
						ID         string
						Number     int
						Title      string
						State      string
						URL        string `graphql:"url"`
						Repository struct {
							Name  string
							Owner struct {
								Login string
							}
						}
					}
				} `graphql:"subIssues(first: 50)"`
			} `graphql:"issue(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	gqlNumber, err := safeGraphQLInt(number)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"repo":   graphql.String(repo),
		"number": gqlNumber,
	}

	err = c.gql.Query("GetSubIssues", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-issues for %s/%s#%d: %w", owner, repo, number, err)
	}

	var subIssues []SubIssue
	for _, node := range query.Repository.Issue.SubIssues.Nodes {
		subIssues = append(subIssues, SubIssue{
			ID:     node.ID,
			Number: node.Number,
			Title:  node.Title,
			State:  node.State,
			URL:    node.URL,
			Repository: Repository{
				Owner: node.Repository.Owner.Login,
				Name:  node.Repository.Name,
			},
		})
	}

	return subIssues, nil
}

// GetSubIssueCounts fetches sub-issue counts for multiple issues in a single query.
// This is more efficient than calling GetSubIssues for each issue individually.
// Returns a map of issue number to sub-issue count.
func (c *Client) GetSubIssueCounts(owner, repo string, numbers []int) (map[int]int, error) {
	if len(numbers) == 0 {
		return make(map[int]int), nil
	}

	// Build a GraphQL query with aliases for each issue
	// Example: query { repository(owner:"o", name:"r") { i1: issue(number:1) { subIssues { totalCount } } } }
	var queryParts []string
	for i, num := range numbers {
		queryParts = append(queryParts, fmt.Sprintf("i%d: issue(number: %d) { subIssues { totalCount } }", i, num))
	}

	query := fmt.Sprintf(`query { repository(owner: %q, name: %q) { %s } }`,
		owner, repo, strings.Join(queryParts, " "))

	// Execute via gh api graphql
	cmd := exec.Command("gh", "api", "graphql",
		"-H", "X-Github-Next: sub_issues",
		"-f", "query="+query)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch sub-issue query: %w", err)
	}

	// Parse the response
	var response struct {
		Data struct {
			Repository map[string]struct {
				SubIssues struct {
					TotalCount int `json:"totalCount"`
				} `json:"subIssues"`
			} `json:"repository"`
		} `json:"data"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse batch sub-issue response: %w", err)
	}

	// Build result map
	result := make(map[int]int)
	for i, num := range numbers {
		alias := fmt.Sprintf("i%d", i)
		if issueData, ok := response.Data.Repository[alias]; ok {
			result[num] = issueData.SubIssues.TotalCount
		} else {
			result[num] = 0
		}
	}

	return result, nil
}

// GetRepositoryIssues fetches issues from a repository with the given state filter
func (c *Client) GetRepositoryIssues(owner, repo, state string) ([]Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// Map state to GraphQL enum values
	var states []graphql.String
	switch state {
	case "open":
		states = []graphql.String{"OPEN"}
	case "closed":
		states = []graphql.String{"CLOSED"}
	case "all", "":
		states = []graphql.String{"OPEN", "CLOSED"}
	default:
		states = []graphql.String{graphql.String(state)}
	}

	var query struct {
		Repository struct {
			Issues struct {
				Nodes []struct {
					ID     string
					Number int
					Title  string
					State  string
					URL    string `graphql:"url"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"issues(first: 100, states: $states)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"repo":   graphql.String(repo),
		"states": states,
	}

	err := c.gql.Query("GetRepositoryIssues", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues from %s/%s: %w", owner, repo, err)
	}

	var issues []Issue
	for _, node := range query.Repository.Issues.Nodes {
		issues = append(issues, Issue{
			ID:     node.ID,
			Number: node.Number,
			Title:  node.Title,
			State:  node.State,
			URL:    node.URL,
			Repository: Repository{
				Owner: owner,
				Name:  repo,
			},
		})
	}

	return issues, nil
}

// GetOpenIssuesByLabel fetches open issues with a specific label
func (c *Client) GetOpenIssuesByLabel(owner, repo, label string) ([]Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Repository struct {
			Issues struct {
				Nodes []struct {
					ID     string
					Number int
					Title  string
					State  string
					URL    string `graphql:"url"`
					Labels struct {
						Nodes []struct {
							Name string
						}
					} `graphql:"labels(first: 10)"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"issues(first: 100, states: [OPEN], labels: [$label])"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner": graphql.String(owner),
		"repo":  graphql.String(repo),
		"label": graphql.String(label),
	}

	err := c.gql.Query("GetOpenIssuesByLabel", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues with label %s from %s/%s: %w", label, owner, repo, err)
	}

	var issues []Issue
	for _, node := range query.Repository.Issues.Nodes {
		var labels []Label
		for _, l := range node.Labels.Nodes {
			labels = append(labels, Label{Name: l.Name})
		}
		issues = append(issues, Issue{
			ID:     node.ID,
			Number: node.Number,
			Title:  node.Title,
			State:  node.State,
			URL:    node.URL,
			Labels: labels,
			Repository: Repository{
				Owner: owner,
				Name:  repo,
			},
		})
	}

	return issues, nil
}

// GetClosedIssuesByLabel fetches closed issues with a specific label
func (c *Client) GetClosedIssuesByLabel(owner, repo, label string) ([]Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Repository struct {
			Issues struct {
				Nodes []struct {
					ID     string
					Number int
					Title  string
					State  string
					URL    string `graphql:"url"`
					Labels struct {
						Nodes []struct {
							Name string
						}
					} `graphql:"labels(first: 10)"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"issues(first: 100, states: [CLOSED], labels: [$label])"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner": graphql.String(owner),
		"repo":  graphql.String(repo),
		"label": graphql.String(label),
	}

	err := c.gql.Query("GetClosedIssuesByLabel", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get closed issues with label %s from %s/%s: %w", label, owner, repo, err)
	}

	var issues []Issue
	for _, node := range query.Repository.Issues.Nodes {
		var labels []Label
		for _, l := range node.Labels.Nodes {
			labels = append(labels, Label{Name: l.Name})
		}
		issues = append(issues, Issue{
			ID:     node.ID,
			Number: node.Number,
			Title:  node.Title,
			State:  node.State,
			URL:    node.URL,
			Labels: labels,
			Repository: Repository{
				Owner: owner,
				Name:  repo,
			},
		})
	}

	return issues, nil
}

// GetParentIssue fetches the parent issue for a given sub-issue
func (c *Client) GetParentIssue(owner, repo string, number int) (*Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Repository struct {
			Issue struct {
				Parent struct {
					ID     string
					Number int
					Title  string
					State  string
					URL    string `graphql:"url"`
				} `graphql:"parent"`
			} `graphql:"issue(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	gqlNumber, err := safeGraphQLInt(number)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"repo":   graphql.String(repo),
		"number": gqlNumber,
	}

	err = c.gql.Query("GetParentIssue", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent issue for %s/%s#%d: %w", owner, repo, number, err)
	}

	// If no parent issue, return nil
	if query.Repository.Issue.Parent.ID == "" {
		return nil, nil
	}

	return &Issue{
		ID:     query.Repository.Issue.Parent.ID,
		Number: query.Repository.Issue.Parent.Number,
		Title:  query.Repository.Issue.Parent.Title,
		State:  query.Repository.Issue.Parent.State,
		URL:    query.Repository.Issue.Parent.URL,
	}, nil
}

// ListProjects fetches all projects for an owner (user or organization)
func (c *Client) ListProjects(owner string) ([]Project, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// First try as user projects
	projects, err := c.listUserProjects(owner)
	if err == nil && len(projects) > 0 {
		return projects, nil
	}

	// If that fails or returns empty, try as organization projects
	orgProjects, err := c.listOrgProjects(owner)
	if err != nil {
		// If both fail, return user error if we had one
		if projects != nil {
			return projects, nil
		}
		return nil, fmt.Errorf("failed to list projects for %s: %w", owner, err)
	}

	return orgProjects, nil
}

func (c *Client) listUserProjects(owner string) ([]Project, error) {
	var query struct {
		User struct {
			ProjectsV2 struct {
				Nodes []struct {
					ID     string
					Number int
					Title  string
					URL    string `graphql:"url"`
					Closed bool
				}
			} `graphql:"projectsV2(first: 20, orderBy: {field: UPDATED_AT, direction: DESC})"`
		} `graphql:"user(login: $owner)"`
	}

	variables := map[string]interface{}{
		"owner": graphql.String(owner),
	}

	err := c.gql.Query("ListUserProjects", &query, variables)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, node := range query.User.ProjectsV2.Nodes {
		if node.Closed {
			continue // Skip closed projects
		}
		projects = append(projects, Project{
			ID:     node.ID,
			Number: node.Number,
			Title:  node.Title,
			URL:    node.URL,
			Closed: node.Closed,
			Owner: ProjectOwner{
				Type:  "User",
				Login: owner,
			},
		})
	}

	return projects, nil
}

// Comment represents an issue comment
type Comment struct {
	ID        string
	Author    string
	Body      string
	CreatedAt string
}

// GetIssueComments fetches comments for an issue
func (c *Client) GetIssueComments(owner, repo string, number int) ([]Comment, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Repository struct {
			Issue struct {
				Comments struct {
					Nodes []struct {
						ID        string
						Body      string
						CreatedAt string
						Author    struct {
							Login string
						}
					}
				} `graphql:"comments(first: 50)"`
			} `graphql:"issue(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	gqlNumber, err := safeGraphQLInt(number)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"owner":  graphql.String(owner),
		"repo":   graphql.String(repo),
		"number": gqlNumber,
	}

	err = c.gql.Query("GetIssueComments", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments for %s/%s#%d: %w", owner, repo, number, err)
	}

	var comments []Comment
	for _, node := range query.Repository.Issue.Comments.Nodes {
		comments = append(comments, Comment{
			ID:        node.ID,
			Author:    node.Author.Login,
			Body:      node.Body,
			CreatedAt: node.CreatedAt,
		})
	}

	return comments, nil
}

func (c *Client) listOrgProjects(owner string) ([]Project, error) {
	var query struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes []struct {
					ID     string
					Number int
					Title  string
					URL    string `graphql:"url"`
					Closed bool
				}
			} `graphql:"projectsV2(first: 20, orderBy: {field: UPDATED_AT, direction: DESC})"`
		} `graphql:"organization(login: $owner)"`
	}

	variables := map[string]interface{}{
		"owner": graphql.String(owner),
	}

	err := c.gql.Query("ListOrgProjects", &query, variables)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, node := range query.Organization.ProjectsV2.Nodes {
		if node.Closed {
			continue // Skip closed projects
		}
		projects = append(projects, Project{
			ID:     node.ID,
			Number: node.Number,
			Title:  node.Title,
			URL:    node.URL,
			Closed: node.Closed,
			Owner: ProjectOwner{
				Type:  "Organization",
				Login: owner,
			},
		})
	}

	return projects, nil
}
