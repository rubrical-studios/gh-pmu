# GitHub Workflow Integration

This command configures Claude to automatically manage GitHub issues during development sessions.

---

## Project Configuration

**Read from `.gh-pm.yml`** in the repository root:

```yaml
project:
    owner: {owner}      # GitHub username or org
    number: {number}    # Project board number
repositories:
    - {owner}/{repo}    # Repository in owner/repo format
```

**Derived values:**
- **Repository:** `repositories[0]`
- **Project Board:** `https://github.com/users/{project.owner}/projects/{project.number}/views/1`

If `.gh-pm.yml` doesn't exist, run `gh pm init` to create it.

---

## Prerequisites

The following GitHub CLI extensions must be installed:

```bash
gh extension install yahsan2/gh-pm        # Project status management
gh extension install yahsan2/gh-sub-issue # Sub-issue linking
```

**Status values:** `backlog`, `ready`, `in_progress`, `in_review`, `done`

---

## Workflow Instructions

When this command is executed, Claude should follow these workflows for the remainder of the session.

**Important:** All `--repo` flags should use the repository from `.gh-pm.yml` (`repositories[0]`).

See the full gh-workflow.md template in the IDPF Framework for complete workflow details.
