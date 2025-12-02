# Testing Approach Selection Guide

**Purpose:** Help teams select appropriate testing methodologies for their project.

---

## Overview

| Approach | Required/Optional | Focus | When to Use |
|----------|-------------------|-------|-------------|
| TDD | **Required** | Code correctness | All development (unless tech stack prevents it) |
| ATDD | Optional | User requirements | When precise acceptance criteria needed upfront |
| BDD | Optional | System behavior | When stakeholder collaboration on behavior is valuable |

---

## TDD (Test-Driven Development)

**Status:** Required for IDPF-Structured, IDPF-Agile, and IDPF-LTS

### What is TDD?
Test-Driven Development follows a RED-GREEN-REFACTOR cycle:
1. **RED:** Write a failing test
2. **GREEN:** Write minimal code to pass the test
3. **REFACTOR:** Improve code while keeping tests green

### When TDD May Not Be Feasible
- Hardware-dependent embedded systems without simulators
- Legacy codebases with no test infrastructure
- Exploratory prototyping (use IDPF-Vibe instead)

### TDD Answers
"Does the code work correctly?"

---

## ATDD (Acceptance Test-Driven Development)

**Status:** Optional for all frameworks

### What is ATDD?
ATDD defines acceptance criteria before development:
1. Stakeholders, developers, and testers collaborate on acceptance criteria
2. Acceptance tests are written before coding
3. Development is complete when acceptance tests pass

### When to Use ATDD
- Requirements must be precisely validated
- Stakeholder sign-off is required
- Regulatory or compliance requirements exist
- Complex business rules need formal verification

### ATDD Answers
"Are we building the right thing?"

### Documentation
- Create Acceptance & Behavior Specification in `PRD/Specs/`
- Organize by REQ-ID (Structured/LTS) or Feature name (Agile)

---

## BDD (Behavior-Driven Development)

**Status:** Optional for all frameworks

### What is BDD?
BDD describes system behavior using Given-When-Then format:
1. **Three Amigos** session: Product owner, developer, tester collaborate
2. Behavior scenarios written in Gherkin (human-readable)
3. Scenarios become executable specifications in codebase

### When to Use BDD
- Cross-functional collaboration is valuable
- System behavior needs shared understanding
- Living documentation is desired
- Non-technical stakeholders need to understand tests

### BDD Answers
"How should the system behave?"

### Documentation
- Create initial planning in `PRD/Specs/` using Acceptance & Behavior Specification template
- Transfer finalized scenarios to feature files in codebase (e.g., `features/`, `specs/`)

---

## Combining Approaches

### TDD + ATDD
- ATDD wraps TDD: acceptance tests define outer boundary
- TDD used for implementation details within acceptance scope
- Recommended for projects with formal requirements

### TDD + BDD
- BDD scenarios drive high-level behavior tests
- TDD used for unit-level implementation
- Recommended for collaborative teams with business stakeholders

### TDD + ATDD + BDD
- Full coverage: acceptance criteria (ATDD) + behavior specs (BDD) + unit tests (TDD)
- Recommended for complex, regulated, or high-stakes projects

---

## Framework-Specific Guidance

### IDPF-Structured
- TDD: Required
- ATDD: Recommended for formal requirements validation
- BDD: Optional, useful for complex behavior specifications
- Specs organization: By REQ-ID (`PRD/Specs/REQ-001/`)

### IDPF-Agile
- TDD: Required
- ATDD: Optional, useful for stories with complex acceptance criteria
- BDD: Recommended for stakeholder collaboration
- Specs organization: By Feature/Epic name (`PRD/Specs/UserAuthentication/`)

### IDPF-LTS
- TDD: Required for all bug fixes
- ATDD/BDD: Optional, useful for critical bug fixes requiring formal verification
- Specs organization: By REQ-ID (`PRD/Specs/REQ-001/`)

---

## Tools by Language

| Language | TDD | BDD |
|----------|-----|-----|
| Java | JUnit, TestNG | Cucumber |
| C#/.NET | NUnit, xUnit | SpecFlow |
| Python | pytest, unittest | Behave |
| JavaScript | Jest, Mocha | Cucumber.js |
| Ruby | RSpec, Minitest | Cucumber |
| Go | testing package | Godog |

---

## References

- [TDD vs BDD vs ATDD: Key Differences - BrowserStack](https://www.browserstack.com/guide/tdd-vs-bdd-vs-atdd)
- [Can BDD be combined with ATDD? - SpecFlow](https://specflow.org/bdd/bdd-combined-with-atdd/)
- [BDD Essential Guide - Monday.com](https://monday.com/blog/rnd/behavior-driven-development/)
- [Behavior Driven Development - Brainhub](https://brainhub.eu/library/behavior-driven-development)
- [Gherkin, BDD, & Cucumber Guide - TestQuality](https://testquality.com/gherkin-bdd-cucumber-guide-to-behavior-driven-development/)
