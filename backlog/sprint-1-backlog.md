# Sprint 1 Backlog: gh-pm-unified

**Sprint Goal:** Establish project foundation and core project management commands
**Sprint Duration:** 2025-12-02 to TBD
**Total Story Points:** 28

---

## Selected Stories

### Tech Story: Project Scaffolding

**Description:** Set up Go project structure, CI/CD, and development tooling.

**Benefit:** Foundation for all development work.

**Acceptance Criteria:**
- [ ] Go module initialized with proper naming
- [ ] Cobra CLI structure with root command
- [ ] GitHub Actions for test, lint, build
- [ ] Makefile with common targets
- [ ] .goreleaser.yml for releases
- [ ] README with installation instructions

**Story Points:** 5
**Status:** Selected

---

### Tech Story: Configuration Package

**Description:** Implement configuration loading, validation, and caching.

**Benefit:** Shared infrastructure for all commands.

**Acceptance Criteria:**
- [ ] Load `.gh-pmu.yml` with Viper
- [ ] Validate required fields
- [ ] Cache project metadata from GitHub API
- [ ] Support field aliases
- [ ] Environment variable overrides

**Story Points:** 5
**Status:** Selected

---

### Tech Story: GitHub API Client Package

**Description:** Implement GraphQL client with sub-issue feature support.

**Benefit:** Reusable API layer for all commands.

**Acceptance Criteria:**
- [ ] Use go-gh for authentication
- [ ] Support `sub_issues` and `issue_types` feature headers
- [ ] Common queries for projects, issues, fields
- [ ] Error handling with user-friendly messages
- [ ] Rate limiting awareness

**Story Points:** 8
**Status:** Selected

---

### Story 1.1: Project Configuration Initialization

**As a** developer setting up a new project
**I want** to initialize gh-pmu configuration interactively
**So that** I can quickly configure project settings without manual YAML editing

**Acceptance Criteria:**
- [ ] `gh pmuinit` prompts for project owner, number, and repositories
- [ ] Creates `.gh-pmu.yml` with provided values
- [ ] Auto-detects current repository if in a git repo
- [ ] Fetches and caches project field metadata from GitHub API
- [ ] Validates project exists before saving configuration

**Story Points:** 5
**Status:** Selected

---

### Story 1.2: List Issues with Project Metadata

**As a** developer reviewing project status
**I want** to list issues with their project field values
**So that** I can see status, priority, and other fields at a glance

**Acceptance Criteria:**
- [ ] `gh pmulist` displays issues from configured project
- [ ] Shows Title, Status, Priority, Assignees by default
- [ ] Supports `--status`, `--priority` filters
- [ ] Supports `--json` output format
- [ ] Respects repository filter from config

**Story Points:** 5
**Status:** Selected

---

## Sprint Progress

**Completed:** 28 story points
**Remaining:** 0 story points
**Velocity:** 28 points/sprint

---

## Notes & Blockers

*Sprint started 2025-12-02*
*Sprint completed 2025-12-02*

All stories completed successfully.

---
