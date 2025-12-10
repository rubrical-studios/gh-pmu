# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.1] - 2025-12-10

### Added
- Detailed single-file history view with expanded commit information (#256)
  - Shows full commit messages, relative timestamps ("3 days ago"), and line change stats (+/-)
  - Automatically activates for single file paths
  - `--compact` flag to use original compact format for single files

## [0.4.0] - 2025-12-10

### Added
- `gh pmu history` command - Show git commit history for files/directories with issue references (#255)
  - Parse issue/PR references from commit messages (#123, fixes #456)
  - Infer change type from commit prefixes (Fix:, Add:, feat:, etc.)
  - Lipgloss styled terminal output with color-coded change types
  - JSON output format (`--json`)
  - Markdown file output (`--output`) to History/ directory
  - Safety protections: 25 file limit, repo root guard (`--force` to override)
  - Filter options: `--since`, `--limit`

### Dependencies
- Added `github.com/charmbracelet/lipgloss` v1.1.0 for styled terminal output

## [0.3.1] - 2025-12-09

### Changed
- Repository moved from `scooter-indie/gh-pmu` to `rubrical-studios/gh-pmu` (#254)
- Installation command updated: `gh extension install rubrical-studios/gh-pmu`
- Go module path updated to `github.com/rubrical-studios/gh-pmu`

### Note for Existing Users
Existing installations will continue to work via GitHub redirect. For a clean setup, reinstall:
```bash
gh extension remove gh-pmu
gh extension install rubrical-studios/gh-pmu
```

## [0.3.0] - 2025-12-08

### Added
- `gh pmu board` command - Interactive terminal-based project board view (#250)
- `gh pmu close` command with reason aliases (completed, not_planned, duplicate, reopened) (#249)
- Integration testing infrastructure with `internal/testutil` package
- Comprehensive integration tests for all commands (list, view, create, move, intake, triage, split, sub add/create/list/remove, init)
- GitHub Actions workflow for integration tests

### Fixed
- `os.Chmod` error handling for errcheck linter compliance
- CI workflow improvements for gh extension installation and authentication

### Changed
- Disabled integration tests workflow (pending test environment fixes)

### Documentation
- Added shell construct limitations guide to `gh-workflow.md` (#252)
- Added integration testing guide for test fixtures

## [0.2.13] - 2025-12-04

### Added
- 31 command enhancements (#94-124)

### Fixed
- Use `t.Fatal` for nil checks to satisfy staticcheck

## [0.2.12] - 2025-12-04

### Fixed
- Staticcheck lint error in triage tests (SA5011) - use `t.Fatal` for nil checks

### Notes
- v0.2.11 release failed to build binaries due to lint error; this release includes all v0.2.11 changes plus the fix

## [0.2.11] - 2025-12-04

### Added
- `--repo` / `-R` flag to `triage` command for targeting specific repositories (#91)

### Fixed
- `GetProjectItems` now uses cursor-based pagination to fetch all items (#90)
  - Previously limited to first 100 items, causing "issue not in project" errors for large projects

### Documentation
- Clarified distinction between labels and project fields in `gh-workflow.md`
- Added pagination integration test scenario to backlog (IT-2.4)

## [0.2.10] - 2025-12-04

### Added
- Comprehensive tests for `cmd/intake.go` output functions (100% coverage)
- Comprehensive tests for `cmd/split.go` output functions (100% coverage)
- Comprehensive tests for `cmd/init.go` helper functions (75-100% coverage)
- Comprehensive tests for `internal/ui/ui.go` spinner methods (96.9% coverage)
- Comprehensive tests for `cmd/view.go` output functions
- Comprehensive tests for `cmd/create.go` runCreate function
- Comprehensive tests for `cmd/move.go` core logic
- Comprehensive tests for `cmd/sub.go` output functions
- Comprehensive tests for `internal/api/mutations.go`
- Comprehensive tests for `cmd/triage.go` command
- Integration testing proposal (`Proposal/PROPOSAL-Automated-Testing.md`)
- Integration testing backlog with 23 stories, 90 story points

### Changed
- Test coverage increased from ~15% to 63.6%
- Fixed golangci-lint errcheck warnings for `os.Chdir` in deferred calls
- Renamed test fixtures from `.gh-pm.yml` to `.gh-pmu.yml`

### Coverage (v0.2.10)
| Package | Coverage |
|---------|----------|
| `internal/api` | 96.6% |
| `internal/config` | 97.0% |
| `internal/ui` | 96.9% |
| `cmd` | 51.2% |
| **Total** | **63.6%** |

## [0.2.9] - 2025-12-03

### Added
- Comprehensive test coverage for triage command

### Changed
- Format all Go files with `gofmt -s`

## [0.2.8] - 2025-12-03

### Added
- Generate Markdown coverage report instead of HTML (`coverage/README.md`)

### Changed
- Remove HTML coverage report in favor of Markdown

## [0.2.7] - 2025-12-03

### Added
- Coverage report generation on releases

### Fixed
- Use `-short` flag in coverage tests to skip auth-dependent tests
- Calculate box width using visible text length (strip ANSI codes)
- Use rune count for visible width calculation in box formatting

## [0.2.6] - 2025-12-03

### Fixed
- Use binary format in goreleaser for gh extension compatibility
- Add Windows support to goreleaser config

## [0.2.5] - 2025-12-03

### Changed
- Consolidate CI workflows with sequential execution

### Fixed
- Format ui.go and remove unused noColor field in Spinner

## [0.2.4] - 2025-12-03

### Added
- Enhanced init UX with project discovery and styled output (Story 1.13)
  - Auto-detect repository from git remote
  - Query GitHub API for associated projects
  - Present numbered list for project selection
  - Styled output with spinners, boxes, and color coding

### Fixed
- Correct gh-sub-issue attribution to yahsan2
- Format cmd/init.go

## [0.2.3] - 2025-12-03

### Changed
- Switch to goreleaser for releases

## [0.2.2] - 2025-12-03

### Fixed
- Simplify release workflow - use default build

## [0.2.1] - 2025-12-03

### Fixed
- init command now creates complete config file (#71)
- Use t.Fatal for nil flag checks (SA5011)
- Address golangci-lint errors
- Format code and downgrade deps for Go 1.22/1.23 compatibility
- Set go.mod to 1.22 for CI compatibility

## [0.2.0] - 2025-12-03

### Added
- Mirror gh-pm CI/CD workflows (#69)
- Release workflow for gh extension install

### Changed
- Rename to gh-pmu for shorter command

### Fixed
- Skip auth-dependent tests in CI
- Use binary format for gh extension install compatibility

## [0.1.0] - 2025-12-03

### Added - Initial Release
- **Project scaffolding**
  - Go module with Cobra CLI framework
  - Makefile with build, test, lint targets
  - GitHub Actions for CI/CD

- **Configuration package** (`internal/config`)
  - Load `.gh-pmu.yml` with Viper
  - Validate required fields
  - Support field aliases
  - Cache project metadata from GitHub API

- **GitHub API client** (`internal/api`)
  - GraphQL client using go-gh library
  - Support for `sub_issues` and `issue_types` feature headers
  - Queries: GetProject, GetProjectFields, GetProjectItems, GetIssue, GetSubIssues, GetParentIssue, GetRepositoryIssues
  - Mutations: CreateIssue, AddIssueToProject, SetProjectItemField, AddSubIssue, RemoveSubIssue

- **Project Management Commands**
  - `gh pmu init` - Interactive project configuration setup
  - `gh pmu list` - List issues with project field values
  - `gh pmu view` - View issue with all project fields and sub-issues
  - `gh pmu create` - Create issue with project fields pre-populated
  - `gh pmu move` - Update issue project fields

- **Sub-Issue Commands**
  - `gh pmu sub add` - Link existing issue as sub-issue
  - `gh pmu sub create` - Create new sub-issue under parent
  - `gh pmu sub list` - List sub-issues with completion count
  - `gh pmu sub remove` - Unlink sub-issue from parent

- **Batch Operations**
  - `gh pmu intake` - Find and add untracked issues to project
  - `gh pmu triage` - Bulk update issues based on config rules
  - `gh pmu split` - Create sub-issues from checklist or arguments

- **Enhanced Integration**
  - Cross-repository sub-issues
  - Sub-issue progress tracking (progress bar in view)
  - Recursive operations (`--recursive` flag for move)

### Attribution
- Based on [gh-pm](https://github.com/yahsan2/gh-pm) and [gh-sub-issue](https://github.com/yahsan2/gh-sub-issue) by [@yahsan2](https://github.com/yahsan2)

---

## Development History

### Pre-release (2025-11-30 to 2025-12-02)

- 2025-11-30: Initial repository setup, README.md
- 2025-12-01: Add proposal document
- 2025-12-02: Add Agile PRD, product backlog, code integration inventory
- 2025-12-02: Complete Sprint 1 (Foundation), Sprint 2 (Core Commands), Sprint 3 (Batch Operations)
- 2025-12-03: Complete Sprint 5 (Enhanced Integration)
- 2025-12-03: Remove Epic 2 & 4 (GitHub API limitations - no view creation API)

---

## Sprint Summary

| Sprint | Focus | Status |
|--------|-------|--------|
| Sprint 1 | Foundation, init, list | Complete |
| Sprint 2 | view, create, move, sub-issues | Complete |
| Sprint 3 | intake, triage, split | Complete |
| Sprint 4 | Epic 2 (Templates) | Removed - API limitations |
| Sprint 5 | Enhanced Integration | Complete |
| Sprint 6 | Test Coverage | Complete (63.6%) |

---

## API Limitations Discovered

### GitHub Projects V2 API
- **No `createProjectV2View` mutation** - Views cannot be created programmatically
- **Status field reserved** - New projects have a default Status field that cannot be replaced
- **Workflows not accessible** - Project automation workflows cannot be read or written via API

These limitations led to removing Epic 2 (Project Templates) from scope.

[Unreleased]: https://github.com/rubrical-studios/gh-pmu/compare/v0.4.1...HEAD
[0.4.1]: https://github.com/rubrical-studios/gh-pmu/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/rubrical-studios/gh-pmu/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/rubrical-studios/gh-pmu/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.13...v0.3.0
[0.2.13]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.12...v0.2.13
[0.2.12]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.11...v0.2.12
[0.2.11]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.10...v0.2.11
[0.2.10]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.9...v0.2.10
[0.2.9]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.8...v0.2.9
[0.2.8]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.7...v0.2.8
[0.2.7]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/rubrical-studios/gh-pmu/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/rubrical-studios/gh-pmu/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/rubrical-studios/gh-pmu/releases/tag/v0.1.0
