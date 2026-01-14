package api

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
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

// GetProjectFields fetches all fields for a project.
// Uses cursor-based pagination to retrieve all fields regardless of project size.
func (c *Client) GetProjectFields(projectID string) ([]ProjectField, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var allFields []ProjectField
	var cursor *string

	for {
		fields, pInfo, err := c.getProjectFieldsPage(projectID, cursor)
		if err != nil {
			return nil, err
		}

		allFields = append(allFields, fields...)

		if !pInfo.HasNextPage {
			break
		}
		cursor = &pInfo.EndCursor
	}

	return allFields, nil
}

// getProjectFieldsPage fetches a single page of project fields
func (c *Client) getProjectFieldsPage(projectID string, cursor *string) ([]ProjectField, pageInfo, error) {
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
					PageInfo struct {
						HasNextPage bool
						EndCursor   string
					}
				} `graphql:"fields(first: 100, after: $cursor)"`
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

	err := c.gql.Query("GetProjectFields", &query, variables)
	if err != nil {
		return nil, pageInfo{}, fmt.Errorf("failed to get project fields: %w", err)
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

	return fields, pageInfo{
		HasNextPage: query.Node.ProjectV2.Fields.PageInfo.HasNextPage,
		EndCursor:   query.Node.ProjectV2.Fields.PageInfo.EndCursor,
	}, nil
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
	Limit      int    // Maximum number of items to return (0 = no limit)
}

// GetProjectItems fetches all items from a project with their field values.
// Uses cursor-based pagination to retrieve all items regardless of project size.
// If filter.Limit > 0, pagination terminates early once the limit is reached.
func (c *Client) GetProjectItems(projectID string, filter *ProjectItemsFilter) ([]ProjectItem, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var allItems []ProjectItem
	var cursor *string
	limit := 0
	if filter != nil {
		limit = filter.Limit
	}

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

			// Early termination if limit is reached
			if limit > 0 && len(allItems) >= limit {
				return allItems[:limit], nil
			}
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

// BoardItemsFilter allows filtering board items
type BoardItemsFilter struct {
	Repository string // Filter by repository (owner/repo format)
}

// GetProjectItemsForBoard fetches minimal project item data optimized for board display.
// Only retrieves: Number, Title, Status, Priority, and Repository.
// Uses cursor-based pagination to retrieve all items regardless of project size.
func (c *Client) GetProjectItemsForBoard(projectID string, filter *BoardItemsFilter) ([]BoardItem, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var allItems []BoardItem
	var cursor *string

	for {
		items, pInfo, err := c.getBoardItemsPage(projectID, cursor)
		if err != nil {
			return nil, err
		}

		// Filter and process items from this page
		for _, item := range items {
			if filter != nil && filter.Repository != "" {
				if item.Repository != filter.Repository {
					continue
				}
			}
			allItems = append(allItems, item)
		}

		if !pInfo.HasNextPage {
			break
		}
		cursor = &pInfo.EndCursor
	}

	return allItems, nil
}

// getBoardItemsPage fetches a single page of board items with minimal data
func (c *Client) getBoardItemsPage(projectID string, cursor *string) ([]BoardItem, pageInfo, error) {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Items struct {
					Nodes []struct {
						Content struct {
							TypeName string `graphql:"__typename"`
							Issue    struct {
								Number     int
								Title      string
								Repository struct {
									NameWithOwner string
								}
							} `graphql:"... on Issue"`
						}
						FieldValues struct {
							Nodes []struct {
								TypeName string `graphql:"__typename"`
								ProjectV2ItemFieldSingleSelectValue struct {
									Name  string
									Field struct {
										ProjectV2SingleSelectField struct {
											Name string
										} `graphql:"... on ProjectV2SingleSelectField"`
									}
								} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
							}
						} `graphql:"fieldValues(first: 10)"`
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

	err := c.gql.Query("GetProjectItemsForBoard", &query, variables)
	if err != nil {
		return nil, pageInfo{}, fmt.Errorf("failed to get board items: %w", err)
	}

	var items []BoardItem
	for _, node := range query.Node.ProjectV2.Items.Nodes {
		// Skip non-issue items
		if node.Content.TypeName != "Issue" {
			continue
		}

		item := BoardItem{
			Number:     node.Content.Issue.Number,
			Title:      node.Content.Issue.Title,
			Repository: node.Content.Issue.Repository.NameWithOwner,
		}

		// Extract Status and Priority from field values
		for _, fv := range node.FieldValues.Nodes {
			if fv.TypeName == "ProjectV2ItemFieldSingleSelectValue" {
				fieldName := fv.ProjectV2ItemFieldSingleSelectValue.Field.ProjectV2SingleSelectField.Name
				value := fv.ProjectV2ItemFieldSingleSelectValue.Name
				switch fieldName {
				case "Status":
					item.Status = value
				case "Priority":
					item.Priority = value
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

// GetSubIssues fetches all sub-issues for a given issue with pagination support
func (c *Client) GetSubIssues(owner, repo string, number int) ([]SubIssue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var subIssues []SubIssue
	var cursor *graphql.String
	pageCount := 0

	for {
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
						PageInfo struct {
							HasNextPage bool
							EndCursor   string
						}
					} `graphql:"subIssues(first: 100, after: $cursor)"`
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
			"cursor": cursor,
		}

		err = c.gql.Query("GetSubIssues", &query, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to get sub-issues for %s/%s#%d: %w", owner, repo, number, err)
		}

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

		pageCount++
		if !query.Repository.Issue.SubIssues.PageInfo.HasNextPage {
			break
		}

		endCursor := graphql.String(query.Repository.Issue.SubIssues.PageInfo.EndCursor)
		cursor = &endCursor

		// Warn if we're fetching many pages (performance awareness)
		if pageCount >= 2 {
			fmt.Fprintf(os.Stderr, "Note: Issue #%d has many sub-issues (%d+ fetched), this may take a moment...\n", number, len(subIssues))
		}
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

	// Map state to GraphQL enum values (IssueState enum, not String)
	var states []IssueState
	switch state {
	case "open":
		states = []IssueState{IssueStateOpen}
	case "closed":
		states = []IssueState{IssueStateClosed}
	case "all", "":
		states = []IssueState{IssueStateOpen, IssueStateClosed}
	default:
		states = []IssueState{IssueState(state)}
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

// SearchRepositoryIssues searches for issues in a repository using GitHub Search API.
// This is more efficient than fetching all issues when filtering by state, labels, or text.
// The limit parameter controls maximum results (0 = no limit, uses pagination).
func (c *Client) SearchRepositoryIssues(owner, repo string, filters SearchFilters, limit int) ([]Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// Build the search query string
	queryParts := []string{
		fmt.Sprintf("repo:%s/%s", owner, repo),
		"is:issue",
	}

	// Add state filter (default to open if not specified)
	switch filters.State {
	case "closed":
		queryParts = append(queryParts, "is:closed")
	case "all":
		// No state filter - include both open and closed
	case "open", "":
		queryParts = append(queryParts, "is:open")
	default:
		queryParts = append(queryParts, "is:open") // Default to open for unknown states
	}

	// Add label filters
	for _, label := range filters.Labels {
		queryParts = append(queryParts, fmt.Sprintf("label:%q", label))
	}

	// Add assignee filter
	if filters.Assignee != "" {
		queryParts = append(queryParts, fmt.Sprintf("assignee:%s", filters.Assignee))
	}

	// Add free-text search
	if filters.Search != "" {
		queryParts = append(queryParts, filters.Search)
	}

	searchQuery := strings.Join(queryParts, " ")

	var allIssues []Issue
	var cursor *string
	pageSize := 100
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}

	for {
		issues, pageInfo, err := c.searchIssuesPage(searchQuery, pageSize, cursor)
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			// Set repository info since search doesn't always include it
			if issue.Repository.Owner == "" {
				issue.Repository = Repository{Owner: owner, Name: repo}
			}
			allIssues = append(allIssues, issue)

			// Check if we've reached the limit
			if limit > 0 && len(allIssues) >= limit {
				return allIssues[:limit], nil
			}
		}

		if !pageInfo.HasNextPage {
			break
		}
		cursor = &pageInfo.EndCursor
	}

	return allIssues, nil
}

// GetProjectFieldsForIssues fetches project field values for multiple issues in batched queries.
// Returns a map from issue node ID to its field values. Uses batching to avoid query size limits.
func (c *Client) GetProjectFieldsForIssues(projectID string, issueIDs []string) (map[string][]FieldValue, error) {
	if len(issueIDs) == 0 {
		return make(map[string][]FieldValue), nil
	}

	// Batch issues to avoid GraphQL query limits (max ~100 aliases per query)
	const batchSize = 50
	result := make(map[string][]FieldValue)

	for i := 0; i < len(issueIDs); i += batchSize {
		end := i + batchSize
		if end > len(issueIDs) {
			end = len(issueIDs)
		}
		batch := issueIDs[i:end]

		batchResult, err := c.getProjectFieldsForIssuesBatch(projectID, batch)
		if err != nil {
			return nil, err
		}

		for id, fields := range batchResult {
			result[id] = fields
		}
	}

	return result, nil
}

// getProjectFieldsForIssuesBatch fetches project fields for a batch of issues using aliases
func (c *Client) getProjectFieldsForIssuesBatch(projectID string, issueIDs []string) (map[string][]FieldValue, error) {
	// Build a GraphQL query with aliases for each issue
	var queryParts []string
	for i, id := range issueIDs {
		queryParts = append(queryParts, fmt.Sprintf(`n%d: node(id: %q) {
			... on Issue {
				id
				projectItems(first: 10) {
					nodes {
						project { id }
						fieldValues(first: 20) {
							nodes {
								__typename
								... on ProjectV2ItemFieldSingleSelectValue {
									name
									field { ... on ProjectV2SingleSelectField { name } }
								}
								... on ProjectV2ItemFieldTextValue {
									text
									field { ... on ProjectV2Field { name } }
								}
							}
						}
					}
				}
			}
		}`, i, id))
	}

	query := fmt.Sprintf("query { %s }", strings.Join(queryParts, " "))

	// Execute via gh api graphql
	cmd := exec.Command("gh", "api", "graphql", "-f", "query="+query)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch project fields query: %w", err)
	}

	// Parse the response
	var response struct {
		Data map[string]struct {
			ID           string `json:"id"`
			ProjectItems struct {
				Nodes []struct {
					Project struct {
						ID string `json:"id"`
					} `json:"project"`
					FieldValues struct {
						Nodes []struct {
							TypeName string `json:"__typename"`
							Name     string `json:"name"`
							Text     string `json:"text"`
							Field    struct {
								Name string `json:"name"`
							} `json:"field"`
						} `json:"nodes"`
					} `json:"fieldValues"`
				} `json:"nodes"`
			} `json:"projectItems"`
		} `json:"data"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse batch project fields response: %w", err)
	}

	// Build result map
	result := make(map[string][]FieldValue)
	for _, nodeData := range response.Data {
		if nodeData.ID == "" {
			continue
		}

		var fieldValues []FieldValue
		for _, projectItem := range nodeData.ProjectItems.Nodes {
			// Only include field values from the target project
			if projectItem.Project.ID != projectID {
				continue
			}

			for _, fv := range projectItem.FieldValues.Nodes {
				switch fv.TypeName {
				case "ProjectV2ItemFieldSingleSelectValue":
					if fv.Name != "" && fv.Field.Name != "" {
						fieldValues = append(fieldValues, FieldValue{
							Field: fv.Field.Name,
							Value: fv.Name,
						})
					}
				case "ProjectV2ItemFieldTextValue":
					if fv.Text != "" && fv.Field.Name != "" {
						fieldValues = append(fieldValues, FieldValue{
							Field: fv.Field.Name,
							Value: fv.Text,
						})
					}
				}
			}
		}

		result[nodeData.ID] = fieldValues
	}

	return result, nil
}

// searchIssuesPage fetches a single page of search results
func (c *Client) searchIssuesPage(query string, pageSize int, cursor *string) ([]Issue, pageInfo, error) {
	var gqlQuery struct {
		Search struct {
			Nodes []struct {
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
				} `graphql:"... on Issue"`
			}
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
		} `graphql:"search(query: $query, type: ISSUE, first: $first, after: $cursor)"`
	}

	gqlPageSize, err := safeGraphQLInt(pageSize)
	if err != nil {
		return nil, pageInfo{}, err
	}

	variables := map[string]interface{}{
		"query":  graphql.String(query),
		"first":  gqlPageSize,
		"cursor": (*graphql.String)(nil),
	}
	if cursor != nil {
		variables["cursor"] = graphql.String(*cursor)
	}

	err = c.gql.Query("SearchIssues", &gqlQuery, variables)
	if err != nil {
		return nil, pageInfo{}, fmt.Errorf("failed to search issues: %w", err)
	}

	var issues []Issue
	for _, node := range gqlQuery.Search.Nodes {
		if node.TypeName != "Issue" {
			continue
		}

		issue := Issue{
			ID:     node.Issue.ID,
			Number: node.Issue.Number,
			Title:  node.Issue.Title,
			Body:   node.Issue.Body,
			State:  node.Issue.State,
			URL:    node.Issue.URL,
			Author: Actor{Login: node.Issue.Author.Login},
		}

		// Parse repository
		if node.Issue.Repository.NameWithOwner != "" {
			parts := splitRepoName(node.Issue.Repository.NameWithOwner)
			if len(parts) == 2 {
				issue.Repository = Repository{
					Owner: parts[0],
					Name:  parts[1],
				}
			}
		}

		// Parse assignees
		for _, a := range node.Issue.Assignees.Nodes {
			issue.Assignees = append(issue.Assignees, Actor{Login: a.Login})
		}

		// Parse labels
		for _, l := range node.Issue.Labels.Nodes {
			issue.Labels = append(issue.Labels, Label{Name: l.Name, Color: l.Color})
		}

		// Parse milestone
		if node.Issue.Milestone.Title != "" {
			issue.Milestone = &Milestone{Title: node.Issue.Milestone.Title}
		}

		issues = append(issues, issue)
	}

	return issues, pageInfo{
		HasNextPage: gqlQuery.Search.PageInfo.HasNextPage,
		EndCursor:   gqlQuery.Search.PageInfo.EndCursor,
	}, nil
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
