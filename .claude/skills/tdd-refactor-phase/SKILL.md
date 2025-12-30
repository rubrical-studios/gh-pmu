---
name: tdd-refactor-phase
version: 1.0.0
description: Guide through REFACTOR phase - improve code while maintaining green
---
# TDD REFACTOR Phase
## When to Use
- After GREEN phase success
- Test passes, implementation works
- Code quality needs improvement

## REFACTOR Phase Goal
**Improve code quality while keeping tests green.**

## Refactoring Assessment
### When to Refactor
- Code duplication exists
- Functions too long (>20 lines)
- Poor naming
- Obvious improvements
- Code smells detected

### When to Skip
- Code is already clean
- Small, simple implementation
- User wants to move to next feature

## Refactoring Categories
| Category | Examples |
|----------|----------|
| **Naming** | Variables, functions, classes |
| **Structure** | Extract method, inline function |
| **Duplication** | DRY violations |
| **Simplification** | Complex conditionals, magic numbers |

## Workflow
### Step 1: Identify Improvements
- Run tests to establish green baseline
- Identify refactoring opportunities
- Prioritize by impact

### Step 2: Apply Small Changes
**Single Code Block Format:**
```
TASK: Refactor [specific improvement]
STEP 1: Open file
STEP 2: Apply specific refactoring
STEP 3: Save file
STEP 4: Run ALL tests
STEP 5: Verify ALL tests pass
STEP 6: Report: Do all tests still pass?
```

### Step 3: Verify Green Maintained
- [ ] Specific refactoring applied
- [ ] ALL tests still pass
- [ ] No new functionality added

### Step 4: Iterate or Complete
**Tests pass:** Continue or → "Done-Next-Step" → Next RED
**Tests fail:** ROLLBACK immediately → Fix → Retry

## Rollback Protocol
If tests fail after refactoring:
```
TASK: Rollback failed refactoring
STEP 1: Revert changes (git checkout or manual)
STEP 2: Run tests
STEP 3: Verify green restored
STEP 4: Smaller refactoring attempt
```

## Best Practices
- Small, incremental changes
- Run tests after EACH change
- Don't add functionality
- Keep commits atomic

## Anti-Patterns
❌ Large refactoring without test runs
❌ Adding features during refactor
❌ Ignoring test failures
❌ Skipping rollback when tests break

## Flows To
- Back to `tdd-red-phase` - Next feature iteration
- `tdd-failure-recovery` - If unexpected test behavior
