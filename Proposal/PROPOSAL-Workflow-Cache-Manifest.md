# Proposal: Workflow-Based Cache Manifest

**Date:** 2025-12-23
**Status:** Draft
**Related Issue:** #434

---

## Executive Summary

Introduce a GitHub Actions workflow (provided by the IDPF framework) that maintains a JSON manifest file containing release and microsprint tracker data. The `gh pmu` CLI reads this manifest for `list` commands instead of making multiple API calls.

**Benefits:**
- Reduce `release list` / `microsprint list` from ~800ms to ~50ms
- Single HTTP fetch vs multiple GraphQL queries
- Push-based cache (updated on change) vs pull-based (fetched on demand)

**Framework Responsibility:** The workflow is part of IDPF, not gh-pmu. Projects adopting IDPF get the workflow automatically.

**gh-pmu Responsibility:** Support reading from the manifest file when available, with fallback to API calls.

---

## Current State

`gh pmu release list` and `gh pmu microsprint list` make 2 API calls each:

```go
openIssues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
closedIssues, err := client.GetClosedIssuesByLabel(owner, repo, "release")
```

This results in:
- ~0.78s response time for `release list`
- ~0.76s response time for `microsprint list`
- API rate limit consumption on every invocation

---

## Proposed Solution

### Part 1: Framework Workflow (IDPF)

A GitHub Actions workflow that triggers on tracker issue events:

```yaml
# .github/workflows/pmu-cache.yml (provided by IDPF framework)
name: Update PMU Cache

on:
  issues:
    types: [opened, closed, reopened, edited, labeled, unlabeled]

jobs:
  update-cache:
    if: |
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
            -f repo="${{ github.event.repository.name }}" \
            --jq '.data.repository' > .github/pmu-cache.json

      - name: Commit manifest
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add .github/pmu-cache.json
          git diff --cached --quiet || git commit -m "chore: update pmu cache manifest"
          git push
```

### Part 2: Manifest Schema

```json
{
  "updated_at": "2025-12-23T10:30:00Z",
  "releases": [
    {
      "number": 430,
      "title": "Release: release/v0.9.0",
      "state": "OPEN",
      "created_at": "2025-12-20T08:00:00Z",
      "closed_at": null
    }
  ],
  "microsprints": [
    {
      "number": 425,
      "title": "Microsprint: 2025-12-23-cache",
      "state": "OPEN",
      "created_at": "2025-12-23T09:00:00Z",
      "closed_at": null
    }
  ]
}
```

### Part 3: gh-pmu Changes

Modify `release list` and `microsprint list` to:

1. **Check for manifest**: Look for `.github/pmu-cache.json` in repo
2. **Use manifest if fresh**: Read from file if `updated_at` within threshold (configurable, default 1 hour)
3. **Fallback to API**: If manifest missing, stale, or `--refresh` flag used

```go
func runReleaseListWithDeps(...) error {
    // Try manifest first
    if manifest, err := loadManifest(".github/pmu-cache.json"); err == nil {
        if !opts.refresh && manifest.IsFresh(cfg.Cache.MaxAge) {
            return printReleasesFromManifest(cmd, manifest)
        }
    }

    // Fallback to API
    return fetchAndPrintReleases(cmd, cfg, client)
}
```

### Part 4: Configuration

```yaml
# .gh-pmu.yml
cache:
  enabled: true           # Use manifest cache (default: true)
  max_age: 3600           # Max cache age in seconds (default: 1 hour)
  manifest: ".github/pmu-cache.json"  # Manifest location
```

---

## Storage Location Options

| Option | Pros | Cons | Recommendation |
|--------|------|------|----------------|
| `.github/pmu-cache.json` | Versioned, accessible, no setup | Creates commits | **Recommended** |
| GitHub Gist | No repo commits | Requires gist token/management | Not recommended |
| Release asset | Clean separation | Only on releases | Not applicable |
| GitHub Pages | CDN-cached | Requires setup | Over-engineered |

**Decision:** Use `.github/pmu-cache.json` - simple, works with existing permissions, git-tracked.

---

## Workflow Considerations

### Commit Noise

The workflow creates commits on every tracker change. Mitigation options:

1. **Squash on release**: Include cache commits in release prep squash
2. **Branch-based**: Update cache on a dedicated branch, merge periodically
3. **Accept it**: Cache commits are small and infrequent

**Recommendation:** Accept the commits. Tracker changes are infrequent (few per day max).

### Race Conditions

Multiple near-simultaneous issue events could cause conflicts.

**Mitigation:** Workflow uses `git diff --cached --quiet ||` to skip if no changes, and GitHub's built-in concurrency handling.

### Initial Setup

New repos need the workflow installed and initial manifest generated.

**Solution:** `gh pmu init` could install the workflow and trigger initial generation.

---

## Implementation Phases

### Phase 1: Framework (IDPF)
- [ ] Create workflow template in IDPF repository
- [ ] Document workflow installation in framework docs
- [ ] Add workflow to IDPF project templates

### Phase 2: gh-pmu Support
- [ ] Add manifest reading to `release list`
- [ ] Add manifest reading to `microsprint list`
- [ ] Add `cache` config section to `.gh-pmu.yml`
- [ ] Add `--refresh` flag to force API fetch
- [ ] Update `gh pmu init` to install workflow (optional)

### Phase 3: Documentation
- [ ] Update gh-pmu docs with cache configuration
- [ ] Document workflow in IDPF framework docs
- [ ] Add troubleshooting guide for cache issues

---

## Acceptance Criteria

- [ ] Workflow triggers on release/microsprint label events
- [ ] Manifest contains all tracker issues (open and closed)
- [ ] `gh pmu release list` reads from manifest when available
- [ ] `gh pmu microsprint list` reads from manifest when available
- [ ] `--refresh` flag bypasses cache and fetches from API
- [ ] Graceful fallback when manifest missing or stale
- [ ] Performance: <100ms for cached reads

---

## Alternatives Considered

### 1. Local Cache in `.gh-pmu.yml`
Update local config on `start`/`close`/`reopen` commands.

**Rejected:** Only works for single-user workflows. Multi-user teams would have stale local caches.

### 2. Background Subprocess Refresh
Spawn detached process to refresh cache after returning cached data.

**Rejected:** Adds complexity, still requires local state, doesn't solve multi-user.

### 3. API Response Caching (HTTP-level)
Cache GraphQL responses locally with TTL.

**Rejected:** Still requires initial API call, doesn't help first-run performance.

---

## Open Questions

1. Should `gh pmu init` auto-install the workflow, or keep it opt-in?
2. Should the manifest include more data (e.g., issue counts per release)?
3. Should there be a `gh pmu cache` command group for manual management?

---

## References

- Issue #434: Enhancement: Cache release/microsprint list data locally
- Current implementation: `cmd/release.go:1065`, `cmd/microsprint.go:1336`
