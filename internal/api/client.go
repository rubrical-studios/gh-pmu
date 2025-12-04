package api

import (
	"github.com/cli/go-gh/v2/pkg/api"
)

// FeatureSubIssues is the GitHub API preview header for sub-issues
const FeatureSubIssues = "sub_issues"

// FeatureIssueTypes is the GitHub API preview header for issue types
const FeatureIssueTypes = "issue_types"

// GraphQLClient interface allows mocking the GitHub GraphQL client for testing
type GraphQLClient interface {
	Query(name string, query interface{}, variables map[string]interface{}) error
	Mutate(name string, mutation interface{}, variables map[string]interface{}) error
}

// Client wraps the GitHub GraphQL API client with project management features
type Client struct {
	gql  GraphQLClient
	opts ClientOptions
}

// ClientOptions configures the API client
type ClientOptions struct {
	// Host is the GitHub hostname (default: github.com)
	Host string

	// EnableSubIssues enables the sub_issues feature preview
	EnableSubIssues bool

	// EnableIssueTypes enables the issue_types feature preview
	EnableIssueTypes bool
}

// NewClient creates a new API client with default options
func NewClient() *Client {
	return NewClientWithOptions(ClientOptions{
		EnableSubIssues:  true,
		EnableIssueTypes: true,
	})
}

// NewClientWithOptions creates a new API client with custom options
func NewClientWithOptions(opts ClientOptions) *Client {
	// Build headers with feature previews
	headers := make(map[string]string)

	// Add GraphQL feature preview headers
	// These enable beta features in the GitHub API
	featureHeaders := []string{}
	if opts.EnableSubIssues {
		featureHeaders = append(featureHeaders, FeatureSubIssues)
	}
	if opts.EnableIssueTypes {
		featureHeaders = append(featureHeaders, FeatureIssueTypes)
	}

	if len(featureHeaders) > 0 {
		// GitHub uses X-Github-Next for feature previews
		headers["X-Github-Next"] = joinFeatures(featureHeaders)
	}

	// Create GraphQL client options
	apiOpts := api.ClientOptions{
		Headers: headers,
	}

	if opts.Host != "" {
		apiOpts.Host = opts.Host
	}

	// Create the GraphQL client
	gql, err := api.NewGraphQLClient(apiOpts)
	if err != nil {
		// If we can't create a client (e.g., not authenticated),
		// return a client with nil gql - methods will return errors
		return &Client{opts: opts}
	}

	return &Client{
		gql:  gql,
		opts: opts,
	}
}

// NewClientWithGraphQL creates a Client with a custom GraphQL client (for testing)
func NewClientWithGraphQL(gql GraphQLClient) *Client {
	return &Client{gql: gql}
}

// joinFeatures joins feature names with commas
func joinFeatures(features []string) string {
	if len(features) == 0 {
		return ""
	}
	result := features[0]
	for i := 1; i < len(features); i++ {
		result += "," + features[i]
	}
	return result
}
