# GitHub Workflow Integration

This command configures Claude to automatically manage GitHub issues during development sessions.

---

## Project Configuration

**Read from `.gh-pmu.yml`** in the repository root:

```yaml
project:
    owner: {owner}      # GitHub username or org
    number: {number}    # Project board number
repositories:
    - {owner}/{repo}    # Repository in owner/repo format
```

If `.gh-pmu.yml` doesn't exist, run `gh pmu init` to create it.

---

## Workflow Instructions

When this command is executed, Claude should follow these workflows for the remainder of the session.

### Bug Workflow
- Trigger: "I found an issue...", "There's a bug...", "finding:"
- Create issue with "bug" label automatically
- Wait for "work issue #X" to start
- Move to in_progress, then in_review after commit
- Close when user says "Done"

### Enhancement Workflow
- Trigger: "I would like...", "Can you add...", "New feature..."
- Create issue with "enhancement" label automatically
- Same flow as Bug workflow

### Sub-Issue Workflow
- Trigger: "Create sub-issues for...", "Break this into phases..."
- Create sub-issues and link to parent via `gh sub-issue add`
- Ask about epic label for parent

---

**Note:** Replace {owner}, {repo}, {number} placeholders after running `gh pmu init`.
