# Release v0.13.2

## Summary

This patch release renames the project text field from "Release" to "Branch" to align with the updated branch semantics where any branch (not just release branches) can be used for tracking work. Full backward compatibility is maintained for existing projects.

## What's New

### Field Renamed: Release to Branch (#603)

The project field used for tracking branch assignments has been renamed from "Release" to "Branch". This aligns with the previously deprecated `--release` flag being replaced by `--branch`.

**Backward Compatibility:**
- Existing projects with a "Release" field continue to work
- New projects created with `gh pmu init` get a "Branch" field
- Commands automatically detect which field exists and use it

**No action required** for existing users - the change is transparent.

### Updated Help Text

All command help text and documentation now references "Branch" instead of "Release" for consistency.

## Installation

```bash
gh extension upgrade rubrical-studios/gh-pmu
```

## Full Changelog

https://github.com/rubrical-studios/gh-pmu/compare/v0.13.1...v0.13.2
