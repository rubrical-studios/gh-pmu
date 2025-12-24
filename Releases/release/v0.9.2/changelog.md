# Changelog for v0.9.2

## [0.9.2] - 2025-12-23

### Added
- `/prepare-release` workflow now updates GitHub release notes from CHANGELOG (#439)
  - Parses CHANGELOG.md and updates corresponding GitHub release body
  - Cleans up old release assets, keeping only 3 most recent tagged releases
  - New scripts: `update-release-notes.js`, `cleanup-release-assets.js`

### Performance
- `gh pmu move` now caches project fields before bulk updates (#451)
  - Eliminates N+1 API calls when updating multiple fields per issue
  - 68-80% reduction in API calls for typical multi-field updates

### Fixed
- `release current` now shows accurate issue count (#449)
  - Uses project items with field filtering instead of label-based query
  - Fixes discrepancy between `release current` and `list --release current`
