# Proposal: MCP Server Mode for gh-pmu

**Date:** 2025-12-25
**Status:** Draft
**Related Issue:** TBD

---

## Executive Summary

Add an MCP (Model Context Protocol) server mode to gh-pmu, enabling AI assistants to interact with GitHub Projects through structured tool interfaces rather than CLI invocation.

**Key Benefits:**
- Eliminate per-invocation startup overhead (~275-960ms per command)
- Accumulate error patterns for proactive failure prevention
- Provide structured responses vs. text parsing
- Enable stateful caching and batch operations

**Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                     gh pmu mcp serve                        │
├─────────────────────────────────────────────────────────────┤
│  Long-Running State                                         │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────────────┐ │
│  │ Config   │  │ GraphQL  │  │ Project Items Cache (TTL) │ │
│  │ (1x)     │  │ Client   │  │ + Error Pattern DB        │ │
│  └──────────┘  └──────────┘  └───────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│  MCP Tools                       MCP Resources              │
│  ├─ move_issue                   ├─ board://current        │
│  ├─ create_issue                 ├─ issue://{number}       │
│  ├─ list_issues                  └─ errors://recent        │
│  └─ microsprint_*                                           │
└─────────────────────────────────────────────────────────────┘
```

---

## Problem Statement

### Current State: CLI Invocation

Each `gh pmu` command invocation incurs:

| Phase | Latency | Notes |
|-------|---------|-------|
| Process spawn + Go runtime | 50-100ms | OS-dependent |
| Config load (`.gh-pmu.yml`) | 5-10ms | YAML parse + validation |
| GraphQL client init | 20-50ms | Auth context from `gh` CLI |
| API call (typical) | 200-800ms | Network latency + GitHub processing |
| **Total per command** | **275-960ms** | |

A typical Claude Code session runs 20-30 `gh pmu` commands, resulting in **8-29 seconds** of pure overhead.

### Error Handling Gaps

Current implementation (`internal/api/errors.go`):

```go
// Detection exists, but no recovery or learning
var ErrRateLimited = errors.New("API rate limit exceeded")

func IsRateLimited(err error) bool {
    // Returns true/false, but no retry logic follows
}
```

**What's missing:**
- No exponential backoff / retry logic
- No error pattern accumulation across sessions
- No preflight checks based on historical failures
- Same errors repeat session after session

### Observed Error Categories

From usage patterns:

| Error Type | Frequency | Preventable? |
|------------|-----------|--------------|
| Status validation (unchecked criteria) | High | Yes, with preflight |
| Rate limiting | Medium | Yes, with backoff + batching |
| Auth expiration | Low | Yes, with proactive refresh |
| Network timeout | Low | Yes, with retry |
| Invalid field values | Medium | Yes, with validation cache |

---

## Proposed Solution: MCP Server Mode

### New Command

```bash
gh pmu mcp serve [--port 8080] [--transport stdio|sse]
```

Starts a long-running MCP server that:
1. Loads config once
2. Maintains persistent GraphQL client
3. Caches project items with TTL-based refresh
4. Accumulates error patterns in SQLite
5. Exposes operations as MCP tools

### Transport Options

| Transport | Use Case | Configuration |
|-----------|----------|---------------|
| stdio | Claude Code, local AI tools | Default, no config |
| SSE | Web-based clients, remote access | `--transport sse --port 8080` |

---

## Go MCP SDK Landscape

### Official SDK (Recommended)

**Repository:** [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)

| Attribute | Value |
|-----------|-------|
| Version | v1.2.0 (Dec 22, 2025) |
| Maintainers | Go team + Anthropic |
| Stars | 3.5k |
| Dependents | 536 projects |
| Transport | stdio, command-based |
| Status | Production-ready |

**Example pattern:**
```go
import "github.com/modelcontextprotocol/go-sdk/mcp"

server := mcp.NewServer("gh-pmu", "1.0.0")

server.AddTool("move_issue", "Move issue to new status", func(ctx context.Context, input MoveInput) (*MoveOutput, error) {
    // Implementation
})

server.Run(mcp.StdioTransport{})
```

### Community Alternative

**Repository:** [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)

| Attribute | Value |
|-----------|-------|
| Dependents | 400+ packages, 200+ modules |
| Transport | stdio, SSE, WebSocket, gRPC |
| Status | Mature, battle-tested |

**Consideration:** More transport options, but official SDK preferred for long-term maintenance.

### Recommendation

Use **official SDK** (`modelcontextprotocol/go-sdk`) for:
- Anthropic alignment
- Long-term support guarantees
- Simpler API surface

---

## MCP Tool Definitions

### Core Tools

```go
// Tool: move_issue
type MoveIssueInput struct {
    Issue    string `json:"issue" jsonschema:"description=Issue reference (#123 or owner/repo#123)"`
    Status   string `json:"status,omitempty" jsonschema:"description=New status value"`
    Priority string `json:"priority,omitempty" jsonschema:"description=New priority value"`
}

type MoveIssueOutput struct {
    Success   bool     `json:"success"`
    Issue     int      `json:"issue"`
    Updated   []string `json:"updated_fields"`
    Warnings  []string `json:"warnings,omitempty"`  // Preflight warnings
}

// Tool: list_issues
type ListIssuesInput struct {
    Status   string `json:"status,omitempty"`
    Priority string `json:"priority,omitempty"`
    Limit    int    `json:"limit,omitempty" jsonschema:"default=50"`
}

type ListIssuesOutput struct {
    Issues     []Issue `json:"issues"`
    TotalCount int     `json:"total_count"`
    Cached     bool    `json:"cached"`  // Indicates if from cache
    CacheAge   string  `json:"cache_age,omitempty"`
}

// Tool: create_issue
type CreateIssueInput struct {
    Title    string   `json:"title"`
    Body     string   `json:"body,omitempty"`
    Labels   []string `json:"labels,omitempty"`
    Status   string   `json:"status,omitempty"`
    Priority string   `json:"priority,omitempty"`
}

// Tool: microsprint_start, microsprint_add, microsprint_close, etc.
// Tool: release_start, release_add, release_close, etc.
```

### MCP Resources

```go
// Resource: board://current
// Returns current Kanban board state as structured JSON

// Resource: issue://{number}
// Returns detailed issue information including project fields

// Resource: errors://recent
// Returns recent error patterns for debugging

// Resource: config://current
// Returns current configuration (sanitized, no secrets)
```

---

## Error Intelligence System

### SQLite Schema

Location: `~/.config/gh-pmu/errors.db`

```sql
CREATE TABLE error_patterns (
    id INTEGER PRIMARY KEY,
    operation TEXT NOT NULL,           -- "move_issue", "create_issue"
    error_type TEXT NOT NULL,          -- "validation", "rate_limit", "auth"
    error_message TEXT,
    context_hash TEXT,                 -- Hash of relevant context
    context_json TEXT,                 -- Full context for analysis
    frequency INTEGER DEFAULT 1,
    first_seen TEXT NOT NULL,
    last_seen TEXT NOT NULL,
    resolution TEXT,                   -- What fixed it (if known)

    UNIQUE(operation, error_type, context_hash)
);

CREATE TABLE resolutions (
    id INTEGER PRIMARY KEY,
    error_pattern_id INTEGER REFERENCES error_patterns(id),
    description TEXT NOT NULL,
    success_count INTEGER DEFAULT 0,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_operation ON error_patterns(operation);
CREATE INDEX idx_recent ON error_patterns(last_seen DESC);
CREATE INDEX idx_frequency ON error_patterns(frequency DESC);
```

### Preflight Check Flow

```go
func (s *MCPServer) PreflightCheck(op string, input any) (*PreflightResult, error) {
    ctx := extractContext(input)

    // Query historical patterns
    patterns, _ := s.db.Query(`
        SELECT error_type, error_message, frequency, resolution
        FROM error_patterns
        WHERE operation = ?
          AND (context_hash = ? OR context_hash IS NULL)
          AND last_seen > datetime('now', '-7 days')
        ORDER BY frequency DESC
        LIMIT 5
    `, op, hashContext(ctx))

    result := &PreflightResult{Safe: true}

    for _, p := range patterns {
        if p.Frequency >= 3 {
            result.Safe = false
            result.Warnings = append(result.Warnings, Warning{
                Type:       p.ErrorType,
                Message:    fmt.Sprintf("This operation failed %d times recently: %s", p.Frequency, p.ErrorMessage),
                Suggestion: p.Resolution,
            })
        }
    }

    return result, nil
}
```

### Learning from Errors

```go
func (s *MCPServer) RecordError(op string, input any, err error) {
    ctx := extractContext(input)
    errType := classifyError(err)

    s.db.Exec(`
        INSERT INTO error_patterns (operation, error_type, error_message, context_hash, context_json, first_seen, last_seen)
        VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))
        ON CONFLICT(operation, error_type, context_hash) DO UPDATE SET
            frequency = frequency + 1,
            last_seen = datetime('now'),
            error_message = excluded.error_message
    `, op, errType, err.Error(), hashContext(ctx), jsonEncode(ctx))
}

func (s *MCPServer) RecordResolution(patternID int, description string) {
    s.db.Exec(`
        INSERT INTO resolutions (error_pattern_id, description, created_at)
        VALUES (?, ?, datetime('now'))
    `, patternID, description)

    // Update pattern with resolution
    s.db.Exec(`
        UPDATE error_patterns SET resolution = ? WHERE id = ?
    `, description, patternID)
}
```

### Practical Example

**Session 1:**
```
Claude: move_issue(42, status="done")
Error: "Cannot move to done: 2 unchecked acceptance criteria"
→ Server records: {op: "move_issue", error: "validation", context: {to_status: "done"}}
```

**Session 2:**
```
Claude: move_issue(57, status="done")
→ Preflight detects: "move_issue to done" failed 1 time
→ Still proceeds (threshold not met)
→ Error again
→ Frequency now 2
```

**Session 3:**
```
Claude: move_issue(99, status="done")
→ Preflight: "Warning: Moving to 'done' has failed 2 times recently due to unchecked criteria.
              Suggestion: Verify acceptance criteria are checked first."
→ Claude decides to check criteria before proceeding
→ Error prevented
```

---

## Caching Architecture

### In-Memory Cache

```go
type ProjectCache struct {
    mu           sync.RWMutex
    items        map[string]*ProjectItem  // key: "owner/repo#number"
    lastRefresh  time.Time
    ttl          time.Duration            // Default: 60 seconds
    refreshing   bool
}

func (c *ProjectCache) Get(key string) (*ProjectItem, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if time.Since(c.lastRefresh) > c.ttl {
        go c.backgroundRefresh()  // Non-blocking refresh
    }

    item, ok := c.items[key]
    return item, ok
}

func (c *ProjectCache) backgroundRefresh() {
    c.mu.Lock()
    if c.refreshing {
        c.mu.Unlock()
        return
    }
    c.refreshing = true
    c.mu.Unlock()

    defer func() {
        c.mu.Lock()
        c.refreshing = false
        c.mu.Unlock()
    }()

    items, err := c.client.GetProjectItems()
    if err != nil {
        return  // Keep stale data
    }

    c.mu.Lock()
    c.items = items
    c.lastRefresh = time.Now()
    c.mu.Unlock()
}
```

### Cache Invalidation

```go
// Invalidate after mutations
func (s *MCPServer) MoveIssue(ctx context.Context, input MoveIssueInput) (*MoveIssueOutput, error) {
    // ... perform mutation ...

    // Invalidate specific item
    s.cache.Invalidate(input.Issue)

    // Or mark for refresh
    s.cache.MarkDirty(input.Issue)

    return output, nil
}
```

---

## Performance Comparison

### Latency (20 operations)

| Metric | CLI Mode | MCP Server |
|--------|----------|------------|
| Process spawns | 20 | 1 |
| Config loads | 20 | 1 |
| Auth contexts | 20 | 1 |
| API calls | 20-40 | 5-10 (cached) |
| Total overhead | 8-29 sec | ~2 sec |
| Error memory | None | Persistent |

### Resource Usage

| Resource | CLI (per call) | MCP Server (persistent) |
|----------|----------------|-------------------------|
| Memory | 20-30MB (each) | 50-100MB (total) |
| File handles | Open/close each | Persistent connections |
| Auth tokens | Loaded each time | Cached |

---

## Configuration

### Server Configuration

```yaml
# .gh-pmu.yml
mcp:
  enabled: true
  transport: stdio              # stdio | sse
  port: 8080                    # For SSE transport
  cache:
    ttl: 60                     # Seconds
    max_items: 1000
  errors:
    database: ~/.config/gh-pmu/errors.db
    retention_days: 30
    preflight_threshold: 3      # Failures before warning
```

### Claude Code Integration

```json
// ~/.claude/claude_desktop_config.json (or equivalent)
{
  "mcpServers": {
    "gh-pmu": {
      "command": "gh",
      "args": ["pmu", "mcp", "serve"]
    }
  }
}
```

---

## Implementation Plan

### Phase 1: Foundation
- [ ] Add `modelcontextprotocol/go-sdk` dependency
- [ ] Create `cmd/mcp.go` with `serve` subcommand
- [ ] Implement stdio transport
- [ ] Wrap existing command logic as MCP tools

### Phase 2: Core Tools
- [ ] `move_issue` tool with structured I/O
- [ ] `list_issues` tool with filtering
- [ ] `create_issue` tool
- [ ] `view_issue` tool
- [ ] `board` resource

### Phase 3: Caching
- [ ] In-memory project items cache
- [ ] TTL-based refresh
- [ ] Cache invalidation on mutations
- [ ] Cache status in tool responses

### Phase 4: Error Intelligence
- [ ] SQLite error pattern database
- [ ] Error recording after failures
- [ ] Preflight check system
- [ ] Resolution tracking
- [ ] `errors://recent` resource

### Phase 5: Advanced Features
- [ ] Microsprint tools
- [ ] Release tools
- [ ] SSE transport option
- [ ] Batch operations
- [ ] Rate limit handling with backoff

---

## Backward Compatibility

### CLI Remains Primary

The MCP server is **additive**. All existing CLI commands continue to work:

```bash
# These still work exactly as before
gh pmu move 42 --status done
gh pmu list --status in_progress
gh pmu board
```

### Shared Core

```
┌─────────────────────────────────────────┐
│  gh pmu CLI (user-facing)               │
├─────────────────────────────────────────┤
│  gh pmu mcp serve (MCP server mode)     │
├─────────────────────────────────────────┤
│  internal/operations/* (shared logic)   │
└─────────────────────────────────────────┘
```

Both CLI and MCP server use the same underlying operation functions.

---

## Security Considerations

### Authentication

- MCP server inherits auth from `gh` CLI (same as current)
- No additional credentials required
- Token refresh handled transparently

### Local-Only by Default

- stdio transport: No network exposure
- SSE transport: Binds to localhost by default
- Explicit `--host 0.0.0.0` required for remote access

### Error Database

- Stored in user config directory
- Contains operation context (issue numbers, field values)
- No secrets or tokens stored
- Auto-cleanup via retention policy

---

## Acceptance Criteria

- [ ] `gh pmu mcp serve` starts MCP server with stdio transport
- [ ] Core tools (move, list, create, view) functional
- [ ] Project items cached with configurable TTL
- [ ] Error patterns recorded to SQLite
- [ ] Preflight warnings for repeat failures
- [ ] CLI commands unaffected
- [ ] Documentation for Claude Code integration
- [ ] Unit tests for MCP tool handlers
- [ ] Integration test with mock MCP client

---

## Open Questions

1. **Transport priority:** Focus on stdio only for v1, or include SSE?
2. **Error sharing:** Should error patterns sync across team (like cache manifest)?
3. **Metrics:** Add telemetry for cache hit rates, error frequencies?
4. **Graceful degradation:** How to handle server crashes mid-session?

---

## References

- [Official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) - v1.2.0
- [MCP Specification](https://modelcontextprotocol.io/docs)
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Community SDK
- PROPOSAL-Local-SQLite-Cache.md - Related caching architecture
- `internal/api/errors.go` - Current error handling
- `internal/config/config.go` - Current config loading

