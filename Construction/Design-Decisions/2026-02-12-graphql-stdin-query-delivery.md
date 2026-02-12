# Design Decision: GraphQL Query Delivery via stdin

**Date:** 2026-02-12
**Status:** Accepted
**Context:** Issue #673 — [Bug]: branch close fails on Windows — command line too long for GraphQL queries

## Decision

Refactored all `exec.Command("gh", "api", "graphql", "-f", "query="+query)` call sites to use `--input -` with `cmd.Stdin`, passing the query as a JSON request body via stdin instead of a CLI argument.

Introduced a shared `buildGraphQLRequestBody(query string) (string, error)` helper in `queries.go` that wraps a GraphQL query string into the `{"query": "..."}` JSON format required by `gh api graphql --input -`.

## Rationale

Windows has a ~32KB command-line argument limit. When `gh pmu branch close` processes 30+ issues, the dynamically constructed GraphQL query (with ~800-1000 chars per alias) exceeds this limit, causing `fork/exec: The filename or extension is too long` errors.

stdin has no practical size limit, making it the correct approach for unbounded query payloads. The existing `executeBatchMutation()` function already used this pattern successfully.

## Alternatives Considered

- **Per-function `executeGraphQLQuery` method on Client:** Would centralize both request building and exec.Command execution. Rejected because it would make the method harder to test (exec.Command can't be mocked in the existing test suite) and would mix testable logic with untestable I/O.

- **GraphQL variables extraction:** Instead of embedding values in the query string, extract them as separate `variables` in the JSON body. Rejected because it would require significant query restructuring for no practical benefit — the query strings themselves are well under any size limit; the issue is purely the CLI argument delivery mechanism.

- **Batch size reduction:** Limit queries to smaller batches to stay under the CLI limit. Rejected because it adds complexity, increases API calls, and doesn't address the root cause.

## Consequences

- **Positive:** All GraphQL exec.Command call sites now use stdin, eliminating the Windows command-line length limit entirely
- **Positive:** Consistent pattern across all batch query functions — matches the existing `executeBatchMutation()` approach
- **Positive:** The `buildGraphQLRequestBody` helper is independently testable with `json.Marshal` for safe escaping
- **Trade-off:** Slightly more code per call site (3 extra lines for request body building and stdin assignment)

## Issues Encountered

- The large payload test initially used 50 aliases (~30KB) which was under the 32KB threshold. Increased to 60 aliases to reliably exceed the limit for the test assertion.
