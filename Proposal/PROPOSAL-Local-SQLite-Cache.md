# Proposal: Local Database Cache for Releases and Microsprints

**Date:** 2024-12-24
**Status:** Draft
**Related Issue:** #455

---

## Executive Summary

Implement a local embedded database cache for release and microsprint tracker data, synchronized via GitHub Actions. This proposal evaluates SQLite and BoltDB as storage options.

**Architecture:**
```
GitHub Actions ──▶ .github/pmu-cache.json ──▶ git pull ──▶ Local DB
     (push)              (manifest)            (sync)       (query)
```

**Goal:** Sub-50ms queries with offline capability and multi-developer consistency.

---

## Problem Statement

Current `gh pmu release list` and `gh pmu microsprint list` commands:
- Require 2 API calls each (~800ms latency)
- Fail without network connectivity
- Consume API rate limits on every invocation
- No team synchronization

---

## Option 1: SQLite

### Overview

Store tracker metadata in a local SQLite database at `~/.config/gh-pmu/cache.db`.

### Schema

```sql
CREATE TABLE trackers (
    id INTEGER PRIMARY KEY,
    repo TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('release', 'microsprint')),
    number INTEGER NOT NULL,
    title TEXT NOT NULL,
    state TEXT NOT NULL CHECK(state IN ('open', 'closed')),
    created_at TEXT,
    closed_at TEXT,
    synced_at TEXT NOT NULL,
    UNIQUE(repo, type, number)
);

CREATE TABLE sync_metadata (
    repo TEXT PRIMARY KEY,
    manifest_hash TEXT,
    last_sync TEXT NOT NULL
);

CREATE INDEX idx_repo_type_state ON trackers(repo, type, state);
CREATE INDEX idx_repo_synced ON trackers(repo, synced_at);
```

### Implementation

```go
// internal/cache/sqlite.go
import "modernc.org/sqlite"  // Pure Go, no CGO

type SQLiteCache struct {
    db *sql.DB
}

func (c *SQLiteCache) Sync(manifest Manifest) error {
    tx, _ := c.db.Begin()
    defer tx.Rollback()

    for _, tracker := range manifest.Releases {
        tx.Exec(`INSERT OR REPLACE INTO trackers
            (repo, type, number, title, state, created_at, closed_at, synced_at)
            VALUES (?, 'release', ?, ?, ?, ?, ?, ?)`,
            manifest.Repo, tracker.Number, tracker.Title,
            tracker.State, tracker.CreatedAt, tracker.ClosedAt, time.Now())
    }

    tx.Exec(`INSERT OR REPLACE INTO sync_metadata (repo, manifest_hash, last_sync)
        VALUES (?, ?, ?)`, manifest.Repo, manifest.Hash(), time.Now())

    return tx.Commit()
}

func (c *SQLiteCache) GetReleases(repo string, state string) ([]Tracker, error) {
    rows, _ := c.db.Query(`
        SELECT number, title, state, created_at, closed_at
        FROM trackers
        WHERE repo = ? AND type = 'release' AND (? = '' OR state = ?)
        ORDER BY number DESC`,
        repo, state, state)
    // ... scan rows
}
```

### Pros

| Benefit | Description |
|---------|-------------|
| Rich queries | Filter by state, date ranges, counts, aggregations |
| ACID transactions | Atomic sync, no corruption |
| Battle-tested | Billions of deployments worldwide |
| Incremental sync | Update only changed records |
| Pure Go option | `modernc.org/sqlite` - no CGO required |

### Cons

| Concern | Mitigation |
|---------|------------|
| Binary size +3-5MB | Acceptable for CLI tool |
| Schema migrations | Version table, auto-migrate on startup |
| Learning curve | Well-documented, standard SQL |

### Query Capabilities

```sql
-- Open releases
SELECT * FROM trackers WHERE type = 'release' AND state = 'open';

-- Releases created this month
SELECT * FROM trackers
WHERE type = 'release'
AND created_at >= date('now', 'start of month');

-- Count by state
SELECT state, COUNT(*) FROM trackers WHERE type = 'microsprint' GROUP BY state;

-- Stale cache detection
SELECT repo FROM sync_metadata WHERE last_sync < datetime('now', '-1 hour');
```

---

## Option 2: BoltDB (bbolt)

### Overview

Use BoltDB (pure Go key-value store) for structured local storage at `~/.config/gh-pmu/cache.bolt`.

### Data Structure

```
Bucket: "trackers"
  └── Key: "{repo}:release:{number}"
      Value: JSON encoded Tracker

Bucket: "sync"
  └── Key: "{repo}"
      Value: JSON encoded SyncMetadata
```

### Implementation

```go
// internal/cache/bolt.go
import "go.etcd.io/bbolt"

type BoltCache struct {
    db *bbolt.DB
}

func (c *BoltCache) Sync(manifest Manifest) error {
    return c.db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("trackers"))

        for _, tracker := range manifest.Releases {
            key := fmt.Sprintf("%s:release:%d", manifest.Repo, tracker.Number)
            data, _ := json.Marshal(tracker)
            b.Put([]byte(key), data)
        }

        syncBucket := tx.Bucket([]byte("sync"))
        meta := SyncMetadata{Hash: manifest.Hash(), LastSync: time.Now()}
        data, _ := json.Marshal(meta)
        syncBucket.Put([]byte(manifest.Repo), data)

        return nil
    })
}

func (c *BoltCache) GetReleases(repo string, state string) ([]Tracker, error) {
    var results []Tracker

    c.db.View(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("trackers"))
        prefix := []byte(fmt.Sprintf("%s:release:", repo))

        c := b.Cursor()
        for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
            var t Tracker
            json.Unmarshal(v, &t)
            if state == "" || t.State == state {
                results = append(results, t)
            }
        }
        return nil
    })

    return results, nil
}
```

### Pros

| Benefit | Description |
|---------|-------------|
| Pure Go | No CGO, easy cross-compilation |
| ACID | Transactional guarantees |
| Single file | Simple deployment |
| Mature | Battle-tested in etcd, InfluxDB |
| Small binary | +1MB vs +3-5MB for SQLite |

### Cons

| Concern | Mitigation |
|---------|------------|
| Key-value only | Prefix scanning for queries |
| No SQL | Manual query implementation |
| Limited filtering | Must scan and filter in Go |
| Single writer | Not an issue for CLI |

### Query Capabilities

```go
// Open releases - must scan and filter
releases, _ := cache.GetReleases(repo, "open")

// Count by state - must load all and count in Go
all, _ := cache.GetReleases(repo, "")
openCount := countByState(all, "open")

// Date filtering - must implement in Go
releases = filterByDate(releases, startDate, endDate)
```

---

## Comparison Matrix

| Criteria | SQLite | BoltDB |
|----------|--------|--------|
| Query flexibility | High (SQL) | Low (scan + filter) |
| Binary size impact | +3-5MB | +1MB |
| CGO required | No (modernc.org/sqlite) | No |
| ACID transactions | Yes | Yes |
| Concurrent readers | Yes | Yes |
| Concurrent writers | Yes (WAL mode) | No (single writer) |
| Learning curve | Low (SQL) | Medium (K/V patterns) |
| Aggregations | Native (GROUP BY) | Manual |
| Date range queries | Native | Manual |
| Full-text search | FTS5 extension | Not available |

---

## GitHub Actions Sync Architecture

### Workflow: Manifest Generator

Triggers on tracker issue events, generates shared manifest.

```yaml
# .github/workflows/pmu-cache.yml
name: Update PMU Cache Manifest

on:
  issues:
    types: [opened, closed, reopened, edited, labeled, unlabeled]
  schedule:
    - cron: '0 */6 * * *'  # Refresh every 6 hours
  workflow_dispatch:

jobs:
  update-manifest:
    if: |
      github.event_name != 'issues' ||
      contains(github.event.issue.labels.*.name, 'release') ||
      contains(github.event.issue.labels.*.name, 'microsprint')
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4

      - name: Generate manifest
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          gh api graphql -f query='
            query($owner: String!, $repo: String!) {
              repository(owner: $owner, name: $repo) {
                releases: issues(labels: ["release"], first: 100, states: [OPEN, CLOSED]) {
                  nodes { number title state createdAt closedAt }
                }
                microsprints: issues(labels: ["microsprint"], first: 100, states: [OPEN, CLOSED]) {
                  nodes { number title state createdAt closedAt }
                }
              }
            }
          ' -f owner="${{ github.repository_owner }}" \
            -f repo="${{ github.event.repository.name }}" > /tmp/data.json

          jq '{
            version: 1,
            updated_at: (now | strftime("%Y-%m-%dT%H:%M:%SZ")),
            repo: "${{ github.repository }}",
            releases: .data.repository.releases.nodes,
            microsprints: .data.repository.microsprints.nodes
          }' /tmp/data.json > .github/pmu-cache.json

      - name: Commit manifest
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add .github/pmu-cache.json
          git diff --cached --quiet || git commit -m "chore: update pmu cache manifest"
          git push
```

### Manifest Schema

```json
{
  "version": 1,
  "updated_at": "2024-12-24T10:00:00Z",
  "repo": "rubrical-studios/gh-pmu",
  "releases": [
    {
      "number": 430,
      "title": "Release: v0.9.0",
      "state": "OPEN",
      "createdAt": "2024-12-20T08:00:00Z",
      "closedAt": null
    }
  ],
  "microsprints": [
    {
      "number": 425,
      "title": "Microsprint: cache-impl",
      "state": "CLOSED",
      "createdAt": "2024-12-23T09:00:00Z",
      "closedAt": "2024-12-24T15:00:00Z"
    }
  ]
}
```

### Sync Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     GitHub Actions                               │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐         │
│  │ Issue Event  │──▶│ GraphQL      │──▶│ Commit       │         │
│  │ or Cron      │   │ Query        │   │ Manifest     │         │
│  └──────────────┘   └──────────────┘   └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    .github/pmu-cache.json (in repo)
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
        Developer A     Developer B     Developer C
              │               │               │
              ▼               ▼               ▼
        ┌─────────────────────────────────────────┐
        │           git pull (periodic)            │
        └─────────────────────────────────────────┘
              │               │               │
              ▼               ▼               ▼
        ┌─────────┐     ┌─────────┐     ┌─────────┐
        │ Local   │     │ Local   │     │ Local   │
        │ DB Sync │     │ DB Sync │     │ DB Sync │
        └─────────┘     └─────────┘     └─────────┘
              │               │               │
              ▼               ▼               ▼
        SQLite/Bolt     SQLite/Bolt     SQLite/Bolt
        (~/.config/     (~/.config/     (~/.config/
         gh-pmu/)        gh-pmu/)        gh-pmu/)
              │               │               │
              ▼               ▼               ▼
        gh pmu list     gh pmu list     gh pmu list
        (<5ms)          (<5ms)          (<5ms)
```

### CLI Sync Logic

```go
// internal/cache/sync.go
func (c *Cache) SyncFromManifest(repoPath string) error {
    manifestPath := filepath.Join(repoPath, ".github", "pmu-cache.json")

    // Check if manifest exists
    data, err := os.ReadFile(manifestPath)
    if err != nil {
        return c.FallbackToAPI()  // No manifest, use API
    }

    var manifest Manifest
    json.Unmarshal(data, &manifest)

    // Check if already synced (hash comparison)
    if c.IsSynced(manifest.Hash()) {
        return nil  // Already up to date
    }

    // Sync to local database
    return c.db.Sync(manifest)
}

// Called on: gh pmu release list, gh pmu microsprint list
func runListWithCache(cmd *cobra.Command, repo string) error {
    cache := GetCache()

    // Try sync from manifest (fast, local file read)
    cache.SyncFromManifest(getRepoPath())

    // Query local database
    releases, err := cache.GetReleases(repo, "")
    if err != nil || len(releases) == 0 {
        // Fallback to API if cache empty/corrupt
        return fetchFromAPI(cmd, repo)
    }

    return printReleases(cmd, releases)
}
```

---

## Multi-Developer Sync

### How It Works

1. **Developer A** creates/closes a release issue
2. **GitHub Actions** triggers, regenerates manifest, commits to repo
3. **Developer B** runs `git pull` (or it happens automatically)
4. **Developer B** runs `gh pmu release list`
5. **CLI** detects manifest changed (hash mismatch), syncs to local DB
6. **CLI** queries local DB (<5ms)

### Consistency Model

| Scenario | Latency | Consistency |
|----------|---------|-------------|
| Same developer, after action | Immediate | Strong |
| Other developer, after git pull | Seconds | Strong |
| Other developer, before git pull | Up to 6 hours | Eventual |
| Offline | N/A | Last synced state |

### Auto-Pull Integration (Optional)

```yaml
# .gh-pmu.yml
cache:
  auto_pull: true           # Run git pull before list commands
  auto_pull_interval: 300   # Only if last pull >5 min ago
```

```go
func maybePull(repoPath string, cfg Config) {
    if !cfg.Cache.AutoPull {
        return
    }

    lastPull := getLastPullTime(repoPath)
    if time.Since(lastPull) > cfg.Cache.AutoPullInterval {
        exec.Command("git", "-C", repoPath, "pull", "--ff-only").Run()
    }
}
```

---

## Recommendation

**SQLite with `modernc.org/sqlite`** (pure Go driver)

| Reason | Explanation |
|--------|-------------|
| Query flexibility | SQL enables filtering, aggregations, date ranges natively |
| Future-proof | Easy to add new queries without code changes |
| Familiar | Standard SQL, easy to debug with `sqlite3` CLI |
| Same binary size class | Both add ~1-5MB, not a deciding factor |
| Battle-tested | SQLite is the most deployed database in the world |

**BoltDB** is a solid alternative if:
- Minimizing binary size is critical
- Query needs remain simple (list all, filter by state)
- Team prefers K/V patterns over SQL

---

## Implementation Plan

### Phase 1: Foundation
- [ ] Add `modernc.org/sqlite` dependency
- [ ] Create schema and migration logic
- [ ] Implement `Cache` interface with SQLite backend

### Phase 2: Sync
- [ ] Implement manifest reader
- [ ] Add hash-based change detection
- [ ] Integrate sync into `release list` and `microsprint list`

### Phase 3: GitHub Actions
- [ ] Create workflow template in IDPF framework
- [ ] Document workflow installation
- [ ] Add `gh pmu init` workflow installation option

### Phase 4: Configuration
- [ ] Add `cache` section to `.gh-pmu.yml`
- [ ] Implement `--refresh` flag for force API fetch
- [ ] Add `gh pmu cache` command group (status, clear, sync)

---

## Configuration

```yaml
# .gh-pmu.yml
cache:
  enabled: true                    # Use local cache (default: true)
  backend: sqlite                  # sqlite | bolt
  path: ~/.config/gh-pmu/cache.db  # Override default location
  auto_sync: true                  # Sync from manifest on list commands
  auto_pull: false                 # Git pull before sync
  auto_pull_interval: 300          # Seconds between auto-pulls
```

---

## Acceptance Criteria

- [ ] SQLite cache with pure Go driver (no CGO)
- [ ] Manifest sync from `.github/pmu-cache.json`
- [ ] Hash-based change detection (skip sync if unchanged)
- [ ] `gh pmu release list` reads from cache (<10ms)
- [ ] `gh pmu microsprint list` reads from cache (<10ms)
- [ ] `--refresh` flag bypasses cache
- [ ] Graceful fallback to API when cache unavailable
- [ ] GitHub Actions workflow template provided
- [ ] Works offline after initial sync

---

## References

- PROPOSAL-Workflow-Cache-Manifest.md (#434) - Original manifest concept
- modernc.org/sqlite - Pure Go SQLite driver
- go.etcd.io/bbolt - BoltDB fork maintained by etcd team
