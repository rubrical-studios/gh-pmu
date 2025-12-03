# Sprint 2 Summary: gh-pm-unified

**Sprint Goal:** Complete issue CRUD operations and establish sub-issue foundation
**Sprint Duration:** 2025-12-02

---

## Completed Stories

- Story 1.3: View Issue with Project Fields - 3 points
- Story 1.4: Create Issue with Project Fields - 5 points
- Story 1.5: Move/Update Issue Fields - 3 points
- Story 1.8: Add Sub-Issue Link - 3 points
- Story 1.9: Create Sub-Issue - 5 points
- Story 1.10: List Sub-Issues - 3 points
- Story 1.11: Remove Sub-Issue Link - 2 points

**Total Completed:** 24 story points

---

## Incomplete Stories

None - all planned stories completed.

**Carried Over:** 0 story points

---

## Velocity

**Planned:** 24 points
**Completed:** 24 points
**Velocity:** 24 points/sprint
**Average Velocity (2 sprints):** 26 points/sprint

---

## Key Achievements

- `gh pmuview <issue>` - View issue with project metadata, sub-issues, and parent info
- `gh pmucreate` - Create issues with automatic project integration and field setting
- `gh pmumove <issue>` - Update project fields (Status, Priority) with alias support
- `gh pmusub add` - Link existing issues as sub-issues
- `gh pmusub create` - Create new issues directly as sub-issues with label inheritance
- `gh pmusub list` - List sub-issues with progress tracking (X/Y complete)
- `gh pmusub remove` - Unlink sub-issues from parents

---

## Technical Highlights

- GraphQL mutations for issue creation (`createIssue`)
- Project item field updates (`updateProjectV2ItemFieldValue`)
- Sub-issue API integration (`addSubIssue`, `removeSubIssue`)
- Issue reference parsing supporting multiple formats (123, #123, owner/repo#123)
- JSON output support for all commands
- Label inheritance from parent issues

---

## Challenges Encountered

- GraphQL `parent` field naming (API uses `parent`, not `parentIssue`)
- GraphQL mutation response structure differences between queries
- Windows path handling for Go executable
- Integration testing constraints (created test projects #17 and #18)

---

## New Stories Discovered

- Add `--assignee` flag to create command
- Add milestone inheritance to sub-issue creation
- Add `gh pmusub move` to re-parent sub-issues
- Add batch sub-issue creation from template
- Add progress bar visualization for sub-issue lists

---

## Commands Available After Sprint 2

| Command | Description |
|---------|-------------|
| `gh pmuinit` | Initialize project configuration |
| `gh pmulist` | List issues with project metadata |
| `gh pmuview <issue>` | View issue details |
| `gh pmucreate` | Create new issue |
| `gh pmumove <issue>` | Update project fields |
| `gh pmusub add` | Link as sub-issue |
| `gh pmusub create` | Create sub-issue |
| `gh pmusub list` | List sub-issues |
| `gh pmusub remove` | Unlink sub-issue |
