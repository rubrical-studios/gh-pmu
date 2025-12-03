# Sprint 2 Backlog: gh-pm-unified

**Sprint Goal:** Complete issue CRUD operations and establish sub-issue foundation
**Sprint Duration:** 2025-12-02 to TBD
**Total Story Points:** 24

---

## Selected Stories

### Story 1.3: View Issue with Project Fields

**As a** developer investigating an issue
**I want** to view an issue with all its project metadata
**So that** I can see the complete context including custom fields

**Acceptance Criteria:**
- [ ] `gh pmuview <issue>` displays issue details
- [ ] Shows all project field values (Status, Priority, custom fields)
- [ ] Shows sub-issues if any exist
- [ ] Shows parent issue if this is a sub-issue
- [ ] Supports `--json` output format

**Story Points:** 3
**Status:** Done

---

### Story 1.4: Create Issue with Project Fields

**As a** developer creating a new issue
**I want** to set project fields during creation
**So that** the issue is properly categorized from the start

**Acceptance Criteria:**
- [ ] `gh pmucreate` opens editor for issue body
- [ ] `--title`, `--body` flags for non-interactive creation
- [ ] `--status`, `--priority` set project field values
- [ ] Automatically adds issue to configured project
- [ ] Applies default labels from config
- [ ] Returns issue URL on success

**Story Points:** 5
**Status:** Done

---

### Story 1.5: Move/Update Issue Fields

**As a** developer updating issue status
**I want** to change project fields from the command line
**So that** I can update status without opening the web UI

**Acceptance Criteria:**
- [ ] `gh pmumove <issue> --status <value>` updates status
- [ ] `gh pmumove <issue> --priority <value>` updates priority
- [ ] Supports field aliases from config (e.g., `in_progress` → "In Progress")
- [ ] Can update multiple fields in one command
- [ ] Shows confirmation of changes made

**Story Points:** 3
**Status:** Done

---

### Story 1.8: Add Sub-Issue Link

**As a** developer organizing work
**I want** to link an existing issue as a sub-issue of another
**So that** I can create issue hierarchies

**Acceptance Criteria:**
- [ ] `gh pmusub add <parent> <child>` links issues
- [ ] Validates both issues exist
- [ ] Uses GraphQL API with `sub_issues` feature header
- [ ] Shows confirmation with parent and child titles
- [ ] Errors gracefully if already linked

**Story Points:** 3
**Status:** Done

---

### Story 1.9: Create Sub-Issue

**As a** developer breaking down work
**I want** to create a new issue directly as a sub-issue
**So that** I can add child tasks without manual linking

**Acceptance Criteria:**
- [ ] `gh pmusub create --parent <id> --title <title>` creates sub-issue
- [ ] Inherits labels from parent (configurable in settings)
- [ ] Inherits assignees from parent (configurable)
- [ ] Inherits milestone from parent (configurable)
- [ ] Automatically links to parent
- [ ] Returns new issue URL

**Story Points:** 5
**Status:** Done

---

### Story 1.10: List Sub-Issues

**As a** developer reviewing task breakdown
**I want** to list all sub-issues of a parent issue
**So that** I can see the full scope of work

**Acceptance Criteria:**
- [x] `gh pmusub list <parent>` shows sub-issues
- [x] Displays title, status, assignee for each
- [x] Shows completion count (X of Y done)
- [x] Supports `--json` output format
- [x] Shows "no sub-issues" if none exist

**Story Points:** 3
**Status:** Done

---

### Story 1.11: Remove Sub-Issue Link

**As a** developer reorganizing work
**I want** to unlink a sub-issue from its parent
**So that** I can restructure issue hierarchies

**Acceptance Criteria:**
- [x] `gh pmusub remove <parent> <child>` unlinks issues
- [x] Does not delete the child issue, only removes link
- [x] Shows confirmation of unlink
- [x] Errors gracefully if not linked

**Story Points:** 2
**Status:** Done

---

## Sprint Progress

**Completed:** 24 story points
**Remaining:** 0 story points
**Velocity:** -

---

## Notes & Blockers

- Sub-issue stories (1.8-1.11) require GraphQL `sub_issues` feature header
- Reference gh-sub-issue source for mutation patterns
