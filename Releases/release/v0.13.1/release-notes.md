# Release v0.13.1

## Summary

This patch release adds auto-creation of default labels when assigning labels that don't exist in the repository. Labels defined in `defaults.yml` are automatically created, improving the workflow for new repositories and reducing manual setup.

## What's New

### Auto-Create Default Labels (#596)
When `gh pmu create` or similar commands assign a label that doesn't exist in the repository, gh-pmu now checks if the label is defined in the embedded `defaults.yml`. If found, it automatically creates the label with the correct color and description.

Labels not in defaults will show a warning instead of being silently skipped.

### Expanded Default Labels (#595)
Added `prd`, `bug`, and `enhancement` to the default label set:
- `prd` - Product Requirements Document (color: #f9d0c4)
- `bug` - Something isn't working (color: #d73a4a)
- `enhancement` - New feature or request (color: #a2eeef)

### Visual Improvements
- Epic label color changed to lighter green (#3fb950) for better visibility against dark backgrounds

## Installation

```bash
gh extension upgrade rubrical-studios/gh-pmu
```

## Full Changelog

https://github.com/rubrical-studios/gh-pmu/compare/v0.13.0...v0.13.1
