# Product Backlog: gh-pm-unified

**Revision:** 3
**Last Updated:** 2025-12-03
**Project Vision:** Create a unified GitHub CLI extension combining project management, sub-issue hierarchy, and project templating for comprehensive GitHub Projects v2 management.

**Major Update:** Stories 2.2, 2.4-2.8 and Epic 4 blocked due to GitHub API limitation - views cannot be created programmatically.

---

## Definition of Done (Global)

All stories must meet these criteria:
- [ ] All acceptance criteria met
- [ ] Unit tests written and passing
- [ ] Code follows Go conventions and project patterns
- [ ] No known bugs
- [ ] Command help text documented
- [ ] README updated (if user-facing feature)

---

## Epic: Core Unification

**Epic Goal:** Merge gh-pm and gh-sub-issue functionality into a unified extension with consistent command structure and shared configuration.

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
**Priority:** High
**Status:** Selected
**Sprint:** 1

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
**Priority:** High
**Status:** Selected
**Sprint:** 1

---

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
**Priority:** Medium
**Status:** Selected
**Sprint:** 2

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
**Priority:** High
**Status:** Selected
**Sprint:** 2

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
**Priority:** High
**Status:** Selected
**Sprint:** 2

---

### Story 1.6: Issue Intake - Find Untracked Issues

**As a** project manager
**I want** to find issues not yet added to the project
**So that** I can ensure all work is tracked on the project board

**Acceptance Criteria:**
- [ ] `gh pmuintake` finds open issues not in the project
- [ ] Shows list of untracked issues with titles
- [ ] `--apply` flag adds them to project with default fields
- [ ] `--dry-run` shows what would be added
- [ ] Respects repository filter from config

**Story Points:** 5
**Priority:** Medium
**Status:** Selected
**Sprint:** 3

---

### Story 1.7: Triage - Bulk Process Issues

**As a** project manager
**I want** to bulk update issues matching certain criteria
**So that** I can efficiently maintain project hygiene

**Acceptance Criteria:**
- [ ] `gh pmutriage <config-name>` runs named triage config
- [ ] Triage configs defined in `.gh-pmu.yml` with query and apply rules
- [ ] Supports applying labels, status, priority changes
- [ ] `--interactive` flag prompts for each issue
- [ ] `--dry-run` shows what would be changed
- [ ] Reports summary of changes made

**Story Points:** 8
**Priority:** Medium
**Status:** Selected
**Sprint:** 3

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
**Priority:** High
**Status:** Selected
**Sprint:** 2

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
**Priority:** High
**Status:** Selected
**Sprint:** 2

---

### Story 1.10: List Sub-Issues

**As a** developer reviewing task breakdown
**I want** to list all sub-issues of a parent issue
**So that** I can see the full scope of work

**Acceptance Criteria:**
- [ ] `gh pmusub list <parent>` shows sub-issues
- [ ] Displays title, status, assignee for each
- [ ] Shows completion count (X of Y done)
- [ ] Supports `--json` output format
- [ ] Shows "no sub-issues" if none exist

**Story Points:** 3
**Priority:** Medium
**Status:** Selected
**Sprint:** 2

---

### Story 1.11: Remove Sub-Issue Link

**As a** developer reorganizing work
**I want** to unlink a sub-issue from its parent
**So that** I can restructure issue hierarchies

**Acceptance Criteria:**
- [ ] `gh pmusub remove <parent> <child>` unlinks issues
- [ ] Does not delete the child issue, only removes link
- [ ] Shows confirmation of unlink
- [ ] Errors gracefully if not linked

**Story Points:** 2
**Priority:** Low
**Status:** Selected
**Sprint:** 2

---

### Story 1.12: Split Issue into Sub-Issues

**As a** developer breaking down an epic
**I want** to split an issue's checklist into sub-issues
**So that** I can convert task lists into trackable issues

**Acceptance Criteria:**
- [ ] `gh pmusplit <issue> --from=body` parses checklist from issue body
- [ ] `gh pmusplit <issue> --from=file.md` parses from external file
- [ ] `gh pmusplit <issue> "Task 1" "Task 2"` creates from arguments
- [ ] Each checklist item becomes a sub-issue
- [ ] Sub-issues linked to parent automatically
- [ ] Shows summary of created sub-issues

**Story Points:** 8
**Priority:** Medium
**Status:** Selected
**Sprint:** 3

---

## Epic: Project Templates & Creation

**Epic Goal:** Enable declarative project creation from YAML templates and existing GitHub projects.

**API Limitation (Discovered 2025-12-03):**
> GitHub API does not support creating project views programmatically. No `createProjectV2View` mutation exists.
> Views can only be created through UI or via `copyProjectV2` mutation.
>
> **Impact:** Stories 2.2, 2.4-2.8 are blocked until GitHub adds view creation API.
> **Recommendation:** Use `--from-project` instead of `--from-template` for full project cloning.

### Story 2.1: Create Project from Existing GitHub Project

**As a** team lead setting up a new project
**I want** to copy an existing project's structure
**So that** I can replicate proven project configurations

**Status:** Done (Sprint 4)

**Acceptance Criteria:**
- [x] `gh pmu project create --from-project <owner>/<number>` copies project
- [x] Copies all custom fields with options
- [x] Copies all views with configurations
- [x] `--title` sets the new project name
- [x] `--owner` specifies target owner (defaults to current user)
- [x] `--include-drafts` optionally copies draft issues
- [x] Returns new project URL

**Story Points:** 8
**Priority:** High

---

### Story 2.2: Create Project from YAML Template

**Status:** Won't Do - GitHub API does not support creating views

**As a** developer starting a new project
**I want** to create a project from a YAML template file
**So that** I can use version-controlled project definitions

**Reason for Won't Do:** GitHub GraphQL API has no `createProjectV2View` mutation. Views cannot be created programmatically. Use `--from-project` to clone existing projects instead.

**Original Acceptance Criteria:**
- [ ] ~~`gh pmu project create --from-template <path>` creates project~~
- [ ] ~~Parses YAML template schema~~
- [ ] ~~Creates all defined fields with options and colors~~
- [ ] ~~Creates all defined views~~ - **BLOCKED**
- [ ] ~~Supports Go template variables~~
- [ ] ~~`--var KEY=VALUE` sets template variables~~
- [ ] ~~`--dry-run` shows what would be created~~
- [ ] ~~Returns new project URL~~

**Story Points:** 13
**Priority:** N/A

---

### Story 2.3: Export Project to YAML Template

**As a** developer who configured a good project setup
**I want** to export my project structure to YAML
**So that** I can reuse it or share it with others

**Status:** Done (Sprint 4)

**Note:** Useful for documentation. To recreate a project, use `--from-project` instead.

**Acceptance Criteria:**
- [x] `gh pmu project export <owner>/<number>` exports to YAML
- [x] `--output <path>` writes to file (default: stdout)
- [x] Exports all custom fields with options
- [x] Exports all views with configurations
- [x] `--minimal` exports fields and views only
- [x] Output validates against template schema

**Removed from scope:**
- ~~`--include-drafts`~~ - Drafts are project-local items, not relevant for templates
- ~~`--include-workflows`~~ - Workflows not accessible via API

**Story Points:** 8
**Priority:** Medium

---

### Story 2.4: Validate Template Syntax

**Status:** Won't Do - No use case without Story 2.2

**As a** template author
**I want** to validate my template before using it
**So that** I can catch errors early

**Reason for Won't Do:** Template validation has no practical use since templates cannot create full projects (Story 2.2 blocked).

**Original Acceptance Criteria:**
- [ ] ~~`gh pmu template validate <path>` validates template~~
- [ ] ~~Checks YAML syntax~~
- [ ] ~~Validates against template schema~~
- [ ] ~~Reports field count, view count, etc.~~
- [ ] ~~Shows detailed errors with line numbers~~
- [ ] ~~Exit code 0 for valid, non-zero for invalid~~

**Story Points:** 5
**Priority:** N/A

---

### Story 2.5: List Available Templates

**Status:** Blocked - Depends on Story 2.2

**As a** developer exploring options
**I want** to list available project templates
**So that** I can see what's available to use

**Blocked because:** No point listing templates that cannot create full projects.

**Story Points:** 5
**Priority:** N/A

---

### Story 2.6: Show Template Details

**Status:** Blocked - Depends on Story 2.2

**As a** developer evaluating a template
**I want** to see detailed template contents
**So that** I can decide if it fits my needs

**Blocked because:** No point showing template details that cannot create full projects.

**Story Points:** 3
**Priority:** N/A

---

### Story 2.7: Built-in Project Templates

**Status:** Blocked - Depends on Story 2.2

**As a** developer wanting quick project setup
**I want** built-in templates for common workflows
**So that** I don't need to create templates from scratch

**Blocked because:** Built-in templates have no value if they cannot create full projects with views.

**Workaround:** Create template projects on GitHub and use `--from-project` to clone them.

**Story Points:** 8
**Priority:** N/A

---

### Story 2.8: Initialize with Template

**Status:** Blocked - Depends on Story 2.2

**As a** developer setting up a new repository
**I want** to init configuration and create project from template together
**So that** I can bootstrap a project in one command

**Blocked because:** Template-based initialization depends on Story 2.2.

**Note:** `--from-project <owner>/<number>` could still be implemented as it uses `copyProjectV2`. Consider splitting into separate story.

**Story Points:** 5
**Priority:** N/A

---

## Epic: Enhanced Integration

**Epic Goal:** Deep integration between sub-issues and project management with cross-repo support.

### Story 3.1: Native Sub-Issue Handling in Split

**As a** developer using split command
**I want** split to work without external extensions
**So that** I don't need separate gh-sub-issue installed

**Acceptance Criteria:**
- [ ] `gh pmusplit` uses internal sub-issue API code
- [ ] No dependency on gh-sub-issue extension
- [ ] Same functionality as current split + sub-issue combo
- [ ] Maintains backward compatibility

**Story Points:** 5
**Priority:** High
**Status:** Backlog
**Sprint:** -

---

### Story 3.2: Cross-Repository Sub-Issues

**As a** developer with multi-repo projects
**I want** to create sub-issues in different repositories
**So that** I can organize work across my codebase

**Acceptance Criteria:**
- [ ] `gh pmusub add` works across repositories
- [ ] `gh pmusub create --repo <owner/repo>` creates in specified repo
- [ ] Parent can be in different repo than child
- [ ] Validates repos are in same project
- [ ] Shows repo info in sub list output

**Story Points:** 8
**Priority:** Medium
**Status:** Backlog
**Sprint:** -

---

### Story 3.3: Sub-Issue Progress Tracking

**As a** project manager reviewing epics
**I want** to see sub-issue completion percentages
**So that** I can track progress on large work items

**Acceptance Criteria:**
- [ ] `gh pmuview <issue>` shows progress bar for parents
- [ ] Shows "3 of 5 sub-issues complete (60%)"
- [ ] `gh pmulist --has-sub-issues` filters to parent issues
- [ ] Progress based on closed/total sub-issue count

**Story Points:** 5
**Priority:** Medium
**Status:** Backlog
**Sprint:** -

---

### Story 3.4: Recursive Operations on Issue Trees

**As a** developer managing large epics
**I want** to perform bulk operations on issue trees
**So that** I can update parent and all sub-issues together

**Acceptance Criteria:**
- [ ] `gh pmumove <issue> --recursive` updates all sub-issues
- [ ] Works with status, priority, labels changes
- [ ] Shows confirmation of all issues to be updated
- [ ] `--dry-run` shows what would be changed
- [ ] Respects depth limit to prevent runaway

**Story Points:** 8
**Priority:** Low
**Status:** Backlog
**Sprint:** -

---

## Epic: Template Ecosystem

**Epic Goal:** Build a template sharing and discovery ecosystem for the community.

**Status:** Blocked - Depends on Epic 2 template functionality

This epic is entirely blocked because it depends on the ability to create projects from templates (Story 2.2), which is not possible due to GitHub API limitations (no view creation API).

### Story 4.1: Remote Template Registry

**As a** developer looking for templates
**I want** to browse and use community templates
**So that** I can benefit from others' project configurations

**Acceptance Criteria:**
- [ ] `gh pmutemplate list --remote` shows registry templates
- [ ] `gh pmutemplate search <query>` searches registry
- [ ] `gh pmuproject create --from-template registry:<name>` uses remote
- [ ] Registry hosted on GitHub (repo or gist-based)
- [ ] Templates verified for schema compliance

**Story Points:** 13
**Priority:** Low
**Status:** Backlog
**Sprint:** -

---

### Story 4.2: Template Inheritance

**As a** template author
**I want** to extend existing templates
**So that** I can build on common configurations

**Acceptance Criteria:**
- [ ] Templates support `extends: <template-name>` field
- [ ] Child templates inherit fields, views from parent
- [ ] Child can override or add to parent definitions
- [ ] Works with built-in and local parent templates
- [ ] Circular inheritance detected and prevented

**Story Points:** 8
**Priority:** Low
**Status:** Backlog
**Sprint:** -

---

### Story 4.3: Template Versioning

**As a** template user
**I want** templates to have semantic versions
**So that** I can manage template updates safely

**Acceptance Criteria:**
- [ ] Templates have `schema_version` field
- [ ] Tool validates compatibility with template schema
- [ ] Migration guidance for breaking schema changes
- [ ] Warning for deprecated fields

**Story Points:** 5
**Priority:** Low
**Status:** Backlog
**Sprint:** -

---

### Story 4.4: Publish Template to Registry

**As a** template author
**I want** to share my template with the community
**So that** others can benefit from my configuration

**Acceptance Criteria:**
- [ ] `gh pmutemplate publish <path>` submits to registry
- [ ] Validates template before submission
- [ ] Requires template metadata (name, description, author)
- [ ] Creates PR to registry repo (or similar mechanism)
- [ ] Author can update/deprecate published templates

**Story Points:** 8
**Priority:** Low
**Status:** Backlog
**Sprint:** -

---

## Technical Debt & Improvements

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
**Priority:** High
**Status:** Selected
**Sprint:** 1

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
**Priority:** High
**Status:** Selected
**Sprint:** 1

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
**Priority:** High
**Status:** Selected
**Sprint:** 1

---

### Tech Story: Release Build & Distribution

**Description:** Configure GitHub Actions to build and publish the extension as installable releases, mirroring gh-pm workflows.

**Benefit:** Users can install gh-pmu via `gh extension install` from published releases.

**Acceptance Criteria:**
- [x] Build workflow: multi-OS (ubuntu, macos) + Go 1.22/1.23 matrix
- [x] Lint workflow: golangci-lint + gofmt + go vet
- [x] Test workflow: Go version matrix + coverage threshold
- [x] Release workflow: cli/gh-extension-precompile
- [x] `gh extension install rubrical-studios/gh-pmu` works from release
- [x] README updated with installation instructions

**Story Points:** 5
**Priority:** High
**Status:** In Progress
**Sprint:** 6

---

### Tech Story: Increase Test Coverage to 80%

**Description:** Add comprehensive unit tests to achieve 80% code coverage.

**Benefit:** Higher confidence in code quality and fewer regressions.

**Acceptance Criteria:**
- [ ] cmd package coverage >= 80%
- [ ] internal/api package coverage >= 80%
- [ ] Mock GitHub API calls for testing run* functions
- [ ] Update CI threshold from 15% to 80%

**Story Points:** 13
**Priority:** Medium
**Status:** Backlog
**Sprint:** -

---

### Story 1.13: Enhanced Init UX with Project Discovery

**As a** developer initializing gh-pmu in a repository
**I want** the init command to automatically detect my repository and suggest associated projects
**So that** I can complete setup with minimal manual input and enjoy a polished experience

**Acceptance Criteria:**
- [x] Auto-detect current repository from git remote (already implemented)
- [x] Query GitHub API for projects associated with the repository owner
- [x] Present a numbered list of available projects for selection
- [x] Allow user to select project by number instead of typing project number manually
- [x] Show project title and number in the selection list
- [x] Fall back to manual entry if no projects found or user chooses "other"
- [x] Support both user and organization projects
- [x] Handle repositories with multiple associated projects gracefully

**UX Requirements:**
- [x] Show a spinner/progress indicator while fetching data from GitHub API
- [x] Use color coding: cyan for prompts, green for success, yellow for warnings
- [x] Display a styled header/banner for the init wizard
- [x] Group related information in visual boxes or sections
- [x] Show checkmarks (✓) for completed steps
- [x] Provide clear step numbers (Step 1 of 3, etc.)
- [x] Gracefully handle terminal width constraints
- [x] Final summary displayed in a formatted box with all configured values

**Example Output:**
```
╭─────────────────────────────────────────╮
│  gh-pmu init                            │
│  Configure project management settings  │
╰─────────────────────────────────────────╯

✓ Detected repository: rubrical-studios/gh-pmu

⠋ Fetching projects for rubrical-studios...

Step 1 of 2: Select Project
┌─────────────────────────────────────────┐
│  1. gh-pmu (#11)                        │
│  2. dotfiles (#8)                       │
│  3. website (#5)                        │
│  0. Enter project number manually       │
└─────────────────────────────────────────┘
Select [1]:

✓ Project: gh-pmu (#11)

Step 2 of 2: Confirm Repository
Repository [rubrical-studios/gh-pmu]:

⠋ Fetching project fields...

╭─────────────────────────────────────────╮
│  ✓ Configuration saved                  │
│                                         │
│  Project:    gh-pmu (#11)               │
│  Repository: rubrical-studios/gh-pmu    │
│  Fields:     12 cached                  │
│  Config:     .gh-pmu.yml                │
╰─────────────────────────────────────────╯
```

**Story Points:** 8
**Priority:** Medium
**Status:** Done
**Sprint:** 7

---

## Icebox (Future Considerations)

Stories that are not prioritized but worth capturing:

- AI-assisted template generation from project description
- Slack/Teams notifications for project changes
- Time tracking integration
- Burndown chart generation
- Sprint planning assistant
- Dependency tracking between issues
- Auto-assignment based on workload
- Template marketplace with ratings/reviews
