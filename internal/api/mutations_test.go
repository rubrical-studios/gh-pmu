package api

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// ============================================================================
// Mock GraphQL Client for Testing
// ============================================================================

// mockGraphQLClient implements GraphQLClient interface for testing
type mockGraphQLClient struct {
	queryFunc  func(name string, query interface{}, variables map[string]interface{}) error
	mutateFunc func(name string, mutation interface{}, variables map[string]interface{}) error
}

func (m *mockGraphQLClient) Query(name string, query interface{}, variables map[string]interface{}) error {
	if m.queryFunc != nil {
		return m.queryFunc(name, query, variables)
	}
	return nil
}

func (m *mockGraphQLClient) Mutate(name string, mutation interface{}, variables map[string]interface{}) error {
	if m.mutateFunc != nil {
		return m.mutateFunc(name, mutation, variables)
	}
	return nil
}

// createMockWithField creates a mock that returns a project with a specific field type
func createMockWithField(fieldName, fieldType string, options []FieldOption) *mockGraphQLClient {
	return &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectFields" {
				// Use reflection to populate the query response
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				fields := projectV2.FieldByName("Fields")
				nodes := fields.FieldByName("Nodes")

				// Create a new slice with one element
				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)
				newNode := reflect.New(nodeType).Elem()

				// Set the typename
				if fieldType == "SINGLE_SELECT" {
					newNode.FieldByName("TypeName").SetString("ProjectV2SingleSelectField")
					singleSelect := newNode.FieldByName("ProjectV2SingleSelectField")
					singleSelect.FieldByName("ID").SetString("field-123")
					singleSelect.FieldByName("Name").SetString(fieldName)
					singleSelect.FieldByName("DataType").SetString(fieldType)

					// Set options
					if len(options) > 0 {
						optionsField := singleSelect.FieldByName("Options")
						optType := optionsField.Type().Elem()
						optSlice := reflect.MakeSlice(optionsField.Type(), len(options), len(options))
						for i, opt := range options {
							optVal := reflect.New(optType).Elem()
							optVal.FieldByName("ID").SetString(opt.ID)
							optVal.FieldByName("Name").SetString(opt.Name)
							optSlice.Index(i).Set(optVal)
						}
						optionsField.Set(optSlice)
					}
				} else {
					newNode.FieldByName("TypeName").SetString("ProjectV2Field")
					field := newNode.FieldByName("ProjectV2Field")
					field.FieldByName("ID").SetString("field-123")
					field.FieldByName("Name").SetString(fieldName)
					field.FieldByName("DataType").SetString(fieldType)
				}

				newNodes.Index(0).Set(newNode)
				nodes.Set(newNodes)
			}
			return nil
		},
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			return nil
		},
	}
}

// ============================================================================
// Nil Client Tests - All mutations should check for nil gql
// ============================================================================

func TestCreateIssue_NilClient(t *testing.T) {
	// Create client with nil gql
	client := &Client{gql: nil}

	_, err := client.CreateIssue("owner", "repo", "title", "body", nil)
	if err == nil {
		t.Fatal("Expected error when gql is nil")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected 'GraphQL client not initialized' error, got: %v", err)
	}
}

func TestAddIssueToProject_NilClient(t *testing.T) {
	client := &Client{gql: nil}

	_, err := client.AddIssueToProject("proj-id", "issue-id")
	if err == nil {
		t.Fatal("Expected error when gql is nil")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected 'GraphQL client not initialized' error, got: %v", err)
	}
}

func TestSetProjectItemField_NilClient(t *testing.T) {
	client := &Client{gql: nil}

	err := client.SetProjectItemField("proj-id", "item-id", "Status", "Done")
	if err == nil {
		t.Fatal("Expected error when gql is nil")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected 'GraphQL client not initialized' error, got: %v", err)
	}
}

func TestAddSubIssue_NilClient(t *testing.T) {
	client := &Client{gql: nil}

	err := client.AddSubIssue("parent-id", "child-id")
	if err == nil {
		t.Fatal("Expected error when gql is nil")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected 'GraphQL client not initialized' error, got: %v", err)
	}
}

func TestRemoveSubIssue_NilClient(t *testing.T) {
	client := &Client{gql: nil}

	err := client.RemoveSubIssue("parent-id", "child-id")
	if err == nil {
		t.Fatal("Expected error when gql is nil")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected 'GraphQL client not initialized' error, got: %v", err)
	}
}

func TestAddLabelToIssue_NilClient(t *testing.T) {
	client := &Client{gql: nil}

	err := client.AddLabelToIssue("issue-id", "bug")
	if err == nil {
		t.Fatal("Expected error when gql is nil")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected 'GraphQL client not initialized' error, got: %v", err)
	}
}

// ============================================================================
// SetProjectItemField Tests with Mocking
// ============================================================================

func TestSetProjectItemField_FieldNotFound(t *testing.T) {
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			// Return empty fields - no matching field will be found
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	err := client.SetProjectItemField("proj-id", "item-id", "NonExistentField", "value")

	if err == nil {
		t.Fatal("Expected error when field not found")
	}
	if !strings.Contains(err.Error(), "field \"NonExistentField\" not found") {
		t.Errorf("Expected 'field not found' error, got: %v", err)
	}
}

func TestSetProjectItemField_GetFieldsError(t *testing.T) {
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("network error")
		},
	}

	client := NewClientWithGraphQL(mock)
	err := client.SetProjectItemField("proj-id", "item-id", "Status", "Done")

	if err == nil {
		t.Fatal("Expected error when GetProjectFields fails")
	}
	if !strings.Contains(err.Error(), "failed to get project fields") {
		t.Errorf("Expected 'failed to get project fields' error, got: %v", err)
	}
}

func TestSetProjectItemField_SingleSelectField_Success(t *testing.T) {
	options := []FieldOption{
		{ID: "opt-1", Name: "Todo"},
		{ID: "opt-2", Name: "In Progress"},
		{ID: "opt-3", Name: "Done"},
	}
	mock := createMockWithField("Status", "SINGLE_SELECT", options)

	client := NewClientWithGraphQL(mock)
	err := client.SetProjectItemField("proj-id", "item-id", "Status", "Done")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestSetProjectItemField_SingleSelectField_OptionNotFound(t *testing.T) {
	options := []FieldOption{
		{ID: "opt-1", Name: "Todo"},
		{ID: "opt-2", Name: "Done"},
	}
	mock := createMockWithField("Status", "SINGLE_SELECT", options)

	client := NewClientWithGraphQL(mock)
	err := client.SetProjectItemField("proj-id", "item-id", "Status", "Invalid Option")

	if err == nil {
		t.Fatal("Expected error when option not found")
	}
	if !strings.Contains(err.Error(), "option \"Invalid Option\" not found") {
		t.Errorf("Expected 'option not found' error, got: %v", err)
	}
}

func TestSetProjectItemField_TextField_Success(t *testing.T) {
	mock := createMockWithField("Notes", "TEXT", nil)

	client := NewClientWithGraphQL(mock)
	err := client.SetProjectItemField("proj-id", "item-id", "Notes", "Some notes")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestSetProjectItemField_NumberField_Success(t *testing.T) {
	mock := createMockWithField("Points", "NUMBER", nil)

	client := NewClientWithGraphQL(mock)
	err := client.SetProjectItemField("proj-id", "item-id", "Points", "5")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestSetProjectItemField_UnsupportedFieldType(t *testing.T) {
	mock := createMockWithField("Date", "DATE", nil)

	client := NewClientWithGraphQL(mock)
	err := client.SetProjectItemField("proj-id", "item-id", "Date", "2024-01-15")

	if err == nil {
		t.Fatal("Expected error for unsupported field type")
	}
	if !strings.Contains(err.Error(), "unsupported field type") {
		t.Errorf("Expected 'unsupported field type' error, got: %v", err)
	}
}

func TestSetProjectItemField_MutationError(t *testing.T) {
	mock := createMockWithField("Notes", "TEXT", nil)
	mock.mutateFunc = func(name string, mutation interface{}, variables map[string]interface{}) error {
		return errors.New("mutation failed")
	}

	client := NewClientWithGraphQL(mock)
	err := client.SetProjectItemField("proj-id", "item-id", "Notes", "Some notes")

	if err == nil {
		t.Fatal("Expected error when mutation fails")
	}
	if !strings.Contains(err.Error(), "failed to set") {
		t.Errorf("Expected 'failed to set' error, got: %v", err)
	}
}

// ============================================================================
// AddIssueToProject Tests with Mocking
// ============================================================================

func TestAddIssueToProject_Success(t *testing.T) {
	mock := &mockGraphQLClient{
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			// Verify the mutation name
			if name != "AddProjectV2ItemById" {
				t.Errorf("Expected mutation name 'AddProjectV2ItemById', got '%s'", name)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	_, err := client.AddIssueToProject("proj-id", "issue-id")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAddIssueToProject_MutationError(t *testing.T) {
	mock := &mockGraphQLClient{
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			return errors.New("mutation failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	_, err := client.AddIssueToProject("proj-id", "issue-id")

	if err == nil {
		t.Fatal("Expected error when mutation fails")
	}
	if !strings.Contains(err.Error(), "failed to add issue to project") {
		t.Errorf("Expected 'failed to add issue to project' error, got: %v", err)
	}
}

// ============================================================================
// AddSubIssue Tests with Mocking
// ============================================================================

func TestAddSubIssue_Success(t *testing.T) {
	mock := &mockGraphQLClient{
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			if name != "AddSubIssue" {
				t.Errorf("Expected mutation name 'AddSubIssue', got '%s'", name)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	err := client.AddSubIssue("parent-id", "child-id")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAddSubIssue_MutationError(t *testing.T) {
	mock := &mockGraphQLClient{
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			return errors.New("mutation failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	err := client.AddSubIssue("parent-id", "child-id")

	if err == nil {
		t.Fatal("Expected error when mutation fails")
	}
	if !strings.Contains(err.Error(), "failed to add sub-issue") {
		t.Errorf("Expected 'failed to add sub-issue' error, got: %v", err)
	}
}

// ============================================================================
// RemoveSubIssue Tests with Mocking
// ============================================================================

func TestRemoveSubIssue_Success(t *testing.T) {
	mock := &mockGraphQLClient{
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			if name != "RemoveSubIssue" {
				t.Errorf("Expected mutation name 'RemoveSubIssue', got '%s'", name)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	err := client.RemoveSubIssue("parent-id", "child-id")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestRemoveSubIssue_MutationError(t *testing.T) {
	mock := &mockGraphQLClient{
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			return errors.New("mutation failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	err := client.RemoveSubIssue("parent-id", "child-id")

	if err == nil {
		t.Fatal("Expected error when mutation fails")
	}
	if !strings.Contains(err.Error(), "failed to remove sub-issue") {
		t.Errorf("Expected 'failed to remove sub-issue' error, got: %v", err)
	}
}

// ============================================================================
// CreateIssue Tests with Mocking
// ============================================================================

func TestCreateIssue_GetRepositoryIDError(t *testing.T) {
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("repo not found")
		},
	}

	client := NewClientWithGraphQL(mock)
	_, err := client.CreateIssue("owner", "repo", "title", "body", nil)

	if err == nil {
		t.Fatal("Expected error when getRepositoryID fails")
	}
	if !strings.Contains(err.Error(), "failed to get repository ID") {
		t.Errorf("Expected 'failed to get repository ID' error, got: %v", err)
	}
}

func TestCreateIssue_MutationError(t *testing.T) {
	callCount := 0
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			// First call is getRepositoryID - succeed
			return nil
		},
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			callCount++
			return errors.New("create issue failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	_, err := client.CreateIssue("owner", "repo", "title", "body", nil)

	if err == nil {
		t.Fatal("Expected error when mutation fails")
	}
	if !strings.Contains(err.Error(), "failed to create issue") {
		t.Errorf("Expected 'failed to create issue' error, got: %v", err)
	}
}

func TestCreateIssue_Success(t *testing.T) {
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return nil
		},
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			if name != "CreateIssue" {
				t.Errorf("Expected mutation name 'CreateIssue', got '%s'", name)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	issue, err := client.CreateIssue("owner", "repo", "title", "body", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if issue == nil {
		t.Fatal("Expected issue to be returned")
	}
	// The issue will have empty fields since our mock doesn't populate them
	if issue.Repository.Owner != "owner" {
		t.Errorf("Expected owner 'owner', got '%s'", issue.Repository.Owner)
	}
	if issue.Repository.Name != "repo" {
		t.Errorf("Expected repo 'repo', got '%s'", issue.Repository.Name)
	}
}

func TestCreateIssue_WithLabels_SkipsInvalidLabels(t *testing.T) {
	queryCount := 0
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			queryCount++
			if name == "GetLabelID" {
				// Label lookups fail
				return errors.New("label not found")
			}
			// getRepositoryID succeeds
			return nil
		},
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	_, err := client.CreateIssue("owner", "repo", "title", "body", []string{"bug", "enhancement"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Should have called GetRepositoryID once and GetLabelID twice
	if queryCount != 3 {
		t.Errorf("Expected 3 query calls (1 repo + 2 labels), got %d", queryCount)
	}
}

// ============================================================================
// getLabelID Tests with Mocking
// ============================================================================

func TestGetLabelID_Success(t *testing.T) {
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetLabelID" {
				// Use reflection to populate the label ID
				v := reflect.ValueOf(query).Elem()
				repo := v.FieldByName("Repository")
				label := repo.FieldByName("Label")
				label.FieldByName("ID").SetString("label-123")
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	labelID, err := client.getLabelID("owner", "repo", "bug")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if labelID != "label-123" {
		t.Errorf("Expected label ID 'label-123', got '%s'", labelID)
	}
}

func TestGetLabelID_QueryError(t *testing.T) {
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("network error")
		},
	}

	client := NewClientWithGraphQL(mock)
	_, err := client.getLabelID("owner", "repo", "bug")

	if err == nil {
		t.Fatal("Expected error when query fails")
	}
	if !strings.Contains(err.Error(), "failed to get label ID") {
		t.Errorf("Expected 'failed to get label ID' error, got: %v", err)
	}
}

func TestGetLabelID_LabelNotFound(t *testing.T) {
	mock := &mockGraphQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			// Don't populate the label ID - leave it empty
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	_, err := client.getLabelID("owner", "repo", "nonexistent")

	if err == nil {
		t.Fatal("Expected error when label not found")
	}
	if !strings.Contains(err.Error(), "label \"nonexistent\" not found") {
		t.Errorf("Expected 'label not found' error, got: %v", err)
	}
}

// ============================================================================
// Input Type Tests - Verify structs have correct fields
// ============================================================================

func TestCreateIssueInput_HasRequiredFields(t *testing.T) {
	// Verify the struct can be created with expected fields
	input := CreateIssueInput{
		RepositoryID: "repo-id",
		Title:        "Test Issue",
		Body:         "Test body",
	}

	if input.RepositoryID != "repo-id" {
		t.Errorf("Expected RepositoryID 'repo-id', got '%s'", input.RepositoryID)
	}
	if input.Title != "Test Issue" {
		t.Errorf("Expected Title 'Test Issue', got '%s'", input.Title)
	}
}

func TestAddProjectV2ItemByIdInput_HasRequiredFields(t *testing.T) {
	input := AddProjectV2ItemByIdInput{
		ProjectID: "proj-id",
		ContentID: "content-id",
	}

	if input.ProjectID != "proj-id" {
		t.Errorf("Expected ProjectID 'proj-id', got '%s'", input.ProjectID)
	}
	if input.ContentID != "content-id" {
		t.Errorf("Expected ContentID 'content-id', got '%s'", input.ContentID)
	}
}

func TestUpdateProjectV2ItemFieldValueInput_HasRequiredFields(t *testing.T) {
	input := UpdateProjectV2ItemFieldValueInput{
		ProjectID: "proj-id",
		ItemID:    "item-id",
		FieldID:   "field-id",
		Value: ProjectV2FieldValue{
			Text: "test value",
		},
	}

	if input.ProjectID != "proj-id" {
		t.Errorf("Expected ProjectID 'proj-id', got '%s'", input.ProjectID)
	}
	if input.ItemID != "item-id" {
		t.Errorf("Expected ItemID 'item-id', got '%s'", input.ItemID)
	}
	if input.FieldID != "field-id" {
		t.Errorf("Expected FieldID 'field-id', got '%s'", input.FieldID)
	}
	if input.Value.Text != "test value" {
		t.Errorf("Expected Value.Text 'test value', got '%s'", input.Value.Text)
	}
}

func TestProjectV2FieldValue_AllFieldTypes(t *testing.T) {
	// Test that all field types can be set
	textValue := ProjectV2FieldValue{Text: "text"}
	if textValue.Text != "text" {
		t.Errorf("Expected Text 'text', got '%s'", textValue.Text)
	}

	numberValue := ProjectV2FieldValue{Number: 42.5}
	if numberValue.Number != 42.5 {
		t.Errorf("Expected Number 42.5, got %f", numberValue.Number)
	}

	dateValue := ProjectV2FieldValue{Date: "2024-01-15"}
	if dateValue.Date != "2024-01-15" {
		t.Errorf("Expected Date '2024-01-15', got '%s'", dateValue.Date)
	}

	selectValue := ProjectV2FieldValue{SingleSelectOptionId: "option-id"}
	if selectValue.SingleSelectOptionId != "option-id" {
		t.Errorf("Expected SingleSelectOptionId 'option-id', got '%s'", selectValue.SingleSelectOptionId)
	}

	iterValue := ProjectV2FieldValue{IterationId: "iter-id"}
	if iterValue.IterationId != "iter-id" {
		t.Errorf("Expected IterationId 'iter-id', got '%s'", iterValue.IterationId)
	}
}

func TestAddSubIssueInput_HasRequiredFields(t *testing.T) {
	input := AddSubIssueInput{
		IssueID:    "parent-id",
		SubIssueID: "child-id",
	}

	if input.IssueID != "parent-id" {
		t.Errorf("Expected IssueID 'parent-id', got '%s'", input.IssueID)
	}
	if input.SubIssueID != "child-id" {
		t.Errorf("Expected SubIssueID 'child-id', got '%s'", input.SubIssueID)
	}
}

func TestRemoveSubIssueInput_HasRequiredFields(t *testing.T) {
	input := RemoveSubIssueInput{
		IssueID:    "parent-id",
		SubIssueID: "child-id",
	}

	if input.IssueID != "parent-id" {
		t.Errorf("Expected IssueID 'parent-id', got '%s'", input.IssueID)
	}
	if input.SubIssueID != "child-id" {
		t.Errorf("Expected SubIssueID 'child-id', got '%s'", input.SubIssueID)
	}
}

// ============================================================================
// CreateIssueInput Optional Fields Tests
// ============================================================================

func TestCreateIssueInput_OptionalFields(t *testing.T) {
	// Test with optional fields set
	labelIDs := []interface{}{"label-1", "label-2"}
	milestoneID := interface{}("milestone-id")

	input := CreateIssueInput{
		RepositoryID: "repo-id",
		Title:        "Test Issue",
		Body:         "Test body",
	}

	// Labels are optional
	if input.LabelIDs != nil {
		t.Error("Expected LabelIDs to be nil by default")
	}

	// Test setting labels
	labels := make([]interface{}, len(labelIDs))
	copy(labels, labelIDs)
	// Note: The actual type is *[]graphql.ID, this is just struct verification

	// Milestone is optional
	if input.MilestoneID != nil {
		t.Error("Expected MilestoneID to be nil by default")
	}
	_ = milestoneID // Verify it can be assigned
}
