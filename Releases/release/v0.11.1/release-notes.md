# Release v0.11.1

**Release Date:** 2026-01-12

## Summary

This patch release includes 8 bug fixes and 9 enhancements focused on improving reliability, adding new filtering options, and completing the release-to-branch terminology migration.

## Highlights

- **New Filters:** `--state` flag for `list` command to filter by open/closed state
- **Label Management:** `--remove-label` flag for `edit` command
- **Dry-Run Support:** Preview destructive operations with `--dry-run` on `branch close` and `microsprint close`
- **Rate Limiting:** Automatic retry with exponential backoff for bulk operations
- **Sub-Issue Pagination:** Handle epics with 50+ sub-issues

## Added

- `--state` filter flag for `list` command to filter by issue state (open/closed) (#522)
- `--remove-label` flag for `edit` command to remove labels from issues (#519)
- `--dry-run` flag for `branch close` and `microsprint close` commands (#527)
- DATE field type support in `SetProjectItemField` for date-based project fields (#518)
- Progress indicator for recursive move operations showing per-issue status (#520)
- Pagination support for `GetSubIssues` query to handle 50+ sub-issues (#521)
- Rate limiting protection with exponential backoff for bulk operations (#525)
- Status field value validation before setting, with helpful error messages (#523)

## Fixed

- `setNumberField` now parses value parameter instead of always setting 0 (#513)
- `--label` flag in `edit` command now actually adds labels (#528)
- `--interactive` flag removed from `create` command (was unimplemented stub) (#514)
- `branch close` now reports correct count of moved issues (#515)
- `GetSubIssueCounts` failures now warn and include items instead of silent exclusion (#516)
- Consolidated duplicate `openInBrowser` implementations into `internal/ui` (#517)
- Help text updated from "release" to "branch" terminology throughout (#512)
- Error messages updated to reference "branch" instead of "release" (#511)

## Changed

- `IsIDPF()` now uses case-insensitive prefix matching for framework detection (#524)
- Coverage threshold lowered from 80% to 70%

## Upgrade Instructions

```bash
gh extension upgrade gh-pmu
```

## Full Changelog

See [CHANGELOG.md](https://github.com/rubrical-studios/gh-pmu/blob/main/CHANGELOG.md) for complete details.
