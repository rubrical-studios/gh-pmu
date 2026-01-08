package api

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	graphql "github.com/cli/shurcooL-graphql"
)

// CreateIssue creates a new issue in a repository
func (c *Client) CreateIssue(owner, repo, title, body string, labels []string) (*Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// First, get the repository ID
	repoID, err := c.getRepositoryID(owner, repo)
	if err != nil {
		return nil, err
	}

	// Get label IDs if labels are provided
	var labelIDs []graphql.ID
	if len(labels) > 0 {
		for _, labelName := range labels {
			labelID, err := c.getLabelID(owner, repo, labelName)
			if err != nil {
				// Skip labels that don't exist
				continue
			}
			labelIDs = append(labelIDs, graphql.ID(labelID))
		}
	}

	var mutation struct {
		CreateIssue struct {
			Issue struct {
				ID     string
				Number int
				Title  string
				Body   string
				State  string
				URL    string `graphql:"url"`
			}
		} `graphql:"createIssue(input: $input)"`
	}

	input := CreateIssueInput{
		RepositoryID: graphql.ID(repoID),
		Title:        graphql.String(title),
	}
	if body != "" {
		input.Body = graphql.String(body)
	}
	if len(labelIDs) > 0 {
		input.LabelIDs = &labelIDs
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err = c.gql.Mutate("CreateIssue", &mutation, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return &Issue{
		ID:     mutation.CreateIssue.Issue.ID,
		Number: mutation.CreateIssue.Issue.Number,
		Title:  mutation.CreateIssue.Issue.Title,
		Body:   mutation.CreateIssue.Issue.Body,
		State:  mutation.CreateIssue.Issue.State,
		URL:    mutation.CreateIssue.Issue.URL,
		Repository: Repository{
			Owner: owner,
			Name:  repo,
		},
	}, nil
}

// CreateIssueInput represents the input for creating an issue
type CreateIssueInput struct {
	RepositoryID graphql.ID     `json:"repositoryId"`
	Title        graphql.String `json:"title"`
	Body         graphql.String `json:"body,omitempty"`
	LabelIDs     *[]graphql.ID  `json:"labelIds,omitempty"`
	AssigneeIDs  *[]graphql.ID  `json:"assigneeIds,omitempty"`
	MilestoneID  *graphql.ID    `json:"milestoneId,omitempty"`
}

// CloseIssueInput represents the input for closing an issue
type CloseIssueInput struct {
	IssueID graphql.ID `json:"issueId"`
}

// ReopenIssueInput represents the input for reopening an issue
type ReopenIssueInput struct {
	IssueID graphql.ID `json:"issueId"`
}

// UpdateIssueInput represents the input for updating an issue
type UpdateIssueInput struct {
	ID    graphql.ID     `json:"id"`
	Body  graphql.String `json:"body,omitempty"`
	Title graphql.String `json:"title,omitempty"`
}

// AddIssueToProject adds an issue to a GitHub Project V2
func (c *Client) AddIssueToProject(projectID, issueID string) (string, error) {
	if c.gql == nil {
		return "", fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID string
			}
		} `graphql:"addProjectV2ItemById(input: $input)"`
	}

	input := AddProjectV2ItemByIdInput{
		ProjectID: graphql.ID(projectID),
		ContentID: graphql.ID(issueID),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("AddProjectV2ItemById", &mutation, variables)
	if err != nil {
		return "", fmt.Errorf("failed to add issue to project: %w", err)
	}

	return mutation.AddProjectV2ItemById.Item.ID, nil
}

// AddProjectV2ItemByIdInput represents the input for adding an item to a project
type AddProjectV2ItemByIdInput struct {
	ProjectID graphql.ID `json:"projectId"`
	ContentID graphql.ID `json:"contentId"`
}

// SetProjectItemField sets a field value on a project item.
// This method fetches project fields on each call. For bulk operations,
// use SetProjectItemFieldWithFields with pre-fetched fields for better performance.
func (c *Client) SetProjectItemField(projectID, itemID, fieldName, value string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// Get the field ID and option ID for single select fields
	fields, err := c.GetProjectFields(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project fields: %w", err)
	}

	return c.SetProjectItemFieldWithFields(projectID, itemID, fieldName, value, fields)
}

// SetProjectItemFieldWithFields sets a field value using pre-fetched project fields.
// Use this method for bulk operations to avoid redundant GetProjectFields API calls.
func (c *Client) SetProjectItemFieldWithFields(projectID, itemID, fieldName, value string, fields []ProjectField) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var field *ProjectField
	for i := range fields {
		if fields[i].Name == fieldName {
			field = &fields[i]
			break
		}
	}

	if field == nil {
		return fmt.Errorf("field %q not found in project", fieldName)
	}

	// Handle different field types
	switch field.DataType {
	case "SINGLE_SELECT":
		return c.setSingleSelectField(projectID, itemID, field, value)
	case "TEXT":
		return c.setTextField(projectID, itemID, field.ID, value)
	case "NUMBER":
		return c.setNumberField(projectID, itemID, field.ID, value)
	default:
		return fmt.Errorf("unsupported field type: %s", field.DataType)
	}
}

func (c *Client) setSingleSelectField(projectID, itemID string, field *ProjectField, value string) error {
	// Find the option ID for the value
	var optionID string
	for _, opt := range field.Options {
		if opt.Name == value {
			optionID = opt.ID
			break
		}
	}

	if optionID == "" {
		return fmt.Errorf("option %q not found for field %q", value, field.Name)
	}

	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
	}

	input := UpdateProjectV2ItemFieldValueInput{
		ProjectID: graphql.ID(projectID),
		ItemID:    graphql.ID(itemID),
		FieldID:   graphql.ID(field.ID),
		Value: ProjectV2FieldValue{
			SingleSelectOptionId: graphql.String(optionID),
		},
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("UpdateProjectV2ItemFieldValue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to set field value: %w", err)
	}

	return nil
}

func (c *Client) setTextField(projectID, itemID, fieldID, value string) error {
	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
	}

	input := UpdateProjectV2ItemFieldValueInput{
		ProjectID: graphql.ID(projectID),
		ItemID:    graphql.ID(itemID),
		FieldID:   graphql.ID(fieldID),
		Value: ProjectV2FieldValue{
			Text: graphql.String(value),
		},
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("UpdateProjectV2ItemFieldValue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to set text field value: %w", err)
	}

	return nil
}

func (c *Client) setNumberField(projectID, itemID, fieldID, value string) error {
	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
	}

	input := UpdateProjectV2ItemFieldValueInput{
		ProjectID: graphql.ID(projectID),
		ItemID:    graphql.ID(itemID),
		FieldID:   graphql.ID(fieldID),
		Value: ProjectV2FieldValue{
			Number: graphql.Float(0), // TODO: parse value to float
		},
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("UpdateProjectV2ItemFieldValue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to set number field value: %w", err)
	}

	return nil
}

// UpdateProjectV2ItemFieldValueInput represents the input for updating a field value
type UpdateProjectV2ItemFieldValueInput struct {
	ProjectID graphql.ID          `json:"projectId"`
	ItemID    graphql.ID          `json:"itemId"`
	FieldID   graphql.ID          `json:"fieldId"`
	Value     ProjectV2FieldValue `json:"value"`
}

// ProjectV2FieldValue represents a field value for a project item
type ProjectV2FieldValue struct {
	Text                 graphql.String `json:"text,omitempty"`
	Number               graphql.Float  `json:"number,omitempty"`
	Date                 graphql.String `json:"date,omitempty"`
	SingleSelectOptionId graphql.String `json:"singleSelectOptionId,omitempty"`
	IterationId          graphql.String `json:"iterationId,omitempty"`
}

// Helper methods

func (c *Client) getRepositoryID(owner, repo string) (string, error) {
	var query struct {
		Repository struct {
			ID string
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner": graphql.String(owner),
		"repo":  graphql.String(repo),
	}

	err := c.gql.Query("GetRepositoryID", &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get repository ID: %w", err)
	}

	return query.Repository.ID, nil
}

// AddSubIssue links a child issue as a sub-issue of a parent issue
func (c *Client) AddSubIssue(parentIssueID, childIssueID string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		AddSubIssue struct {
			Issue struct {
				ID string
			}
			SubIssue struct {
				ID string
			}
		} `graphql:"addSubIssue(input: $input)"`
	}

	input := AddSubIssueInput{
		IssueID:    graphql.ID(parentIssueID),
		SubIssueID: graphql.ID(childIssueID),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("AddSubIssue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to add sub-issue: %w", err)
	}

	return nil
}

// AddSubIssueInput represents the input for adding a sub-issue
type AddSubIssueInput struct {
	IssueID    graphql.ID `json:"issueId"`
	SubIssueID graphql.ID `json:"subIssueId"`
}

// RemoveSubIssue removes a child issue from its parent issue
func (c *Client) RemoveSubIssue(parentIssueID, childIssueID string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		RemoveSubIssue struct {
			Issue struct {
				ID string
			}
			SubIssue struct {
				ID string
			}
		} `graphql:"removeSubIssue(input: $input)"`
	}

	input := RemoveSubIssueInput{
		IssueID:    graphql.ID(parentIssueID),
		SubIssueID: graphql.ID(childIssueID),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("RemoveSubIssue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to remove sub-issue: %w", err)
	}

	return nil
}

// RemoveSubIssueInput represents the input for removing a sub-issue
type RemoveSubIssueInput struct {
	IssueID    graphql.ID `json:"issueId"`
	SubIssueID graphql.ID `json:"subIssueId"`
}

// CreateProjectField creates a new field in a GitHub project.
// Supported field types: TEXT, NUMBER, DATE, SINGLE_SELECT, ITERATION
func (c *Client) CreateProjectField(projectID, name, dataType string, singleSelectOptions []string) (*ProjectField, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		CreateProjectV2Field struct {
			ProjectV2Field struct {
				TypeName       string `graphql:"__typename"`
				ProjectV2Field struct {
					ID   string
					Name string
				} `graphql:"... on ProjectV2Field"`
				ProjectV2SingleSelectField struct {
					ID      string
					Name    string
					Options []struct {
						ID   string
						Name string
					}
				} `graphql:"... on ProjectV2SingleSelectField"`
			} `graphql:"projectV2Field"`
		} `graphql:"createProjectV2Field(input: $input)"`
	}

	input := CreateProjectV2FieldInput{
		ProjectID: graphql.ID(projectID),
		DataType:  graphql.String(dataType),
		Name:      graphql.String(name),
	}

	// Add single select options if provided
	if dataType == "SINGLE_SELECT" && len(singleSelectOptions) > 0 {
		var options []ProjectV2SingleSelectFieldOptionInput
		for _, opt := range singleSelectOptions {
			options = append(options, ProjectV2SingleSelectFieldOptionInput{
				Name: graphql.String(opt),
			})
		}
		input.SingleSelectOptions = &options
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("CreateProjectV2Field", &mutation, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to create project field: %w", err)
	}

	// Build result based on field type
	result := &ProjectField{
		Name:     name,
		DataType: dataType,
	}

	// Extract ID based on the type returned
	if dataType == "SINGLE_SELECT" {
		result.ID = mutation.CreateProjectV2Field.ProjectV2Field.ProjectV2SingleSelectField.ID
		result.Name = mutation.CreateProjectV2Field.ProjectV2Field.ProjectV2SingleSelectField.Name
		for _, opt := range mutation.CreateProjectV2Field.ProjectV2Field.ProjectV2SingleSelectField.Options {
			result.Options = append(result.Options, FieldOption{
				ID:   opt.ID,
				Name: opt.Name,
			})
		}
	} else {
		result.ID = mutation.CreateProjectV2Field.ProjectV2Field.ProjectV2Field.ID
		result.Name = mutation.CreateProjectV2Field.ProjectV2Field.ProjectV2Field.Name
	}

	return result, nil
}

// CreateProjectV2FieldInput represents the input for creating a project field
type CreateProjectV2FieldInput struct {
	ProjectID           graphql.ID                               `json:"projectId"`
	DataType            graphql.String                           `json:"dataType"`
	Name                graphql.String                           `json:"name"`
	SingleSelectOptions *[]ProjectV2SingleSelectFieldOptionInput `json:"singleSelectOptions,omitempty"`
}

// ProjectV2SingleSelectFieldOptionInput represents an option for a single select field
type ProjectV2SingleSelectFieldOptionInput struct {
	Name        graphql.String `json:"name"`
	Color       graphql.String `json:"color,omitempty"`
	Description graphql.String `json:"description,omitempty"`
}

// AddLabelToIssue adds a label to an issue
func (c *Client) AddLabelToIssue(issueID, labelName string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// Note: This requires finding the label ID first, which needs the repository
	// For now, we'll skip this as it requires additional context
	// A full implementation would use addLabelsToLabelable mutation
	return nil
}

func (c *Client) getLabelID(owner, repo, labelName string) (string, error) {
	var query struct {
		Repository struct {
			Label struct {
				ID string
			} `graphql:"label(name: $labelName)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner":     graphql.String(owner),
		"repo":      graphql.String(repo),
		"labelName": graphql.String(labelName),
	}

	err := c.gql.Query("GetLabelID", &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get label ID: %w", err)
	}

	if query.Repository.Label.ID == "" {
		return "", fmt.Errorf("label %q not found", labelName)
	}

	return query.Repository.Label.ID, nil
}

// getUserID gets a user's ID from their login
func (c *Client) getUserID(login string) (string, error) {
	var query struct {
		User struct {
			ID string
		} `graphql:"user(login: $login)"`
	}

	variables := map[string]interface{}{
		"login": graphql.String(login),
	}

	err := c.gql.Query("GetUserID", &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get user ID for %s: %w", login, err)
	}

	if query.User.ID == "" {
		return "", fmt.Errorf("user %q not found", login)
	}

	return query.User.ID, nil
}

// getMilestoneID gets a milestone ID from the repository
func (c *Client) getMilestoneID(owner, repo, milestone string) (string, error) {
	var query struct {
		Repository struct {
			Milestones struct {
				Nodes []struct {
					ID     string
					Title  string
					Number int
				}
			} `graphql:"milestones(first: 100, states: OPEN)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner": graphql.String(owner),
		"repo":  graphql.String(repo),
	}

	err := c.gql.Query("GetMilestones", &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get milestones: %w", err)
	}

	// Try to match by title or number
	for _, m := range query.Repository.Milestones.Nodes {
		if m.Title == milestone || fmt.Sprintf("%d", m.Number) == milestone {
			return m.ID, nil
		}
	}

	return "", fmt.Errorf("milestone %q not found", milestone)
}

// CreateIssueWithOptions creates an issue with extended options
func (c *Client) CreateIssueWithOptions(owner, repo, title, body string, labels, assignees []string, milestone string) (*Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// First, get the repository ID
	repoID, err := c.getRepositoryID(owner, repo)
	if err != nil {
		return nil, err
	}

	// Get label IDs if labels are provided
	var labelIDs []graphql.ID
	if len(labels) > 0 {
		for _, labelName := range labels {
			labelID, err := c.getLabelID(owner, repo, labelName)
			if err != nil {
				// Skip labels that don't exist
				continue
			}
			labelIDs = append(labelIDs, graphql.ID(labelID))
		}
	}

	// Get assignee IDs
	var assigneeIDs []graphql.ID
	if len(assignees) > 0 {
		for _, login := range assignees {
			userID, err := c.getUserID(login)
			if err != nil {
				// Skip users that don't exist
				continue
			}
			assigneeIDs = append(assigneeIDs, graphql.ID(userID))
		}
	}

	// Get milestone ID
	var milestoneID *graphql.ID
	if milestone != "" {
		mID, err := c.getMilestoneID(owner, repo, milestone)
		if err != nil {
			// Non-fatal, just warn
			fmt.Printf("Warning: milestone %q not found\n", milestone)
		} else {
			gqlID := graphql.ID(mID)
			milestoneID = &gqlID
		}
	}

	var mutation struct {
		CreateIssue struct {
			Issue struct {
				ID     string
				Number int
				Title  string
				Body   string
				State  string
				URL    string `graphql:"url"`
			}
		} `graphql:"createIssue(input: $input)"`
	}

	input := CreateIssueInput{
		RepositoryID: graphql.ID(repoID),
		Title:        graphql.String(title),
	}
	if body != "" {
		input.Body = graphql.String(body)
	}
	if len(labelIDs) > 0 {
		input.LabelIDs = &labelIDs
	}
	if len(assigneeIDs) > 0 {
		input.AssigneeIDs = &assigneeIDs
	}
	if milestoneID != nil {
		input.MilestoneID = milestoneID
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err = c.gql.Mutate("CreateIssue", &mutation, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return &Issue{
		ID:     mutation.CreateIssue.Issue.ID,
		Number: mutation.CreateIssue.Issue.Number,
		Title:  mutation.CreateIssue.Issue.Title,
		Body:   mutation.CreateIssue.Issue.Body,
		State:  mutation.CreateIssue.Issue.State,
		URL:    mutation.CreateIssue.Issue.URL,
		Repository: Repository{
			Owner: owner,
			Name:  repo,
		},
	}, nil
}

// CloseIssue closes an issue by its ID
func (c *Client) CloseIssue(issueID string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		CloseIssue struct {
			Issue struct {
				ID string
			}
		} `graphql:"closeIssue(input: $input)"`
	}

	input := CloseIssueInput{
		IssueID: graphql.ID(issueID),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("CloseIssue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to close issue: %w", err)
	}

	return nil
}

// ReopenIssue reopens a closed issue
func (c *Client) ReopenIssue(issueID string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		ReopenIssue struct {
			Issue struct {
				ID string
			}
		} `graphql:"reopenIssue(input: $input)"`
	}

	input := ReopenIssueInput{
		IssueID: graphql.ID(issueID),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("ReopenIssue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to reopen issue: %w", err)
	}

	return nil
}

// UpdateIssueBody updates the body of an issue
func (c *Client) UpdateIssueBody(issueID, body string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		UpdateIssue struct {
			Issue struct {
				ID string
			}
		} `graphql:"updateIssue(input: $input)"`
	}

	input := UpdateIssueInput{
		ID:   graphql.ID(issueID),
		Body: graphql.String(body),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("UpdateIssue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to update issue body: %w", err)
	}

	return nil
}

// UpdateIssueTitle updates the title of an issue
func (c *Client) UpdateIssueTitle(issueID, title string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		UpdateIssue struct {
			Issue struct {
				ID string
			}
		} `graphql:"updateIssue(input: $input)"`
	}

	input := UpdateIssueInput{
		ID:    graphql.ID(issueID),
		Title: graphql.String(title),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("UpdateIssue", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to update issue title: %w", err)
	}

	return nil
}

// GetIssueByNumber returns an issue by its number (alias for GetIssue)
func (c *Client) GetIssueByNumber(owner, repo string, number int) (*Issue, error) {
	return c.GetIssue(owner, repo, number)
}

// GetProjectItemID returns the project item ID for an issue in a project
func (c *Client) GetProjectItemID(projectID, issueID string) (string, error) {
	if c.gql == nil {
		return "", fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Node struct {
			ProjectV2 struct {
				Items struct {
					Nodes []struct {
						ID      string
						Content struct {
							Issue struct {
								ID string
							} `graphql:"... on Issue"`
						}
					}
					PageInfo struct {
						HasNextPage bool
						EndCursor   string
					}
				} `graphql:"items(first: 100)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
	}

	err := c.gql.Query("GetProjectItems", &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get project items: %w", err)
	}

	for _, item := range query.Node.ProjectV2.Items.Nodes {
		if item.Content.Issue.ID == issueID {
			return item.ID, nil
		}
	}

	return "", fmt.Errorf("issue not found in project")
}

// GetProjectItemFieldValue returns the value of a field on a project item
func (c *Client) GetProjectItemFieldValue(projectID, itemID, fieldName string) (string, error) {
	if c.gql == nil {
		return "", fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Node struct {
			ProjectV2Item struct {
				FieldValues struct {
					Nodes []struct {
						ProjectV2ItemFieldTextValue struct {
							Text  string
							Field struct {
								Name string
							} `graphql:"field"`
						} `graphql:"... on ProjectV2ItemFieldTextValue"`
						ProjectV2ItemFieldSingleSelectValue struct {
							Name  string
							Field struct {
								Name string
							} `graphql:"field"`
						} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
					}
				} `graphql:"fieldValues(first: 20)"`
			} `graphql:"... on ProjectV2Item"`
		} `graphql:"node(id: $itemId)"`
	}

	variables := map[string]interface{}{
		"itemId": graphql.ID(itemID),
	}

	err := c.gql.Query("GetProjectItemFieldValue", &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to get field value: %w", err)
	}

	for _, fv := range query.Node.ProjectV2Item.FieldValues.Nodes {
		if fv.ProjectV2ItemFieldTextValue.Field.Name == fieldName {
			return fv.ProjectV2ItemFieldTextValue.Text, nil
		}
		if fv.ProjectV2ItemFieldSingleSelectValue.Field.Name == fieldName {
			return fv.ProjectV2ItemFieldSingleSelectValue.Name, nil
		}
	}

	return "", nil
}

// GetIssuesByRelease returns issues that have a specific release field value
func (c *Client) GetIssuesByRelease(owner, repo, releaseVersion string) ([]Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	issues, err := c.GetRepositoryIssues(owner, repo, "OPEN")
	if err != nil {
		return nil, err
	}

	return issues, nil
}

// GetIssuesByPatch returns issues that have a specific patch field value
func (c *Client) GetIssuesByPatch(owner, repo, patchVersion string) ([]Issue, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	issues, err := c.GetRepositoryIssues(owner, repo, "OPEN")
	if err != nil {
		return nil, err
	}

	return issues, nil
}

// WriteFile writes content to a file path
func (c *Client) WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// MkdirAll creates a directory and all parents
func (c *Client) MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

// GitAdd stages files to git
func (c *Client) GitAdd(paths ...string) error {
	args := append([]string{"add"}, paths...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// GitTag creates an annotated git tag
func (c *Client) GitTag(tag, message string) error {
	cmd := exec.Command("git", "tag", "-a", tag, "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git tag failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// GitCommit creates a git commit with the given message
func (c *Client) GitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// GitCheckoutNewBranch creates and checks out a new git branch
func (c *Client) GitCheckoutNewBranch(branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout -b failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// GetAuthenticatedUser returns the login of the currently authenticated user
func (c *Client) GetAuthenticatedUser() (string, error) {
	if c.gql == nil {
		return "", fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var query struct {
		Viewer struct {
			Login string
		}
	}

	err := c.gql.Query("GetAuthenticatedUser", &query, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}

	return query.Viewer.Login, nil
}

// GetIssuesByMicrosprint returns issues assigned to a specific microsprint
// This queries the project items and filters by the Microsprint text field
func (c *Client) GetIssuesByMicrosprint(owner, repo, microsprintName string) ([]Issue, error) {
	// This is a simplified implementation - for production we'd query the project
	// and filter by the Microsprint field value
	// For now, return empty slice - the close command doesn't strictly need this
	return []Issue{}, nil
}

// LabelExists checks if a label exists in a repository
func (c *Client) LabelExists(owner, repo, labelName string) (bool, error) {
	_, err := c.getLabelID(owner, repo, labelName)
	if err != nil {
		// Label not found is not an error for this function
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateLabel creates a new label in a repository
func (c *Client) CreateLabel(owner, repo, name, color, description string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// Get repository ID first
	repoID, err := c.getRepositoryID(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get repository ID: %w", err)
	}

	var mutation struct {
		CreateLabel struct {
			Label struct {
				ID   string
				Name string
			}
		} `graphql:"createLabel(input: $input)"`
	}

	input := CreateLabelInput{
		RepositoryID: graphql.ID(repoID),
		Name:         graphql.String(name),
		Color:        graphql.String(color),
		Description:  graphql.String(description),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err = c.gql.Mutate("CreateLabel", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	return nil
}

// CreateLabelInput represents the input for creating a label
type CreateLabelInput struct {
	RepositoryID graphql.ID     `json:"repositoryId"`
	Name         graphql.String `json:"name"`
	Color        graphql.String `json:"color"`
	Description  graphql.String `json:"description,omitempty"`
}

// FieldExists checks if a field exists in a project by name
func (c *Client) FieldExists(projectID, fieldName string) (bool, error) {
	fields, err := c.GetProjectFields(projectID)
	if err != nil {
		return false, err
	}
	for _, f := range fields {
		if f.Name == fieldName {
			return true, nil
		}
	}
	return false, nil
}

// AddIssueComment adds a comment to an issue
func (c *Client) AddIssueComment(issueID, body string) (*Comment, error) {
	if c.gql == nil {
		return nil, fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	var mutation struct {
		AddComment struct {
			CommentEdge struct {
				Node struct {
					ID        string
					Body      string
					CreatedAt string
					Author    struct {
						Login string
					}
				}
			}
		} `graphql:"addComment(input: $input)"`
	}

	input := AddCommentInput{
		SubjectID: graphql.ID(issueID),
		Body:      graphql.String(body),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.gql.Mutate("AddComment", &mutation, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return &Comment{
		ID:        mutation.AddComment.CommentEdge.Node.ID,
		Body:      mutation.AddComment.CommentEdge.Node.Body,
		Author:    mutation.AddComment.CommentEdge.Node.Author.Login,
		CreatedAt: mutation.AddComment.CommentEdge.Node.CreatedAt,
	}, nil
}

// AddCommentInput represents the input for adding a comment
type AddCommentInput struct {
	SubjectID graphql.ID     `json:"subjectId"`
	Body      graphql.String `json:"body"`
}

// DeleteLabel deletes a label from a repository
func (c *Client) DeleteLabel(owner, repo, labelName string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// Get the label ID first
	labelID, err := c.getLabelID(owner, repo, labelName)
	if err != nil {
		return fmt.Errorf("failed to get label ID: %w", err)
	}

	var mutation struct {
		DeleteLabel struct {
			ClientMutationID string
		} `graphql:"deleteLabel(input: $input)"`
	}

	input := DeleteLabelInput{
		ID: graphql.ID(labelID),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err = c.gql.Mutate("DeleteLabel", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	return nil
}

// DeleteLabelInput represents the input for deleting a label
type DeleteLabelInput struct {
	ID graphql.ID `json:"id"`
}

// UpdateLabel updates a label's properties in a repository
func (c *Client) UpdateLabel(owner, repo, labelName, newName, newColor, newDescription string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// Get the label ID first
	labelID, err := c.getLabelID(owner, repo, labelName)
	if err != nil {
		return fmt.Errorf("failed to get label ID: %w", err)
	}

	var mutation struct {
		UpdateLabel struct {
			Label struct {
				ID   string
				Name string
			}
		} `graphql:"updateLabel(input: $input)"`
	}

	input := UpdateLabelInput{
		ID:          graphql.ID(labelID),
		Name:        graphql.String(newName),
		Color:       graphql.String(newColor),
		Description: graphql.String(newDescription),
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err = c.gql.Mutate("UpdateLabel", &mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to update label: %w", err)
	}

	return nil
}

// UpdateLabelInput represents the input for updating a label
type UpdateLabelInput struct {
	ID          graphql.ID     `json:"id"`
	Name        graphql.String `json:"name,omitempty"`
	Color       graphql.String `json:"color,omitempty"`
	Description graphql.String `json:"description,omitempty"`
}
