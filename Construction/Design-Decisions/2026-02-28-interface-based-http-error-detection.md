# Design Decision: Interface-Based HTTP Error Detection

**Date:** 2026-02-28
**Issue:** #710
**Status:** Accepted

## Context

The `internal/api` package needs to detect HTTP 403/429 status codes and extract `Retry-After` headers from errors returned by go-gh's `api.HTTPError`. However, importing `github.com/cli/go-gh/v2/pkg/api` directly causes a package name collision — both the dependency and our own package are named `api`.

## Decision

Use Go interface-based duck typing instead of direct type assertion on go-gh's `HTTPError`:

- `httpStatusCoder` interface: `HTTPStatusCode() int`
- `retryAfterProvider` interface: `RetryAfterSeconds() string`

go-gh's `HTTPError` satisfies both interfaces without requiring an import. Detection uses `errors.As(err, &sc)` to extract the interface from wrapped errors.

## Alternatives Considered

1. **Import with alias** (`ghapi "github.com/cli/go-gh/v2/pkg/api"`) — works but adds a dependency coupling between packages and requires type assertion on a concrete foreign type.
2. **String parsing only** — unreliable; status codes and headers not available in error messages.
3. **HTTP middleware/transport** — intercepting at transport layer would catch all calls but requires modifying the HTTP client setup and doesn't cover `exec.Command` subprocess calls.

## Consequences

- Decoupled from go-gh's concrete types — survives dependency upgrades as long as the interface contract holds.
- Tests use local mock types satisfying the same interfaces, avoiding test dependency on go-gh.
- Pattern 2 (`exec.Command` subprocess calls) is not covered by this approach — those errors are exit codes + stderr, not `HTTPError`. Separate detection needed if subprocess retry is added later.
