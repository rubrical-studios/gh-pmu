---
name: tdd-failure-recovery
description: Guide experienced developers through TDD failure scenarios and recovery procedures when tests behave unexpectedly
version: "0.38.0"
license: Complete terms in LICENSE.txt
---

# TDD Failure Recovery

This Skill guides experienced developers through diagnosing and recovering from unexpected test behaviors in the TDD cycle.

## When to Use This Skill

Invoke this Skill when:
- RED phase test passes unexpectedly (should fail)
- GREEN phase test still fails (should pass)
- REFACTOR phase breaks tests (should stay green)
- Tests behave unpredictably or inconsistently
- Need to rollback to previous working state
- Need diagnostic guidance for test failures

## Prerequisites

- Understanding of expected TDD cycle behavior
- Knowledge of current TDD phase (RED/GREEN/REFACTOR)
- Test execution capability via Claude Code
- Version control or ability to undo changes

## Failure Scenarios and Recovery

### Scenario 1: RED Phase Test Passes Unexpectedly

**Expected:** Test should FAIL because feature not implemented
**Actual:** Test PASSES immediately

**From frameworks:**
> "RED phase test passes unexpectedly: The test is invalid; ASSISTANT revises the test"

#### Diagnosis

**Possible causes:**

1. **Feature already exists**
   - Code from previous iteration still present
   - Similar feature implemented earlier
   - Accidentally wrote implementation before test

2. **Test is too permissive**
   - Assertion too broad (always passes)
   - Testing wrong thing
   - Mock/stub returns success by default

3. **Test setup incorrect**
   - Not actually exercising the code
   - Bypassing actual implementation
   - Testing test data instead of real code

#### Recovery Steps

**Step 1: Verify test is executing**
- Add intentional failure to test
- Confirm test can fail
- Remove intentional failure

**Step 2: Check for existing implementation**
- Search codebase for feature
- If found: Delete implementation OR write test for different behavior
- Re-run test, verify it fails

**Step 3: Review test logic**
- Is assertion correct?
- Is test calling actual implementation?
- Does test accurately represent requirement?

**Step 4: Revise test**
- ASSISTANT provides corrected test
- User executes via Claude Code
- Verify test now fails as expected

**Step 5: Resume TDD cycle**
- Proceed autonomously to GREEN phase

### Scenario 2: GREEN Phase Test Still Fails

**Expected:** Implementation should make test PASS
**Actual:** Test still FAILS after implementation

**From frameworks:**
> "GREEN phase test fails: Implementation is incomplete; ASSISTANT revises the code"

#### Diagnosis

**Possible causes:**

1. **Implementation incomplete**
   - Missing edge case handling
   - Incorrect logic
   - Wrong return value/behavior

2. **Implementation has bugs**
   - Syntax errors
   - Logic errors
   - Type mismatches

3. **Test expectations misunderstood**
   - Implemented wrong behavior
   - Misread test requirements
   - Partial implementation

4. **Environmental issues**
   - Dependencies not available
   - Database/file system issues
   - Configuration problems

#### Recovery Steps

**Step 1: Read failure message carefully**
- What does test expect?
- What did implementation provide?
- What's the specific mismatch?

**Step 2: Verify implementation**
- Does code execute without errors?
- Does code match test requirements?
- Are there syntax/logic errors?

**Step 3: Check test requirements**
- Re-read test assertions
- Understand exact expectations
- Verify test data and setup

**Step 4: Revise implementation**
- ASSISTANT provides corrected implementation
- User executes via Claude Code
- Verify test now passes

**Step 5: Run full test suite**
- Ensure no regressions
- All tests should be green
- If any fail, address before proceeding

**Step 6: Resume TDD cycle**
- Proceed autonomously to REFACTOR phase

### Scenario 3: REFACTOR Phase Breaks Tests

**Expected:** Refactoring should keep all tests GREEN
**Actual:** One or more tests FAIL after refactoring

**From frameworks:**
> "Refactoring breaks tests: Roll back refactoring; tests must stay green"

#### Diagnosis

**Possible causes:**

1. **Behavioral change introduced**
   - Refactoring accidentally changed logic
   - Different code path taken
   - Side effects changed

2. **Breaking change in API**
   - Function signature changed
   - Return type modified
   - Expected interface violated

3. **Incomplete refactoring**
   - Updated some call sites, missed others
   - Renamed in some places, not all
   - Partial extraction left inconsistency

4. **Test dependency on implementation**
   - Test was coupled to specific implementation
   - Refactoring changed what test relied on
   - Test needs updating (smell: test too brittle)

#### Recovery Steps

**Step 1: IMMEDIATE ROLLBACK**
- Undo refactoring changes
- Return to last green state
- Verify tests are green again

**Critical principle:**
```
TESTS MUST STAY GREEN
If refactoring breaks tests → ROLLBACK
Do not proceed with broken tests
```

**Step 2: Analyze what broke**
- Which test(s) failed?
- What was the failure message?
- What specific refactoring caused it?

**Step 3: Decide next action**

**Option A: Skip refactoring**
- Refactoring too risky
- Tests remain green
- Defer improvement to later

**Option B: Smaller refactoring**
- Break refactoring into smaller steps
- Apply minimal change
- Test after each micro-step
- Stop at first failure

**Option C: Fix test (if test is brittle)**
- Test was over-coupled to implementation
- Update test to test behavior, not implementation
- Re-apply refactoring
- Verify tests pass

**Step 4: Resume TDD cycle**
- If refactoring skipped or completed successfully
- Proceed autonomously to next behavior or complete story

### Scenario 4: Rollback to Previous State

When tests break unexpectedly or a change makes things worse, rollback to the previous working state.

#### When to Use Rollback

**Valid rollback scenarios:**
- Refactoring broke tests
- Implementation made things worse
- Went down wrong path
- Need to return to known good state

#### Rollback Procedure

**ASSISTANT provides to user in single code block:**

```
TASK: Rollback to previous working state

STEP 1: [Identify changes to undo]
STEP 2: [Restore previous code version]
STEP 3: [Verify file state matches pre-change]
STEP 4: [Run full test suite]
STEP 5: [Verify all tests GREEN]
STEP 6: [Report back: Tests green and rollback complete?]
```

**After rollback:**
- All tests should be green
- Code returned to last known good state
- Ready to try different approach
- User decides next step (retry or skip)

### Scenario 5: Inconsistent Test Results

**Problem:** Tests pass sometimes, fail other times

#### Diagnosis

**Possible causes:**

1. **Test order dependency**
   - Test relies on another test running first
   - Shared state between tests
   - Improper test isolation

2. **Timing issues**
   - Race conditions
   - Async operations not properly awaited
   - Timeouts too short

3. **External dependencies**
   - Database state varies
   - File system changes
   - Network/API calls unreliable

4. **Random data in tests**
   - Test uses random values
   - Different data causes different outcomes
   - Non-deterministic behavior

#### Recovery Steps

**Step 1: Isolate the test**
- Run failing test alone
- Run in different order
- Identify if order-dependent

**Step 2: Check test isolation**
- Does test clean up after itself?
- Does test set up its own data?
- Does test depend on external state?

**Step 3: Fix test isolation**
- Add proper setup/teardown
- Use fixtures for test data
- Mock external dependencies
- Ensure deterministic behavior

**Step 4: Verify consistency**
- Run test multiple times
- Run full suite multiple times
- Confirm consistent pass/fail

## Diagnostic Flowchart

### When Test Fails Unexpectedly

```
Test failed unexpectedly
    ↓
What phase?
    ↓
┌───────────┬────────────┬─────────────┐
RED         GREEN        REFACTOR
↓           ↓            ↓
Should      Should       Should
fail        pass         stay green
↓           ↓            ↓
But         But          But
passes      fails        fails
↓           ↓            ↓
Test        Impl.        ROLLBACK
invalid     incomplete   immediately
↓           ↓            ↓
Revise      Revise       Try smaller
test        impl.        or skip
```

## Prevention Strategies

### Strategy 1: Verify Each Phase

**RED phase:**
- Always run test and verify it fails
- Never assume test will fail
- Check failure message is correct

**GREEN phase:**
- Run test and verify it passes
- Run full suite to check regressions
- Don't skip verification step

**REFACTOR phase:**
- Run full suite after every change
- Small steps, test after each
- Rollback immediately if any test fails

### Strategy 2: Clear Communication

**User → ASSISTANT:**
- Report exact test results
- Include failure messages
- Describe unexpected behavior

**ASSISTANT → User:**
- Clear recovery instructions
- Single code block format
- Explicit verification steps

### Strategy 3: Maintain Green State

**Golden rule:**
```
Tests should ALWAYS be green except during RED phase

RED phase: Intentionally red (one test)
GREEN phase: Return to all green
REFACTOR phase: Stay all green
Between features: All green

If not green when expected → STOP and recover
```

## Common Recovery Patterns

### Pattern 1: The Reset

```
Situation: Confused state, unclear what's wrong
Action: Rollback to last known green
Outcome: Clean slate, try again
```

### Pattern 2: The Minimal Fix

```
Situation: Small issue, clear fix
Action: Targeted correction
Outcome: Tests green again, proceed
```

### Pattern 3: The Skip

```
Situation: Risk > reward for current change
Action: Skip problematic change
Outcome: Defer to later, maintain green
```

### Pattern 4: The Divide and Conquer

```
Situation: Large change broke something
Action: Break into smaller incremental changes
Outcome: Identify exactly what breaks
```

## Resources

See `resources/` directory for:
- `failure-diagnostic-flowchart.md` - Visual decision tree
- `recovery-procedures.md` - Step-by-step recovery for each scenario
- `test-isolation-guide.md` - Ensuring test independence

## Relationship to Other Skills

**Supports all phases:**
- `tdd-red-phase` - RED phase failures
- `tdd-green-phase` - GREEN phase failures
- `tdd-refactor-phase` - REFACTOR phase failures

**Related skills:**
- `test-writing-patterns` - Avoiding test-related failures

## Expected Outcome

After successful failure recovery:
- Tests returned to expected state (all green)
- Understanding of what went wrong
- Correction applied or change reverted
- Ready to resume TDD cycle
- Lessons learned for future prevention

---

**End of TDD Failure Recovery Skill**
