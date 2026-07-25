# VEDA Agent Runtime Implementation Guide

This document defines the engineering execution standards that all implementations must follow to maintain architectural consistency and quality. It serves as the engineering handbook to be consulted before beginning any implementation work.

## 1. Repository Workflow

### Branch Strategy
- **Main Branch**: `main` - Always contains production-ready code
- **Development Branches**: 
  - `feature/vXX.X-description` for new features (where XX.X is the target milestone version)
  - `fix/issue-number-description` for bug fixes
  - `refactor/component-description` for refactoring work
  - `docs/topic-description` for documentation changes
- **Release Branches**: `release/vX.Y.Z` created from `main` for release preparation
- **Hotfix Branches**: `hotfix/description` branched from `main` for critical production fixes

### Commit Strategy
- **Atomic Commits**: Each commit should represent a single logical change
- **Commit Messages**: Follow conventional commits format:
  ```
  <type>(<scope>): <subject>
  
  <body>
  
  <footer>
  ```
  - Types: feat, fix, docs, style, refactor, perf, test, chore, ci
  - Scope: Component or module being changed (e.g., kernel, execution, memory)
  - Subject: Imperative mood, max 50 characters
  - Body: Detailed explanation, wrap at 72 characters
  - Footer: References to issues, breaking changes
- **Sign-off**: All commits must include `Signed-off-by` line agreeing to Developer Certificate of Origin

### Pull Request Expectations
- **Size Limit**: PRs should be reviewable in under 60 minutes (typically <400 lines changed)
- **Single Responsibility**: Each PR should address only one concern (feature, fix, refactor)
- **Description**: Must include:
  - Clear summary of changes
  - Related issue/ticket numbers
  - Impact assessment (performance, security, breaking changes)
  - Testing performed
  - Screenshots or examples for UI changes
- **Checks**: All CI checks must pass before review can begin
- **Draft PRs**: Use for work-in-progress feedback; mark as ready for review when complete

### Review Workflow
1. **Author**: Creates draft PR, runs local tests, requests initial feedback if needed
2. **Reviewer Assignment**: Automatically assigned based on CODEOWNERS file
3. **Review Process**:
   - First pass: Architecture and design review (by architect or tech lead)
   - Second pass: Code quality and correctness (by peer developers)
   - Third pass: Testing and documentation (by QA or designated reviewer)
4. **Approval Requirements**: 
   - Minimum 2 approvals (including 1 architecture review for interface changes)
   - No requested changes
   - All status checks passing
5. **Merge Method**: Squash and merge to maintain clean history
6. **Post-Merge**: Delete branch, update documentation if needed

## 2. Coding Standards

### Go Conventions
- Follow [Effective Go](https://golang.org/doc/effective_go.html) and [Go CodeReview Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `goimports` on save (configured via editor settings)
- Target Go 1.20+ for all development
- Enable `go vet` and `staticcheck` in CI pipeline

### Naming Conventions
- **Packages**: lowercase, single words, no underscores (e.g., `kernel`, `execution`)
- **Interfaces**: 
  - Single-method interfaces: named after method + `er` (e.g., `Executor`, `Validator`)
  - Multi-method interfaces: noun or noun phrase (e.g., `Registry`, `Store`)
- **Structs**: MixedCase (PascalCase) for exported, camelCase for unexported
- **Functions**: MixedCase (PascalCase) for exported, camelCase for unexported
- **Variables**: camelCase, descriptive but concise
- **Constants**: UPPER_SNAKE_CASE for exported, camelCase for unexported
- **Error Variables**: `Err` prefix followed by description (e.g., `ErrNotFound`)
- **Factory Functions**: `New` followed by type being created (e.g., `NewStore()`)

### Package Organization
- **Internal Packages**: 
  - `internal/` - Application private code (not importable outside)
  - `pkg/` - Library code safe for external use
- **Layer Separation**: 
  - `/interfaces` - Interface definitions only
  - `/impl` or `/internal` - Implementation details
  - Avoid putting implementation in interface packages
- **Dependency Direction**: 
  - Dependencies point inward toward domain core
  - Higher-level packages depend on lower-level packages
  - Never depend on implementation packages from another layer
- **Circular Dependencies**: Strictly prohibited; detected by CI

### Error Handling
- **Error Values**: Use `error` interface for all error conditions
- **Error Creation**: 
  - `errors.New()` for static messages
  - `fmt.Errorf()` for formatted messages (use `%w` to wrap errors)
  - Define error variables for common errors (e.g., `var ErrNotFound = errors.New("not found")`)
- **Error Checking**: 
  - Check errors immediately after function calls
  - Never ignore errors with `_` unless intentionally discarding
  - Use `errors.Is()` and `errors.As()` for error inspection
- **Error Propagation**: 
  - Wrap errors with context when propagating up
  - Don't lose original error information when wrapping
  - Use `%w` verb to maintain error chain
- **Panics**: 
  - Recover from panics at appropriate boundaries
  - Never let panics escape public APIs
  - Convert panics to errors when appropriate using `defer` and `recover()`

### Context Usage
- **Propagation**: 
  - Pass `context.Context` as first parameter to functions
  - Never store `context.Context` in structs (except middleware)
  - Always propagate context to downstream calls
- **Cancellation**: 
  - Respect `ctx.Done()` for cancellation
  - Use `ctx.Err()` to determine cancellation reason
  - Clean up resources on cancellation
- **Timeouts**: 
  - Use `context.WithTimeout` or `context.WithDeadline` for operations with time limits
  - Default to reasonable timeouts (5-30 seconds) unless domain-specific
- **Values**: 
  - Use context values only for request-scoped data (trace IDs, user info, etc.)
  - Avoid using for optional parameters; use function overloads instead
  - Define key types to avoid collisions: `type ctxKey int`

### Dependency Injection
- **Constructor Injection**: 
  - Dependencies passed via constructor functions
  - Use option pattern for optional dependencies: `func NewService(opts ...Option) *Service`
  - Define functional options for configuration: `type Option func(*Service)`
- **Interface-Based**: 
  - Depend on interfaces, not concrete implementations
  - Allow mocking for testing
  - Enable swapping implementations
- **Avoid**: 
  - Global variables for dependencies
  - Service locator patterns
  - Hard-coded instantiation of dependencies
- **Wire**: Consider using Google Wire for complex dependency graphs (optional)

### Logging
- **Structured Logging**: 
  - Use key-value pairs (JSON format preferred)
  - Include standard fields: `timestamp`, `level`, `message`, `logger`
  - Add contextual fields: `requestID`, `userID`, `operation`, `duration`
- **Log Levels**: 
  - `TRACE`: Extremely detailed information (disabled in production)
  - `DEBUG`: Diagnostic information (typically disabled in production)
  - `INFO`: General information about system operation
  - `WARN`: Warning about potentially harmful situations
  - `ERROR`: Error events that might still allow application to continue
  - `FATAL`: Severe errors causing premature termination
- **Implementation**: 
  - Use structured logging library (zap, zerolog, etc.)
  - Avoid string concatenation in hot paths
  - Implement sampling for high-volume logs
  - Never log sensitive information (passwords, tokens, PII)
- **Configuration**: 
  - Make log level configurable at runtime
  - Support multiple outputs (file, stdout, syslog, etc.)
  - Implement log rotation and retention policies

### Concurrency
- **Goroutine Usage**: 
  - Launch goroutines only when necessary for concurrency
  - Always have a clear plan for stopping goroutines
  - Use `context.Context` for cancellation and timeouts
  - Avoid leaking goroutines
- **Channels**: 
  - Use channels for communication between goroutines
  - Prefer unbuffered channels unless buffering is specifically needed
  - Close channels only from the sender side
  - Never close nil or already closed channels
- **Shared State**: 
  - Prefer communication over shared memory
  - When sharing state, use appropriate synchronization (mutex, rwmutex, atomic)
  - Minimize time spent holding locks
  - Consider using `sync/atomic` for simple cases
- **Patterns**: 
  - Worker pools for limiting concurrency
  - Fan-in/fan-out for distributing and collecting work
  - Pipelines for staged processing
  - Context cancellation for graceful shutdown

### Testing
- **Test Organization**: 
  - Table-driven tests for functions with multiple input/output combinations
  - Test both happy paths and error conditions
  - Test boundary conditions and edge cases
  - Use subtests (`t.Run`) for related test cases
- **Assertions**: 
  - Use built-in testing package for assertions
  - Consider using assertion libraries for complex comparisons (testify, etc.)
  - Make failure messages clear and informative
- **Mocking**: 
  - Use interfaces to enable mocking
  - Consider using mocking libraries (gomock, testify/mock, etc.)
  - Keep mocks simple and focused
  - Verify interactions when behavior is important
- **Benchmarks**: 
  - Write benchmarks for performance-critical code
  - Use `testing.B` for benchmarking
  - Report meaningful metrics (ns/op, allocations, etc.)
  - Compare against baselines

## 3. Architecture Rules

### Absolute Prohibitions
Developers are **NEVER** allowed to:

1. **Create Circular Dependencies**: 
   - No package A depending on package B which depends on package A
   - Detected by `go list -f '{{.Deps}}'` and CI checks
   - Violations block merge immediately

2. **Create Direct Subsystem Coupling**: 
   - No direct calls between major subsystems (kernel, execution, lifecycle, etc.)
   - All communication must go through defined interfaces or event bus
   - Violations require architecture review and justification

3. **Use Hidden Singleton State**: 
   - No package-level variables that create implicit global state
   - State must be explicitly passed or managed through scoped containers
   - Exceptions: read-only configuration, metrics registrars (with approval)

4. **Bypass Interfaces**: 
   - No direct access to implementation details behind interfaces
   - No type assertions to concrete types unless absolutely necessary (and documented)
   - Violations require architecture review

5. **Use Package-Level Mutable Globals**: 
   - No mutable global variables at package level
   - State must be encapsulated in objects with clear ownership
   - Exceptions: 
     - Metrics collectors (registered once, immutable after)
     - Loggers (configured once, immutable after)
     - Singletons explicitly approved by architecture board

6. **Implement Before Dependency Milestone Completes**: 
   - No implementation work on a milestone until its dependencies reach "Done" status
   - Verified through dependency tracking in project management system
   - Violations block merge and require rework

7. **Modify Frozen Interfaces Without Approval**: 
   - Interfaces marked as frozen in the implementation plan cannot be changed
   - Changes require architecture review board approval
   - Violations are rejected immediately

8. **Introduce Hidden Control Flow**: 
   - No hidden callbacks, hooks, or magic behavior
   - All control flow must be explicit and visible in code
   - Event-driven communication must use the defined event bus

9. **Violate Dependency Direction**: 
   - Dependencies must point inward toward domain core
   - Higher-level packages depending on lower-level packages is forbidden
   - Detected by dependency analysis tools

10. **Use Reflection for Business Logic**: 
    - Reflection only allowed for framework infrastructure (serialization, ORM, etc.)
    - Never use reflection to bypass type safety in business logic
    - Requires explicit justification and review

## 4. Implementation Workflow

For every engineering task, follow this exact sequence:

### Step 1: Read Architecture
- Review relevant sections in `VEDA_AGENT_RUNTIME_ARCHITECTURE.md`
- Identify all interfaces, contracts, and architectural constraints
- Note any specific requirements or prohibitions for the component
- Check for any open architecture decisions that might affect implementation

### Step 2: Read Milestone
- Review the corresponding milestone in `VEDA_AGENT_RUNTIME_IMPLEMENTATION_PLAN.md`
- Understand the goal, deliverables, and acceptance criteria
- Identify dependencies on other milestones
- Note any risk checkpoints or special validation requirements

### Step 3: Implement Task
- Implement only the specific engineering task as defined in the backlog
- Follow all coding standards and architecture rules
- Make minimal, focused changes
- Do not add extra functionality beyond acceptance criteria
- Do not refactor unrelated code (save for separate refactoring task)

### Step 4: Write Tests
- Write unit tests for all new functionality
- Achieve specified coverage targets
- Test both success and failure paths
- Include edge cases and boundary conditions
- Ensure tests are deterministic and don't have external dependencies
- Update or add tests for any modified existing functionality

### Step 5: Run Validation
- Run all local checks: `go build ./...`, `go test ./...`, `go vet ./...`, `staticcheck ./...`
- Verify performance benchmarks are met (if applicable)
- Check that documentation is updated
- Confirm no linting errors
- Run any specific validation scripts mentioned in the task

### Step 6: Architecture Review
- For interface changes or high-risk tasks: request architecture review
- Reviewer checks for:
  - No architectural drift
  - No dependency violations
  - No interface bypassing
  - Compliance with dependency direction rules
  - No hidden singleton state
- Address all review comments before proceeding

### Step 7: Merge
- Request code review through standard pull request process
- Address all review comments
- Ensure all checks pass
- Merge using squash and merge
- Delete feature branch
- Update any relevant documentation or tracking systems

## 5. Testing Workflow

### Unit Tests
- **Scope**: Individual functions, methods, and types in isolation
- **Dependencies**: All external dependencies mocked or stubbed
- **Execution**: `go test ./...` in CI and locally
- **Coverage**: 
  - >90% for complex logic (algorithms, business rules)
  - >80% for simple logic (getters, setters, simple wrappers)
  - Measured per package, not globally
- **Location**: Same package as code under test, `_test.go` files
- **Frameworks**: Standard `testing` table, optionally enhanced with testify

### Contract Tests
- **Scope**: Verify implementations conform to defined interfaces
- **Focus**: 
  - Interface method signatures and behavior
  - Contractual guarantees and invariants
  - Backward and forward compatibility
  - Message schema and format compliance
- **Execution**: Part of test suite, tagged `contract`
- **Tools**: Interface reflection, custom contract testing frameworks
- **Frequency**: Run on every change to interfaces or implementations

### Integration Tests
- **Scope**: Interactions between components and subsystems
- **Focus**: 
  - Component-to-component interactions
  - Subsystem integration points
  - Message flow between services
  - Event publishing and consumption
- **Environment**: 
  - Limited external dependencies (use test doubles for expensive resources)
  - Real implementations where possible, mocks where necessary
  - In-memory databases, mock network services
- **Execution**: `go test -tags=integration ./...`
- **Duration**: Target <1 second per test, <30 seconds for full suite
- **Location**: `*_integration_test.go` files or `/integration` directory

### Benchmark Tests
- **Scope**: Performance-critical code paths
- **Focus**: 
  - Latency measurements for key operations
  - Throughput under various load conditions
  - Resource utilization (CPU, memory, I/O)
  - Bottleneck identification
- **Execution**: `go test -bench=. ./...` or `go test -benchmem ./...`
- **Reporting**: 
  - ns/op, allocations, bytes/op
  - Comparison against baseline/commit
  - Regression alerts for >10% degradation
- **Location**: `*_benchmark.go` files or `/benchmark` directory

### Regression Tests
- **Scope**: Previously fixed bugs and edge cases
- **Purpose**: Ensure fixes remain effective and don't reappear
- **Maintenance**: 
  - Add test for every fixed bug before closing issue
  - Tag with `regression` or include in relevant test suite
  - Review periodically for obsolescence
- **Execution**: Part of standard test suite

### Performance Tests
- **Scope**: System-level performance characteristics
- **Focus**: 
  - End-to-end workflows (agent creation → task execution → result)
  - Performance under load (simulated users, concurrent operations)
  - Scaling characteristics
  - Resource efficiency
- **Environment**: 
  - Production-like (may be scaled down)
  - Realistic workloads and data patterns
  - Monitoring and profiling enabled
- **Tools**: Load testing frameworks (k6, vegeta, etc.), profilers (pprof)
- **Frequency**: 
  - Nightly in CI for basic benchmarks
  - Weekly for full performance sprintly for comprehensive performance test suites
  - Before major releases

## 6. Code Review Checklist

Every pull request must verify:

### ✅ Architecture Preserved
- [ ] No circular dependencies introduced
- [ ] No direct subsystem coupling (all communication via interfaces/events)
- [ ] No hidden singleton state
- [ ] No interface bypassing
- [ ] No package-level mutable globals (except approved exceptions)
- [ ] Dependency direction maintained (inward toward domain core)
- [ ] No violation of frozen interfaces without approval

### ✅ Interfaces Unchanged
- [ ] No changes to frozen interfaces without architecture review approval
- [ ] Interface method signatures preserved
- [ ] Behavioral contracts maintained
- [ ] Backward compatibility preserved (unless breaking change approved)
- [ ] All interface implementations updated consistently

### ✅ No Dependency Violations
- [ ] All dependencies point inward toward domain core
- [ ] No higher-level depending on lower-level without justification
- [ ] No implementation depending on another implementation (only on interfaces)
- [ ] External dependencies approved and tracked
- [ ] No forbidden dependencies (checked via dependency analysis)

### ✅ Tests Added
- [ ] Unit tests for all new functionality
- [ ] Tests achieve specified coverage targets
- [ ] Both positive and negative test cases included
- [ ] Edge cases and boundary conditions tested
- [ ] Tests are deterministic and don't require external setup
- [ ] Existing tests still pass (no regressions)

### ✅ Documentation Updated
- [ ] Code comments added for complex logic
- [ ] Public functions and types have godoc comments
- [ ] Package-level documentation present and accurate
- [ ] Architecture decision records updated if applicable
- [ ] User-facing documentation updated for API/behavior changes
- [ ] Examples and tutorials updated if affected

### ✅ Performance Acceptable
- [ ] Performance benchmarks meet or exceed targets
- [ ] No significant regression in related performance metrics
- [ ] Resource usage (memory, CPU) within expected bounds
- [ ] No unnecessary allocations in hot paths
- [ ] Benchmark results documented in PR description

### ✅ No Duplicated Logic
- [ ] No copy-pasted code from elsewhere in the codebase
- [ ] Similar functionality extracted to shared utilities or functions
- [ ] DRY principle applied where appropriate
- [ ] Abstractions created only when there are ≥2 implementations
- [ ] Premature generalization avoided

### ✅ Code Quality
- [ ] Follows Go formatting standards (gofmt, goimports)
- [ ] No linting errors (golangci-lint passes)
- [ ] Meaningful variable and function names
- [ ] Appropriate error handling and context propagation
- [ ] No commented-out code or debug statements
- [ ] Proper handling of edge cases and error conditions

## 7. Definition of Ready

Before starting any task, verify:

### ✅ Dependencies Complete
- [ ] All dependent milestones show "Done" status in tracking system
- [ ] Required interfaces are implemented and available
- [ ] Required infrastructure (databases, services) is provisioned
- [ ] Blocking issues or bugs are resolved

### ✅ Architecture Understood
- [ ] Relevant architecture sections reviewed and understood
- [ ] Interface contracts clear and documented
- [ ] Architectural constraints and prohibitions known
- [ ] Any open questions resolved with architect or tech lead

### ✅ Acceptance Criteria Clear
- [ ] Task description understood completely
- [ ] Acceptance criteria are specific, measurable, and testable
- [ ] Definition of Done understood and agreed upon
- [ ] Success criteria well-defined
- [ ] Edge cases and error conditions considered

### ✅ Interfaces Frozen
- [ ] No pending interface changes that would affect this task
- [ ] Interface stability confirmed for the duration of the task
- [ ] Any planned interface changes scheduled after this task
- [ ] Notification process understood for interface changes

### ✅ Resources Available
- [ ] Development environment set up and working
- [ ] Required tools, libraries, and SDKs installed
- [ ] Access to necessary repositories, documentation, and systems
- [ ] Team capacity and support available as needed

### ✅ Estimation Agreed
- [ ] Task effort estimated and agreed upon
- [ ] Dependencies and risks factored into estimate
- [ ] Timeline and scheduling confirmed
- [ ] Contingency for unknowns considered

## 8. Definition of Done

A task is considered complete only when:

### ✅ Task Complete
- [ ] All acceptance criteria met
- [ ] Implementation matches design and requirements
- [ ] No known defects or issues remaining
- [ ] Code is clean, readable, and maintainable

### ✅ Tests Pass
- [ ] All unit tests pass (`go test ./...`)
- [ ] Coverage meets or exceeds target percentage
- [ ] All integration tests pass (if applicable)
- [ ] No test regressions introduced
- [ ] Tests are deterministic and reliable

### ✅ Coverage Acceptable
- [ ] Line coverage ≥90% for complex logic, ≥80% for simple logic
- [ ] Branch coverage ≥80% for complex logic, ≥70% for simple logic
- [ ] No critical paths untested
- [ ] Coverage report reviewed and approved

### ✅ Documentation Complete
- [ ] Code comments present for non-obvious logic
- [ ] Public API fully documented with godoc
- [ ] Package-level documentation accurate and complete
- [ ] Any architecture decision records updated
- [ ] User guides, tutorials, or examples updated if needed

### ✅ Review Approved
- [ ] All code review comments addressed
- [ ] Required number of approvals obtained (minimum 2)
- [ ] Architecture review approval obtained (if required)
- [ ] No outstanding reviewer concerns
- [ ] Review feedback incorporated satisfactorily

### ✅ Validation Passed
- [ ] Build succeeds: `go build ./...` with no errors
- [ ] All tests pass: `go test ./...` with no failures
- [ ] Linting passes: `golangci-lint run` with no errors
- [ ] Security scans pass: no new vulnerabilities introduced
- [ ] Performance benchmarks meet targets
- [ ] Any specific validation scripts pass

## 9. Architecture Drift Prevention

### Identifying Architecture Drift
Watch for these warning signs:
- **Creeping Coupling**: Direct calls between subsystems that should communicate via interfaces
- **Hidden State**: Appearance of global or static variables managing state
- **Interface Erosion**: Type assertions or reflection bypassing interface contracts
- **Dependency Violations**: Layers depending on layers they shouldn't
- **God Objects**: Components taking on too many responsibilities
- **Magic Numbers/Strings**: Undocumented constants affecting behavior
- **Inconsistent Patterns**: Different implementations solving same problem differently
- **Performance Regressions**: Unexplained slowdowns indicating architectural issues
- **Testability Issues**: Increasing difficulty in writing unit tests

### Escalation Process
1. **Self-Check**: Developer runs architecture validation checklist before submitting PR
2. **Peer Review**: Reviewer identifies potential drift during code review
3. **Automated Detection**: CI runs dependency analysis and architectural linting
4. **Architecture Review**: Suspected drift escalated to architecture review board
5. **Investigation**: Team investigates scope and impact of drift
6. **Decision**: 
   - If minor and isolated: fix immediately as part of current work
   - If systemic: create technical debt item for planned remediation
   - If beneficial evolution: propose architecture update through ADR process

### When to Update Architecture Documents
Update `VEDA_AGENT_RUNTIME_ARCHITECTURE.md` only when:
- **Fundamental Change**: Core architectural principles or patterns change
- **Interface Evolution**: Public contracts change in backward-incompatible way
- **New Major Component**: Significant new subsystem added to architecture
- **Pattern Deprecation**: Established pattern is replaced with better alternative
- **Technology Shift**: Fundamental technology choice changes (language, framework, etc.)
- **Scalability Evolution**: Architecture adapts to new scale requirements

Process for updates:
1. **Proposal**: Architect or tech lead drafts change with justification
2. **Review**: Architecture review board reviews and provides feedback
3. **Discussion**: Team discusses implications and alternatives
4. **Decision**: Consensus or decision by architecture lead
5. **Implementation**: Update document and communicate changes
6. **Migration**: Plan and execute migration of existing code if needed

## 10. Technical Debt Policy

### When Technical Debt is Allowed
Technical debt may be incurred intentionally only when:
- **Prototyping**: Spike or proof-of-concept to validate approach
- **Deadline-Driven**: Critical business deadline with agreed-upon payback plan
- **Blocking Dependency**: Waiting on external dependency or decision
- **Learning Opportunity**: Intentional sacrifice to gain knowledge for better solution
- **Emergency Fix**: Critical production issue requiring immediate workaround

### When Technical Debt is NOT Allowed
Technical debt is prohibited when:
- **Architectural Violation**: Creates coupling, violates dependencies, or breaks encapsulation
- **Security Risk**: Introduces potential vulnerabilities or weakens defenses
- **Reliability Threat**: Significantly increases failure risk or reduces fault tolerance
- **Performance Catastrophe**: Causes unacceptable degradation in latency or throughput
- **Maintenance Nightmare**: Makes future changes exponentially more difficult or risky
- **Legal/Compliance Issue**: Violates regulatory requirements or licensing terms

### Documentation Requirements
All intentional technical debt must be documented with:
- **Debt Item**: Clear description of what was compromised and why
- **Rationale**: Business or technical justification for taking on debt
- **Impact Assessment**: Estimated consequences if not addressed
- **Interest Estimate**: "Interest" cost (extra effort) incurred per time period
- **Payback Plan**: Specific steps, estimates, and timeline for remediation
- **Visibility**: Added to project backlog with appropriate priority and tags
- **Review**: Must be approved by tech lead or architecture board

### Debt Management
- **Visibility**: All technical debt items visible in project backlog
- **Regular Review**: Debt reviewed in sprint planning and retrospectives
- **Payback Priority**: 
  - High interest debt paid first
  - Blocking debt resolved before dependent work
  - Scheduled during capacity planning
- **Prevention**: 
  - Definition of Done includes debt prevention checks
  - Code reviews flag potential debt accumulation
  - Architecture reviews prevent architectural debt
- **Metrics**: 
  - Track debt ratio (debt effort / new feature effort)
  - Monitor interest payments (extra time due to debt)
  - Set targets for debt reduction over time

## 11. Decision Record Process

### When ADRs are Required
Create an Architecture Decision Record (ADR) for:
- **Architecturally Significant Decisions**: 
  - Changes to core patterns or principles
  - Selection of major frameworks or technologies
  - Major changes to data flow or communication mechanisms
  - Changes affecting multiple subsystems or teams
- **Irreversible or Costly-to-Reverse Decisions**: 
  - Data model changes requiring migration
  - API changes affecting external consumers
  - Infrastructure changes with significant migration effort
- **Decisions with Significant Trade-offs**: 
  - Performance vs. simplicity trade-offs
  - Consistency vs. availability trade-offs (CAP theorem)
  - Security vs. usability trade-offs
- **Precedent-Setting Decisions**: 
  - First use of a particular pattern or technology
  - Establishment of new conventions or standards
  - Choices that will influence future similar decisions

### When ADRs are NOT Required
No ADR needed for:
- **Implementation Details**: 
  - Choice of algorithm within a function
  - Specific data structure selection (unless architecturally significant)
  - Minor refactoring or cleanup
  - Bug fixes that don't change behavior or architecture
- **Routine Technical Work**: 
  - Dependency updates (unless major version with breaking changes)
  - Configuration changes
  - Test additions or improvements
  - Documentation updates
- **Reversible Decisions**: 
  - Feature flags with short-term lifespan
  - Experimental branches that can be discarded
  - Local optimizations with clear rollback path

### ADR Template
```
# [ADR-XXX]: Title

## Status
Proposed | Accepted | Superseded | Deprecated

## Context
What is the issue that we're seeing that is motivating this decision or change?

## Decision
What is the change that we're proposing and/or doing?

## Consequences
What becomes easier or more difficult to do because of this change?
- Positive consequences (+)
- Negative consequences (-)

## Related Decisions
- ADR-YYY: Related decision that influenced this one
- ADR-ZZZ: Related decision that is influenced by this one

## Notes
Any additional information or links to supporting materials
```

### Approval Process
1. **Draft**: Author creates ADR in `docs/adr/` directory following template
2. **Review**: 
   - Technical review by relevant team leads
   - Architecture review by architecture board
   - Stakeholder review if affecting external interfaces
3. **Discussion**: 
   - Asynchronous discussion in PR comments
   - Synchronous discussion in architecture sync if needed
4. **Decision**: 
   - Approved by architecture lead or consensus of architecture board
   - Explicit rejection with feedback if not approved
5. **Record**: 
   - Upon acceptance: move to `docs/adr/accepted/` and assign ADR number
   - Upon rejection: move to `docs/adr/rejected/` with reason
   - Upon supersession: link to new ADR that replaces it
6. **Communication**: 
   - Announced in team changelog or newsletter
   - Referenced in related implementation tasks
   - Considered in future decision-making processes

## 12. Implementation Principles

### Core Principles
1. **Keep Tasks Small**: 
   - No task should exceed 2 days estimated effort
   - Break down large features into smallest valuable increments
   - Small tasks enable faster feedback and easier course correction

2. **Prefer Composition Over Inheritance**: 
   - Build complex behaviors by combining simple components
   - Use interfaces and dependency injection for flexibility
   - Avoid deep inheritance hierarchies
   - Favor "has-a" relationships over "is-a" relationships

3. **Avoid Premature Optimization**: 
   - Make it work, then make it right, then make it fast
   - Optimize only after profiling identifies actual bottlenecks
   - Prioritize clarity and correctness over premature performance gains
   - Use profiling data to guide optimization efforts

4. **Keep Interfaces Stable**: 
   - Treat published interfaces as contracts
   - Avoid changing interfaces without strong justification
   - Use versioning or extension mechanisms for evolution
   - Deprecate before removing (when applicable)

5. **Implement Only One Task at a Time**: 
   - Finish current task before starting next
   - Avoid context switching between unrelated tasks
   - Complete the full Definition of Done before moving on
   - Use WIP limits to prevent overload

6. **Never Mix Refactoring with Feature Work**: 
   - Separate refactoring commits from feature commits
   - If refactoring needed, do it in dedicated PR first
   - Feature work should implement against clean, stable code
   - Exceptions: minor cleanup directly related to the change being made

### Additional Principles
- **Make Invalid States Unrepresentable**: 
  - Use type system to prevent impossible states
  - Validate inputs at boundaries
  - Use domain-specific types instead of primitives when appropriate
  
- **Fail Fast, Fail Loud**: 
  - Validate inputs early and explicitly
  - Panic only for truly unrecoverable internal inconsistencies
  - Return meaningful errors for recoverable problems
  - Log errors with sufficient context for debugging

- **Be Explicit About Assumptions**: 
  - Document preconditions, postconditions, and invariants
  - Use comments to explain non-obvious business logic
  - Make timing and ordering assumptions explicit in code
  
- **Optimize for Readability**: 
  - Code is read more often than written
  - Favor clarity over cleverness
  - Use descriptive names even if longer
  - Break complex expressions into intermediate variables
  
- **Test Behavior, Not Implementation**: 
  - Tests should verify what the code does, not how it does it
  - Refactoring should not break tests unless behavior changes
  - Focus on public contracts and observable outcomes
  
- **Embrace Simplicity**: 
  - Simple solutions are easier to understand, test, and maintain
  - Avoid unnecessary abstraction layers
  - Remove dead code and unused dependencies aggressively
  - Regularly simplify and refactor accumulating complexity