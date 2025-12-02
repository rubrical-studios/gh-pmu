# Product Requirements Documents (PRD)

This directory contains product requirements documents for this project.

## Structure

```
PRD/
+-- PRD-[ProjectName].md    # Main requirements document
+-- Templates/              # PRD templates (from framework)
+-- Specs/                  # ATDD/BDD specifications (if using)
    +-- Templates/          # Spec templates
```

## Getting Started

1. Copy the appropriate template from `Templates/`:
   - `PRD-Structured-Comprehensive.md` - Full requirements coverage
   - `PRD-Structured-Moderate.md` - Key sections only
   - `PRD-Agile-Lightweight.md` - Epic/Story input

2. Rename to `PRD-[YourProjectName].md`

3. Fill in the requirements

## Testing Approach

See `Templates/Testing-Approach-Selection-Guide.md` for guidance on:
- TDD (required for development)
- ATDD (optional, for acceptance criteria)
- BDD (optional, for behavior specifications)
