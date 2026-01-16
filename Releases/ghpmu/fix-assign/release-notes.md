# Release v0.12.1

**Release Date:** 2026-01-16
**Branch:** ghpmu/fix-assign
**Tracker:** #552

## Summary

Patch release fixing a critical bug in the batch mutation system that caused HTTP 400 errors when using `gh pmu move` with `--branch` or `--status` flags.

## Fixed

- **Batch mutation HTTP 400 error** - The `executeBatchMutation()` function was generating invalid JSON for GraphQL requests. The value and input objects used GraphQL literal syntax (unquoted keys) instead of valid JSON required by the `--input` flag. Additionally, the request format incorrectly mixed `-f query=` with `--input -`. Fixed by using proper JSON format and restructuring to pass the complete GraphQL request body via stdin. (#551)

## Upgrade Notes

This is a patch release with no breaking changes. Upgrade recommended for all users experiencing batch mutation failures.

## Contributors

- Claude Opus 4.5 (Co-Author)
