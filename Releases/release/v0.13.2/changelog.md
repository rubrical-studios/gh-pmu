# Changelog for v0.13.2

## Changed

- Renamed project text field from "Release" to "Branch" in defaults (#603)
- Updated all command help text to reference "Branch" field
- Added backward compatibility for existing projects with "Release" field

## Technical Details

### Files Modified
- `internal/defaults/defaults.yml` - Field name change
- `cmd/fieldnames.go` - New file with field name constants and resolution
- `cmd/branch.go` - Updated field references and help text
- `cmd/move.go` - Updated field references
- `cmd/create.go` - Updated field references and interface
- `cmd/list.go` - Added dual-field filtering
- `cmd/validation.go` - Updated context builder

### Migration Notes
No migration required. The change is backward compatible:
1. Commands check for "Branch" field first
2. Fall back to "Release" field for existing projects
3. New projects get "Branch" field from init
