# gh-pmu

A GitHub CLI extension for project management and sub-issue hierarchy.

## Features

📋 **Project Management** - List, view, create, and update issues with project field values in one command

🔗 **Sub-Issue Hierarchy** - Create and manage parent-child issue relationships with progress tracking

⚡ **Batch Operations** - Intake untracked issues, triage with rules, split checklists into sub-issues

📊 **Board View** - Terminal Kanban board visualization

🔄 **Cross-Repository** - Work with sub-issues across multiple repositories

## Installation

```bash
gh extension install rubrical-studios/gh-pmu
```

## Upgrade

```bash
gh extension upgrade gh-pmu
```

## Quick Start

```bash
# Initialize configuration
gh pmu init

# List issues with project metadata
gh pmu list

# View issue with project fields and sub-issue progress
gh pmu view 42

# Update status
gh pmu move 42 --status in_progress

# Create sub-issue
gh pmu sub create --parent 42 --title "Subtask"
```

## Documentation

| Guide | Description |
|-------|-------------|
| [Configuration](docs/configuration.md) | Setup `.gh-pmu.yml`, field aliases, triage rules |
| [Commands](docs/commands.md) | Complete command reference with examples |
| [Sub-Issues](docs/sub-issues.md) | Parent-child hierarchies, epics, progress tracking |
| [Batch Operations](docs/batch-operations.md) | Intake, triage, and split workflows |
| [gh vs gh pmu](docs/gh-comparison.md) | When to use each CLI |
| [Development](docs/development.md) | Building, testing, contributing |

## Commands

```
Project:    init, list, view, create, move, close, board, field
Sub-Issues: sub add, sub create, sub list, sub remove
Batch:      intake, triage, split
Utilities:  filter, history
```

Run `gh pmu --help` for full command list.

## Unique Capabilities

Flags and features not available in base `gh` CLI:

| Command | Unique Flags | Purpose |
|---------|--------------|---------|
| `list` | `--status`, `--priority`, `--has-sub-issues` | Filter by project fields |
| `create` | `--status`, `--priority`, `--from-file` | Set project fields on create |
| `close` | `--update-status` | Move to 'done' before closing |
| `move` | `--recursive`, `--dry-run`, `--depth` | Cascade updates to sub-issues |
| `sub create` | `--inherit-labels`, `--inherit-milestone` | Inherit from parent issue |
| `split` | `--from`, `--dry-run` | Create sub-issues from checklist |

See [gh vs gh pmu](docs/gh-comparison.md) for detailed comparison.

## Attribution

This project builds upon work from [@yahsan2](https://github.com/yahsan2):
- [gh-pm](https://github.com/yahsan2/gh-pm)
- [gh-sub-issue](https://github.com/yahsan2/gh-sub-issue)

## License

MIT
