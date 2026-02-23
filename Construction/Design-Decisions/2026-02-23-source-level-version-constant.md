# Design Decision: Source-Level Version Constant

**Date:** 2026-02-23
**Status:** Accepted
**Context:** Issue #700 — [Enhancement]: Add gh-pmu version number to .gh-pmu.yml to detect if init needs to be run

## Decision

Use a source-level version constant in `internal/version/version.go` as the primary version source, updated by `/prepare-release` at release time. The ldflags injection via goreleaser remains as a fallback override.

The acceptance gate distinguishes dev builds by checking the raw `version` variable (empty when no ldflags injected), not `getVersion()` which always returns a real version string.

The `/prepare-release` instruction for updating version.go is placed in the `pre-commit` USER-EXTENSION block rather than the core spec table, so it survives framework upgrades.

## Rationale

The initial approach relied solely on ldflags (`-ldflags "-X cmd.version=..."`) set at build time by goreleaser. This created a chicken-and-egg problem: during development `version` was `"dev"`, meaning `gh pmu init` would stamp `version: dev` into `.gh-pmu.yml` — useless for upgrade detection.

By placing the version constant in source code and having `/prepare-release` update it before tagging, the version is always a real semver string. The constant is committed to the repo, so even source-built binaries report the correct version.

## Alternatives Considered

- **ldflags only**: Rejected — `version: dev` would be stamped during development, making upgrade detection impossible without a production build.
- **Read version from goreleaser config or git tags at runtime**: Rejected — adds runtime complexity and external dependencies. A simple constant is more reliable.
- **Modify core prepare-release spec table**: Rejected — the spec is framework-managed and would be overwritten on framework upgrades. Extension points exist for this purpose.

## Consequences

- `/prepare-release` must update `internal/version/version.go` before tagging (enforced via `pre-commit` extension)
- Dev builds report the last-released version (acceptable — the acceptance gate bypasses entirely for dev builds via `version == ""`)
- Existing `.gh-pmu.yml` files without `version` field load cleanly (`omitempty` tag)

## Issues Encountered

- Changing `version` from `"dev"` to `""` broke the acceptance gate and many cmd tests. The gate checked `version == "dev"` to skip for dev builds; after the change, `getVersion()` returned `"0.16.0"` so the gate enforced acceptance in tests. Fixed by checking the raw ldflags variable (`version == ""`) instead of the effective version.
- `TestAcceptCommand_RecordsVersion` compared against raw `version` (now empty) but `accept.go` writes `getVersion()`. Fixed by updating the test to compare against `getVersion()`.
