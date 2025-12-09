# Sprint 6 Backlog

**Sprint Goal:** Release readiness - build and distribute gh-pmu as installable extension

**Sprint Duration:** TBD
**Total Story Points:** 5

---

## Selected Stories

### Tech Story: Release Build & Distribution
**Points:** 5 | **Priority:** High

**Description:** Configure GoReleaser and GitHub Actions to build and publish the extension as installable releases.

**Benefit:** Users can install gh-pmu via `gh extension install` from published releases.

**Acceptance Criteria:**
- [ ] `.goreleaser.yml` configured for gh extension format
- [ ] GitHub Actions workflow builds on tag push
- [ ] Produces binaries for linux/darwin/windows (amd64/arm64)
- [ ] Release includes checksums and changelog
- [ ] `gh extension install rubrical-studios/gh-pmu` works from release
- [ ] README updated with installation instructions

**Technical Notes:**
- Use goreleaser archive format compatible with gh extension
- Binary must be named `gh-pmu` (matches extension name)
- Consider using `gh extension create` scaffold as reference
- Test installation locally before publishing

---

## Sprint Metrics

| Metric | Value |
|--------|-------|
| Total Points | 5 |
| Stories | 1 |
| High Priority | 1 |

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Release workflow tested with dry-run
- [ ] Installation verified on at least one platform
- [ ] README includes install/upgrade/uninstall instructions

## Risks

1. **GoReleaser configuration** - May need iteration to get archive format correct for gh extension
2. **Cross-platform testing** - Limited ability to test all OS/arch combinations
