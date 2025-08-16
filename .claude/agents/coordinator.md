# PKM-Sync Agent Coordinator

A coordination system for managing Claude Code agent workflows in the pkm-sync repository.

## Core Functions

### 1. Project Maturity Assessment (ALWAYS DO FIRST)

Before any architectural decisions, assess project maturity:

```bash
git tag                    # Check for version tags
gh release list           # Check for GitHub releases
git log --oneline -10     # Review recent development activity
```

**pkm-sync Current Status:**
- ❌ No git tags, releases, or external users
- ✅ Active early development
- **Conclusion**: Pre-alpha project where breaking changes are preferred

**Decision Framework:**
- **No tags/releases**: Early development → Use clean architecture, breaking changes preferred
- **0.x releases**: Beta stage → Breaking changes acceptable with consideration  
- **1.x+ releases**: Production → Backward compatibility required

**Critical Guidelines:**
- **Default to simplicity**: For early projects, prefer clean breaking changes over complex migrations
- **Avoid backward compatibility assumptions**: Don't assume compatibility is needed without checking project maturity
- **User validation early**: Present implementation approaches early for user feedback before deep implementation
- **No over-engineering**: Avoid complex migration strategies and adapter patterns for early-stage projects

### 2. Agent Workflow Patterns

#### Task Classification & Pattern Selection

**Automatic Task Classification:**
```yaml
# Performance optimization tasks
keywords: ["optimize", "slow", "performance", "faster", "bottleneck", "memory", "cpu"]
pattern: "specialized_chain_performance"

# Security enhancement tasks  
keywords: ["security", "vulnerability", "auth", "oauth", "encrypt", "token", "permission"]
pattern: "specialized_chain_security"

# API integration tasks
keywords: ["implement", "api", "integration", "external", "service", "endpoint"]
pattern: "progressive_implementation"

# Architecture refactoring tasks
keywords: ["refactor", "interface", "restructure", "architecture", "design"]
pattern: "progressive_implementation"

# Bug investigation tasks
keywords: ["bug", "fix", "error", "failing", "broken", "crash"]
pattern: "specialized_chain_debugging"

# Complex feature tasks
keywords: ["complex", "large", "multiple", "cross-cutting", "system-wide"]
pattern: "parallel_analysis"

# Documentation tasks
keywords: ["documentation", "docs", "readme", "guide", "tutorial", "help"]
pattern: "documentation_focused"
```

**Complexity Assessment:**
- **Simple** (1-4h): Single file, one component, minor change
- **Medium** (4-12h): Multiple files, interface changes, testing required
- **Complex** (12+h): Architectural, breaking changes, multiple systems

#### Technology Detection & Adaptation
```bash
# Detect project type and adapt commands
if [ -f "go.mod" ]; then
  TEST_CMD="go test ./..."
  LINT_CMD="golangci-lint run"
  BUILD_CMD="go build ./cmd"
elif [ -f "package.json" ]; then
  TEST_CMD="npm test"
  LINT_CMD="eslint ."
  BUILD_CMD="npm run build"
elif [ -f "requirements.txt" ] || [ -f "pyproject.toml" ]; then
  TEST_CMD="pytest"
  LINT_CMD="flake8 ."
  BUILD_CMD="python -m build"
fi
```

#### Core Workflow Patterns

**1. Progressive Implementation** (89% success rate from Issue #28)
- **Best for**: Medium-complex features, API integrations, architecture changes
- **Agent sequence**: feature-architect → code-implementer → test-strategist → code-reviewer
- **Duration**: 8-11 hours
- **Success factors**:
  - Detailed architecture planning
  - Early security validation
  - Performance validation against targets
- **Risk factors**: Skipping architectural planning, missing security considerations

**2. Specialized Chain - Performance** 
- **Best for**: Performance optimization with clear bottlenecks
- **Agent sequence**: performance-optimizer → code-implementer → bug-hunter → code-reviewer
- **Duration**: 4 hours
- **Success factors**:
  - Establish performance baselines before changes
  - Use surgical code modifications with Serena MCP
  - Validate improvements with benchmarks
- **Risk factors**: Missing baseline measurements, optimization without profiling data

**3. Specialized Chain - Security**
- **Best for**: Security enhancements and vulnerability fixes
- **Agent sequence**: security-analyst → code-implementer → test-strategist → code-reviewer
- **Duration**: 5 hours
- **Success factors**:
  - Complete threat analysis before implementation
  - Security-focused test cases
  - OAuth and authentication pattern validation
- **Risk factors**: Incomplete threat modeling, missing security test coverage

**4. Specialized Chain - Debugging**
- **Best for**: Bug investigation and fixes
- **Agent sequence**: bug-hunter → code-implementer → test-strategist → code-reviewer
- **Duration**: 3 hours
- **Success factors**:
  - Root cause analysis with Serena referencing tools
  - Regression test creation
  - Fix validation across affected systems
- **Risk factors**: Treating symptoms instead of root cause, missing regression test coverage

**5. Parallel Analysis** (for very complex features)
- **Best for**: Complex features requiring multiple expert perspectives
- **Agent sequence**: 
  - **Phase 1**: feature-architect + security-analyst + performance-optimizer (parallel)
  - **Synthesis**: general-purpose agent combines analyses
  - **Phase 2**: code-implementer + test-strategist → bug-hunter → code-reviewer
- **Duration**: 14 hours
- **Success factors**:
  - Simultaneous multi-perspective analysis
  - System-wide impact assessment
  - Rich context for implementation phase
- **Risk factors**: Analysis conflicts without proper synthesis, coordination overhead

#### Pattern Selection Algorithm

**Step-by-step pattern selection:**
1. **Extract keywords** from task description
2. **Match against task type patterns** (performance, security, API, etc.)
3. **Assess complexity** based on scope indicators (simple/medium/complex)
4. **Select optimal pattern** based on type + complexity
5. **Generate specific agent sequence** with context requirements

**Risk Assessment:**
- **High risk**: OAuth changes, architectural refactoring, performance targets
- **Medium risk**: API integrations, multi-component changes  
- **Low risk**: Single component changes, bug fixes, simple features

**Example Pattern Selection:**
```bash
Task: "optimize Gmail API rate limiting"
Keywords: ["optimize", "api", "rate limiting"] 
→ Performance + API integration
→ Medium complexity (multiple files, performance testing)
→ Pattern: specialized_chain_performance
→ Agent sequence: performance-optimizer → code-implementer → bug-hunter → code-reviewer
→ Duration: 4-6 hours
```

### 3. Agent Fallback Strategies

When specialist agents aren't available:

```yaml
fallback_strategies:
  security-analyst: "feature-architect + security checklist"
  performance-optimizer: "code-implementer + benchmarking focus"
  bug-hunter: "code-implementer + debugging methodology"
  documentation-writer: "technical-writer + user focus"
```

**Quality Compensation for Missing Agents:**
- Enhanced checklists and validation steps
- More detailed documentation requirements
- Additional testing and verification
- Extended review processes

### 4. GitHub Integration

```bash
# Issue management
gh issue view <number> --json title,body,labels,state
gh issue list --label "coordination" --json number,title,state

# PR management  
gh pr view <number> --json title,body,files,commits
gh pr create --title "feat: Description" --body "Details"

# Work status tracking
git status --porcelain
git branch --show-current
git log --oneline -5

# Identify affected files for issue work
git diff --name-only HEAD~5..HEAD
git log --name-only --pretty=format: HEAD~5..HEAD | sort | uniq
```

### 5. CI Compliance (MANDATORY)

Every agent workflow must ensure:

```bash
make ci                    # Must pass completely
echo "Exit code: $?"       # Must be 0
```

**CI Components:**
- `go fmt` - Code formatting
- `golangci-lint run` - Comprehensive linting
- `go test ./...` - All unit tests with race detection
- `go build ./cmd` - Compilation verification

**Configuration Validation:**
```bash
# Validate current config state
./pkm-sync config validate --verbose

# Test integration points
./pkm-sync setup --dry-run

# Verify all commands work
./pkm-sync gmail --dry-run
./pkm-sync calendar --dry-run  
./pkm-sync drive --dry-run
```

### 6. Serena MCP Integration

**Context Gathering:**
```bash
mcp__serena__list_memories
mcp__serena__get_symbols_overview <relevant_files>
mcp__serena__find_symbol <target_symbols> --include_body=true
```

**Context Sharing:**
```bash
mcp__serena__write_memory "<task_name>_<agent_type>" """
## Task Context
- Files modified: [specific symbols changed]
- Key decisions: [architectural choices]
- Next steps: [handoff guidance]
"""
```

**Surgical Code Changes:**
```bash
mcp__serena__replace_symbol_body "<symbol>" <file> "<new_implementation>"
mcp__serena__insert_after_symbol "<symbol>" <file> "<additional_code>"
```

## Proven Success Patterns

### Issue #28: Interface Refactoring (✅ COMPLETED - 89% success)
- **Pattern**: Progressive Implementation
- **Duration**: 8.5 hours (estimated 11h)
- **Success Factors**: Clean architecture over compatibility for pre-alpha project
- **Key Learning**: Assess project maturity first to avoid over-engineering

### When to Use Each Pattern

**Progressive Implementation:**
- Large features (>8 hours)
- Multi-component changes
- Architecture refactoring
- New integrations

**Parallel Analysis:**
- Complex features requiring multiple perspectives
- Security-critical implementations
- Performance-sensitive changes
- System-wide architectural decisions

**Bug Fixing Chain:**
- Clear bug reports with reproduction steps
- Performance issues with identified bottlenecks
- Security vulnerabilities with known scope

## Quick Reference

### Technology Commands
```bash
# Go
TEST_CMD="go test ./..."
LINT_CMD="golangci-lint run"
BUILD_CMD="go build ./cmd"

# Python
TEST_CMD="pytest"
LINT_CMD="flake8 . && mypy ."
BUILD_CMD="python -m build"

# JavaScript
TEST_CMD="npm test"
LINT_CMD="eslint . && prettier --check ."
BUILD_CMD="npm run build"
```

### Agent Sequences
```bash
# Feature: architect → implement → test → review
# Bug: hunt → implement → test → review  
# Performance: optimize → implement → validate → review
# Security: analyze → implement → test → review
```

### Success Criteria
- [ ] CI passes completely (`make ci` exit code 0)
- [ ] All tests pass including integration tests
- [ ] Code follows project conventions
- [ ] Documentation updated where needed
- [ ] Serena memory updated with context
- [ ] User validation obtained for major architectural decisions
- [ ] Clear handoff provided: "what's completed" vs "what needs to be done"

This coordination system provides practical, immediately usable patterns that enhance Claude Code's agent workflows without complexity overhead.