# gh-pmu

A unified GitHub CLI extension for project management, sub-issue hierarchy, and project templating.

## Features

- **Project Management**: List, view, create, and update issues with project field values
- **Sub-Issue Hierarchy**: Create and manage parent-child issue relationships
- **Project Templates**: Create projects from YAML templates or existing projects

## Installation

### From GitHub Releases

```bash
gh extension install scooter-indie/gh-pm-unified
```

### From Source

```bash
git clone https://github.com/scooter-indie/gh-pm-unified.git
cd gh-pm-unified
make install-extension
```

## Quick Start

1. Initialize configuration in your repository:

```bash
gh pmu init
```

2. List issues with project metadata:

```bash
gh pmu list
```

3. View an issue with all project fields:

```bash
gh pmu view 123
```

## Usage

```bash
gh pmu [command]

Available Commands:
  init        Initialize gh-pmu configuration
  list        List issues with project metadata
  view        View issue with project fields
  create      Create issue with project fields
  move        Update issue project fields
  sub         Manage sub-issues
  project     Manage project templates

Flags:
  -h, --help      help for gh-pmu
  -v, --version   version for gh-pmu
```

## Configuration

gh-pmu uses a `.gh-pmu.yml` file in your repository root:

```yaml
project:
  owner: your-username
  number: 1
repositories:
  - your-username/your-repo
defaults:
  priority: p2
  status: backlog
```

## Development

### Prerequisites

- Go 1.21+
- GitHub CLI (`gh`)

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## Attribution

This project builds upon and unifies functionality from the following open-source projects by [@yahsan2](https://github.com/yahsan2):

- **[gh-pm](https://github.com/yahsan2/gh-pm)** - GitHub project management CLI extension
- **[gh-sub-issue](https://github.com/yahsan2/gh-sub-issue)** - Sub-issue hierarchy management

Thank you to the original author for the foundational work.

## License

MIT
