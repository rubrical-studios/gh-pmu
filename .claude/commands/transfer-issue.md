---
version: "v0.33.3"
allowed-tools: Bash
description: Transfer issue between releases/sprints (project)
argument-hint: "<issue-number> [--to branch]"
---

<!-- MANAGED -->
Transfer an issue between releases or sprints.

Run the transfer-issue script:

```bash
node .claude/scripts/shared/transfer-issue.js "$ARGUMENTS"
```

Report the result to the user.
