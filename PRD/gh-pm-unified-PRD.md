# Product Requirements Document: gh-pm-unified

**Version:** 1.0
**Date:** 2025-12-02
**Author:** rubrical-studios
**Status:** Draft | Ready for Backlog Creation

---

## Product Vision

### Vision Statement
Create a unified GitHub CLI extension that combines project management, sub-issue hierarchy, and project templating into a single, cohesive tool for comprehensive GitHub Projects v2 management from the command line.

### Target Users
- **CLI-first developers** who prefer command-line workflows over web UI
- **Project managers** who need batch operations and automation
- **DevOps engineers** who integrate project management into CI/CD pipelines
- **Open source maintainers** managing issues across multiple repositories

### Key Value Proposition
Eliminate the need for multiple fragmented extensions (`gh-pm`, `gh-sub-issue`) by providing a unified tool with consistent command patterns, shared configuration, and new project templating capabilities not available in core `gh` CLI.

---

## Feature Areas (Future Epics)

### Epic 1: Core Unification
**Description:** Merge functionality from gh-pm and gh-sub-issue into a unified extension with consistent command structure.

**Capabilities:**
- Initialize project configuration (`gh pmuinit`)
- List issues in project with project metadata (`gh pmulist`)
- View issue with project fields (`gh pmuview`)
- Create issue with project fields pre-populated (`gh pmucreate`)
- Move/update issue status and priority (`gh pmumove`)
- Find and add untracked issues to project (`gh pmuintake`)
- Bulk process issues with configurable rules (`gh pmutriage`)
- Add existing issue as sub-issue (`gh pmusub add`)
- Create new sub-issue under parent (`gh pmusub create`)
- List sub-issues of a parent issue (`gh pmusub list`)
- Remove/unlink sub-issue from parent (`gh pmusub remove`)
- Split issue into sub-issues from checklist or arguments (`gh pmusplit`)

**Success Criteria:**
- All commands from both source extensions work in unified tool
- Single `.gh-pmu.yml` configuration file serves all commands
- Consistent flag patterns across all commands
- Backward-compatible with existing gh-pm configurations

### ~~Epic 2: Project Templates & Creation~~ - REMOVED

**Status:** Removed from scope (2025-12-03)

**Reason:** This epic provided no unique value over native `gh project` commands:
- `gh pmu project create --from-project` → Use `gh project copy` instead
- `gh pmu project export` → Use `gh project field-list` + `gh project view` instead
- Template-based creation → Blocked by GitHub API (no view creation API)

**Recommendation:** Use native `gh project` commands for project management. Focus gh-pmu on unique value: sub-issue hierarchy, intake/triage automation, and enhanced integration.

### Epic 3: Enhanced Integration
**Description:** Deep integration between sub-issues and project management features.

**Capabilities:**
- Native sub-issue handling in split command (remove external dependency)
- Cross-repository sub-issue support
- Progress tracking showing sub-issue completion percentages
- Recursive operations on full issue trees (bulk status updates, etc.)

**Success Criteria:**
- Split command works without gh-sub-issue installed
- Sub-issues can be created across different repositories
- Parent issues show accurate completion status
- Bulk operations respect issue hierarchies

### ~~Epic 4: Template Ecosystem~~ - REMOVED

**Status:** Removed from scope (2025-12-03)

**Reason:** This epic depended entirely on Epic 2 template functionality, which was removed.

---

## Non-Functional Expectations

| Category | Expectation |
|----------|-------------|
| Performance | Commands complete within 2-3 seconds for typical operations |
| Security | Respect GitHub token scopes; never store tokens in config files |
| Usability | Consistent with `gh` CLI conventions; helpful error messages |
| Compatibility | Support Go 1.21+; work with GitHub.com and GitHub Enterprise |
| Extensibility | Clean package structure enabling future capability additions |

---

## Constraints

- **GitHub API Limitations:** Sub-issues require special GraphQL headers (`sub_issues`, `issue_types`)
- **View Creation API:** GitHub Projects v2 views CANNOT be created via API - no `createProjectV2View` mutation exists (verified 2025-12-03). Views can only be created through UI or by copying an existing project via `copyProjectV2`.
- **Status Field Reserved:** New projects have a default Status field that cannot be replaced via API.
- **Workflow API:** GitHub Project workflows may not be fully creatable via API (needs verification)
- **License Compliance:** Must maintain MIT license and proper attribution to source projects
- **Backward Compatibility:** Existing `.gh-pmu.yml` files must continue to work

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Consistent with gh CLI, source projects, and extension ecosystem |
| CLI Framework | Cobra | Industry standard, used by gh CLI and source projects |
| Configuration | Viper + YAML | Flexible parsing, environment variable support |
| Templating | Go text/template | Built-in, no external dependencies |
| API Client | go-gh | Official gh extension library with auth handling |

---

## Testing Approach

| Approach | Status | Notes |
|----------|--------|-------|
| TDD | Required | All business logic test-driven |
| ATDD | [ ] Used / [x] Not Used | CLI testing via integration tests |
| BDD | [ ] Used / [x] Not Used | Not applicable for CLI tool |

**Testing Strategy:**
- Unit tests for all `pkg/` packages (config, template, API clients)
- Integration tests for commands using mock GitHub API
- End-to-end tests against test GitHub projects (optional, CI only)

---

## Dependencies

### Source Projects (MIT Licensed)
- [gh-pm](https://github.com/yahsan2/gh-pm) by @yahsan2
- [gh-sub-issue](https://github.com/yahsan2/gh-sub-issue) by @yahsan2

### Go Dependencies
- `github.com/cli/go-gh` - Official gh extension library
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

---

## Open Questions (Require Resolution)

| # | Question | Impact | Decision Needed By |
|---|----------|--------|-------------------|
| 1 | Final extension name? | Branding, installation command | Sprint 1 |
| 2 | Collaborate with original author or independent fork? | Community, maintenance | Sprint 1 |
| 3 | How long to maintain backward compatibility aliases? | Migration path complexity | Sprint 2 |
| 4 | Can GitHub Project workflows be fully created via API? | Template feature scope | Sprint 2 |
| 5 | Should there be a central template registry? | Epic 4 scope | Sprint 3+ |

---

## Notes for Backlog Creation

- Feature Areas → Epics
- Capabilities → User Stories
- Success Criteria → Acceptance Criteria
- Constraints → Considered during Sprint Planning
- Open Questions → May block certain stories until resolved

**Suggested Sprint Structure:**
- **Sprint 1:** Core unification (Epic 1 partial) - establish foundation
- **Sprint 2:** Core unification continued (Epic 1)
- **Sprint 3:** Core unification complete (Epic 1) - intake, triage, split
- **Sprint 4:** ~~Project templates~~ - REMOVED (redundant with native gh project commands)
- **Future:** Enhanced integration (Epic 3) - sub-issue hierarchy, cross-repo support

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-02 | rubrical-studios | Initial PRD from PROPOSAL.md |
| 1.1 | 2025-12-03 | rubrical-studios | Removed Epic 2 & 4 - redundant with native gh project commands, template features blocked by API |
