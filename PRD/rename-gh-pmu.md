# PRD: Rename CLI Command from `gh pm` to `gh pmu`

**Author:** Developer
**Date:** 2025-12-02
**Status:** Completed
**Priority:** High

---

## Problem Statement

The current command name `gh pm` conflicts with the existing `gh-pm` extension by yahsan2. During development, we need a distinct command name to avoid confusion and allow both tools to coexist. The name `gh pmu` (Project Management Unified) clearly differentiates this tool.

---

## Goals

1. Rename all CLI invocations from `gh pm` to `gh pmu`
2. Update all documentation, help text, and examples
3. Maintain backward compatibility during transition (optional)
4. Ensure all tests pass with new command name

---

## Non-Goals

- Changing any functionality
- Modifying the API or internal architecture
- Adding new features

---

## Scope of Changes

### 1. Binary/Extension Name

| Location | Current | New |
|----------|---------|-----|
| `main.go` | `gh-pm` | `gh-pmu` |
| `go.mod` | `github.com/rubrical-studios/gh-pm` | `github.com/rubrical-studios/gh-pmu` |
| Binary output | `gh-pm` | `gh-pmu` |

### 2. Command Help Text

Files to update:
- `cmd/root.go` - Root command Use and descriptions
- `cmd/init.go` - Examples
- `cmd/list.go` - Examples
- `cmd/view.go` - Examples
- `cmd/create.go` - Examples
- `cmd/move.go` - Examples
- `cmd/sub.go` - All subcommand examples

### 3. Configuration

| Location | Current | New |
|----------|---------|-----|
| Config filename | `.gh-pm.yml` | `.gh-pmu.yml` |
| `internal/config/config.go` | `ConfigFileName = ".gh-pm.yml"` | `ConfigFileName = ".gh-pmu.yml"` |

### 4. Documentation

Files to update:
- `README.md`
- `docs/testing.md`
- `backlog/*.md` - All sprint docs and backlog references
- `CLAUDE.md` - Framework instructions
- `.claude/commands/gh-workflow.md`

### 5. GitHub Workflow

- Update `.claude/commands/gh-workflow.md` references
- Update any CI/CD workflows referencing the command

### 6. Tests

- Update all test files that reference command names in assertions
- Verify help text tests pass with new name

---

## Acceptance Criteria

- [x] `gh pmu --help` displays correct command name and examples
- [x] `gh pmu init` creates `.gh-pmu.yml` configuration file
- [x] All subcommands work: `gh pmu list`, `gh pmu view`, `gh pmu create`, `gh pmu move`, `gh pmu sub`
- [x] All tests pass with new command name
- [x] Documentation reflects new command name throughout
- [x] `go build` produces `gh-pmu` binary
- [ ] Extension installs as `gh extension install` and runs as `gh pmu`

---

## Migration Considerations

### Option A: Hard Rename (Recommended)
- Simple find/replace across codebase
- Clean break from old name
- No backward compatibility overhead

### Option B: Alias Support
- Keep `gh pm` as alias to `gh pmu`
- Add deprecation warning
- Remove alias in future version
- More complex, not recommended for development phase

**Recommendation:** Option A - Hard rename since we're still in development.

---

## File Inventory

### Must Update (Code)
```
main.go
go.mod
cmd/root.go
cmd/init.go
cmd/list.go
cmd/view.go
cmd/create.go
cmd/move.go
cmd/sub.go
internal/config/config.go
```

### Tests (No Changes Needed)

Tests use Cobra's command structure directly (`NewRootCommand()`, `cmd.SetArgs()`) rather than string-based command invocations. No test files require updates.

### Must Update (Docs)
```
README.md
CLAUDE.md
STARTUP.md
docs/testing.md
.claude/commands/gh-workflow.md
backlog/product-backlog.md
backlog/sprint-1-backlog.md
backlog/sprint-2-backlog.md
backlog/sprint-1-summary.md
backlog/sprint-2-summary.md
backlog/sprint-2-retro.md
PRD/README.md
```

### Must Rename (Files)
```
.gh-pm.yml -> .gh-pmu.yml (if exists in repo)
```

---

## Estimated Effort

**Story Points:** 3

This is primarily a find/replace operation with verification:
1. Update go.mod module path
2. Find/replace `gh pm` → `gh pmu` in all files
3. Find/replace `gh-pm` → `gh-pmu` in all files
4. Find/replace `.gh-pm.yml` → `.gh-pmu.yml`
5. Run tests to verify
6. Build and manually verify help text

---

## Risks

1. **Missed references** - Grep thoroughly before considering complete
2. **Import path changes** - go.mod change affects all internal imports
3. **Existing config files** - Users with `.gh-pm.yml` need to rename

---

## Success Metrics

- All tests pass
- `gh pmu --version` works
- No references to `gh pm` or `gh-pm` remain in codebase (except LICENSE attribution)
