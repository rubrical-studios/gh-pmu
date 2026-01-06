# Milestones: gh-pmu

**Last Updated:** 2026-01-04

---

## Current Focus

**Target:** v0.10.0 (Released)
**Status:** Complete
**Target Date:** 2026-01-03

### Objectives

- [x] Add `--body-stdout` flag for streamlined body export
- [x] Add `--body-stdin` flag for piped body input
- [x] Improve edit workflow for AI-assisted development

---

## Milestone Roadmap

| Milestone | Target Date | Key Deliverables | Status |
|-----------|-------------|------------------|--------|
| v0.10.0 | 2026-01-03 | Body editing enhancements | Complete |
| v0.9.7 | Ongoing | Patch fixes, stability | In Progress |
| v0.11.0 | TBD | Based on backlog priority | Planning |

---

## Milestone Details

### v0.10.0 - Body Editing Enhancements

**Target:** 2026-01-03
**Status:** Complete

**Scope:**
- `--body-stdout` flag for `gh pmu view`
- `--body-stdin` flag for `gh pmu edit`
- Streamlined piping workflow

**Success Criteria:**
- [x] Flags implemented and tested
- [x] Documentation updated
- [x] CHANGELOG entry added

**Dependencies:**
- None

---

### v0.9.7 - Patch Track

**Target:** Ongoing
**Status:** In Progress

**Scope:**
- Bug fixes from user reports
- Stability improvements
- Documentation corrections

**Success Criteria:**
- [ ] All P0 bugs resolved
- [ ] No regressions

**Dependencies:**
- User feedback via GitHub Issues

---

## Release Planning

| Version | Type | Target | Key Features |
|---------|------|--------|--------------|
| v0.10.0 | Minor | Released | Body editing |
| v0.9.7 | Patch | In Progress | Bug fixes |
| v0.11.0 | Minor | TBD | Backlog-driven |

---

## Risks to Timeline

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| GitHub API changes | Low | High | Version pinning, tests |
| Breaking gh changes | Low | Medium | Pin go-gh version |
| Scope creep | Medium | Medium | Strict PR review |

---

*See also: Scope-Boundaries.md, Charter-Details.md*
