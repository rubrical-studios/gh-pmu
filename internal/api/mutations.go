package api

import (
	"fmt"

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
	RepositoryID graphql.ID      `json:"repositoryId"`
	Title        graphql.String  `json:"title"`
	Body         graphql.String  `json:"body,omitempty"`
	LabelIDs     *[]graphql.ID   `json:"labelIds,omitempty"`
	AssigneeIDs  *[]graphql.ID   `json:"assigneeIds,omitempty"`
	MilestoneID  *graphql.ID     `json:"milestoneId,omitempty"`
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

// SetProjectItemField sets a field value on a project item
func (c *Client) SetProjectItemField(projectID, itemID, fieldName, value string) error {
	if c.gql == nil {
		return fmt.Errorf("GraphQL client not initialized - are you authenticated with gh?")
	}

	// Get the field ID and option ID for single select fields
	fields, err := c.GetProjectFields(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project fields: %w", err)
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
	ProjectID graphql.ID           `json:"projectId"`
	ItemID    graphql.ID           `json:"itemId"`
	FieldID   graphql.ID           `json:"fieldId"`
	Value     ProjectV2FieldValue  `json:"value"`
}

// ProjectV2FieldValue represents a field value for a project item
type ProjectV2FieldValue struct {
	Text                 graphql.String  `json:"text,omitempty"`
	Number               graphql.Float   `json:"number,omitempty"`
	Date                 graphql.String  `json:"date,omitempty"`
	SingleSelectOptionId graphql.String  `json:"singleSelectOptionId,omitempty"`
	IterationId          graphql.String  `json:"iterationId,omitempty"`
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
