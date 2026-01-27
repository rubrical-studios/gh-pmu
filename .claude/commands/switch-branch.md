---
version: "v0.33.2"
allowed-tools: Bash
description: Switch branch/sprint context (project)
argument-hint: "[branch-name]"
---

<!-- MANAGED -->
Switch between branch and sprint contexts.

Run the switch-branch script:

```bash
node .claude/scripts/shared/switch-branch.js "$ARGUMENTS"
```

After switching, report the new context to the user.
