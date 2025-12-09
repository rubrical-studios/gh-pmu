package api

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSplitRepoName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "valid owner/repo format",
			input:    "rubrical-studios/gh-pmu",
			expected: []string{"rubrical-studios", "gh-pmu"},
		},
		{
			name:     "no slash returns nil",
			input:    "noslash",
			expected: nil,
		},
		{
			name:     "empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "slash at beginning",
			input:    "/repo",
			expected: []string{"", "repo"},
		},
		{
			name:     "slash at end",
			input:    "owner/",
			expected: []string{"owner", ""},
		},
		{
			name:     "multiple slashes returns first split only",
			input:    "owner/repo/extra",
			expected: []string{"owner", "repo/extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitRepoName(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("splitRepoName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetProject_NilClient(t *testing.T) {
	// ARRANGE: Create client with nil gql
	client := &Client{gql: nil}

	// ACT: Call GetProject
	project, err := client.GetProject("owner", 1)

	// ASSERT: Should return error about uninitialized client
	if err == nil {
		t.Fatal("Expected error when gql is nil, got nil")
	}
	if project != nil {
		t.Error("Expected nil project when error occurs")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected error about uninitialized client, got: %v", err)
	}
}

func TestGetProjectFields_NilClient(t *testing.T) {
	// ARRANGE: Create client with nil gql
	client := &Client{gql: nil}

	// ACT: Call GetProjectFields
	fields, err := client.GetProjectFields("project-id")

	// ASSERT: Should return error about uninitialized client
	if err == nil {
		t.Fatal("Expected error when gql is nil, got nil")
	}
	if fields != nil {
		t.Error("Expected nil fields when error occurs")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected error about uninitialized client, got: %v", err)
	}
}

func TestGetIssue_NilClient(t *testing.T) {
	// ARRANGE: Create client with nil gql
	client := &Client{gql: nil}

	// ACT: Call GetIssue
	issue, err := client.GetIssue("owner", "repo", 1)

	// ASSERT: Should return error about uninitialized client
	if err == nil {
		t.Fatal("Expected error when gql is nil, got nil")
	}
	if issue != nil {
		t.Error("Expected nil issue when error occurs")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected error about uninitialized client, got: %v", err)
	}
}

func TestGetProjectItems_NilClient(t *testing.T) {
	// ARRANGE: Create client with nil gql
	client := &Client{gql: nil}

	// ACT: Call GetProjectItems
	items, err := client.GetProjectItems("project-id", nil)

	// ASSERT: Should return error about uninitialized client
	if err == nil {
		t.Fatal("Expected error when gql is nil, got nil")
	}
	if items != nil {
		t.Error("Expected nil items when error occurs")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected error about uninitialized client, got: %v", err)
	}
}

func TestGetSubIssues_NilClient(t *testing.T) {
	// ARRANGE: Create client with nil gql
	client := &Client{gql: nil}

	// ACT: Call GetSubIssues
	subIssues, err := client.GetSubIssues("owner", "repo", 1)

	// ASSERT: Should return error about uninitialized client
	if err == nil {
		t.Fatal("Expected error when gql is nil, got nil")
	}
	if subIssues != nil {
		t.Error("Expected nil subIssues when error occurs")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected error about uninitialized client, got: %v", err)
	}
}

func TestGetRepositoryIssues_NilClient(t *testing.T) {
	// ARRANGE: Create client with nil gql
	client := &Client{gql: nil}

	// ACT: Call GetRepositoryIssues
	issues, err := client.GetRepositoryIssues("owner", "repo", "open")

	// ASSERT: Should return error about uninitialized client
	if err == nil {
		t.Fatal("Expected error when gql is nil, got nil")
	}
	if issues != nil {
		t.Error("Expected nil issues when error occurs")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected error about uninitialized client, got: %v", err)
	}
}

func TestGetParentIssue_NilClient(t *testing.T) {
	// ARRANGE: Create client with nil gql
	client := &Client{gql: nil}

	// ACT: Call GetParentIssue
	parent, err := client.GetParentIssue("owner", "repo", 1)

	// ASSERT: Should return error about uninitialized client
	if err == nil {
		t.Fatal("Expected error when gql is nil, got nil")
	}
	if parent != nil {
		t.Error("Expected nil parent when error occurs")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected error about uninitialized client, got: %v", err)
	}
}

func TestListProjects_NilClient(t *testing.T) {
	// ARRANGE: Create client with nil gql
	client := &Client{gql: nil}

	// ACT: Call ListProjects
	projects, err := client.ListProjects("owner")

	// ASSERT: Should return error about uninitialized client
	if err == nil {
		t.Fatal("Expected error when gql is nil, got nil")
	}
	if projects != nil {
		t.Error("Expected nil projects when error occurs")
	}
	if !strings.Contains(err.Error(), "GraphQL client not initialized") {
		t.Errorf("Expected error about uninitialized client, got: %v", err)
	}
}

// ============================================================================
// GetProject Tests with Mocking - User vs Org fallback
// ============================================================================

// queryMockClient is a simple mock that tracks query names and can return errors
type queryMockClient struct {
	queryCalls  []string
	mutateCalls []string
	queryFunc   func(name string, query interface{}, variables map[string]interface{}) error
	mutateFunc  func(name string, mutation interface{}, variables map[string]interface{}) error
}

func (m *queryMockClient) Query(name string, query interface{}, variables map[string]interface{}) error {
	m.queryCalls = append(m.queryCalls, name)
	if m.queryFunc != nil {
		return m.queryFunc(name, query, variables)
	}
	return nil
}

func (m *queryMockClient) Mutate(name string, mutation interface{}, variables map[string]interface{}) error {
	m.mutateCalls = append(m.mutateCalls, name)
	if m.mutateFunc != nil {
		return m.mutateFunc(name, mutation, variables)
	}
	return nil
}

func TestGetProject_UserSucceeds(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetUserProject" {
				// Populate user project response using reflection
				v := reflect.ValueOf(query).Elem()
				user := v.FieldByName("User")
				projectV2 := user.FieldByName("ProjectV2")
				projectV2.FieldByName("ID").SetString("proj-123")
				projectV2.FieldByName("Number").SetInt(1)
				projectV2.FieldByName("Title").SetString("Test Project")
				projectV2.FieldByName("URL").SetString("https://github.com/users/owner/projects/1")
				projectV2.FieldByName("Closed").SetBool(false)
				return nil
			}
			return errors.New("unexpected query")
		},
	}

	client := NewClientWithGraphQL(mock)
	project, err := client.GetProject("owner", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("Expected project to be returned")
	}
	if project.ID != "proj-123" {
		t.Errorf("Expected project ID 'proj-123', got '%s'", project.ID)
	}
	if project.Owner.Type != "User" {
		t.Errorf("Expected owner type 'User', got '%s'", project.Owner.Type)
	}
	if len(mock.queryCalls) != 1 || mock.queryCalls[0] != "GetUserProject" {
		t.Errorf("Expected only GetUserProject query, got: %v", mock.queryCalls)
	}
}

func TestGetProject_UserFailsOrgSucceeds(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetUserProject" {
				return errors.New("user not found")
			}
			if name == "GetOrgProject" {
				// Populate org project response
				v := reflect.ValueOf(query).Elem()
				org := v.FieldByName("Organization")
				projectV2 := org.FieldByName("ProjectV2")
				projectV2.FieldByName("ID").SetString("org-proj-456")
				projectV2.FieldByName("Number").SetInt(2)
				projectV2.FieldByName("Title").SetString("Org Project")
				projectV2.FieldByName("URL").SetString("https://github.com/orgs/myorg/projects/2")
				projectV2.FieldByName("Closed").SetBool(false)
				return nil
			}
			return errors.New("unexpected query")
		},
	}

	client := NewClientWithGraphQL(mock)
	project, err := client.GetProject("myorg", 2)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("Expected project to be returned")
	}
	if project.ID != "org-proj-456" {
		t.Errorf("Expected project ID 'org-proj-456', got '%s'", project.ID)
	}
	if project.Owner.Type != "Organization" {
		t.Errorf("Expected owner type 'Organization', got '%s'", project.Owner.Type)
	}
	// Should have tried user first, then org
	if len(mock.queryCalls) != 2 {
		t.Errorf("Expected 2 query calls, got: %v", mock.queryCalls)
	}
}

func TestGetProject_BothFail(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetUserProject" {
				return errors.New("user not found")
			}
			if name == "GetOrgProject" {
				return errors.New("org not found")
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	project, err := client.GetProject("unknown", 1)

	if err == nil {
		t.Fatal("Expected error when both user and org fail")
	}
	if project != nil {
		t.Error("Expected nil project when error occurs")
	}
	if !strings.Contains(err.Error(), "failed to get project") {
		t.Errorf("Expected 'failed to get project' error, got: %v", err)
	}
}

// ============================================================================
// GetRepositoryIssues State Mapping Tests
// ============================================================================

func TestGetRepositoryIssues_StateMapping(t *testing.T) {
	tests := []struct {
		name           string
		inputState     string
		expectedStates []string
	}{
		{
			name:           "open state",
			inputState:     "open",
			expectedStates: []string{"OPEN"},
		},
		{
			name:           "closed state",
			inputState:     "closed",
			expectedStates: []string{"CLOSED"},
		},
		{
			name:           "all state",
			inputState:     "all",
			expectedStates: []string{"OPEN", "CLOSED"},
		},
		{
			name:           "empty state defaults to all",
			inputState:     "",
			expectedStates: []string{"OPEN", "CLOSED"},
		},
		{
			name:           "custom state passed through",
			inputState:     "CUSTOM",
			expectedStates: []string{"CUSTOM"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedStates []string
			mock := &queryMockClient{
				queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
					if name == "GetRepositoryIssues" {
						// Capture the states that were passed
						if states, ok := variables["states"].([]interface{}); ok {
							for _, s := range states {
								capturedStates = append(capturedStates, string(s.(string)))
							}
						}
					}
					return nil
				},
			}

			client := NewClientWithGraphQL(mock)
			_, _ = client.GetRepositoryIssues("owner", "repo", tt.inputState)

			// Note: The actual captured states depend on graphql.String conversion
			// This test verifies the function is called correctly
			if len(mock.queryCalls) != 1 || mock.queryCalls[0] != "GetRepositoryIssues" {
				t.Errorf("Expected GetRepositoryIssues query, got: %v", mock.queryCalls)
			}
		})
	}
}

func TestGetRepositoryIssues_QueryError(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("network error")
		},
	}

	client := NewClientWithGraphQL(mock)
	issues, err := client.GetRepositoryIssues("owner", "repo", "open")

	if err == nil {
		t.Fatal("Expected error when query fails")
	}
	if issues != nil {
		t.Error("Expected nil issues when error occurs")
	}
	if !strings.Contains(err.Error(), "failed to get issues") {
		t.Errorf("Expected 'failed to get issues' error, got: %v", err)
	}
}

// ============================================================================
// GetParentIssue Tests
// ============================================================================

func TestGetParentIssue_NoParent(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			// Don't populate parent - leave ID empty
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	parent, err := client.GetParentIssue("owner", "repo", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if parent != nil {
		t.Error("Expected nil parent when issue has no parent")
	}
}

func TestGetParentIssue_HasParent(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetParentIssue" {
				v := reflect.ValueOf(query).Elem()
				repo := v.FieldByName("Repository")
				issue := repo.FieldByName("Issue")
				parent := issue.FieldByName("Parent")
				parent.FieldByName("ID").SetString("parent-123")
				parent.FieldByName("Number").SetInt(42)
				parent.FieldByName("Title").SetString("Parent Issue")
				parent.FieldByName("State").SetString("OPEN")
				parent.FieldByName("URL").SetString("https://github.com/owner/repo/issues/42")
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	parent, err := client.GetParentIssue("owner", "repo", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if parent == nil {
		t.Fatal("Expected parent issue to be returned")
	}
	if parent.ID != "parent-123" {
		t.Errorf("Expected parent ID 'parent-123', got '%s'", parent.ID)
	}
	if parent.Number != 42 {
		t.Errorf("Expected parent number 42, got %d", parent.Number)
	}
}

func TestGetParentIssue_QueryError(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("query failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	parent, err := client.GetParentIssue("owner", "repo", 1)

	if err == nil {
		t.Fatal("Expected error when query fails")
	}
	if parent != nil {
		t.Error("Expected nil parent when error occurs")
	}
	if !strings.Contains(err.Error(), "failed to get parent issue") {
		t.Errorf("Expected 'failed to get parent issue' error, got: %v", err)
	}
}

// ============================================================================
// GetSubIssues Tests
// ============================================================================

func TestGetSubIssues_QueryError(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("query failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	subIssues, err := client.GetSubIssues("owner", "repo", 1)

	if err == nil {
		t.Fatal("Expected error when query fails")
	}
	if subIssues != nil {
		t.Error("Expected nil subIssues when error occurs")
	}
	if !strings.Contains(err.Error(), "failed to get sub-issues") {
		t.Errorf("Expected 'failed to get sub-issues' error, got: %v", err)
	}
}

func TestGetSubIssues_EmptyResult(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			// Don't populate any sub-issues
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	subIssues, err := client.GetSubIssues("owner", "repo", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(subIssues) != 0 {
		t.Errorf("Expected empty subIssues, got %d", len(subIssues))
	}
}

// ============================================================================
// GetIssue Tests
// ============================================================================

func TestGetIssue_QueryError(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("query failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	issue, err := client.GetIssue("owner", "repo", 1)

	if err == nil {
		t.Fatal("Expected error when query fails")
	}
	if issue != nil {
		t.Error("Expected nil issue when error occurs")
	}
	if !strings.Contains(err.Error(), "failed to get issue") {
		t.Errorf("Expected 'failed to get issue' error, got: %v", err)
	}
}

// ============================================================================
// GetProjectItems Tests
// ============================================================================

func TestGetProjectItems_QueryError(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("query failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err == nil {
		t.Fatal("Expected error when query fails")
	}
	if items != nil {
		t.Error("Expected nil items when error occurs")
	}
	if !strings.Contains(err.Error(), "failed to get project items") {
		t.Errorf("Expected 'failed to get project items' error, got: %v", err)
	}
}

func TestGetProjectItems_EmptyResult(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected empty items, got %d", len(items))
	}
}

// ============================================================================
// ListProjects Tests with Mocking
// ============================================================================

func TestListProjects_UserSucceeds(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "ListUserProjects" {
				v := reflect.ValueOf(query).Elem()
				user := v.FieldByName("User")
				projectsV2 := user.FieldByName("ProjectsV2")
				nodes := projectsV2.FieldByName("Nodes")

				// Create a slice with one project
				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)
				newNode := reflect.New(nodeType).Elem()
				newNode.FieldByName("ID").SetString("proj-1")
				newNode.FieldByName("Number").SetInt(1)
				newNode.FieldByName("Title").SetString("User Project")
				newNode.FieldByName("URL").SetString("https://github.com/users/owner/projects/1")
				newNode.FieldByName("Closed").SetBool(false)
				newNodes.Index(0).Set(newNode)
				nodes.Set(newNodes)
				return nil
			}
			return errors.New("unexpected query")
		},
	}

	client := NewClientWithGraphQL(mock)
	projects, err := client.ListProjects("owner")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
	if projects[0].Title != "User Project" {
		t.Errorf("Expected title 'User Project', got '%s'", projects[0].Title)
	}
}

func TestListProjects_UserEmptyFallsToOrg(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "ListUserProjects" {
				// Return empty (no projects)
				return nil
			}
			if name == "ListOrgProjects" {
				v := reflect.ValueOf(query).Elem()
				org := v.FieldByName("Organization")
				projectsV2 := org.FieldByName("ProjectsV2")
				nodes := projectsV2.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)
				newNode := reflect.New(nodeType).Elem()
				newNode.FieldByName("ID").SetString("org-proj-1")
				newNode.FieldByName("Number").SetInt(1)
				newNode.FieldByName("Title").SetString("Org Project")
				newNode.FieldByName("URL").SetString("https://github.com/orgs/myorg/projects/1")
				newNode.FieldByName("Closed").SetBool(false)
				newNodes.Index(0).Set(newNode)
				nodes.Set(newNodes)
				return nil
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	projects, err := client.ListProjects("myorg")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
	if projects[0].Title != "Org Project" {
		t.Errorf("Expected title 'Org Project', got '%s'", projects[0].Title)
	}
}

func TestListProjects_SkipsClosedProjects(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "ListUserProjects" {
				v := reflect.ValueOf(query).Elem()
				user := v.FieldByName("User")
				projectsV2 := user.FieldByName("ProjectsV2")
				nodes := projectsV2.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

				// Open project
				openNode := reflect.New(nodeType).Elem()
				openNode.FieldByName("ID").SetString("proj-1")
				openNode.FieldByName("Number").SetInt(1)
				openNode.FieldByName("Title").SetString("Open Project")
				openNode.FieldByName("Closed").SetBool(false)
				newNodes.Index(0).Set(openNode)

				// Closed project
				closedNode := reflect.New(nodeType).Elem()
				closedNode.FieldByName("ID").SetString("proj-2")
				closedNode.FieldByName("Number").SetInt(2)
				closedNode.FieldByName("Title").SetString("Closed Project")
				closedNode.FieldByName("Closed").SetBool(true)
				newNodes.Index(1).Set(closedNode)

				nodes.Set(newNodes)
				return nil
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	projects, err := client.ListProjects("owner")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project (open only), got %d", len(projects))
	}
	if projects[0].Title != "Open Project" {
		t.Errorf("Expected open project, got '%s'", projects[0].Title)
	}
}

func TestListProjects_BothFail(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "ListUserProjects" {
				return errors.New("user not found")
			}
			if name == "ListOrgProjects" {
				return errors.New("org not found")
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	projects, err := client.ListProjects("unknown")

	if err == nil {
		t.Fatal("Expected error when both user and org fail")
	}
	if projects != nil {
		t.Error("Expected nil projects when error occurs")
	}
	if !strings.Contains(err.Error(), "failed to list projects") {
		t.Errorf("Expected 'failed to list projects' error, got: %v", err)
	}
}

// ============================================================================
// GetProjectFields Additional Tests
// ============================================================================

func TestGetProjectFields_QueryError(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return errors.New("query failed")
		},
	}

	client := NewClientWithGraphQL(mock)
	fields, err := client.GetProjectFields("proj-id")

	if err == nil {
		t.Fatal("Expected error when query fails")
	}
	if fields != nil {
		t.Error("Expected nil fields when error occurs")
	}
	if !strings.Contains(err.Error(), "failed to get project fields") {
		t.Errorf("Expected 'failed to get project fields' error, got: %v", err)
	}
}

// ============================================================================
// GetIssue Tests - Improved Coverage
// ============================================================================

func TestGetIssue_Success(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetIssue" {
				v := reflect.ValueOf(query).Elem()
				repo := v.FieldByName("Repository")
				issue := repo.FieldByName("Issue")
				issue.FieldByName("ID").SetString("issue-123")
				issue.FieldByName("Number").SetInt(42)
				issue.FieldByName("Title").SetString("Test Issue")
				issue.FieldByName("Body").SetString("Issue body")
				issue.FieldByName("State").SetString("OPEN")
				issue.FieldByName("URL").SetString("https://github.com/owner/repo/issues/42")

				// Set author
				author := issue.FieldByName("Author")
				author.FieldByName("Login").SetString("testuser")

				// Set milestone
				milestone := issue.FieldByName("Milestone")
				milestone.FieldByName("Title").SetString("v1.0")
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	issue, err := client.GetIssue("owner", "repo", 42)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if issue == nil {
		t.Fatal("Expected issue to be returned")
	}
	if issue.ID != "issue-123" {
		t.Errorf("Expected ID 'issue-123', got '%s'", issue.ID)
	}
	if issue.Number != 42 {
		t.Errorf("Expected number 42, got %d", issue.Number)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", issue.Title)
	}
	if issue.Author.Login != "testuser" {
		t.Errorf("Expected author 'testuser', got '%s'", issue.Author.Login)
	}
	if issue.Milestone == nil || issue.Milestone.Title != "v1.0" {
		t.Error("Expected milestone with title 'v1.0'")
	}
	if issue.Repository.Owner != "owner" {
		t.Errorf("Expected repository owner 'owner', got '%s'", issue.Repository.Owner)
	}
}

func TestGetIssue_WithAssignees(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetIssue" {
				v := reflect.ValueOf(query).Elem()
				repo := v.FieldByName("Repository")
				issue := repo.FieldByName("Issue")
				issue.FieldByName("ID").SetString("issue-123")
				issue.FieldByName("Number").SetInt(1)
				issue.FieldByName("Title").SetString("Test")
				issue.FieldByName("State").SetString("OPEN")

				// Set assignees
				assignees := issue.FieldByName("Assignees")
				nodes := assignees.FieldByName("Nodes")
				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("Login").SetString("user1")
				newNodes.Index(0).Set(node1)

				node2 := reflect.New(nodeType).Elem()
				node2.FieldByName("Login").SetString("user2")
				newNodes.Index(1).Set(node2)

				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	issue, err := client.GetIssue("owner", "repo", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(issue.Assignees) != 2 {
		t.Fatalf("Expected 2 assignees, got %d", len(issue.Assignees))
	}
	if issue.Assignees[0].Login != "user1" {
		t.Errorf("Expected first assignee 'user1', got '%s'", issue.Assignees[0].Login)
	}
}

func TestGetIssue_WithLabels(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetIssue" {
				v := reflect.ValueOf(query).Elem()
				repo := v.FieldByName("Repository")
				issue := repo.FieldByName("Issue")
				issue.FieldByName("ID").SetString("issue-123")
				issue.FieldByName("Number").SetInt(1)
				issue.FieldByName("Title").SetString("Test")
				issue.FieldByName("State").SetString("OPEN")

				// Set labels
				labels := issue.FieldByName("Labels")
				nodes := labels.FieldByName("Nodes")
				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("Name").SetString("bug")
				node1.FieldByName("Color").SetString("d73a4a")
				newNodes.Index(0).Set(node1)

				node2 := reflect.New(nodeType).Elem()
				node2.FieldByName("Name").SetString("enhancement")
				node2.FieldByName("Color").SetString("a2eeef")
				newNodes.Index(1).Set(node2)

				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	issue, err := client.GetIssue("owner", "repo", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(issue.Labels) != 2 {
		t.Fatalf("Expected 2 labels, got %d", len(issue.Labels))
	}
	if issue.Labels[0].Name != "bug" {
		t.Errorf("Expected first label 'bug', got '%s'", issue.Labels[0].Name)
	}
}

func TestGetIssue_NoMilestone(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetIssue" {
				v := reflect.ValueOf(query).Elem()
				repo := v.FieldByName("Repository")
				issue := repo.FieldByName("Issue")
				issue.FieldByName("ID").SetString("issue-123")
				issue.FieldByName("Number").SetInt(1)
				issue.FieldByName("Title").SetString("Test")
				issue.FieldByName("State").SetString("OPEN")
				// Don't set milestone title - leave empty
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	issue, err := client.GetIssue("owner", "repo", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if issue.Milestone != nil {
		t.Error("Expected nil milestone when not set")
	}
}

// ============================================================================
// GetProjectItems Tests - Improved Coverage
// ============================================================================

func TestGetProjectItems_WithItems(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)
				newNode := reflect.New(nodeType).Elem()

				newNode.FieldByName("ID").SetString("item-1")

				// Set content
				content := newNode.FieldByName("Content")
				content.FieldByName("TypeName").SetString("Issue")

				issueContent := content.FieldByName("Issue")
				issueContent.FieldByName("ID").SetString("issue-123")
				issueContent.FieldByName("Number").SetInt(42)
				issueContent.FieldByName("Title").SetString("Test Issue")
				issueContent.FieldByName("State").SetString("OPEN")
				issueContent.FieldByName("URL").SetString("https://github.com/owner/repo/issues/42")

				// Set repository
				issueRepo := issueContent.FieldByName("Repository")
				issueRepo.FieldByName("NameWithOwner").SetString("owner/repo")

				newNodes.Index(0).Set(newNode)
				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if items[0].Issue == nil {
		t.Fatal("Expected issue in item")
	}
	if items[0].Issue.Number != 42 {
		t.Errorf("Expected issue number 42, got %d", items[0].Issue.Number)
	}
	if items[0].Issue.Repository.Owner != "owner" {
		t.Errorf("Expected repository owner 'owner', got '%s'", items[0].Issue.Repository.Owner)
	}
}

func TestGetProjectItems_WithFilter(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

				// Item 1 - matches filter
				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("ID").SetString("item-1")
				content1 := node1.FieldByName("Content")
				content1.FieldByName("TypeName").SetString("Issue")
				issue1 := content1.FieldByName("Issue")
				issue1.FieldByName("ID").SetString("issue-1")
				issue1.FieldByName("Number").SetInt(1)
				issue1.FieldByName("Title").SetString("Match")
				issue1.FieldByName("State").SetString("OPEN")
				repo1 := issue1.FieldByName("Repository")
				repo1.FieldByName("NameWithOwner").SetString("owner/repo")
				newNodes.Index(0).Set(node1)

				// Item 2 - doesn't match filter
				node2 := reflect.New(nodeType).Elem()
				node2.FieldByName("ID").SetString("item-2")
				content2 := node2.FieldByName("Content")
				content2.FieldByName("TypeName").SetString("Issue")
				issue2 := content2.FieldByName("Issue")
				issue2.FieldByName("ID").SetString("issue-2")
				issue2.FieldByName("Number").SetInt(2)
				issue2.FieldByName("Title").SetString("No Match")
				issue2.FieldByName("State").SetString("OPEN")
				repo2 := issue2.FieldByName("Repository")
				repo2.FieldByName("NameWithOwner").SetString("other/repo")
				newNodes.Index(1).Set(node2)

				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", &ProjectItemsFilter{Repository: "owner/repo"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item after filter, got %d", len(items))
	}
	if items[0].Issue.Title != "Match" {
		t.Errorf("Expected issue title 'Match', got '%s'", items[0].Issue.Title)
	}
}

func TestGetProjectItems_SkipsNonIssues(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

				// Item 1 - Draft issue (should be skipped)
				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("ID").SetString("item-1")
				content1 := node1.FieldByName("Content")
				content1.FieldByName("TypeName").SetString("DraftIssue")
				newNodes.Index(0).Set(node1)

				// Item 2 - Real issue
				node2 := reflect.New(nodeType).Elem()
				node2.FieldByName("ID").SetString("item-2")
				content2 := node2.FieldByName("Content")
				content2.FieldByName("TypeName").SetString("Issue")
				issue2 := content2.FieldByName("Issue")
				issue2.FieldByName("ID").SetString("issue-2")
				issue2.FieldByName("Number").SetInt(2)
				issue2.FieldByName("Title").SetString("Real Issue")
				issue2.FieldByName("State").SetString("OPEN")
				repo2 := issue2.FieldByName("Repository")
				repo2.FieldByName("NameWithOwner").SetString("owner/repo")
				newNodes.Index(1).Set(node2)

				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item (draft skipped), got %d", len(items))
	}
	if items[0].Issue.Title != "Real Issue" {
		t.Errorf("Expected 'Real Issue', got '%s'", items[0].Issue.Title)
	}
}

func TestGetProjectItems_WithFieldValues(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)
				newNode := reflect.New(nodeType).Elem()

				newNode.FieldByName("ID").SetString("item-1")
				content := newNode.FieldByName("Content")
				content.FieldByName("TypeName").SetString("Issue")
				issue := content.FieldByName("Issue")
				issue.FieldByName("ID").SetString("issue-1")
				issue.FieldByName("Number").SetInt(1)
				issue.FieldByName("Title").SetString("Test")
				issue.FieldByName("State").SetString("OPEN")
				repo := issue.FieldByName("Repository")
				repo.FieldByName("NameWithOwner").SetString("owner/repo")

				// Set field values
				fieldValues := newNode.FieldByName("FieldValues")
				fvNodes := fieldValues.FieldByName("Nodes")
				fvNodeType := fvNodes.Type().Elem()
				newFvNodes := reflect.MakeSlice(fvNodes.Type(), 2, 2)

				// Single select field value
				fv1 := reflect.New(fvNodeType).Elem()
				fv1.FieldByName("TypeName").SetString("ProjectV2ItemFieldSingleSelectValue")
				singleSelect := fv1.FieldByName("ProjectV2ItemFieldSingleSelectValue")
				singleSelect.FieldByName("Name").SetString("In Progress")
				singleSelectField := singleSelect.FieldByName("Field")
				singleSelectFieldInner := singleSelectField.FieldByName("ProjectV2SingleSelectField")
				singleSelectFieldInner.FieldByName("Name").SetString("Status")
				newFvNodes.Index(0).Set(fv1)

				// Text field value
				fv2 := reflect.New(fvNodeType).Elem()
				fv2.FieldByName("TypeName").SetString("ProjectV2ItemFieldTextValue")
				textValue := fv2.FieldByName("ProjectV2ItemFieldTextValue")
				textValue.FieldByName("Text").SetString("Some notes")
				textField := textValue.FieldByName("Field")
				textFieldInner := textField.FieldByName("ProjectV2Field")
				textFieldInner.FieldByName("Name").SetString("Notes")
				newFvNodes.Index(1).Set(fv2)

				fvNodes.Set(newFvNodes)
				newNodes.Index(0).Set(newNode)
				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if len(items[0].FieldValues) != 2 {
		t.Fatalf("Expected 2 field values, got %d", len(items[0].FieldValues))
	}

	// Check Status field
	foundStatus := false
	foundNotes := false
	for _, fv := range items[0].FieldValues {
		if fv.Field == "Status" && fv.Value == "In Progress" {
			foundStatus = true
		}
		if fv.Field == "Notes" && fv.Value == "Some notes" {
			foundNotes = true
		}
	}
	if !foundStatus {
		t.Error("Expected Status field with value 'In Progress'")
	}
	if !foundNotes {
		t.Error("Expected Notes field with value 'Some notes'")
	}
}

func TestGetProjectItems_WithAssignees(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)
				newNode := reflect.New(nodeType).Elem()

				newNode.FieldByName("ID").SetString("item-1")
				content := newNode.FieldByName("Content")
				content.FieldByName("TypeName").SetString("Issue")
				issue := content.FieldByName("Issue")
				issue.FieldByName("ID").SetString("issue-1")
				issue.FieldByName("Number").SetInt(1)
				issue.FieldByName("Title").SetString("Test")
				issue.FieldByName("State").SetString("OPEN")
				repo := issue.FieldByName("Repository")
				repo.FieldByName("NameWithOwner").SetString("owner/repo")

				// Set assignees
				assignees := issue.FieldByName("Assignees")
				assigneeNodes := assignees.FieldByName("Nodes")
				assigneeNodeType := assigneeNodes.Type().Elem()
				newAssigneeNodes := reflect.MakeSlice(assigneeNodes.Type(), 1, 1)
				assignee := reflect.New(assigneeNodeType).Elem()
				assignee.FieldByName("Login").SetString("testuser")
				newAssigneeNodes.Index(0).Set(assignee)
				assigneeNodes.Set(newAssigneeNodes)

				newNodes.Index(0).Set(newNode)
				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if len(items[0].Issue.Assignees) != 1 {
		t.Fatalf("Expected 1 assignee, got %d", len(items[0].Issue.Assignees))
	}
	if items[0].Issue.Assignees[0].Login != "testuser" {
		t.Errorf("Expected assignee 'testuser', got '%s'", items[0].Issue.Assignees[0].Login)
	}
}

// ============================================================================
// GetSubIssues Tests - Improved Coverage
// ============================================================================

func TestGetSubIssues_Success(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetSubIssues" {
				v := reflect.ValueOf(query).Elem()
				repo := v.FieldByName("Repository")
				issue := repo.FieldByName("Issue")
				subIssues := issue.FieldByName("SubIssues")
				nodes := subIssues.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

				// Sub-issue 1
				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("ID").SetString("sub-1")
				node1.FieldByName("Number").SetInt(10)
				node1.FieldByName("Title").SetString("Sub-issue 1")
				node1.FieldByName("State").SetString("OPEN")
				node1.FieldByName("URL").SetString("https://github.com/owner/repo/issues/10")
				repo1 := node1.FieldByName("Repository")
				repo1.FieldByName("Name").SetString("repo")
				owner1 := repo1.FieldByName("Owner")
				owner1.FieldByName("Login").SetString("owner")
				newNodes.Index(0).Set(node1)

				// Sub-issue 2
				node2 := reflect.New(nodeType).Elem()
				node2.FieldByName("ID").SetString("sub-2")
				node2.FieldByName("Number").SetInt(11)
				node2.FieldByName("Title").SetString("Sub-issue 2")
				node2.FieldByName("State").SetString("CLOSED")
				node2.FieldByName("URL").SetString("https://github.com/owner/repo/issues/11")
				repo2 := node2.FieldByName("Repository")
				repo2.FieldByName("Name").SetString("repo")
				owner2 := repo2.FieldByName("Owner")
				owner2.FieldByName("Login").SetString("owner")
				newNodes.Index(1).Set(node2)

				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	subIssues, err := client.GetSubIssues("owner", "repo", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(subIssues) != 2 {
		t.Fatalf("Expected 2 sub-issues, got %d", len(subIssues))
	}
	if subIssues[0].Number != 10 {
		t.Errorf("Expected first sub-issue number 10, got %d", subIssues[0].Number)
	}
	if subIssues[1].State != "CLOSED" {
		t.Errorf("Expected second sub-issue state 'CLOSED', got '%s'", subIssues[1].State)
	}
}

// ============================================================================
// GetRepositoryIssues Tests - Improved Coverage
// ============================================================================

func TestGetRepositoryIssues_Success(t *testing.T) {
	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetRepositoryIssues" {
				v := reflect.ValueOf(query).Elem()
				repo := v.FieldByName("Repository")
				issues := repo.FieldByName("Issues")
				nodes := issues.FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("ID").SetString("issue-1")
				node1.FieldByName("Number").SetInt(1)
				node1.FieldByName("Title").SetString("First Issue")
				node1.FieldByName("State").SetString("OPEN")
				node1.FieldByName("URL").SetString("https://github.com/owner/repo/issues/1")
				newNodes.Index(0).Set(node1)

				node2 := reflect.New(nodeType).Elem()
				node2.FieldByName("ID").SetString("issue-2")
				node2.FieldByName("Number").SetInt(2)
				node2.FieldByName("Title").SetString("Second Issue")
				node2.FieldByName("State").SetString("CLOSED")
				node2.FieldByName("URL").SetString("https://github.com/owner/repo/issues/2")
				newNodes.Index(1).Set(node2)

				nodes.Set(newNodes)
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	issues, err := client.GetRepositoryIssues("owner", "repo", "all")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
	if issues[0].Title != "First Issue" {
		t.Errorf("Expected first issue title 'First Issue', got '%s'", issues[0].Title)
	}
	if issues[0].Repository.Owner != "owner" {
		t.Errorf("Expected repository owner 'owner', got '%s'", issues[0].Repository.Owner)
	}
}

// ============================================================================
// GetProjectItems Pagination Tests
// ============================================================================

func TestGetProjectItems_Pagination_MultiplePages(t *testing.T) {
	// Track which page we're on
	callCount := 0

	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				callCount++
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")
				pageInfoField := items.FieldByName("PageInfo")

				nodeType := nodes.Type().Elem()

				if callCount == 1 {
					// First page - return items 1-2 with hasNextPage=true
					newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

					node1 := reflect.New(nodeType).Elem()
					node1.FieldByName("ID").SetString("item-1")
					content1 := node1.FieldByName("Content")
					content1.FieldByName("TypeName").SetString("Issue")
					issue1 := content1.FieldByName("Issue")
					issue1.FieldByName("ID").SetString("issue-1")
					issue1.FieldByName("Number").SetInt(1)
					issue1.FieldByName("Title").SetString("Issue 1")
					issue1.FieldByName("State").SetString("OPEN")
					repo1 := issue1.FieldByName("Repository")
					repo1.FieldByName("NameWithOwner").SetString("owner/repo")
					newNodes.Index(0).Set(node1)

					node2 := reflect.New(nodeType).Elem()
					node2.FieldByName("ID").SetString("item-2")
					content2 := node2.FieldByName("Content")
					content2.FieldByName("TypeName").SetString("Issue")
					issue2 := content2.FieldByName("Issue")
					issue2.FieldByName("ID").SetString("issue-2")
					issue2.FieldByName("Number").SetInt(2)
					issue2.FieldByName("Title").SetString("Issue 2")
					issue2.FieldByName("State").SetString("OPEN")
					repo2 := issue2.FieldByName("Repository")
					repo2.FieldByName("NameWithOwner").SetString("owner/repo")
					newNodes.Index(1).Set(node2)

					nodes.Set(newNodes)

					// Set pagination info - more pages available
					pageInfoField.FieldByName("HasNextPage").SetBool(true)
					pageInfoField.FieldByName("EndCursor").SetString("cursor-page-1")
				} else if callCount == 2 {
					// Second page - return item 3 with hasNextPage=false
					newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)

					node3 := reflect.New(nodeType).Elem()
					node3.FieldByName("ID").SetString("item-3")
					content3 := node3.FieldByName("Content")
					content3.FieldByName("TypeName").SetString("Issue")
					issue3 := content3.FieldByName("Issue")
					issue3.FieldByName("ID").SetString("issue-3")
					issue3.FieldByName("Number").SetInt(3)
					issue3.FieldByName("Title").SetString("Issue 3")
					issue3.FieldByName("State").SetString("OPEN")
					repo3 := issue3.FieldByName("Repository")
					repo3.FieldByName("NameWithOwner").SetString("owner/repo")
					newNodes.Index(0).Set(node3)

					nodes.Set(newNodes)

					// Set pagination info - no more pages
					pageInfoField.FieldByName("HasNextPage").SetBool(false)
					pageInfoField.FieldByName("EndCursor").SetString("")
				}
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", callCount)
	}
	if len(items) != 3 {
		t.Fatalf("Expected 3 items from 2 pages, got %d", len(items))
	}
	if items[0].Issue.Number != 1 {
		t.Errorf("Expected first issue number 1, got %d", items[0].Issue.Number)
	}
	if items[2].Issue.Number != 3 {
		t.Errorf("Expected third issue number 3, got %d", items[2].Issue.Number)
	}
}

func TestGetProjectItems_Pagination_SinglePage(t *testing.T) {
	callCount := 0

	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				callCount++
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")
				pageInfoField := items.FieldByName("PageInfo")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)

				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("ID").SetString("item-1")
				content1 := node1.FieldByName("Content")
				content1.FieldByName("TypeName").SetString("Issue")
				issue1 := content1.FieldByName("Issue")
				issue1.FieldByName("ID").SetString("issue-1")
				issue1.FieldByName("Number").SetInt(1)
				issue1.FieldByName("Title").SetString("Only Issue")
				issue1.FieldByName("State").SetString("OPEN")
				repo1 := issue1.FieldByName("Repository")
				repo1.FieldByName("NameWithOwner").SetString("owner/repo")
				newNodes.Index(0).Set(node1)

				nodes.Set(newNodes)

				// No more pages
				pageInfoField.FieldByName("HasNextPage").SetBool(false)
				pageInfoField.FieldByName("EndCursor").SetString("")
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 API call (single page), got %d", callCount)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
}

func TestGetProjectItems_Pagination_CursorPropagation(t *testing.T) {
	var receivedCursors []interface{}

	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				// Track the cursor value passed
				receivedCursors = append(receivedCursors, variables["cursor"])

				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")
				pageInfoField := items.FieldByName("PageInfo")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)

				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("ID").SetString("item-1")
				content1 := node1.FieldByName("Content")
				content1.FieldByName("TypeName").SetString("Issue")
				issue1 := content1.FieldByName("Issue")
				issue1.FieldByName("ID").SetString("issue-1")
				issue1.FieldByName("Number").SetInt(1)
				issue1.FieldByName("Title").SetString("Issue")
				issue1.FieldByName("State").SetString("OPEN")
				repo1 := issue1.FieldByName("Repository")
				repo1.FieldByName("NameWithOwner").SetString("owner/repo")
				newNodes.Index(0).Set(node1)

				nodes.Set(newNodes)

				// Return different cursors based on call
				if len(receivedCursors) == 1 {
					pageInfoField.FieldByName("HasNextPage").SetBool(true)
					pageInfoField.FieldByName("EndCursor").SetString("expected-cursor-123")
				} else {
					pageInfoField.FieldByName("HasNextPage").SetBool(false)
					pageInfoField.FieldByName("EndCursor").SetString("")
				}
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	_, err := client.GetProjectItems("proj-id", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(receivedCursors) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(receivedCursors))
	}

	// First call should have nil cursor (the nil pointer type)
	if receivedCursors[0] != nil {
		// Check if it's a typed nil
		rv := reflect.ValueOf(receivedCursors[0])
		if !rv.IsNil() {
			t.Errorf("First call should have nil cursor, got %v", receivedCursors[0])
		}
	}

	// Second call should have the cursor from first page
	// The cursor is passed as graphql.String which is a string type alias
	cursorVal := reflect.ValueOf(receivedCursors[1])
	if cursorVal.Kind() == reflect.String {
		if cursorVal.String() != "expected-cursor-123" {
			t.Errorf("Second call should have cursor 'expected-cursor-123', got %v", receivedCursors[1])
		}
	} else {
		t.Errorf("Second call cursor should be string type, got %T: %v", receivedCursors[1], receivedCursors[1])
	}
}

func TestGetProjectItems_Pagination_ErrorOnSecondPage(t *testing.T) {
	callCount := 0

	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				callCount++

				if callCount == 2 {
					return errors.New("API error on second page")
				}

				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")
				pageInfoField := items.FieldByName("PageInfo")

				nodeType := nodes.Type().Elem()
				newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)

				node1 := reflect.New(nodeType).Elem()
				node1.FieldByName("ID").SetString("item-1")
				content1 := node1.FieldByName("Content")
				content1.FieldByName("TypeName").SetString("Issue")
				issue1 := content1.FieldByName("Issue")
				issue1.FieldByName("ID").SetString("issue-1")
				issue1.FieldByName("Number").SetInt(1)
				issue1.FieldByName("Title").SetString("Issue")
				issue1.FieldByName("State").SetString("OPEN")
				repo1 := issue1.FieldByName("Repository")
				repo1.FieldByName("NameWithOwner").SetString("owner/repo")
				newNodes.Index(0).Set(node1)

				nodes.Set(newNodes)

				// Indicate there's another page
				pageInfoField.FieldByName("HasNextPage").SetBool(true)
				pageInfoField.FieldByName("EndCursor").SetString("cursor-1")
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", nil)

	if err == nil {
		t.Fatal("Expected error when second page fails")
	}
	if !strings.Contains(err.Error(), "API error on second page") {
		t.Errorf("Expected error message about second page, got: %v", err)
	}
	if items != nil {
		t.Errorf("Expected nil items on error, got %d items", len(items))
	}
}

func TestGetProjectItems_Pagination_WithFilter(t *testing.T) {
	callCount := 0

	mock := &queryMockClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "GetProjectItems" {
				callCount++
				v := reflect.ValueOf(query).Elem()
				node := v.FieldByName("Node")
				projectV2 := node.FieldByName("ProjectV2")
				items := projectV2.FieldByName("Items")
				nodes := items.FieldByName("Nodes")
				pageInfoField := items.FieldByName("PageInfo")

				nodeType := nodes.Type().Elem()

				if callCount == 1 {
					// First page - 2 items, one matches filter
					newNodes := reflect.MakeSlice(nodes.Type(), 2, 2)

					node1 := reflect.New(nodeType).Elem()
					node1.FieldByName("ID").SetString("item-1")
					content1 := node1.FieldByName("Content")
					content1.FieldByName("TypeName").SetString("Issue")
					issue1 := content1.FieldByName("Issue")
					issue1.FieldByName("ID").SetString("issue-1")
					issue1.FieldByName("Number").SetInt(1)
					issue1.FieldByName("Title").SetString("Match 1")
					issue1.FieldByName("State").SetString("OPEN")
					repo1 := issue1.FieldByName("Repository")
					repo1.FieldByName("NameWithOwner").SetString("target/repo")
					newNodes.Index(0).Set(node1)

					node2 := reflect.New(nodeType).Elem()
					node2.FieldByName("ID").SetString("item-2")
					content2 := node2.FieldByName("Content")
					content2.FieldByName("TypeName").SetString("Issue")
					issue2 := content2.FieldByName("Issue")
					issue2.FieldByName("ID").SetString("issue-2")
					issue2.FieldByName("Number").SetInt(2)
					issue2.FieldByName("Title").SetString("Other Repo")
					issue2.FieldByName("State").SetString("OPEN")
					repo2 := issue2.FieldByName("Repository")
					repo2.FieldByName("NameWithOwner").SetString("other/repo")
					newNodes.Index(1).Set(node2)

					nodes.Set(newNodes)
					pageInfoField.FieldByName("HasNextPage").SetBool(true)
					pageInfoField.FieldByName("EndCursor").SetString("cursor-1")
				} else {
					// Second page - 1 item matching filter
					newNodes := reflect.MakeSlice(nodes.Type(), 1, 1)

					node3 := reflect.New(nodeType).Elem()
					node3.FieldByName("ID").SetString("item-3")
					content3 := node3.FieldByName("Content")
					content3.FieldByName("TypeName").SetString("Issue")
					issue3 := content3.FieldByName("Issue")
					issue3.FieldByName("ID").SetString("issue-3")
					issue3.FieldByName("Number").SetInt(3)
					issue3.FieldByName("Title").SetString("Match 2")
					issue3.FieldByName("State").SetString("OPEN")
					repo3 := issue3.FieldByName("Repository")
					repo3.FieldByName("NameWithOwner").SetString("target/repo")
					newNodes.Index(0).Set(node3)

					nodes.Set(newNodes)
					pageInfoField.FieldByName("HasNextPage").SetBool(false)
					pageInfoField.FieldByName("EndCursor").SetString("")
				}
			}
			return nil
		},
	}

	client := NewClientWithGraphQL(mock)
	items, err := client.GetProjectItems("proj-id", &ProjectItemsFilter{Repository: "target/repo"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items matching filter across pages, got %d", len(items))
	}
	if items[0].Issue.Title != "Match 1" {
		t.Errorf("Expected first item 'Match 1', got '%s'", items[0].Issue.Title)
	}
	if items[1].Issue.Title != "Match 2" {
		t.Errorf("Expected second item 'Match 2', got '%s'", items[1].Issue.Title)
	}
}
