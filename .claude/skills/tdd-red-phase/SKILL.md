---
name: tdd-red-phase
version: 1.0.0
description: Guide through RED phase - writing failing tests and verifying failures
---
# TDD RED Phase
## When to Use
- Starting new feature implementation
- "First-Step" command (IDPF-Structured)
- "Start-Story [ID]" command (IDPF-Agile)
- Beginning TDD iteration

## RED Phase Goal
**Write a test that fails for the right reason.**
**Correct failure:** Feature doesn't exist, behavior not implemented
**Incorrect failure:** Syntax error, missing imports, test passes unexpectedly

## Workflow
### Step 1: Identify Testable Behavior
One test per behavior, one behavior per test.
```
✓ "Function returns sum of two numbers"
✓ "GET /users returns 200 status"
✗ "User management works" (too broad)
```

### Step 2: Write Failing Test
**Test structure (AAA):**
1. ARRANGE: Set up preconditions
2. ACT: Execute behavior
3. ASSERT: Verify outcome

**Single Code Block Format:**
```
TASK: [description]
STEP 1: Create/open test file
STEP 2: Add imports
STEP 3: Write complete test function
STEP 4: Save file
STEP 5: Run test command
STEP 6: Verify test FAILS with expected message
STEP 7: Report: Did test fail as expected?
```

### Step 3: Execute and Verify
- [ ] Test executed without syntax errors
- [ ] Test failed (not passed)
- [ ] Failure indicates missing implementation
- [ ] Failure message is clear

### Step 4: Analyze
**Fails as expected:** RED complete → "Done-Next-Step" → GREEN phase
**Passes unexpectedly:** Test is invalid → Revise test
**Errors instead of fails:** Fix test code → Retry

## Best Practices
- Write minimal tests (one assertion)
- Use clear test names: `test_[feature]_[scenario]_[expected]`
- Write descriptive assertions with failure messages

## Anti-Patterns
❌ Writing implementation first
❌ Skipping failure verification
❌ Tolerating test errors ("fix in GREEN")

## Integration
**IDPF-Structured:** User: "First-Step" → RED phase test → Verify failure → "Done-Next-Step"
**IDPF-Agile:** User: "Start-Story [ID]" → Break into behaviors → RED phase for first

## Flows To
- `tdd-green-phase` - Next phase after RED success
- `tdd-failure-recovery` - Handle unexpected results
