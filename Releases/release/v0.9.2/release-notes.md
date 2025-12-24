# Release v0.9.2

## Highlights

- **Performance**: `gh pmu move` now caches project fields, reducing API calls by 68-80%
- **Accuracy**: `release current` issue count now matches `list --release current`
- **Automation**: `/prepare-release` workflow updates GitHub release notes and cleans up old assets

## What's Changed

### Added
- `/prepare-release` workflow now updates GitHub release notes from CHANGELOG (#439)
  - Parses CHANGELOG.md and updates corresponding GitHub release body
  - Cleans up old release assets, keeping only 3 most recent tagged releases

### Performance
- `gh pmu move` now caches project fields before bulk updates (#451)
  - Eliminates N+1 API calls when updating multiple fields per issue

### Fixed
- `release current` now shows accurate issue count (#449)
  - Uses project items with field filtering instead of label-based query
