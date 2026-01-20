# Changelog for v0.13.1

## [0.13.1] - 2026-01-19

### Added
- Auto-create default labels when assigning non-existent labels (#596)
  - Labels defined in `defaults.yml` are automatically created if missing from repository
  - Warns user when label doesn't exist and isn't in defaults
- Added `prd`, `bug`, and `enhancement` labels to default label set (#595)
- `GetLabel()` method in defaults package for label lookup

### Changed
- Epic label color changed to lighter green (#3fb950) for better visibility

## Commits

- chore: change epic label to lighter green
- chore: change epic label color to green
- feat(defaults): add bug and enhancement labels to defaults
- feat(api): auto-create default labels when assigning non-existent labels
- docs: move E2E Test Script proposal to Implemented
- docs: mark PRD-E2E-Test-Script as Complete
