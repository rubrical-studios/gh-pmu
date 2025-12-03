# Product Backlog: gh-pm-unified

**Revision:** 2
**Last Updated:** 2025-12-02
**Project Vision:** Create a unified GitHub CLI extension combining project management, sub-issue hierarchy, and project templating for comprehensive GitHub Projects v2 management.

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

### Story 2.1: Create Project from Existing GitHub Project

**As a** team lead setting up a new project
**I want** to copy an existing project's structure
**So that** I can replicate proven project configurations

**Acceptance Criteria:**
- [ ] `gh pmuproject create --from-project <owner>/<number>` copies project
- [ ] Copies all custom fields with options
- [ ] Copies all views with configurations
- [ ] `--title` sets the new project name
- [ ] `--owner` specifies target owner (defaults to current user)
- [ ] `--include-drafts` optionally copies draft issues
- [ ] Returns new project URL

**Story Points:** 8
**Priority:** High
**Status:** Backlog
**Sprint:** -

---

### Story 2.2: Create Project from YAML Template

**As a** developer starting a new project
**I want** to create a project from a YAML template file
**So that** I can use version-controlled project definitions

**Acceptance Criteria:**
- [ ] `gh pmuproject create --from-template <path>` creates project
- [ ] Parses YAML template schema
- [ ] Creates all defined fields with options and colors
- [ ] Creates all defined views
- [ ] Supports Go template variables (`{{.ProjectName}}`, etc.)
- [ ] `--var KEY=VALUE` sets template variables
- [ ] `--dry-run` shows what would be created
- [ ] Returns new project URL

**Story Points:** 13
**Priority:** High
**Status:** Backlog
**Sprint:** -

---

### Story 2.3: Export Project to YAML Template

**As a** developer who configured a good project setup
**I want** to export my project structure to YAML
**So that** I can reuse it or share it with others

**Acceptance Criteria:**
- [ ] `gh pmuproject export <number>` exports to YAML
- [ ] `--output <path>` writes to file (default: stdout)
- [ ] Exports all custom fields with options
- [ ] Exports all views with configurations
- [ ] `--include-drafts` includes draft issues
- [ ] `--include-workflows` includes workflow definitions
- [ ] `--minimal` exports fields and views only
- [ ] Output validates against template schema

**Story Points:** 8
**Priority:** Medium
**Status:** Backlog
**Sprint:** -

---

### Story 2.4: Validate Template Syntax

**As a** template author
**I want** to validate my template before using it
**So that** I can catch errors early

**Acceptance Criteria:**
- [ ] `gh pmutemplate validate <path>` validates template
- [ ] Checks YAML syntax
- [ ] Validates against template schema
- [ ] Reports field count, view count, etc.
- [ ] Shows detailed errors with line numbers
- [ ] Exit code 0 for valid, non-zero for invalid

**Story Points:** 5
**Priority:** Medium
**Status:** Backlog
**Sprint:** -

---

### Story 2.5: List Available Templates

**As a** developer exploring options
**I want** to list available project templates
**So that** I can see what's available to use

**Acceptance Criteria:**
- [ ] `gh pmutemplate list` shows all templates
- [ ] `--builtin` shows only built-in templates
- [ ] `--local` shows only local templates (from config path)
- [ ] Displays name, description, field count for each
- [ ] Built-in templates embedded in binary

**Story Points:** 5
**Priority:** Medium
**Status:** Backlog
**Sprint:** -

---

### Story 2.6: Show Template Details

**As a** developer evaluating a template
**I want** to see detailed template contents
**So that** I can decide if it fits my needs

**Acceptance Criteria:**
- [ ] `gh pmutemplate show <name>` displays template details
- [ ] Shows all fields with types and options
- [ ] Shows all views with configurations
- [ ] Shows workflow definitions if present
- [ ] Works with built-in and local templates

**Story Points:** 3
**Priority:** Low
**Status:** Backlog
**Sprint:** -

---

### Story 2.7: Built-in Project Templates

**As a** developer wanting quick project setup
**I want** built-in templates for common workflows
**So that** I don't need to create templates from scratch

**Acceptance Criteria:**
- [ ] `kanban` template: Simple To Do → In Progress → Done
- [ ] `scrum` template: Sprint-based with story points
- [ ] `bug-tracker` template: Severity, resolution tracking
- [ ] `feature-roadmap` template: Quarters, themes
- [ ] Templates embedded in binary using Go embed
- [ ] Each template has description and tags

**Story Points:** 8
**Priority:** Medium
**Status:** Backlog
**Sprint:** -

---

### Story 2.8: Initialize with Template

**As a** developer setting up a new repository
**I want** to init configuration and create project from template together
**So that** I can bootstrap a project in one command

**Acceptance Criteria:**
- [ ] `gh pmuinit --from-template <path>` creates project and config
- [ ] Creates project using template
- [ ] Creates `.gh-pmu.yml` configured for new project
- [ ] Caches field metadata automatically
- [ ] `--from-project <owner>/<number>` works similarly

**Story Points:** 5
**Priority:** Medium
**Status:** Backlog
**Sprint:** -

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
