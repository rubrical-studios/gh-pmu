---
name: tdd-failure-recovery
version: 0.16.0
description: Handle unexpected test behaviors and recover from TDD issues
---
# TDD Failure Recovery
## When to Use
- Test behaves unexpectedly in any TDD phase
- "Rollback-Previous-Step" command
- Need to recover from TDD cycle issues

## Failure Scenarios
### RED Phase Failures
| Problem | Diagnosis | Recovery |
|---------|-----------|----------|
| Test passes immediately | Wrong test or feature exists | Verify feature doesn't exist; revise test |
| Syntax error | Test code buggy | Fix imports/syntax; retry |
| Wrong failure message | Assertion incorrect | Review assertion; fix test |

### GREEN Phase Failures
| Problem | Diagnosis | Recovery |
|---------|-----------|----------|
| Test still fails | Implementation incomplete | Analyze failure; add missing logic |
| Different error | Wrong approach | Review test requirement; try different impl |
| Other tests break | Regression | Fix without breaking new test |

### REFACTOR Phase Failures
| Problem | Diagnosis | Recovery |
|---------|-----------|----------|
| Test fails after refactor | Refactoring broke behavior | ROLLBACK immediately |
| Multiple tests fail | Structural change broke things | Revert; smaller changes |

## Recovery Commands
### Rollback Protocol
```
TASK: Rollback to last known good state
STEP 1: git status (check what changed)
STEP 2: git diff (review changes)
STEP 3: git checkout -- <files> (revert)
STEP 4: Run tests
STEP 5: Verify green restored
STEP 6: Identify smaller safe change
```

### Test Isolation
```
TASK: Isolate failing test
STEP 1: Run only failing test
STEP 2: Add debug output
STEP 3: Identify exact failure point
STEP 4: Fix specific issue
STEP 5: Run full suite
```

## Diagnostic Questions
**For any unexpected behavior:**
1. What phase were we in?
2. What was expected?
3. What actually happened?
4. What changed since last green?

## Prevention
- Run tests frequently
- Small incremental changes
- Commit at each green state
- Don't skip verification steps

## Flows To
- Back to current TDD phase after recovery
- `tdd-red-phase` if starting fresh iteration
