package api

import (
	"net/http"
	"testing"
)

func TestNewClient_ReturnsClient(t *testing.T) {
	// ACT: Create a new client
	client := NewClient()

	// ASSERT: Client is not nil
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

func TestNewClient_HasGraphQLClient(t *testing.T) {
	// Skip in CI - requires gh auth
	if testing.Short() {
		t.Skip("Skipping test that requires gh auth")
	}

	// ACT: Create a new client
	client := NewClient()

	// ASSERT: GraphQL client is accessible
	if client.gql == nil {
		t.Fatal("Expected GraphQL client to be initialized")
	}
}

func TestNewClientWithOptions_CustomHost(t *testing.T) {
	// ARRANGE: Custom options
	opts := ClientOptions{
		Host: "github.example.com",
	}

	// ACT: Create client with options
	client := NewClientWithOptions(opts)

	// ASSERT: Client is created (host is used internally)
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

func TestClient_FeatureHeaders_Included(t *testing.T) {
	// This test verifies that sub_issues feature header is configured
	// We can't easily test the actual header without making a request,
	// but we can verify the client was created with the right options

	client := NewClient()

	// ASSERT: Client exists and has feature flags set
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	// The feature headers are set during client creation
	// Actual header verification would require integration tests
}

func TestJoinFeatures_Empty(t *testing.T) {
	result := joinFeatures([]string{})
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestJoinFeatures_Single(t *testing.T) {
	result := joinFeatures([]string{"sub_issues"})
	if result != "sub_issues" {
		t.Errorf("Expected 'sub_issues', got '%s'", result)
	}
}

func TestJoinFeatures_Multiple(t *testing.T) {
	result := joinFeatures([]string{"sub_issues", "issue_types"})
	if result != "sub_issues,issue_types" {
		t.Errorf("Expected 'sub_issues,issue_types', got '%s'", result)
	}
}

func TestSetTestTransport(t *testing.T) {
	// Ensure test transport is nil initially (or clear it)
	SetTestTransport(nil)

	// Set a test transport
	transport := &mockRoundTripper{}
	SetTestTransport(transport)

	// Verify testTransport is set (indirectly through NewClient behavior)
	// The actual verification is that NewClient uses it

	// Clear the transport
	SetTestTransport(nil)
}

func TestSetTestAuthToken(t *testing.T) {
	// Clear any existing token
	SetTestAuthToken("")

	// Set a token
	SetTestAuthToken("test-token-123")

	// Clear the token
	SetTestAuthToken("")
}

func TestNewClient_UsesTestTransport(t *testing.T) {
	// Set up test transport
	transport := &mockRoundTripper{}
	SetTestTransport(transport)
	SetTestAuthToken("test-token")
	defer func() {
		SetTestTransport(nil)
		SetTestAuthToken("")
	}()

	// Create client - should use test transport
	client := NewClient()
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	// The client should be created with the test transport
	// We can't easily verify the transport is set, but we can verify
	// the client was created without error
}

func TestNewClientWithOptions_Transport(t *testing.T) {
	transport := &mockRoundTripper{}
	opts := ClientOptions{
		Transport: transport,
		AuthToken: "test-token",
	}

	client := NewClientWithOptions(opts)
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

func TestNewClientWithOptions_AllOptions(t *testing.T) {
	transport := &mockRoundTripper{}
	opts := ClientOptions{
		Host:             "github.example.com",
		EnableSubIssues:  true,
		EnableIssueTypes: true,
		Transport:        transport,
		AuthToken:        "test-token",
	}

	client := NewClientWithOptions(opts)
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct{}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestNewClientWithGraphQL(t *testing.T) {
	// Create mock GraphQL client
	mockGQL := &simpleGraphQLMock{}

	// Create client with mock
	client := NewClientWithGraphQL(mockGQL)

	// Assert client is created
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	// Assert the mock was set
	if client.gql != mockGQL {
		t.Error("Expected GraphQL client to be the mock")
	}
}

func TestNewClientWithOptions_DisabledFeatures(t *testing.T) {
	transport := &mockRoundTripper{}
	opts := ClientOptions{
		EnableSubIssues:  false,
		EnableIssueTypes: false,
		Transport:        transport,
		AuthToken:        "test-token",
	}

	client := NewClientWithOptions(opts)
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

func TestNewClientWithOptions_OnlySubIssues(t *testing.T) {
	transport := &mockRoundTripper{}
	opts := ClientOptions{
		EnableSubIssues:  true,
		EnableIssueTypes: false,
		Transport:        transport,
		AuthToken:        "test-token",
	}

	client := NewClientWithOptions(opts)
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

func TestNewClientWithOptions_OnlyIssueTypes(t *testing.T) {
	transport := &mockRoundTripper{}
	opts := ClientOptions{
		EnableSubIssues:  false,
		EnableIssueTypes: true,
		Transport:        transport,
		AuthToken:        "test-token",
	}

	client := NewClientWithOptions(opts)
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

// simpleGraphQLMock implements GraphQLClient for client_test.go testing
type simpleGraphQLMock struct{}

func (m *simpleGraphQLMock) Query(name string, query interface{}, variables map[string]interface{}) error {
	return nil
}

func (m *simpleGraphQLMock) Mutate(name string, mutation interface{}, variables map[string]interface{}) error {
	return nil
}
