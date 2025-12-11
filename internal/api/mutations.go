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
	RepositoryID graphql.ID     `json:"repositoryId"`
	Title        graphql.String `json:"title"`
	Body         graphql.String `json:"body,omitempty"`
	LabelIDs     *[]graphql.ID  `json:"labelIds,omitempty"`
	AssigneeIDs  *[]graphql.ID  `json:"assigneeIds,omitempty"`
	MilestoneID  *graphql.ID    `json:"milestoneId,omitempty"`
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
			}
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
