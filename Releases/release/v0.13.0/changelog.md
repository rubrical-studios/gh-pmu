# Changelog for v0.13.0

## [0.13.0] - 2026-01-19

### Added
- State filtering for `GetProjectItems` and `GetProjectItemsForBoard` - filter by issue state (OPEN/CLOSED) (#560)
- Auto-create custom fields and labels during `init` command (#587)
- E2E test infrastructure with command integration tests (#582)
- Label verification tests for branch and microsprint trackers

### Fixed
- Cursor-based pagination for `GetRepositoryIssues` to handle large result sets (#557)
- Cursor-based pagination for `GetIssuesByLabel` functions (#558)
- Pagination fallback in `GetSubIssuesBatch` for epics with >100 sub-issues (#559)
- Microsprint add retry for GitHub API eventual consistency
- Use Release field name instead of Branch during init

### Changed
- Extract `buildBatchMutationRequest` for improved testability (#554)
- Update framework to v0.26.3
