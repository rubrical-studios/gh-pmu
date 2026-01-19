# Release v0.13.0

**Release Date:** 2026-01-19

## Highlights

This release focuses on API reliability improvements (pagination fixes) and enhanced project initialization with auto-creation of custom fields and labels.

## What's New

### State Filtering
Filter project items by issue state (OPEN/CLOSED) in `GetProjectItems` and `GetProjectItemsForBoard` APIs.

### Auto-Create Fields and Labels
The `init` command now automatically creates required custom fields (Release, Microsprint) and labels (branch, microsprint, epic, story) if they don't exist.

### E2E Test Infrastructure
Comprehensive end-to-end testing infrastructure with 12 test scenarios covering branch lifecycle, microsprint workflow, board rendering, and more.

## Bug Fixes

- **Pagination fixes** - All API functions now properly handle large result sets:
  - `GetRepositoryIssues` - cursor-based pagination
  - `GetIssuesByLabel` - cursor-based pagination
  - `GetSubIssuesBatch` - fallback for epics with >100 sub-issues
- **Microsprint retry** - Added retry logic for eventual consistency when adding issues to microsprints
- **Init field naming** - Fixed Release field name usage during initialization

## Breaking Changes

None.

## Upgrade

```bash
gh extension upgrade gh-pmu
```
