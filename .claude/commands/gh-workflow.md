# GitHub Workflow Integration

**Framework-only command for maintainers**

This command configures Claude to automatically manage GitHub issues during development sessions.

---

## Project Configuration

**GitHub Project Board:** https://github.com/users/scooter-indie/projects/11
**Repository:** scooter-indie/gh_ext

---

## Prerequisites

The following GitHub CLI extensions must be installed:

```bash
gh extension install yahsan2/gh-pm        # Project status management
gh extension install yahsan2/gh-sub-issue # Sub-issue linking
```

**Configuration:** `.gh-pm.yml` in repo root configures project board status values.

**Status values:** `backlog`, `ready`, `in_progress`, `in_review`, `done`

---

## Workflow Instructions

When this command is executed, Claude should follow these workflows for the remainder of the session.
See the full gh-workflow.md template in the IDPF Framework for complete workflow details.

---

*Note: Replace [GITHUB_PROJECT_BOARD_URL] and [GITHUB_REPOSITORY] with your project values.*
*The startup procedure will prompt for these values if not configured.*
