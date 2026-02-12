---
name: tdd-green-phase
description: Guide experienced developers through GREEN phase of TDD cycle - writing minimal implementation to pass failing tests
version: "0.38.0"
license: Complete terms in LICENSE.txt
---

# TDD GREEN Phase

This Skill guides experienced developers through the GREEN phase of the Test-Driven Development cycle: implementing the minimum code necessary to make a failing test pass.

## When to Use This Skill

Invoke this Skill when:
- RED phase test has been verified as failing
- Proceeding autonomously from RED phase
- Implementing feature to pass test
- Moving from RED to GREEN in TDD cycle

## Prerequisites

- Completed RED phase with verified failing test
- Clear understanding of what test expects
- Implementation environment ready
- Claude Code available for code execution and testing

## GREEN Phase Objectives

The GREEN phase has one critical goal: **Write the minimum code to make the test pass.**

### What "Minimum Code" Means

**✓ Correct approach:**
- Implements exactly what test requires
- No additional features beyond test scope
- Simplest solution that makes test green
- Avoids premature optimization

**✗ Incorrect approach:**
- Implements features not tested
- Over-engineers solution
- Adds "might need later" functionality
- Optimizes before it's necessary

## GREEN Phase Workflow

### Step 1: Understand Test Requirements

**Review the failing test to identify:**
- What behavior is expected?
- What inputs does test provide?
- What output/result does test expect?
- What edge cases does this test cover?

**Key principle:** Test defines the contract. Implementation must fulfill contract, nothing more.

### Step 2: Plan Minimal Implementation

**Before writing code, identify:**
- Minimum logic needed
- Required data structures
- Necessary dependencies
- Expected return values/side effects

**Avoid planning:**
- Features not in test
- Abstractions not yet needed
- Optimizations not required
- Error handling not tested

### Step 3: Implement to Pass Test

**ASSISTANT provides to user in single code block:**

1. Implementation code that makes test pass
2. Test execution command
3. Expected success message

**Critical: Single Code Block Format**

All instructions must be in ONE code block using STEP format:

```
TASK: [Brief description of implementation]

STEP 1: [Open/locate implementation file]
STEP 2: [Navigate to specific location]
STEP 3: [Add/modify implementation code - COMPLETE code block]
STEP 4: [Context about implementation choices]
STEP 5: [Save file]
STEP 6: [Run test command]
STEP 7: [Verify test PASSES]
STEP 8: [Report back: Did test pass?]
```

### Step 4: Execute and Verify Success

**User action:**
- Copy complete code block to Claude Code
- Claude Code implements and executes test
- User verifies test PASSES (green)

**Verification checklist:**
- [ ] Test executed without errors
- [ ] Test passed (green)
- [ ] No other tests broke (if running full suite)
- [ ] Implementation is minimal and clear

### Step 5: Analyze Success

**If test passes:**
→ GREEN phase complete
→ Proceed autonomously to REFACTOR phase

**If test still fails:**
→ Implementation incomplete or incorrect
→ ASSISTANT revises implementation
→ Repeat Step 3

**If test passes but other tests fail:**
→ Implementation broke existing functionality
→ ASSISTANT revises to fix regressions
→ Re-run all tests

## GREEN Phase Best Practices

### Principle 1: YAGNI (You Aren't Gonna Need It)

**Good (minimal):**
```
Implements exactly what test requires
Hard-coded values acceptable if test passes
Simple conditional logic
Direct implementation
```

**Poor (over-engineered):**
```
Implements untested features "just in case"
Complex abstractions for single use case
Premature optimization
Anticipatory design
```

### Principle 2: Simplest Thing That Works

**Example progression (illustrated conceptually):**

```
Test: Function should return sum of two numbers

WRONG (over-engineered):
- Generic calculation engine
- Configuration system
- Plugin architecture
- Complex error handling

RIGHT (minimal):
- Function takes two parameters
- Returns their sum
- Done
```

### Principle 3: Let Tests Drive Design

**Test tells you:**
- Function signature needed
- Parameters required
- Return type expected
- Behavior specification

**Implementation follows test:**
```
Test says: get_user(id) should return user object
Implementation: Function with that exact signature
No more, no less
```

### Principle 4: Hard-Code First, Generalize Later

**Acceptable GREEN phase patterns:**

```
Test expects specific output?
→ Return that specific output (hard-coded)
→ Generalize in future tests

Test expects list with one item?
→ Return list with that one item
→ Handle multiple items when tested

Test checks single condition?
→ Implement that condition only
→ Add more conditions when tested
```

## Common GREEN Phase Mistakes

### Mistake 1: Over-Implementation

**Problem:** Adding features not required by test

**Example:**
```
Test: Should add two numbers
WRONG: Build calculator with +, -, ×, ÷
RIGHT: Function that adds two numbers
```

**Solution:**
- Implement ONLY what test requires
- Trust that future tests will drive future features

### Mistake 2: Premature Abstraction

**Problem:** Creating abstractions before they're needed

**Example:**
```
Test: Store user in memory
WRONG: Build database abstraction layer
RIGHT: Store in variable/collection
```

**Solution:**
- Wait for second or third use case
- Rule of Three: abstract after third duplication

### Mistake 3: Ignoring Test Failure Details

**Problem:** Not reading what test actually expects

**Solution:**
- Read failure message carefully
- Understand exact expectation
- Implement precise requirement

### Mistake 4: Breaking Existing Tests

**Problem:** Making current test pass but breaking others

**Solution:**
- Run full test suite
- Ensure all tests remain green
- Fix regressions before proceeding

## GREEN Phase Anti-Patterns

### Anti-Pattern 1: Feature Creep

```
✗ WRONG:
Test: User can log in
Implementation: Login + password reset + 2FA + OAuth

✓ CORRECT:
Test: User can log in
Implementation: Basic login functionality only
```

### Anti-Pattern 2: Optimization Before Profiling

```
✗ WRONG:
Test: Function returns result
Implementation: Cached, memoized, async, optimized

✓ CORRECT:
Test: Function returns result
Implementation: Straightforward synchronous function
```

### Anti-Pattern 3: Copy-Paste Without Understanding

```
✗ WRONG:
Test fails → Copy random code from internet → Test passes

✓ CORRECT:
Test fails → Understand requirement → Implement minimal solution
```

## Implementation Strategies

### Strategy 1: Fake It (Temporarily)

**When:** Very simple test, clear path to generalization

```
Test expects result: 5
Implementation: return 5

Next test expects different result
→ Now implement real logic
```

**Valid temporary approach, not final solution.**

### Strategy 2: Obvious Implementation

**When:** Solution is straightforward and clear

```
Test: Add two numbers
Implementation: return a + b

Clear, simple, complete.
```

### Strategy 3: Triangulation

**When:** Not sure how to generalize

```
Test 1: add(2, 3) should return 5
Implementation: return 5

Test 2: add(1, 4) should return 5
Still works with: return 5

Test 3: add(2, 2) should return 4
Now fails! Must implement: return a + b
```

**Multiple tests force generalization.**

## Integration with IDPF-Agile

GREEN phase follows RED in story implementation. TDD executes **autonomously**:

```
RED phase verified
ASSISTANT: Implements minimal code to pass test
ASSISTANT: Runs test and verifies pass
ASSISTANT: Proceeds to REFACTOR phase (no user interaction needed)
```

**Workflow Checkpoints:** The only required user interaction is at story completion (In Review → Done).

## Full Test Suite Execution

**During GREEN phase:**
- Run specific test being fixed
- Verify it turns green

**Before completing GREEN phase:**
- Run full test suite (if available)
- Ensure no regressions
- All tests should remain green

**When to run full suite:**
- After implementation complete
- Before committing code
- Before proceeding to REFACTOR phase

## GREEN Phase Checklist

Before proceeding to REFACTOR phase, verify:

- [ ] Implementation code is complete and correct
- [ ] Target test now PASSES (green)
- [ ] Implementation is minimal (no over-engineering)
- [ ] No existing tests broke (full suite green)
- [ ] Code is understandable and clear
- [ ] Implementation matches test requirements exactly
- [ ] No untested features added

## Resources

See `resources/` directory for:
- `green-phase-checklist.md` - Quick reference checklist
- `minimal-implementation-guide.md` - How to identify minimum code
- `triangulation-examples.md` - When and how to use triangulation

## Relationship to Other Skills

**Flows from:**
- `tdd-red-phase` - Previous phase with failing test

**Flows to:**
- `tdd-refactor-phase` - Next phase after GREEN success

**Related skills:**
- `tdd-failure-recovery` - Handle unexpected GREEN phase failures
- `test-writing-patterns` - Understanding test requirements

## Expected Outcome

After successful GREEN phase:
- Test that was red is now green
- Implementation is minimal and clear
- No regressions in existing tests
- Code is ready for refactoring consideration
- Autonomous progression to REFACTOR phase (no user command needed)

---

**End of TDD GREEN Phase Skill**
