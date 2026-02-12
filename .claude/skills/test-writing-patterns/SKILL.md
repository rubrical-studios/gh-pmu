---
name: test-writing-patterns
description: Guide experienced developers on test structure, patterns, assertions, and test doubles for effective test-driven development
version: "0.38.0"
license: Complete terms in LICENSE.txt
---

# Test Writing Patterns

This Skill provides experienced developers with test structure patterns, assertion strategies, test doubles, and organizational practices for effective test-driven development.

## When to Use This Skill

Invoke this Skill when:
- User needs guidance on test structure
- Questions about test organization
- Deciding which type of test to write
- Need test double (mock/stub/fake) guidance
- Uncertainty about assertion strategies
- Want to improve test quality and maintainability

## Prerequisites

- Understanding of testing concepts
- Familiarity with TDD RED-GREEN-REFACTOR cycle
- Experience with at least one testing framework
- Knowledge of testing terminology

## Test Structure Patterns

### AAA Pattern (Arrange-Act-Assert)

**Most common test structure:**

```
ARRANGE: Set up test conditions and inputs
ACT: Execute the behavior being tested
ASSERT: Verify the expected outcome
```

**Benefits:**
- Clear separation of concerns
- Easy to read and understand
- Consistent structure across tests
- Self-documenting

**Example structure (language-agnostic):**

```
test_function_name():
    // ARRANGE
    Setup test data
    Create required objects
    Configure dependencies

    // ACT
    result = call_function_under_test(inputs)

    // ASSERT
    verify(result == expected)
    verify(side_effects_occurred)
```

### Given-When-Then Pattern (BDD Style)

**Alternative structure (common in BDD):**

```
GIVEN: Initial context/preconditions
WHEN: Action/event occurs
THEN: Expected outcomes
```

**Benefits:**
- Natural language alignment
- Stakeholder-friendly
- Clear business context
- Behavior-focused

**Mapping to AAA:**
```
GIVEN = ARRANGE
WHEN = ACT
THEN = ASSERT
```

### Four-Phase Test Pattern

**Extended structure for complex tests:**

```
SETUP: Prepare test fixture
EXERCISE: Execute system under test
VERIFY: Check results
TEARDOWN: Clean up resources
```

**When to use:**
- Tests with expensive setup
- Tests requiring cleanup (files, connections)
- Integration tests
- Tests with shared resources

## Test Organization Strategies

### Test File Organization

**Principle: Mirror production structure**

```
Production:
  src/
    services/
      user_service
      order_service
    utils/
      validator
      formatter

Tests:
  tests/
    services/
      test_user_service
      test_order_service
    utils/
      test_validator
      test_formatter
```

### Test Naming Conventions

**Pattern: test_[unit]_[scenario]_[expected]**

```
test_add_positive_numbers_returns_sum
test_get_user_when_not_found_returns_null
test_create_order_with_invalid_data_raises_exception
test_authenticate_with_wrong_password_fails
```

**Benefits:**
- Clear what is being tested
- Describes the scenario
- States expected outcome
- Readable as documentation

### Test Suite Organization

**By test type:**

```
tests/
  unit/          # Fast, isolated unit tests
  integration/   # Tests with dependencies
  e2e/           # End-to-end tests
  performance/   # Performance/load tests
```

**By feature:**

```
tests/
  user_management/
  order_processing/
  payment_handling/
  reporting/
```

## Assertion Strategies

### Single Concept Per Test

**Good: One assertion or set of related assertions**

```
test_user_creation_sets_properties():
    user = create_user(name="Alice", age=30)

    assert user.name == "Alice"
    assert user.age == 30
    assert user.created_at is not null
    // All verify the same concept: user creation
```

**Poor: Multiple unrelated assertions**

```
test_user_operations():
    assert create_user() works
    assert update_user() works
    assert delete_user() works
    assert list_users() works
    // Split into separate tests
```

### Assertion Quality

**Good assertions:**
- Specific and precise
- Include helpful messages
- Test observable behavior
- Independent of implementation details

**Poor assertions:**
- Generic (just assert True)
- No context when failing
- Test internal state
- Coupled to implementation

### Common Assertion Types

**Equality:**
```
assert actual == expected
assert value equals 5
```

**Truthiness:**
```
assert condition is true
assert value is not null
```

**Comparison:**
```
assert value > 0
assert count <= limit
```

**Collection membership:**
```
assert item in collection
assert collection contains element
```

**Exception/Error:**
```
assert raises(ExpectedException)
assert throws error
```

**Type checking:**
```
assert isinstance(obj, ExpectedType)
assert type matches
```

## Test Doubles

### Types of Test Doubles

**From frameworks:**
> "Test doubles: mocks, stubs, fakes, and spies"

### Stub

**Purpose:** Provide predetermined responses to calls

**When to use:**
- Need to control dependency behavior
- Isolate unit under test
- Avoid external dependencies

**Characteristics:**
- Returns fixed values
- No behavior verification
- Simplest test double

**Example usage:**
```
Stub database to return specific user
Stub API to return fixed response
Stub file system to return test data
```

### Mock

**Purpose:** Verify interactions/calls

**When to use:**
- Need to verify method was called
- Check call parameters
- Verify call count/order

**Characteristics:**
- Records calls made to it
- Can verify expectations
- Used for behavior verification

**Example usage:**
```
Mock logger to verify error logged
Mock email service to verify send called
Mock event publisher to verify event emitted
```

### Fake

**Purpose:** Working implementation (simplified)

**When to use:**
- Real implementation too slow/complex
- Need realistic behavior
- Integration testing

**Characteristics:**
- Functional implementation
- Simpler than production
- Actually works

**Example usage:**
```
Fake in-memory database
Fake file system in memory
Fake message queue
```

### Spy

**Purpose:** Record information about calls while delegating to real object

**When to use:**
- Need real behavior plus verification
- Want to observe interactions
- Partial mocking

**Characteristics:**
- Wraps real object
- Records calls
- Delegates to real implementation

**Example usage:**
```
Spy on cache to verify hits/misses
Spy on validator to track validations
Spy on logger while keeping logs
```

### Test Double Selection Guide

```
Need to control response? → Stub
Need to verify call made? → Mock
Need working but simple version? → Fake
Need real behavior + verification? → Spy
```

## Test Isolation

### Principle: Tests Should Be Independent

**Each test should:**
- Set up its own data
- Clean up after itself
- Run in any order
- Not depend on other tests

**From frameworks:**
> "Tests depend on order: Each test should work independently"

### Achieving Isolation

**Setup/Teardown:**
```
Before each test:
  - Create fresh test data
  - Reset state
  - Initialize dependencies

After each test:
  - Clean up resources
  - Remove test data
  - Close connections
```

**Fixtures:**
```
Define reusable test data/objects
Initialize fresh for each test
Avoid shared mutable state
```

**Database isolation:**
```
Use transactions (rollback after test)
Separate test database
Clear data between tests
Use in-memory database
```

## Test Data Strategies

### Explicit Test Data

**Prefer explicit over random:**

```
Good:
  user = create_user(name="Alice", age=30)
  // Clear what data represents

Poor:
  user = create_user(name=random_string(), age=random_int())
  // Unclear, non-deterministic
```

### Minimal Test Data

**Use simplest data that tests the behavior:**

```
Good:
  data = {name: "A", value: 1}
  // Minimal, clear

Poor:
  data = {
    name: "Very Long Realistic Name",
    value: 42,
    description: "Long realistic description",
    metadata: {...},
    // Lots of irrelevant fields
  }
```

### Test Data Builders

**Pattern for complex object creation:**

```
Builder pattern:
  user = UserBuilder()
    .with_name("Alice")
    .with_role("admin")
    .build()

Benefits:
  - Readable
  - Defaults for irrelevant fields
  - Explicit about what matters
```

## Testing Strategies by Type

### Unit Tests

**Characteristics:**
- Test single unit (function/class/module)
- Fast execution
- No external dependencies
- Use test doubles for dependencies

**Focus:**
- Logic correctness
- Edge cases
- Error handling
- Single responsibility

### Integration Tests

**Characteristics:**
- Test multiple units together
- May use real dependencies
- Slower than unit tests
- Verify component interaction

**Focus:**
- Interface contracts
- Data flow between components
- Error propagation
- Integration points

### End-to-End Tests

**Characteristics:**
- Test complete user workflows
- Use real system
- Slowest tests
- Highest confidence

**Focus:**
- Critical user paths
- Business scenarios
- System behavior
- Actual usage patterns

## Test Coverage Considerations

### What Coverage Means

**Code coverage metrics:**
- Line coverage: Lines executed
- Branch coverage: Conditional paths taken
- Function coverage: Functions called
- Statement coverage: Statements executed

**Coverage ≠ Quality:**
```
100% coverage doesn't mean:
  - All behaviors tested
  - All edge cases covered
  - Tests are good quality
  - Code is correct

Coverage shows:
  - Which code is executed by tests
  - Which code is NOT executed
  - Gaps in test suite
```

### Coverage Goals

**From frameworks (implied):**
- Test coverage analysis and quality metrics are mentioned
- Focus on behavior testing, not coverage percentage
- TDD naturally produces good coverage

**Guidelines:**
- Aim for high coverage of critical paths
- 100% coverage is not always necessary
- Focus on meaningful tests, not coverage numbers
- Use coverage to find untested code
- Don't write tests just to increase coverage

## Parameterized Tests

### Testing Multiple Scenarios

**Pattern: Same test logic, different data**

```
Instead of:
  test_add_positive_numbers
  test_add_negative_numbers
  test_add_mixed_numbers
  test_add_zeros

Use parameterized test:
  test_add_numbers:
    parameters:
      (2, 3, 5)
      (-2, -3, -5)
      (2, -3, -1)
      (0, 0, 0)

    for each (a, b, expected):
      assert add(a, b) == expected
```

**Benefits:**
- Reduces duplication
- Easy to add new cases
- Clear data-driven approach
- Comprehensive coverage

## Test Smells and How to Fix Them

### Smell 1: Test Does Too Much

**Symptom:** Long test with many assertions

**Fix:** Split into multiple focused tests

### Smell 2: Tests Are Brittle

**Symptom:** Tests break with unrelated changes

**Fix:** Test behavior, not implementation details

### Smell 3: Tests Are Slow

**Symptom:** Test suite takes too long

**Fix:** Use test doubles, optimize setup, parallelize

### Smell 4: Tests Are Unclear

**Symptom:** Hard to understand what test does

**Fix:** Better naming, clear AAA structure, comments for context

### Smell 5: Tests Depend on Each Other

**Symptom:** Tests fail when run in different order

**Fix:** Ensure test isolation, proper setup/teardown

### Smell 6: Duplicate Setup Code

**Symptom:** Same setup in many tests

**Fix:** Extract to fixtures, use test data builders

## Framework-Agnostic Principles

### Principles Applicable Across All Languages/Frameworks

**Test structure:**
- AAA pattern works everywhere
- Clear naming conventions
- One concept per test

**Test isolation:**
- Independent tests
- Fresh data each test
- Proper cleanup

**Test quality:**
- Clear assertions
- Meaningful names
- Focused tests

**Test organization:**
- Mirror production structure
- Separate by type or feature
- Consistent conventions

## Resources

See `resources/` directory for:
- `aaa-pattern-template.md` - AAA test template
- `test-doubles-guide.md` - When to use each test double type
- `assertion-patterns.md` - Common assertion patterns by scenario
- `test-organization-examples.md` - Organization structure examples

## Relationship to Other Skills

**Used by:**
- `tdd-red-phase` - Writing test structure
- `tdd-green-phase` - Understanding test requirements
- `tdd-refactor-phase` - Improving test quality

**Independent from:**
- `beginner-testing` - This skill assumes experience

**Complements:**
- `tdd-failure-recovery` - Understanding why tests fail

## Expected Outcome

After applying test writing patterns:
- Tests are well-structured and readable
- Appropriate test doubles used
- Clear, meaningful assertions
- Good test organization
- Independent, isolated tests
- Maintainable test suite

---

**End of Test Writing Patterns Skill**
