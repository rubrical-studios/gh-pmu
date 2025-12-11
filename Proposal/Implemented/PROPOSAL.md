# gh-pm-unified: A Unified GitHub CLI Extension for Project Management

## Proposal Document

**Version:** 1.0.0
**Date:** November 30, 2025
**Status:** Implemented
**Implemented:** December 10, 2025 (v0.4.4)

---

## Implementation Notes

**Phase 1 (Core Unification):** ✅ Complete - All commands implemented and released.

**Phase 2 (Project Templates):** ❌ Descoped - GitHub API does not support programmatic creation of project views or workflows. See [CHANGELOG.md](../CHANGELOG.md) for details.

**Future Roadmap:** See [feature-roadmap-v0.5.md](feature-roadmap-v0.5.md) for planned enhancements.

---

## Executive Summary

This proposal outlines the creation of a unified GitHub CLI extension that combines the functionality of two existing open-source projects into a single, cohesive tool for comprehensive GitHub project and issue management from the command line.

The unified extension provides:
- Full GitHub Projects v2 management
- Issue hierarchy (sub-issues/parent-child relationships)
- Workflow automation (triage, intake, splitting)
- ~~Project templating from GitHub projects and YAML definitions~~ (descoped)
- Extensible architecture for future capabilities

---

## Attribution & Source Projects

This project builds upon and incorporates work from the following open-source repositories:

### gh-pm
- **Repository:** https://github.com/yahsan2/gh-pm
- **Author:** [@yahsan2](https://github.com/yahsan2)
- **License:** MIT
- **Description:** A GitHub CLI extension for project management with GitHub Projects (v2) and Issues. Provides workflow automation including triage, intake, issue splitting, and field management.

### gh-sub-issue
- **Repository:** https://github.com/yahsan2/gh-sub-issue
- **Author:** [@yahsan2](https://github.com/yahsan2)
- **License:** MIT
- **Description:** A GitHub CLI extension for managing sub-issues (child issues). Enables hierarchical task structures by linking issues as parent-child relationships.

Both projects are authored by the same developer and share compatible architectures, making unification straightforward while preserving all existing functionality.

---

## Motivation

### Current State

Users who want comprehensive GitHub project management from the CLI currently need:

1. **Core `gh` CLI** - Basic project CRUD operations
2. **gh-pm extension** - Workflow automation, triage, intake
3. **gh-sub-issue extension** - Issue hierarchy management

This fragmentation creates several issues:
- Multiple extensions to install and maintain
- Potential version compatibility issues
- Duplicated code and dependencies
- Inconsistent command patterns
- No unified configuration

### Gap in Core GitHub CLI

The official GitHub CLI (`gh`) provides `gh project` commands for basic operations but lacks:

| Feature | Core `gh` | Needed |
|---------|-----------|--------|
| Sub-issue management | ❌ | ✅ |
| Bulk triage operations | ❌ | ✅ |
| Issue intake (find untracked issues) | ❌ | ✅ |
| Configuration file support | ❌ | ✅ |
| Friendly field aliases | ❌ | ✅ |
| Issue splitting from checklists | ❌ | ✅ |
| YAML-based project templates | ❌ | ✅ |
| Create project from template (simplified) | Partial | ✅ |

There is an [open issue (#10298)](https://github.com/cli/cli/issues/10298) on the GitHub CLI repository requesting sub-issue support, with 64+ upvotes, indicating strong community demand.

### Gap in Project Templating

The core `gh` CLI supports:
- `gh project copy` - Copy an existing project's structure
- `gh project mark-template` - Mark a project as a template

However, it **lacks**:
- Creating projects from a declarative YAML definition
- Version-controlled project templates
- Portable, shareable project configurations
- Automatic field creation from template

---

## Proposed Solution

### Project Name Options

- `gh-pm-unified`
- `gh-project-manager`
- `gh-pm-plus`
- `gh-projects-extended`

### Command Structure

The unified extension will use a consistent command hierarchy:

```
gh pmu<command> [subcommand] [flags]
```

#### Core Commands (from gh-pm)

```bash
# Configuration
gh pmuinit                    # Initialize project configuration

# Issue Management
gh pmulist                    # List issues in project
gh pmuview <issue>            # View issue with project metadata
gh pmucreate                  # Create issue with project fields
gh pmumove <issue>            # Update issue status/priority

# Workflow Automation
gh pmuintake                  # Find and add untracked issues
gh pmutriage <config>         # Bulk process issues with rules
```

#### Sub-issue Commands (from gh-sub-issue)

```bash
# Sub-issue Management
gh pmusub add <parent> <child>      # Link existing issue as sub-issue
gh pmusub create --parent <id>      # Create new sub-issue
gh pmusub list <parent>             # List sub-issues of parent
gh pmusub remove <parent> <child>   # Unlink sub-issue
```

#### Issue Splitting (combines both)

```bash
# Split parent into sub-issues
gh pmusplit <issue> --from=body     # Parse checklist from issue body
gh pmusplit <issue> --from=file.md  # Parse from external file
gh pmusplit <issue> "Task 1" "Task 2" # Direct task arguments
```

#### Project Template Commands (NEW)

```bash
# Create project from existing GitHub project template
gh pmuproject create --from-project <owner>/<project-number>
gh pmuproject create --from-project github/4247 --title "My Roadmap"

# Create project from YAML template file
gh pmuproject create --from-template ./templates/scrum-board.yml
gh pmuproject create --from-template ./templates/kanban.yml --title "Q1 Sprint"

# Export existing project to YAML template
gh pmuproject export <project-number> --output ./my-template.yml
gh pmuproject export 1 --owner my-org --output ./templates/team-board.yml

# List available templates (local + remote registry)
gh pmutemplate list
gh pmutemplate list --local          # Only local templates
gh pmutemplate list --remote         # Only from template registry

# Validate a template file
gh pmutemplate validate ./my-template.yml

# Initialize with template selection
gh pmuinit --from-template ./templates/scrum.yml
gh pmuinit --from-project my-org/5
```

### Configuration File

Unified `.gh-pmu.yml` configuration:

```yaml
# Project settings
project:
  name: "My Project"
  number: 1
  org: "my-organization"
  owner: "username"

# Repositories
repositories:
  - owner/repo1
  - owner/repo2

# Default values for new issues
defaults:
  priority: p2
  status: backlog
  labels:
    - "pm-tracked"

# Field mappings (aliases → actual values)
fields:
  priority:
    field: "Priority"
    values:
      p0: "P0"
      p1: "P1"
      p2: "P2"
      critical: "P0"
      high: "P1"
      medium: "P2"

  status:
    field: "Status"
    values:
      backlog: "Backlog"
      ready: "Ready"
      in_progress: "In Progress"
      in_review: "In Review"
      done: "Done"

# Sub-issue settings
sub_issues:
  inherit_labels: true
  inherit_assignees: true
  inherit_milestone: true
  exclude_labels:
    - "epic"
    - "meta"

# Triage configurations
triage:
  tracked:
    query: "is:issue is:open -label:pm-tracked"
    apply:
      labels:
        - pm-tracked
      fields:
        status: backlog

  estimate:
    query: "is:issue is:open status:backlog -has:estimate"
    interactive:
      estimate: true

# Metadata cache (auto-generated)
metadata:
  project:
    id: "PVT_xxx"
  fields:
    status:
      id: "PVTSSF_xxx"
      options:
        backlog: "abc123"
        ready: "def456"
```

---

## Project Template Specification

### YAML Template Schema

Project templates are defined in YAML files that describe the complete structure of a GitHub Project v2.

#### Template File: `project-template.yml`

```yaml
# Template metadata
template:
  name: "Scrum Board"
  description: "A Scrum-style project board with sprints and story points"
  version: "1.0.0"
  author: "your-username"
  tags:
    - scrum
    - agile
    - sprint

# Project settings
project:
  title: "{{.ProjectName}}"           # Supports Go template variables
  description: "Project created from Scrum Board template"
  visibility: private                  # private | public
  readme: |
    ## Welcome to {{.ProjectName}}
    
    This project uses a Scrum workflow with the following statuses:
    - **Backlog**: Items waiting to be prioritized
    - **Sprint Backlog**: Items committed to current sprint
    - **In Progress**: Work actively being done
    - **In Review**: Awaiting code review or QA
    - **Done**: Completed items

# Custom fields to create
fields:
  - name: "Status"
    type: single_select
    options:
      - name: "Backlog"
        color: GRAY
        description: "Items waiting to be prioritized"
      - name: "Sprint Backlog"
        color: BLUE
        description: "Committed to current sprint"
      - name: "In Progress"
        color: YELLOW
        description: "Work in progress"
      - name: "In Review"
        color: PURPLE
        description: "Awaiting review"
      - name: "Done"
        color: GREEN
        description: "Completed"

  - name: "Priority"
    type: single_select
    options:
      - name: "P0 - Critical"
        color: RED
      - name: "P1 - High"
        color: ORANGE
      - name: "P2 - Medium"
        color: YELLOW
      - name: "P3 - Low"
        color: GRAY

  - name: "Story Points"
    type: number
    
  - name: "Sprint"
    type: iteration
    duration: 14                       # days
    start_day: monday
    
  - name: "Due Date"
    type: date

  - name: "Team"
    type: single_select
    options:
      - name: "Frontend"
        color: BLUE
      - name: "Backend"
        color: GREEN
      - name: "DevOps"
        color: PURPLE
      - name: "QA"
        color: ORANGE

  - name: "Epic"
    type: text

# Views to create
views:
  - name: "Board"
    type: board
    group_by: "Status"
    sort_by: "Priority"
    filter: "is:open"
    columns:
      - "Status"
      - "Priority"
      - "Assignees"
      - "Sprint"

  - name: "Sprint Backlog"
    type: table
    filter: "status:\"Sprint Backlog\",\"In Progress\",\"In Review\""
    columns:
      - "Title"
      - "Status"
      - "Priority"
      - "Story Points"
      - "Assignees"
    sort_by: "Priority"

  - name: "Roadmap"
    type: roadmap
    date_field: "Sprint"
    group_by: "Team"

  - name: "My Items"
    type: table
    filter: "assignee:@me is:open"
    columns:
      - "Title"
      - "Status"
      - "Priority"
      - "Due Date"

# Workflows (auto-archive, auto-add, etc.)
workflows:
  - name: "Auto-archive done items"
    enabled: true
    trigger:
      type: item_closed
      conditions:
        status: "Done"
    action:
      type: archive
      after_days: 14

  - name: "Auto-add from repository"
    enabled: true
    trigger:
      type: issue_opened
      repositories:
        - "{{.Repository}}"
    action:
      type: add_to_project
      set_fields:
        status: "Backlog"

# Draft issues to create (optional)
draft_issues:
  - title: "📋 Sprint Planning Template"
    body: |
      ## Sprint Goal
      _Define the sprint goal here_
      
      ## Committed Items
      - [ ] Item 1
      - [ ] Item 2
      
      ## Capacity
      | Team Member | Available Days | Notes |
      |-------------|---------------|-------|
      | @member1    | 10            |       |
    labels:
      - "template"
      - "sprint-planning"

  - title: "📊 Retrospective Template"
    body: |
      ## What went well?
      - 
      
      ## What could be improved?
      - 
      
      ## Action items
      - [ ] Action 1
    labels:
      - "template"
      - "retrospective"

# Field aliases for gh-pm configuration
aliases:
  status:
    backlog: "Backlog"
    sprint: "Sprint Backlog"
    in_progress: "In Progress"
    review: "In Review"
    done: "Done"
  
  priority:
    critical: "P0 - Critical"
    high: "P1 - High"
    medium: "P2 - Medium"
    low: "P3 - Low"

# Linked repositories (optional)
repositories:
  - "{{.Repository}}"
```

### Template Variables

Templates support Go template syntax for dynamic values:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.ProjectName}}` | Name provided by user | "Q1 2025 Sprint" |
| `{{.Owner}}` | Target owner (user/org) | "my-org" |
| `{{.Repository}}` | Current repository | "owner/repo" |
| `{{.User}}` | Current authenticated user | "johndoe" |
| `{{.Date}}` | Current date | "2025-11-30" |
| `{{.Year}}` | Current year | "2025" |
| `{{.Quarter}}` | Current quarter | "Q4" |

### Built-in Templates

The extension will include several built-in templates:

```
templates/
├── kanban.yml              # Simple Kanban board
├── scrum.yml               # Scrum with sprints
├── bug-tracker.yml         # Bug tracking workflow
├── feature-roadmap.yml     # Product roadmap
├── release-planning.yml    # Release management
└── team-standup.yml        # Daily standup tracking
```

### Template Commands in Detail

#### Create Project from GitHub Template

```bash
# List available template projects
gh pmuproject list --templates --owner my-org

# Create from existing GitHub project
gh pmuproject create \
  --from-project my-org/5 \
  --title "Q1 Sprint Board" \
  --owner @me \
  --include-drafts

# Interactive mode
gh pmuproject create --from-project
# > Select source owner: my-org
# > Select template project: [5] Scrum Template
# > Enter new project title: Q1 Sprint Board
# > Include draft issues? (y/N): y
# ✓ Created project: https://github.com/users/me/projects/12
```

#### Create Project from YAML Template

```bash
# Create from local template file
gh pmuproject create \
  --from-template ./templates/scrum.yml \
  --title "Q1 Sprint Board" \
  --owner my-org \
  --var "Repository=my-org/main-repo"

# Create from URL
gh pmuproject create \
  --from-template https://example.com/templates/kanban.yml \
  --title "Feature Backlog"

# Create from built-in template
gh pmuproject create \
  --from-template builtin:scrum \
  --title "Team Sprint Board"

# Dry-run to preview what would be created
gh pmuproject create \
  --from-template ./scrum.yml \
  --dry-run

# Output:
# Would create project "Q1 Sprint Board" with:
#   Fields: Status (5 options), Priority (4 options), Story Points, Sprint, Due Date, Team (4 options), Epic
#   Views: Board, Sprint Backlog, Roadmap, My Items
#   Workflows: Auto-archive done items, Auto-add from repository
#   Draft Issues: 2
```

#### Export Project to Template

```bash
# Export current project to YAML template
gh pmuproject export 1 --output ./my-template.yml

# Export with specific options
gh pmuproject export 1 \
  --owner my-org \
  --include-drafts \
  --include-workflows \
  --output ./full-template.yml

# Export to stdout (for piping)
gh pmuproject export 1 --format yaml

# Export minimal template (fields and views only)
gh pmuproject export 1 --minimal --output ./minimal.yml
```

#### Template Management

```bash
# Validate template syntax
gh pmutemplate validate ./my-template.yml
# ✓ Template is valid
# Fields: 6 defined
# Views: 4 defined
# Workflows: 2 defined

# List built-in templates
gh pmutemplate list --builtin
# NAME              DESCRIPTION
# kanban            Simple Kanban board with To Do, In Progress, Done
# scrum             Scrum board with sprints and story points
# bug-tracker       Bug tracking with severity and resolution
# feature-roadmap   Product roadmap with quarters and themes

# Show template details
gh pmutemplate show builtin:scrum
# Name: Scrum Board
# Description: A Scrum-style project board with sprints and story points
# Fields:
#   - Status (single_select): 5 options
#   - Priority (single_select): 4 options
#   - Story Points (number)
#   - Sprint (iteration): 14 days
# Views: Board, Sprint Backlog, Roadmap, My Items
# Workflows: 2 defined
```

---

## Feature Matrix

### Phase 1: Core Unification

| Feature | Source | Command | Status |
|---------|--------|---------|--------|
| Initialize config | gh-pm | `gh pmuinit` | 🔲 |
| List issues | gh-pm | `gh pmulist` | 🔲 |
| View issue | gh-pm | `gh pmuview` | 🔲 |
| Create issue | gh-pm | `gh pmucreate` | 🔲 |
| Move/update issue | gh-pm | `gh pmumove` | 🔲 |
| Issue intake | gh-pm | `gh pmuintake` | 🔲 |
| Triage issues | gh-pm | `gh pmutriage` | 🔲 |
| Add sub-issue | gh-sub-issue | `gh pmusub add` | 🔲 |
| Create sub-issue | gh-sub-issue | `gh pmusub create` | 🔲 |
| List sub-issues | gh-sub-issue | `gh pmusub list` | 🔲 |
| Remove sub-issue | gh-sub-issue | `gh pmusub remove` | 🔲 |
| Split issue | gh-pm | `gh pmusplit` | 🔲 |

### Phase 2: Project Templates & Creation

| Feature | Description | Command | Status |
|---------|-------------|---------|--------|
| Create from GitHub project | Copy existing project | `gh pmuproject create --from-project` | 🔲 |
| Create from YAML template | Create project from YAML definition | `gh pmuproject create --from-template` | 🔲 |
| Export to YAML template | Export project structure to YAML | `gh pmuproject export` | 🔲 |
| Template validation | Validate template YAML syntax | `gh pmutemplate validate` | 🔲 |
| List templates | List built-in and local templates | `gh pmutemplate list` | 🔲 |
| Built-in templates | Ship with common project templates | N/A | 🔲 |
| Template variables | Support Go template syntax | N/A | 🔲 |
| Init from template | Initialize with template selection | `gh pmuinit --from-template` | 🔲 |

### Phase 3: Enhanced Integration

| Feature | Description | Status |
|---------|-------------|--------|
| Unified sub handling | Remove gh-sub-issue dependency from split | 🔲 |
| Cross-repo sub-issues | Full cross-repository support | 🔲 |
| Progress tracking | Show sub-issue completion percentages | 🔲 |
| Recursive operations | Operations on full issue trees | 🔲 |
| Template registry | Remote template repository/registry | 🔲 |
| Template inheritance | Templates that extend other templates | 🔲 |

### Phase 4: Future Capabilities

Reserved for additional capabilities not yet specified:

| Feature | Description | Status |
|---------|-------------|--------|
| TBD | Future enhancement 1 | 🔲 |
| TBD | Future enhancement 2 | 🔲 |
| TBD | Future enhancement 3 | 🔲 |

---

## Technical Architecture

### Language & Dependencies

- **Language:** Go (consistent with both source projects and gh CLI)
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **GitHub API:** [go-gh](https://github.com/cli/go-gh) (official gh extension library)
- **Configuration:** [Viper](https://github.com/spf13/viper) for YAML parsing
- **Templating:** Go `text/template` for variable substitution

### Package Structure

```
gh-pm-unified/
├── cmd/
│   ├── root.go           # Root command
│   ├── init.go           # Configuration initialization
│   ├── list.go           # List issues
│   ├── view.go           # View issue
│   ├── create.go         # Create issue
│   ├── move.go           # Update issue fields
│   ├── intake.go         # Issue intake
│   ├── triage.go         # Bulk triage
│   ├── split.go          # Issue splitting
│   ├── sub/
│   │   ├── add.go        # Add sub-issue
│   │   ├── create.go     # Create sub-issue
│   │   ├── list.go       # List sub-issues
│   │   └── remove.go     # Remove sub-issue
│   ├── project/
│   │   ├── create.go     # Create project (from template)
│   │   ├── export.go     # Export project to template
│   │   └── list.go       # List projects
│   ├── template/
│   │   ├── list.go       # List templates
│   │   ├── show.go       # Show template details
│   │   └── validate.go   # Validate template
│   └── version.go        # Version info
├── pkg/
│   ├── config/           # Configuration management
│   ├── project/          # Project API operations
│   ├── issue/            # Issue API operations
│   ├── subissue/         # Sub-issue API operations
│   ├── template/         # Template parsing and rendering
│   │   ├── schema.go     # YAML schema definitions
│   │   ├── parser.go     # Template parser
│   │   ├── renderer.go   # Variable substitution
│   │   ├── validator.go  # Schema validation
│   │   └── builtin/      # Built-in templates (embedded)
│   ├── filter/           # Query parsing and filtering
│   ├── output/           # Output formatting
│   └── utils/            # Shared utilities
├── templates/            # Built-in template files
│   ├── kanban.yml
│   ├── scrum.yml
│   ├── bug-tracker.yml
│   ├── feature-roadmap.yml
│   └── release-planning.yml
├── .gh-pmu.yml            # Example configuration
├── go.mod
├── go.sum
├── main.go
├── Makefile
├── LICENSE
└── README.md
```

### API Integration

The extension will use GitHub's GraphQL API for most operations, with special handling for sub-issues and project creation:

```go
// Sub-issue mutations require special headers
func (c *Client) AddSubIssue(parentID, childID string) error {
    mutation := `
        mutation($parentId: ID!, $childId: ID!) {
            addSubIssue(input: {
                issueId: $parentId
                subIssueId: $childId
            }) {
                issue { title }
                subIssue { title }
            }
        }`
    
    // Requires: GraphQL-Features: sub_issues header
    return c.graphQLWithFeatures(mutation, variables, 
        "sub_issues", "issue_types")
}

// Create project from template
func (c *Client) CreateProjectFromTemplate(tmpl *Template, owner, title string) (*Project, error) {
    // 1. Create base project
    project, err := c.CreateProject(owner, title)
    if err != nil {
        return nil, err
    }
    
    // 2. Create custom fields
    for _, field := range tmpl.Fields {
        if err := c.CreateProjectField(project.ID, field); err != nil {
            return nil, fmt.Errorf("failed to create field %s: %w", field.Name, err)
        }
    }
    
    // 3. Create views
    for _, view := range tmpl.Views {
        if err := c.CreateProjectView(project.ID, view); err != nil {
            return nil, fmt.Errorf("failed to create view %s: %w", view.Name, err)
        }
    }
    
    // 4. Create draft issues
    for _, draft := range tmpl.DraftIssues {
        if err := c.CreateDraftIssue(project.ID, draft); err != nil {
            return nil, fmt.Errorf("failed to create draft issue: %w", err)
        }
    }
    
    return project, nil
}
```

### Template Rendering

```go
// Render template with variables
func RenderTemplate(tmpl *Template, vars map[string]string) (*Template, error) {
    // Default variables
    if vars == nil {
        vars = make(map[string]string)
    }
    
    // Add automatic variables
    vars["Date"] = time.Now().Format("2006-01-02")
    vars["Year"] = time.Now().Format("2006")
    vars["Quarter"] = fmt.Sprintf("Q%d", (time.Now().Month()-1)/3+1)
    
    // Get current user
    if _, ok := vars["User"]; !ok {
        user, _ := getCurrentUser()
        vars["User"] = user
    }
    
    // Render all string fields using Go templates
    rendered, err := renderFields(tmpl, vars)
    if err != nil {
        return nil, fmt.Errorf("template rendering failed: %w", err)
    }
    
    return rendered, nil
}
```

---

## Migration Path

### For gh-pm Users

```bash
# Remove old extension
gh extension remove pm

# Install unified extension
gh extension install <org>/gh-pm-unified

# Existing .gh-pmu.yml files are compatible
gh pmulist  # Works immediately
```

### For gh-sub-issue Users

```bash
# Remove old extension
gh extension remove sub-issue

# Install unified extension
gh extension install <org>/gh-pm-unified

# Commands move under 'sub' subcommand
gh pmusub add 123 456        # Was: gh sub-issue add 123 456
gh pmusub create --parent 123 --title "Task"
gh pmusub list 123
gh pmusub remove 123 456
```

### Backward Compatibility

Consider providing aliases for transition period:

```bash
# Aliases (deprecated, will show warning)
gh sub-issue add → gh pmusub add
gh sub-issue create → gh pmusub create
gh sub-issue list → gh pmusub list
gh sub-issue remove → gh pmusub remove
```

---

## License

The unified project will be released under the **MIT License**, consistent with both source projects.

```
MIT License

Copyright (c) 2025 [Project Maintainers]

Portions of this software are derived from:
- gh-pm (c) yahsan2 - https://github.com/yahsan2/gh-pm
- gh-sub-issue (c) yahsan2 - https://github.com/yahsan2/gh-sub-issue

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## Next Steps

1. **Contact original author** - Reach out to @yahsan2 regarding collaboration or blessing
2. **Fork repositories** - Create forks of both projects as starting point
3. **Merge codebases** - Integrate gh-sub-issue functionality into gh-pm structure
4. **Refactor commands** - Implement unified command structure
5. **Design template schema** - Finalize YAML template specification
6. **Implement template engine** - Build template parsing and rendering
7. **Create built-in templates** - Develop standard project templates
8. **Update documentation** - Comprehensive README and usage guides
9. **Testing** - Unit tests and integration tests
10. **Release** - Publish to GitHub and gh extension registry

---

## Open Questions

1. **Naming:** What should the final extension be called?
2. **Collaboration:** Should this be a fork or new project with original author involvement?
3. **Scope:** What additional capabilities should be considered for Phase 4?
4. **Compatibility:** How long should backward compatibility aliases be maintained?
5. **Template Registry:** Should there be a central registry for sharing templates?
6. **Template Versioning:** How should template schema versions be managed?
7. **Workflow Support:** Can GitHub Project workflows be fully created via API?

---

## References

- [GitHub CLI Manual](https://cli.github.com/manual/)
- [gh project commands](https://cli.github.com/manual/gh_project)
- [gh project copy](https://cli.github.com/manual/gh_project_copy)
- [gh project mark-template](https://cli.github.com/manual/gh_project_mark-template)
- [GitHub Projects documentation](https://docs.github.com/en/issues/planning-and-tracking-with-projects)
- [Sub-issues documentation](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/adding-sub-issues)
- [GitHub CLI Extension Guide](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions)
- [go-gh library](https://github.com/cli/go-gh)
- [GitHub GraphQL API - Projects](https://docs.github.com/en/graphql/reference/objects#projectv2)

---

## Appendix A: Complete YAML Template Schema

```yaml
# Schema version for forward compatibility
schema_version: "1.0"

# Template metadata
template:
  name: string                    # Required: Template display name
  description: string             # Optional: Template description
  version: string                 # Optional: Semantic version
  author: string                  # Optional: Author username/name
  license: string                 # Optional: License identifier
  tags: [string]                  # Optional: Searchable tags
  source_url: string              # Optional: Source repository URL

# Project configuration
project:
  title: string                   # Required: Project title (supports variables)
  description: string             # Optional: Project description
  visibility: enum                # Optional: private | public (default: private)
  readme: string                  # Optional: Project README content

# Field definitions
fields:
  - name: string                  # Required: Field name
    type: enum                    # Required: single_select | text | number | date | iteration
    description: string           # Optional: Field description
    options:                      # Required for single_select, optional for iteration
      - name: string              # Required: Option name
        color: enum               # Optional: GRAY|RED|ORANGE|YELLOW|GREEN|BLUE|PURPLE|PINK
        description: string       # Optional: Option description
    duration: integer             # For iteration: days per iteration
    start_day: enum               # For iteration: monday|sunday

# View definitions
views:
  - name: string                  # Required: View name
    type: enum                    # Required: board | table | roadmap
    default: boolean              # Optional: Is this the default view?
    group_by: string              # Optional: Field name to group by
    sort_by: string               # Optional: Field name to sort by
    sort_direction: enum          # Optional: asc | desc
    filter: string                # Optional: Filter query
    columns: [string]             # Optional: Field names to display as columns

# Workflow definitions
workflows:
  - name: string                  # Required: Workflow name
    enabled: boolean              # Optional: Is workflow enabled? (default: true)
    trigger:
      type: enum                  # Required: item_added | item_closed | item_reopened
      conditions:                 # Optional: Additional conditions
        status: string
        label: string
    action:
      type: enum                  # Required: set_field | archive | add_to_project
      field: string               # For set_field: target field
      value: string               # For set_field: target value
      after_days: integer         # For archive: days before archiving

# Draft issues to pre-populate
draft_issues:
  - title: string                 # Required: Issue title
    body: string                  # Optional: Issue body (markdown)
    labels: [string]              # Optional: Labels to apply
    assignees: [string]           # Optional: Usernames to assign
    fields:                       # Optional: Project field values
      field_name: value

# Field aliases for gh-pm configuration
aliases:
  field_name:                     # Maps to fields[].name
    alias: actual_value           # alias → options[].name

# Repositories to link
repositories:
  - string                        # Repository in owner/repo format (supports variables)
```

---

*This document is a living proposal and will be updated as the project evolves.*

*Last Updated: November 30, 2025 - Added Project Templates specification (Phase 2)*
