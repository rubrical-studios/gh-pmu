# Proposal: Release Close Next Steps

**Date:** 2025-12-23
**Status:** Draft
**Author:** Claude (via user request)

---

## Executive Summary

After `gh pmu release close` completes, output a "Next steps" section with the `gh release create` command to help users create the GitHub Release.

This bridges the gap between gh-pmu's release planning workflow and GitHub's release publishing, without adding scope to gh-pmu itself.

---

## Current State

Currently, `gh pmu release close` outputs:

```
Closing release: release/v1.0.0
  Tracker issue: #350
  Issues in release: 12 (12 done, 0 incomplete)

Proceed? (y/n): y

✓ Release closed: v1.0.0
✓ Artifacts created in: Releases/v1.0.0
✓ Tag created: v1.0.0
```

Users must then separately run `gh release create` — which is easy to forget or get wrong.

---

## Proposed Change

Add a "Next steps" section to the output when `--tag` is used:

```
✓ Release closed: v1.0.0
✓ Artifacts created in: Releases/v1.0.0
✓ Tag created: v1.0.0

Next steps:
  git push --tags
  gh release create v1.0.0 --notes-file Releases/v1.0.0/release-notes.md
```

### Conditional Output

| Scenario | Output |
|----------|--------|
| `--tag` used | Show `git push --tags` + `gh release create` |
| No `--tag` | Show nothing (user didn't want a tag) |
| Tag already pushed | Show only `gh release create` (detect via `git ls-remote`) |

### Release Notes Path

Use the actual path where release notes were generated:

```go
notesPath := fmt.Sprintf("%s/%s/release-notes.md", cfg.GetArtifactDirectory(), releaseVersion)
```

If release notes generation is disabled in config, omit `--notes-file`:

```
gh release create v1.0.0 --generate-notes
```

---

## Implementation

### Location

Add to `runReleaseCloseWithDeps()` in `cmd/release.go`, after the existing success output.

### Code

```go
// Output next steps for GitHub Release (only when tag was created)
if opts.tag {
    fmt.Fprintln(cmd.OutOrStdout())
    fmt.Fprintln(cmd.OutOrStdout(), "Next steps:")

    // Check if tag is already pushed
    tagPushed := isTagPushed(releaseVersion)
    if !tagPushed {
        fmt.Fprintln(cmd.OutOrStdout(), "  git push --tags")
    }

    // Build gh release create command
    if cfg.ShouldGenerateReleaseNotes() {
        notesPath := fmt.Sprintf("%s/%s/release-notes.md", cfg.GetArtifactDirectory(), releaseVersion)
        fmt.Fprintf(cmd.OutOrStdout(), "  gh release create %s --notes-file %s\n", releaseVersion, notesPath)
    } else {
        fmt.Fprintf(cmd.OutOrStdout(), "  gh release create %s --generate-notes\n", releaseVersion)
    }
}
```

### Helper Function

```go
// isTagPushed checks if a tag exists on the remote
func isTagPushed(tag string) bool {
    cmd := exec.Command("git", "ls-remote", "--tags", "origin", tag)
    output, err := cmd.Output()
    if err != nil {
        return false
    }
    return len(output) > 0
}
```

---

## Examples

### Standard Close with Tag

```
$ gh pmu release close release/v1.0.0 --tag

Closing release: release/v1.0.0
  Tracker issue: #350
  Issues in release: 8 (8 done, 0 incomplete)

Proceed? (y/n): y

✓ Release closed: v1.0.0
✓ Artifacts created in: Releases/v1.0.0
✓ Tag created: v1.0.0

Next steps:
  git push --tags
  gh release create v1.0.0 --notes-file Releases/v1.0.0/release-notes.md
```

### Close Without Tag (No Next Steps)

```
$ gh pmu release close release/v1.0.0

Closing release: release/v1.0.0
  Tracker issue: #350
  Issues in release: 8 (8 done, 0 incomplete)

Proceed? (y/n): y

✓ Release closed: v1.0.0
✓ Artifacts created in: Releases/v1.0.0
```

### Custom Artifact Directory

```yaml
# .gh-pmu.yml
release:
  artifacts:
    directory: docs/releases
```

```
Next steps:
  git push --tags
  gh release create v1.0.0 --notes-file docs/releases/v1.0.0/release-notes.md
```

### Release Notes Disabled

```yaml
# .gh-pmu.yml
release:
  artifacts:
    release_notes: false
```

```
Next steps:
  git push --tags
  gh release create v1.0.0 --generate-notes
```

---

## Acceptance Criteria

- [ ] `gh pmu release close --tag` outputs "Next steps" section after success
- [ ] Shows `git push --tags` if tag not yet pushed to remote
- [ ] Shows `gh release create <version> --notes-file <path>` with correct path
- [ ] Uses `--generate-notes` if release notes generation is disabled
- [ ] No "Next steps" output when `--tag` is not used
- [ ] Path respects `release.artifacts.directory` config

---

## Testing

### Unit Tests

```go
func TestReleaseClose_NextStepsOutput(t *testing.T) {
    tests := []struct {
        name           string
        tagFlag        bool
        notesEnabled   bool
        expectNextStep bool
        expectNotesFile bool
    }{
        {"with tag and notes", true, true, true, true},
        {"with tag no notes", true, false, true, false},
        {"without tag", false, true, false, false},
    }
    // ...
}
```

### Manual Test

```bash
# Create and close a test release
gh pmu release start --branch release/v99.0.0-test
gh pmu release close release/v99.0.0-test --tag --yes

# Verify output includes next steps
# Clean up
git tag -d v99.0.0-test
```

---

## Alternatives Considered

### 1. Auto-create GitHub Release

**Rejected:** Would duplicate `gh release` functionality and require handling binary uploads, draft/prerelease flags, etc.

### 2. `--release` Flag on Close

```bash
gh pmu release close release/v1.0.0 --tag --release
```

**Rejected:** Still limited compared to `gh release create` options. Users need full control.

### 3. No Change

**Rejected:** User requested this improvement. Guidance without automation is the right balance.

---

## Decision

- [ ] Approved - Proceed to implementation
- [ ] Approved with modifications
- [ ] Rejected
- [ ] Needs more information

---

## References

- Related: `cmd/release.go` - existing release close implementation
- Inspiration: User feedback on CI/CD workflow integration
