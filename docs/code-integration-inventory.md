# Code Integration Inventory

**Date:** 2025-12-02
**Purpose:** Document source code to be integrated from gh-pm and gh-sub-issue into gh-pm-unified

---

## Source Repository: gh-pm

**URL:** https://github.com/yahsan2/gh-pm
**License:** MIT
**Author:** @yahsan2

### File Structure (40 Go files)

```
gh-pm/
├── main.go                          # Entry point
├── cmd/
│   ├── root.go                      # Root command setup
│   ├── init.go                      # gh pm init
│   ├── init_test.go
│   ├── list.go                      # gh pm list
│   ├── list_test.go
│   ├── view.go                      # gh pm view
│   ├── create.go                    # gh pm create
│   ├── move.go                      # gh pm move
│   ├── intake.go                    # gh pm intake
│   ├── intake_test.go
│   ├── triage.go                    # gh pm triage
│   ├── triage_test.go
│   ├── split.go                     # gh pm split
│   ├── split_test.go
│   └── version.go                   # gh pm version
├── pkg/
│   ├── args/
│   │   ├── parser.go                # Argument parsing utilities
│   │   └── parser_test.go
│   ├── config/
│   │   ├── config.go                # .gh-pm.yml loading/parsing
│   │   └── config_test.go
│   ├── filter/
│   │   ├── options.go               # Query filter options
│   │   └── options_test.go
│   ├── init/
│   │   ├── detector.go              # Git repo detection
│   │   ├── detector_test.go
│   │   ├── errors.go                # Init-specific errors
│   │   ├── metadata.go              # Project metadata fetching
│   │   ├── metadata_test.go
│   │   └── prompt.go                # Interactive prompts
│   ├── issue/
│   │   ├── client.go                # GitHub API client for issues
│   │   ├── client_test.go
│   │   ├── creator.go               # Issue creation logic
│   │   ├── creator_test.go
│   │   ├── models.go                # Issue data models
│   │   ├── search.go                # Issue search queries
│   │   ├── search_test.go
│   │   ├── batch.go                 # Batch operations
│   │   └── errors.go                # Issue-specific errors
│   ├── output/
│   │   └── formatter.go             # Table/JSON output formatting
│   ├── project/
│   │   ├── project.go               # Project API operations
│   │   ├── url.go                   # Project URL parsing
│   │   ├── url_test.go
│   │   └── userproject.go           # User project handling
│   └── utils/
│       ├── dateconv.go              # Date conversion utilities
│       └── dateconv_test.go
└── test/fixtures/
    └── split_command_sample_tasks.md
```

### Key Components to Integrate

| Package | Purpose | Priority | Notes |
|---------|---------|----------|-------|
| `pkg/config` | Load `.gh-pm.yml`, field aliases, defaults | High | Core infrastructure |
| `pkg/init` | Repo detection, metadata fetching, prompts | High | Used by `init` command |
| `pkg/issue` | Issue CRUD, search, batch operations | High | Core functionality |
| `pkg/project` | Project API, URL parsing | High | Core functionality |
| `pkg/filter` | Query filter parsing | Medium | Used by list/triage |
| `pkg/output` | Table/JSON formatting | Medium | UI consistency |
| `pkg/args` | Argument parsing helpers | Low | May use Cobra directly |
| `pkg/utils` | Date conversion | Low | Small utility |

---

## Source Repository: gh-sub-issue

**URL:** https://github.com/yahsan2/gh-sub-issue
**License:** MIT
**Author:** @yahsan2

### File Structure (10 Go files)

```
gh-sub-issue/
├── main.go                          # Entry point (111 bytes)
├── cmd/
│   ├── root.go                      # Root command setup (680 bytes)
│   ├── add.go                       # gh sub-issue add (8.6 KB)
│   ├── add_test.go                  # (5.8 KB)
│   ├── create.go                    # gh sub-issue create (14.8 KB)
│   ├── create_test.go               # (10.2 KB)
│   ├── list.go                      # gh sub-issue list (19.5 KB)
│   ├── list_test.go                 # (10.8 KB)
│   ├── remove.go                    # gh sub-issue remove (6.5 KB)
│   └── remove_test.go               # (7.0 KB)
```

### Key Components to Integrate

| File | Purpose | Priority | Notes |
|------|---------|----------|-------|
| `cmd/add.go` | Link existing issue as sub-issue | High | GraphQL mutation with feature headers |
| `cmd/create.go` | Create new sub-issue | High | Inherits labels/assignees from parent |
| `cmd/list.go` | List sub-issues of parent | High | Recursive listing, progress tracking |
| `cmd/remove.go` | Unlink sub-issue | Medium | Remove parent-child relationship |

### Critical: GraphQL Feature Headers

gh-sub-issue uses special GraphQL headers required for sub-issue API:

```go
// Required headers for sub-issue mutations
headers := map[string]string{
    "GraphQL-Features": "sub_issues,issue_types",
}
```

This must be preserved in the unified client.

---

## Unified Package Structure

```
gh-pm-unified/
├── main.go
├── cmd/
│   ├── root.go
│   ├── init.go
│   ├── list.go
│   ├── view.go
│   ├── create.go
│   ├── move.go
│   ├── intake.go
│   ├── triage.go
│   ├── split.go
│   ├── version.go
│   ├── sub/                         # Sub-command group
│   │   ├── add.go
│   │   ├── create.go
│   │   ├── list.go
│   │   └── remove.go
│   ├── project/                     # Project sub-commands (NEW)
│   │   ├── create.go
│   │   └── export.go
│   └── template/                    # Template sub-commands (NEW)
│       ├── list.go
│       ├── show.go
│       └── validate.go
├── pkg/
│   ├── config/                      # From gh-pm
│   │   ├── config.go
│   │   └── loader.go
│   ├── github/                      # Unified API client (MERGED)
│   │   ├── client.go                # Base client with auth
│   │   ├── issues.go                # From gh-pm/pkg/issue
│   │   ├── projects.go              # From gh-pm/pkg/project
│   │   ├── subissues.go             # From gh-sub-issue/cmd
│   │   └── graphql.go               # Shared GraphQL utilities
│   ├── template/                    # NEW for Phase 2
│   │   ├── schema.go
│   │   ├── parser.go
│   │   ├── renderer.go
│   │   └── builtin/
│   ├── filter/                      # From gh-pm
│   ├── output/                      # From gh-pm
│   └── utils/                       # From gh-pm
├── templates/                       # Built-in templates (NEW)
│   ├── kanban.yml
│   ├── scrum.yml
│   └── ...
├── .gh-pm.yml                       # Example config
├── go.mod
├── Makefile
├── .goreleaser.yaml
├── LICENSE                          # Include attribution
└── README.md                        # Credit original author
```

---

## Integration Steps

### Phase 1: Scaffolding
1. [ ] Initialize Go module: `github.com/rubrical-studios/gh-pm-unified`
2. [ ] Set up Cobra root command
3. [ ] Copy `.goreleaser.yaml` from gh-pm
4. [ ] Copy `Makefile` from gh-pm
5. [ ] Set up GitHub Actions from gh-pm

### Phase 2: Core Packages
1. [ ] Copy `pkg/config` from gh-pm
2. [ ] Create unified `pkg/github` client
3. [ ] Merge issue operations from gh-pm
4. [ ] Extract sub-issue mutations from gh-sub-issue cmd files
5. [ ] Copy `pkg/output` from gh-pm
6. [ ] Copy `pkg/filter` from gh-pm

### Phase 3: Commands
1. [ ] Implement commands using refactored packages
2. [ ] Add `sub` command group with sub-issue commands
3. [ ] Ensure all tests pass

### Phase 4: Attribution
1. [ ] Update LICENSE with attribution
2. [ ] Add credits to README
3. [ ] Add source references in code comments

---

## Dependencies (from go.mod analysis)

### gh-pm dependencies
- `github.com/cli/go-gh/v2` - Official GitHub CLI library
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration
- `gopkg.in/yaml.v3` - YAML parsing

### gh-sub-issue dependencies
- `github.com/cli/go-gh/v2` - Official GitHub CLI library
- `github.com/spf13/cobra` - CLI framework

### Unified dependencies (no conflicts)
- `github.com/cli/go-gh/v2`
- `github.com/spf13/cobra`
- `github.com/spf13/viper`
- `gopkg.in/yaml.v3`

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| API breaking changes | High | Pin go-gh version, test thoroughly |
| Sub-issue API changes | Medium | Monitor GitHub changelog |
| Different coding styles | Low | Apply consistent formatting |
| Missing test coverage | Medium | Expand tests during integration |

---

*Document created as part of Tech Story #60*
