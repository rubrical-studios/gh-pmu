---
name: tdd-red-phase
description: Guide experienced developers through RED phase of TDD cycle - writing failing tests and verifying expected failures
license: Complete terms in LICENSE.txt
---

# TDD RED Phase

This Skill guides experienced developers through the RED phase of the Test-Driven Development cycle: writing failing tests and verifying expected failures.

## When to Use This Skill

Invoke this Skill when:
- Starting implementation of a new feature or behavior
- User issues "First-Step" command in IDPF-Structured framework
- User issues "Start-Story [ID]" command in IDPF-Agile framework
- Beginning a new TDD iteration
- User needs guidance on writing a failing test

## Prerequisites

- Working development environment with test framework configured
- Clear understanding of the behavior to be tested
- Familiarity with TDD principles
- Claude Code available for test execution

## RED Phase Objectives

The RED phase has one critical goal: **Write a test that fails for the right reason.**

### What "Right Reason" Means

**✓ Correct failure:**
- Test fails because feature doesn't exist yet
- Test fails because behavior not implemented
- Failure message clearly indicates what's missing

**✗ Incorrect failure:**
- Test fails due to syntax error
- Test fails due to missing imports
- Test fails due to incorrect test setup
- Test passes unexpectedly (test is invalid)

## RED Phase Workflow

### Step 1: Identify Testable Behavior

**From frameworks:**
- IDPF-Structured: "Smallest task possible (typically one test case covering one behavior)"
- IDPF-Agile: "Decompose story into testable behaviors"

**Key principle:** One test per behavior, one behavior per test.

**Good behavior identification:**
```
✓ "Function returns sum of two numbers"
✓ "GET /users returns 200 status"
✓ "Invalid email shows validation error"
✓ "User can delete their own posts"
```

**Poor behavior identification:**
```
✗ "User management works" (too broad)
✗ "Fix the bug" (not a behavior)
✗ "Improve performance" (not testable as pass/fail)
```

### Step 2: Write the Failing Test

**Test structure (framework-agnostic):**

```
1. ARRANGE: Set up test data and preconditions
2. ACT: Execute the behavior being tested
3. ASSERT: Verify the expected outcome
```

**ASSISTANT provides to user in single code block:**

1. Clear description of behavior being tested
2. Complete test code (imports, setup, execution, assertion)
3. Test execution command
4. Expected failure message

**Critical: Single Code Block Format**

All instructions must be in ONE code block using STEP format:

```
TASK: [Brief description of test being written]

STEP 1: [Create or open test file]
STEP 2: [Add necessary imports]
STEP 3: [Write complete test function with proper structure]
STEP 4: [Save file]
STEP 5: [Run test command]
STEP 6: [Verify test FAILS with expected message]
STEP 7: [Report back: Did test fail as expected?]
```

### Step 3: Execute and Verify Failure

**User action:**
- Copy complete code block to Claude Code
- Claude Code executes test
- User verifies test FAILS (not passes, not errors)

**Verification checklist:**
- [ ] Test executed without syntax errors
- [ ] Test failed (not passed)
- [ ] Failure message indicates missing implementation
- [ ] Failure message is clear and understandable

### Step 4: Analyze Failure

**If test fails as expected:**
→ RED phase complete
→ User responds "Done-Next-Step" to proceed to GREEN phase

**If test passes unexpectedly:**
→ Test is invalid (probably testing wrong thing)
→ ASSISTANT revises test
→ Repeat Step 2

**If test errors instead of fails:**
→ Test has bugs (syntax, imports, setup)
→ ASSISTANT fixes test code
→ Repeat Step 2

## RED Phase Best Practices

### Write Minimal Tests

**Good (minimal):**
```
Test verifies one specific behavior
Uses simplest possible test data
Single assertion (or closely related assertions)
```

**Poor (excessive):**
```
Test verifies multiple unrelated behaviors
Complex test data setup
Multiple unrelated assertions
```

### Use Clear Test Names

**Framework-agnostic naming conventions:**

```
test_[feature]_[scenario]_[expected_result]

Examples:
- test_add_function_with_positive_numbers_returns_sum
- test_get_users_when_authenticated_returns_200
- test_create_post_with_invalid_data_returns_validation_error
```

### Write Descriptive Assertions

**Good assertions:**
```
Assert with clear comparison
Include helpful failure messages
Test observable behavior, not implementation
```

**Poor assertions:**
```
Assert True without context
No failure messages
Test internal state instead of behavior
```

## Common RED Phase Mistakes

### Mistake 1: Test Passes Immediately

**Problem:** Wrote test for already-implemented feature

**Solution:**
- Verify feature doesn't exist yet
- If exists, delete implementation first
- Re-run test to confirm failure

### Mistake 2: Test Has Syntax Errors

**Problem:** Test won't run due to code errors

**Solution:**
- Fix syntax errors
- Ensure proper imports
- Verify test framework setup
- Run test again

### Mistake 3: Test Too Broad

**Problem:** Test checks multiple behaviors at once

**Solution:**
- Split into multiple tests
- One test per behavior
- Keep tests focused and simple

### Mistake 4: Unclear Failure Message

**Problem:** When test fails, message doesn't explain what's missing

**Solution:**
- Add descriptive assertion messages
- Use clear variable names
- Test one thing at a time

## RED Phase Anti-Patterns

### Anti-Pattern 1: Writing Implementation First

```
✗ WRONG:
1. Write code
2. Write test to verify code
3. Test passes immediately

✓ CORRECT (TDD):
1. Write failing test
2. Write code to pass test
3. Test now passes
```

### Anti-Pattern 2: Skipping Failure Verification

```
✗ WRONG:
1. Write test
2. Assume it will fail
3. Move to GREEN phase

✓ CORRECT:
1. Write test
2. RUN test and VERIFY it fails
3. Confirm failure reason is correct
4. Then move to GREEN phase
```

### Anti-Pattern 3: Tolerating Test Errors

```
✗ WRONG:
Test throws exception → "I'll fix it in GREEN phase"

✓ CORRECT:
Test throws exception → "Fix test now so it fails cleanly"
```

## Integration with Frameworks

### IDPF-Structured

RED phase is Step 1 of TDD iteration:

```
User: "First-Step"
ASSISTANT: Provides RED phase test in single code block
User: Executes via Claude Code
User: Verifies failure
User: "Done-Next-Step" → Proceeds to GREEN phase
```

### IDPF-Agile

RED phase is first step of story implementation:

```
User: "Start-Story [ID]"
ASSISTANT: Breaks story into testable behaviors
ASSISTANT: Provides RED phase test for first behavior
User: Executes via Claude Code
User: Verifies failure
User: "Done-Next-Step" → Proceeds to GREEN phase
```

## RED Phase Checklist

Before proceeding to GREEN phase, verify:

- [ ] Test code is complete and syntactically correct
- [ ] Test executes without errors
- [ ] Test FAILS (does not pass)
- [ ] Failure message clearly indicates missing implementation
- [ ] Test name clearly describes behavior being tested
- [ ] Test is focused on single behavior
- [ ] Test uses minimal, clear test data
- [ ] User has reported failure verification via "Done-Next-Step"

## Resources

See `resources/` directory for:
- `red-phase-checklist.md` - Quick reference checklist
- `test-structure-patterns.md` - Common test structure patterns
- `failure-verification-guide.md` - How to verify test failures correctly

## Relationship to Other Skills

**Flows to:**
- `tdd-green-phase` - Next phase after RED phase success

**Related skills:**
- `test-writing-patterns` - Detailed test structure guidance
- `tdd-failure-recovery` - Handle unexpected RED phase results

**Independent from:**
- `beginner-testing` - This skill assumes TDD experience

## Expected Outcome

After successful RED phase:
- One failing test exists
- Failure is verified and understood
- Test clearly defines expected behavior
- Ready to implement minimum code to pass test (GREEN phase)
- User has communicated readiness via "Done-Next-Step"

---

**End of TDD RED Phase Skill**
