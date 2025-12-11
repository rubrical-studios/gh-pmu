---
name: tdd-green-phase
version: 1.0.0
description: Guide through GREEN phase - minimal implementation to pass test
---
# TDD GREEN Phase
## When to Use
- After RED phase success
- User reports "Done-Next-Step" with verified failing test
- Test exists and fails correctly

## GREEN Phase Goal
**Write minimum code to make the test pass.**

## Core Principle: YAGNI
"You Aren't Gonna Need It" - Implement only what the test requires.
```
✓ Return hardcoded value if test only checks one case
✓ Add minimal logic to pass assertion
✗ Add error handling not tested
✗ Add features not required by test
✗ Optimize before test passes
```

## Workflow
### Step 1: Analyze Failing Test
- What assertion is failing?
- What's the minimum to pass?

### Step 2: Write Minimal Implementation
**Single Code Block Format:**
```
TASK: Implement minimal code to pass test
STEP 1: Open implementation file
STEP 2: Write minimum code to satisfy test
STEP 3: Save file
STEP 4: Run test command
STEP 5: Verify test PASSES
STEP 6: Report: Does test pass now?
```

### Step 3: Execute and Verify
- [ ] Implementation compiled/ran without errors
- [ ] Test NOW PASSES
- [ ] No other tests broken

### Step 4: Analyze Result
**Test passes:** GREEN complete → "Done-Next-Step" → REFACTOR
**Test still fails:** Analyze → Fix implementation → Retry
**Other tests break:** Fix without breaking new test

## Implementation Strategies
| Test Case | Minimum Implementation |
|-----------|----------------------|
| Single input/output | Return constant |
| Multiple inputs | Simple conditional |
| Pattern | Minimal algorithm |

## Best Practices
- Don't add functionality not tested
- Don't optimize yet
- Don't refactor yet
- Keep implementation simple

## Anti-Patterns
❌ Adding "obvious" features not tested
❌ Error handling for cases not tested
❌ Refactoring during GREEN phase
❌ Breaking existing tests

## Flows To
- `tdd-refactor-phase` - Next phase after GREEN success
- `tdd-failure-recovery` - If tests behave unexpectedly
