# Project Charter: gh-pmu

**Status:** Active
**Last Updated:** 2026-01-04

## Vision

A GitHub CLI extension that streamlines project workflows by unifying issue tracking, sub-issue hierarchy, and workflow automation into a single cohesive tool.

## Current Focus

v0.10.0 - Enhanced body editing with `--body-stdout` and `--body-stdin` flags

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.22 |
| Framework | Cobra CLI |
| API | GitHub GraphQL (go-gh, shurcooL-graphql) |

## In Scope (Current)

- Project field management (status, priority, custom fields)
- Sub-issue hierarchy with progress tracking
- Batch operations (intake, triage, split)
- Workflow automation (microsprint, release)
- Terminal Kanban board visualization

---
*See Inception/ for full specifications*
