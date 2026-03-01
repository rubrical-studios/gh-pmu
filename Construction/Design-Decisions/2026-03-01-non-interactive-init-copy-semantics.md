# Design Decision: Non-Interactive Init Copy Semantics

**Date:** 2026-03-01
**Issue:** #714

## Context

The `gh pmu init --non-interactive` command previously used `--project` to specify an existing project number to configure directly. This led to confusion about the flag's purpose — users expected it to behave as a source template for project creation.

## Decision

Rename `--project` to `--source-project` and change non-interactive mode to create a new project by copying from the source project, rather than configuring the source project directly.

## Alternatives Considered

1. **Keep `--project` and add `--source-project` separately** — Would create ambiguity about which flag to use in non-interactive mode. Rejected for API clarity.
2. **Add a `--copy` boolean flag** — Would require both `--project` and `--copy` in non-interactive mode. Rejected as unnecessarily verbose.

## Consequences

- **Breaking change:** Users of `--project` in CI/CD scripts must update to `--source-project`.
- **Interactive mode:** Flag rename applies there too (pre-selects a project), but behavior is unchanged.
- Non-interactive mode now always creates a new project, which is the expected CI/CD use case.
