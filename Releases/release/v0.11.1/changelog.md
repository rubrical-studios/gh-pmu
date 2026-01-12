# Changelog for v0.11.1

## Bug Fixes

| Issue | Title |
|-------|-------|
| #512 | Help text inconsistencies after release→branch rename |
| #513 | setNumberField always sets 0 instead of parsing value |
| #514 | --interactive flag in create command not implemented |
| #515 | Release close reports wrong count of moved issues |
| #516 | Silent failure when GetSubIssueCounts fails in list filtering |
| #517 | Duplicate openInBrowser implementations |
| #528 | --label flag in edit command silently does nothing |

## Enhancements

| Issue | Title |
|-------|-------|
| #511 | Improve no-release-assignment error message |
| #518 | Add DATE field type support in SetProjectItemField |
| #519 | Add --remove-labels flag to edit command |
| #520 | Add progress indicator for recursive move operations |
| #521 | Add pagination support for GetSubIssues query |
| #522 | Add --state filter to list command |
| #523 | Add validation for status field value existence |
| #524 | Case-insensitive framework detection in IsIDPF() |
| #525 | Add rate limiting protection for bulk operations |
| #527 | Add dry-run flag for destructive operations |

## Files Changed

### cmd/
- `branch.go` - Help text updates, dry-run flag
- `create.go` - Remove --interactive flag
- `edit.go` - Add --remove-label flag, fix --label
- `list.go` - Add --state filter, fix GetSubIssueCounts handling
- `microsprint.go` - Add dry-run flag
- `move.go` - Add progress indicator, rate limiting
- `sub.go` - Use shared openInBrowser
- `triage.go` - Rate limiting for bulk operations
- `validation.go` - Status validation
- `view.go` - Use shared openInBrowser

### internal/api/
- `mutations.go` - Fix setNumberField, add DATE support, fix AddLabelToIssue
- `queries.go` - Add GetSubIssues pagination
- `retry.go` - New file for rate limiting with exponential backoff

### internal/config/
- `config.go` - Case-insensitive IsIDPF(), add ValidateFieldValue

### internal/ui/
- `ui.go` - Add shared OpenInBrowser function
