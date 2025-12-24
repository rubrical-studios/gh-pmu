# Proposal: Local SQLite Cache for Releases and Microsprints

**Date:** 2024-12-24
**Status:** Draft
**Related Issue:** #455

---

## Executive Summary

Evaluate SQLite as a local cache for release and microsprint tracker data, with comparison to alternative caching strategies. This proposal analyzes trade-offs for offline-capable, low-latency access to tracker metadata.

**Context:** This proposal complements PROPOSAL-Workflow-Cache-Manifest.md (#434) which proposes a server-side manifest. A local cache could provide additional benefits for offline scenarios and reduced network dependency.

---

## Problem Statement

Current `gh pmu release list` and `gh pmu microsprint list` commands:
- Require 2 API calls each (~800ms latency)
- Fail without network connectivity
- Consume API rate limits on every invocation

**Goal:** Provide sub-50ms responses with offline capability.

---

## Option 1: SQLite (Embedded Database)

### Overview

Store tracker metadata in a local SQLite database at `~/.config/gh-pmu/cache.db` or project-level `.gh-pmu-cache.db`.

### Schema

```sql
CREATE TABLE trackers (
    id INTEGER PRIMARY KEY,
    repo TEXT NOT NULL,
    type TEXT NOT NULL,  -- 'release' | 'microsprint'
    number INTEGER NOT NULL,
    title TEXT NOT NULL,
    state TEXT NOT NULL,  -- 'open' | 'closed'
    created_at TEXT,
    closed_at TEXT,
    updated_at TEXT,
    UNIQUE(repo, type, number)
);

CREATE TABLE cache_metadata (
    key TEXT PRIMARY KEY,
    value TEXT
);

CREATE INDEX idx_repo_type_state ON trackers(repo, type, state);
```

### Pros

- **Rich queries**: Filter by state, date ranges, full-text search
- **Atomic updates**: ACID transactions prevent corruption
- **Battle-tested**: SQLite handles billions of deployments
- **Offline-first**: Full functionality without network
- **Incremental sync**: Update only changed records

### Cons

- **Dependency**: Adds CGO dependency (or use modernc.org/sqlite for pure Go)
- **Binary size**: +2-5MB to binary
- **Schema migrations**: Must handle version upgrades
- **Single-user**: Local cache doesn't sync across team members
- **Staleness risk**: Cache can diverge from server state

### Implementation Complexity

Medium-High. Requires:
- Schema definition and migrations
- Repository pattern for data access
- Sync logic with conflict resolution
- Cache invalidation strategy

---

## Option 2: JSON File Cache

### Overview

Store tracker data in a simple JSON file (similar to the manifest proposal but client-side).

```json
{
  "version": 1,
  "updated_at": "2024-12-24T10:00:00Z",
  "repos": {
    "rubrical-studios/gh-pmu": {
      "releases": [...],
      "microsprints": [...]
    }
  }
}
```

### Pros

- **Zero dependencies**: Native Go encoding/json
- **Human readable**: Easy to inspect and debug
- **Simple**: Minimal code complexity
- **Portable**: Works everywhere

### Cons

- **Full rewrites**: Must read/write entire file on updates
- **No concurrent access**: File locking required
- **Limited queries**: Must load all data into memory
- **Size limits**: Performance degrades with large datasets

### Implementation Complexity

Low. 100-200 lines of code.

---

## Option 3: BoltDB / bbolt

### Overview

Use BoltDB (pure Go key-value store) for structured local storage.

### Pros

- **Pure Go**: No CGO, easy cross-compilation
- **ACID**: Transactional guarantees
- **Embedded**: Single-file database
- **Mature**: Battle-tested in etcd, InfluxDB, etc.

### Cons

- **Key-value only**: No SQL, queries require manual implementation
- **No concurrent writers**: Single writer, multiple readers
- **Dependency**: Additional module to maintain

### Implementation Complexity

Medium. Less than SQLite but more than JSON.

---

## Option 4: In-Memory + Periodic Refresh

### Overview

Cache data in memory during CLI session with background refresh.

### Pros

- **Fastest reads**: Sub-1ms access
- **No disk I/O**: No file management
- **Always fresh**: Periodic refresh keeps data current

### Cons

- **No persistence**: Lost on process exit
- **Memory overhead**: Must hold all data in RAM
- **Cold start**: First call still slow

### Implementation Complexity

Low-Medium. Good for long-running processes, not CLI.

---

## Option 5: Hybrid (Manifest + Local Fallback)

### Overview

Combine PROPOSAL-Workflow-Cache-Manifest.md (server-side) with a local JSON fallback.

1. Check for `.github/pmu-cache.json` in repo (git-tracked, team-shared)
2. Fall back to `~/.config/gh-pmu/cache.json` (local, user-specific)
3. Fall back to API

### Pros

- **Best of both**: Team sync via manifest, offline via local
- **Graceful degradation**: Works in all scenarios
- **Simple implementation**: JSON for both layers

### Cons

- **Two caches**: Potential inconsistency
- **Complexity**: Multiple code paths

### Implementation Complexity

Medium. Two cache layers but same format.

---

## Comparison Matrix

| Criteria | SQLite | JSON File | BoltDB | In-Memory | Hybrid |
|----------|--------|-----------|--------|-----------|--------|
| Read latency | <5ms | <10ms | <5ms | <1ms | <10ms |
| Dependencies | CGO/pure | None | Module | None | None |
| Binary size | +2-5MB | +0 | +1MB | +0 | +0 |
| Offline support | Full | Full | Full | None | Full |
| Team sync | No | No | No | No | Partial |
| Query flexibility | High | Low | Medium | Medium | Low |
| Implementation effort | High | Low | Medium | Low | Medium |
| Maintenance burden | Medium | Low | Low | Low | Low |

---

## Multi-Developer Sync Considerations

Local caches are inherently single-user. Supporting 2..n developers requires additional infrastructure.

### Sync Strategies

#### Strategy A: Git-Tracked Manifest (Recommended)

Use the approach from PROPOSAL-Workflow-Cache-Manifest.md (#434):

```
.github/pmu-cache.json  <-- committed to repo
```

**How it works:**
1. GitHub Actions workflow updates manifest on tracker changes
2. Developers pull latest via `git pull`
3. All team members read same cache file

**Requirements:**
- GitHub Actions workflow (IDPF framework provides this)
- Developers must `git pull` to get updates
- Cache is eventually consistent (seconds to minutes)

**Pros:** Zero additional infrastructure, works with existing git workflow
**Cons:** Creates commits, requires pull to sync

---

#### Strategy B: GitHub API as Source of Truth

Don't sync local caches - treat GitHub as the canonical source.

```
Developer A: local cache -> stale? -> fetch from GitHub API
Developer B: local cache -> stale? -> fetch from GitHub API
```

**How it works:**
1. Each developer maintains independent local cache
2. TTL-based invalidation (e.g., 1 hour)
3. On cache miss or stale, fetch fresh from API

**Requirements:**
- Short TTL to limit staleness window
- `--refresh` flag for immediate sync
- Accept that caches may temporarily diverge

**Pros:** Simple, no sync infrastructure
**Cons:** Eventual consistency, API rate limit consumption

---

#### Strategy C: Shared Cache Server

Deploy a lightweight cache service accessible to all developers.

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│ Developer A │────▶│ Cache Server │◀────│ Developer B │
└─────────────┘     └──────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ GitHub API  │
                    └─────────────┘
```

**How it works:**
1. Central server fetches from GitHub, caches results
2. Developers query cache server instead of GitHub
3. Server handles TTL, invalidation, rate limits

**Requirements:**
- Server infrastructure (could be serverless/edge function)
- Network access from developer machines
- Authentication/authorization

**Pros:** Single source of truth, protects API rate limits
**Cons:** Infrastructure to maintain, network dependency

---

#### Strategy D: Webhook-Triggered Push

GitHub webhooks push updates to developer machines.

**Requirements:**
- Publicly accessible endpoints on dev machines (impractical)
- Or: polling service that developers subscribe to

**Verdict:** Not practical for CLI tool. Dismiss.

---

### Sync Requirements by Cache Type

| Cache Type | Sync Approach | Additional Work |
|------------|---------------|-----------------|
| SQLite | Export/import or Strategy B | Schema for sync metadata, conflict resolution |
| JSON File | Git-tracked (Strategy A) | Merge conflict handling |
| BoltDB | Strategy B only | No practical sync option |
| In-Memory | N/A | Cannot sync |
| Hybrid | Strategy A + B combined | Already designed for this |

---

### Recommended Multi-Dev Approach

**For gh-pmu:** Combine Strategies A and B:

1. **Primary:** Git-tracked manifest (`.github/pmu-cache.json`)
   - Updated by GitHub Actions on tracker changes
   - Shared via normal git operations
   - Team-wide consistency

2. **Fallback:** Local cache with TTL (`~/.config/gh-pmu/cache.json`)
   - Used when manifest unavailable or stale
   - Independent refresh via API
   - Enables offline work

3. **Override:** `--refresh` flag
   - Bypasses all caches, fetches fresh from API
   - Useful when developer knows cache is stale

**Priority order:**
```
1. Check manifest freshness -> use if fresh
2. Check local cache freshness -> use if fresh
3. Fetch from API -> update local cache
```

**Consistency guarantee:** Eventual (seconds to minutes). Acceptable for tracker metadata where real-time sync is not critical.

---

## Recommendation

**Short-term (v1.0):** Implement **Option 2 (JSON File Cache)** as a simple local fallback:
- Zero new dependencies
- Minimal code (150 lines)
- Pairs well with manifest proposal

**Long-term consideration:** If query needs grow (filtering, search, date ranges), evaluate SQLite with pure Go driver (modernc.org/sqlite).

**Not recommended for gh-pmu:**
- SQLite: Overkill for current use case. Binary size and CGO concerns outweigh benefits.
- BoltDB: Middle ground that doesn't offer enough over JSON.
- In-Memory: CLI is short-lived, no persistence benefit.

---

## Implementation Sketch (JSON Cache)

```go
// internal/cache/cache.go
type LocalCache struct {
    path string
    data CacheData
}

type CacheData struct {
    Version   int                          `json:"version"`
    UpdatedAt time.Time                    `json:"updated_at"`
    Repos     map[string]RepoTrackers      `json:"repos"`
}

func (c *LocalCache) GetReleases(repo string) ([]Tracker, bool) {
    // Check freshness, return cached data or indicate miss
}

func (c *LocalCache) SetReleases(repo string, releases []Tracker) error {
    // Update cache, write to disk
}
```

---

## Open Questions

1. Should local cache be per-project or global (`~/.config/gh-pmu/`)?
2. What TTL makes sense for local cache? (suggest: 1 hour, configurable)
3. Should cache invalidation be explicit (`gh pmu cache clear`) or automatic?
4. How does this interact with the manifest proposal? Priority order?

---

## Acceptance Criteria

- [ ] Cache storage location defined and documented
- [ ] Read/write performance meets <50ms target
- [ ] Graceful fallback when cache missing or corrupted
- [ ] Cache TTL configurable in `.gh-pmu.yml`
- [ ] Clear mechanism for cache invalidation
- [ ] Works offline after initial population

---

## References

- PROPOSAL-Workflow-Cache-Manifest.md - Server-side manifest approach
- Issue #434: Enhancement: Cache release/microsprint list data locally
- modernc.org/sqlite - Pure Go SQLite (if SQLite pursued later)
- go.etcd.io/bbolt - BoltDB fork maintained by etcd team
