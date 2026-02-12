---
name: tdd-refactor-phase
description: Guide experienced developers through REFACTOR phase of TDD cycle - improving code quality while maintaining green tests
version: "0.38.0"
license: Complete terms in LICENSE.txt
---

# TDD REFACTOR Phase

This Skill guides experienced developers through the REFACTOR phase of the Test-Driven Development cycle: improving code quality, structure, and clarity while ensuring all tests remain green.

## When to Use This Skill

Invoke this Skill when:
- GREEN phase complete with passing test
- Proceeding autonomously from GREEN phase
- Code works but could be improved
- Evaluating refactoring opportunities

## Prerequisites

- Completed GREEN phase with all tests passing
- Working implementation that satisfies test requirements
- Full test suite available to verify refactoring safety
- Claude Code available for analysis and execution

## REFACTOR Phase Objectives

The REFACTOR phase has dual goals:

1. **Improve code quality** - Make code cleaner, more maintainable, better structured
2. **Keep tests green** - Ensure all improvements maintain existing functionality

### What "Refactoring" Means

**✓ Refactoring IS:**
- Improving code structure without changing behavior
- Making code more readable and maintainable
- Eliminating duplication
- Simplifying complex logic
- Improving naming and organization

**✗ Refactoring IS NOT:**
- Adding new features
- Changing tested behavior
- Fixing bugs (that's a new test + implementation)
- Performance optimization without measurement
- Breaking tests to "improve" code

## REFACTOR Phase Workflow

### Step 1: Analyze Refactoring Opportunities

**ASSISTANT instructs User to ask Claude Code:**
```
"Analyze this code for refactoring opportunities"
```

**User action:**
- Provide instruction to Claude Code
- Claude Code analyzes implementation
- User reports findings back to ASSISTANT

**Claude Code should identify:**
- Code duplication
- Long or complex functions
- Unclear variable/function names
- Missing abstractions
- Violation of design principles (SOLID, DRY, etc.)
- Complex conditional logic
- Magic numbers/strings

### Step 2: Evaluate Refactoring Suggestions

**ASSISTANT evaluates Claude Code's findings:**

**Decision: Refactor Now**
- Clear improvement opportunity
- Low risk, high value
- Directly improves code in this iteration
- Won't over-engineer solution

**Decision: Skip Refactoring**
- Suggestion is premature abstraction
- Risk > reward for current iteration
- Better addressed in future iteration
- Code is already clear enough

**From IDPF frameworks:**
> "ASSISTANT evaluates findings and either:
> Option A: Provides refactored code for User to apply via Claude Code
> Option B: Recommends skipping refactoring"

### Step 3: Apply Refactoring (if approved)

**If ASSISTANT approves refactoring:**

**ASSISTANT provides to user in single code block:**

1. Refactored code
2. Explanation of improvements
3. Test execution command
4. Expected: All tests still pass

**Critical: Single Code Block Format**

```
TASK: [Brief description of refactoring]

STEP 1: [Open implementation file]
STEP 2: [Navigate to code being refactored]
STEP 3: [Apply refactored code - COMPLETE implementation]
STEP 4: [Explanation of what improved and why]
STEP 5: [Save file]
STEP 6: [Run full test suite]
STEP 7: [Verify ALL tests still PASS (green)]
STEP 8: [Report back: All tests green?]
```

### Step 4: Verify Tests Remain Green

**Critical verification:**
- Run FULL test suite (not just recent test)
- ALL tests must pass
- No test failures allowed
- No test errors allowed

**If any test fails:**
→ Refactoring broke something
→ Rollback refactoring immediately
→ Keep tests green
→ Option: Try smaller refactoring

**From IDPF frameworks:**
> "If refactoring is applied, User runs tests via Claude Code to verify tests still pass"

### Step 5: Complete REFACTOR Phase

**If refactoring applied and tests green:**
→ REFACTOR phase complete
→ Code improved and safe
→ Proceed to next behavior or complete story

**If refactoring skipped:**
→ REFACTOR phase complete (no changes)
→ Proceed to next behavior or complete story

**TDD cycle continues autonomously** until story implementation is complete. The only workflow checkpoint is story completion (In Review → Done).

## REFACTOR Phase Best Practices

### Practice 1: Refactor in Small Steps

**Good approach:**
```
1. Extract one variable → Run tests
2. Rename one function → Run tests
3. Extract one function → Run tests

Each step verified independently
```

**Poor approach:**
```
1. Extract variables + rename + restructure all at once
2. Run tests
3. Multiple failures, unclear which change broke what
```

### Practice 2: One Refactoring at a Time

**Focus on one improvement:**
- Eliminate duplication (one instance)
- Improve naming (one variable/function)
- Extract function (one extraction)
- Simplify conditional (one condition)

**Then run tests. Then next refactoring.**

### Practice 3: Keep Tests Green

**Golden rule:**
```
Tests must ALWAYS be green after refactoring.

If refactoring breaks tests:
→ Rollback immediately
→ Tests must stay green
→ Try smaller refactoring
```

**From IDPF frameworks:**
> "Refactoring breaks tests: Roll back refactoring; tests must stay green"

### Practice 4: Refactor for Clarity, Not Cleverness

**Good refactoring:**
```
Makes code easier to understand
Makes intent clearer
Reduces cognitive load
Improves maintainability
```

**Poor refactoring:**
```
Clever one-liners that obscure intent
Over-abstracted "elegant" solutions
Premature design patterns
Showing off language features
```

## Common Refactorings

### Refactoring 1: Extract Variable

**Before:**
```
Calculation or expression embedded in code
Hard to understand what value represents
```

**After:**
```
Value assigned to well-named variable
Intent clear from variable name
```

### Refactoring 2: Extract Function

**Before:**
```
Long function doing multiple things
Logic buried in larger context
```

**After:**
```
Logic extracted to well-named function
Function does one clear thing
Reusable and testable independently
```

### Refactoring 3: Rename for Clarity

**Before:**
```
Unclear variable/function names
Intent not obvious
Abbreviations or generic names
```

**After:**
```
Names clearly express intent
Self-documenting code
No ambiguity about purpose
```

### Refactoring 4: Eliminate Duplication

**Before:**
```
Same code appears multiple places
Changes must be synchronized
Easy to miss one location
```

**After:**
```
Duplicated code extracted to function
Single source of truth
Changes in one place
```

### Refactoring 5: Simplify Conditional Logic

**Before:**
```
Nested conditions
Complex boolean expressions
Hard to follow logic flow
```

**After:**
```
Guard clauses reduce nesting
Early returns simplify flow
Extracted boolean expressions with clear names
```

## When to Skip Refactoring

### Skip Scenario 1: Premature Abstraction

**Indicators:**
- Only one use of the code
- Future needs unclear
- Abstraction more complex than original

**Decision: Skip**
- Wait for second or third occurrence
- Rule of Three before abstracting

### Skip Scenario 2: Code Already Clear

**Indicators:**
- Claude Code suggests minor naming changes
- Current names are already descriptive
- Change doesn't add clarity

**Decision: Skip**
- Current code is good enough
- Don't refactor for sake of refactoring

### Skip Scenario 3: High Risk, Low Value

**Indicators:**
- Refactoring touches many files
- Complex change for minor improvement
- Could introduce bugs

**Decision: Skip or Defer**
- Not worth risk in this iteration
- Consider in future dedicated refactoring session

### Skip Scenario 4: Over-Engineering

**Indicators:**
- Suggestion adds design patterns prematurely
- Creates abstraction for single use case
- "Might need this later"

**Decision: Skip**
- Keep it simple
- Wait for actual need

## REFACTOR Phase Anti-Patterns

### Anti-Pattern 1: Refactoring Without Tests

```
✗ WRONG:
Make changes → Hope nothing broke

✓ CORRECT:
Make changes → Run tests → Verify green → Proceed
```

### Anti-Pattern 2: Accepting Broken Tests

```
✗ WRONG:
Refactor → Tests fail → "I'll fix tests later"

✓ CORRECT:
Refactor → Tests fail → ROLLBACK → Tests green again
```

### Anti-Pattern 3: Big Bang Refactoring

```
✗ WRONG:
Change everything at once
Tests fail, don't know which change broke what

✓ CORRECT:
Small incremental changes
Test after each change
Identify exactly what breaks when
```

### Anti-Pattern 4: Refactoring + Features

```
✗ WRONG:
Refactor existing code + add new feature simultaneously

✓ CORRECT:
Refactor (tests stay green) OR add feature (new test)
Never both at same time
```

## Integration with IDPF-Agile

REFACTOR phase follows GREEN in story implementation. TDD executes **autonomously**:

```
GREEN phase verified, tests passing
ASSISTANT: Analyzes code for refactoring opportunities
ASSISTANT: Evaluates and either:
  - Applies refactoring (Option A)
  - Skips refactoring (Option B)
ASSISTANT: Runs tests, verifies green
ASSISTANT: Proceeds to next behavior OR completes story
```

**Workflow Checkpoints:** The only required user interaction is at story completion (In Review → Done).

## Rollback Procedures

**If refactoring breaks tests:**

1. **Immediate action:** Rollback changes (git checkout or undo)
2. **Verify:** Tests return to green
3. **Options:**
   - Try smaller refactoring
   - Skip refactoring for now
   - Investigate why tests broke

**Rollback is autonomous** — the assistant handles reverting broken changes and maintaining green tests throughout the TDD cycle.

## REFACTOR Phase Checklist

Before proceeding to next feature, verify:

- [ ] Code analyzed for refactoring opportunities
- [ ] Suggestions evaluated
- [ ] If refactoring applied:
  - [ ] Refactored code is clear and improved
  - [ ] All tests run and PASS (green)
  - [ ] No test failures or errors
  - [ ] Behavior unchanged
- [ ] If refactoring skipped:
  - [ ] Valid reason for skipping
  - [ ] Tests still green

## Resources

See `resources/` directory for:
- `refactor-checklist.md` - Quick reference checklist
- `common-refactorings.md` - Catalog of common refactoring patterns
- `when-to-skip-refactoring.md` - Decision guide for skipping refactoring

## Relationship to Other Skills

**Flows from:**
- `tdd-green-phase` - Previous phase with passing tests

**Flows to:**
- `tdd-red-phase` - Next feature starts new RED phase

**Related skills:**
- `tdd-failure-recovery` - Handle broken tests during refactoring

## Expected Outcome

After successful REFACTOR phase:
- Code quality improved (if refactored) OR intentionally left as-is (if skipped)
- All tests remain green
- No behavioral changes
- Ready to start next feature with RED phase
- Autonomous progression to next behavior or story completion

---

**End of TDD REFACTOR Phase Skill**
