# Release v0.12.0

## Highlights

This release focuses on **performance optimizations** across multiple commands, significantly reducing API calls and improving response times for large projects.

## What's New

### Performance Optimizations

- **Batch field mutations** - `move` command now batches field updates, reducing API calls from O(N) to O(N/50)
- **Batch sub-issue queries** - Recursive operations fetch sub-issues in batches per level instead of per-issue
- **Repository-scoped queries** - `branch current`, `branch close`, `filter`, and `intake` commands now use repository filtering to reduce data transfer
- **Targeted issue queries** - `filter` command uses `GetProjectItemsByIssues` for stdin-provided issues instead of fetching all project items

### API Improvements

- **`GetProjectItemsByIssues`** - New targeted query method for fetching specific issues by reference
- **`GetSubIssuesBatch`** - New batch query for fetching sub-issues across multiple parents
- **`BatchUpdateProjectItemFields`** - New batch mutation for updating multiple fields in a single request

### Other Changes

- Updated IDPF framework to v0.25.0
- Fixed hook error message to use `--branch` instead of deprecated `--release`
- Auto-label creation for missing labels

## Upgrade Notes

No breaking changes. All optimizations are backward-compatible.

## Issues Included

- #541 - Optimize move command with targeted queries
- #542 - Batch sub-issue queries for recursive operations
- #543 - Batch field mutations for move command
- #545 - Optimize filter command with targeted queries
- #546 - Optimize branch current command
- #547 - Optimize branch close command
- #548 - Optimize intake command
