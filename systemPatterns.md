# System Patterns: pre-push

## System Architecture
**Layered Architecture with Clear Separation of Concerns**

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Layer (cmd/pre-push)                 │
├─────────────────────────────────────────────────────────────┤
│                  Library API (pkg/)                         │
├─────────────────────────────────────────────────────────────┤
│                Internal Packages                            │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │   config/   │    exec/    │    uses/    │   version/  │  │
│  │             │             │             │             │  │
│  │   install/  │     ui/     │             │             │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                External Dependencies                        │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │   Git API   │ version-go  │   YAML     │   Context   │  │
│  │             │  library    │  parsing   │  handling   │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Key Technical Decisions
- **DAG-based execution**: Parallel execution of independent checks with explicit dependencies
- **YAML configuration**: GitHub Actions-inspired config format for familiarity and flexibility
- **Context-aware execution**: All operations respect context cancellation and timeouts
- **Error policy system**: Configurable stop/warn behavior per check step
- **Variable interpolation**: GitHub-style `${{ }}` syntax for dynamic configuration
- **Self-installing hooks**: Automatic installation and update of Git pre-push hooks
- **Static binary distribution**: CGO-disabled builds for maximum compatibility

## Design Patterns in Use
- **Command Pattern**: CLI commands as separate handlers (test, list-uses, install)
- **Strategy Pattern**: Different execution strategies for built-in vs. custom actions
- **Observer Pattern**: Step execution monitoring and status reporting
- **Factory Pattern**: Action creation based on configuration type (uses: vs run:)
- **Builder Pattern**: DAG construction and execution plan building
- **Template Method**: Common execution flow with customizable steps

## Component Relationships
**Core Flow:**
1. **CLI** → **Config Parser** → **DAG Builder** → **Executor** → **UI Renderer**
2. **Config Parser** → **Variable Resolver** → **Action Factory** → **Step Creation**
3. **Executor** → **Parallel Scheduler** → **Step Runners** → **Result Aggregator**
4. **Step Runners** → **Built-in Actions** (git@*) or **External Commands** (run:)

**Key Interactions:**
- **config/**: Parses YAML, validates schema, resolves variables
- **exec/**: Builds DAG, schedules execution, manages concurrency
- **uses/**: Implements built-in actions (git@untracked, git@uncommitted, etc.)
- **version/**: Integrates with version-go library for semantic versioning
- **ui/**: Handles colored output, status reporting, error formatting
- **install/**: Manages Git hook installation and updates

## Built-in Actions Architecture
**Single Group Design - IMPLEMENTED:**
- **Group**: `git@` only (no multiple groups planned)
- **Actions**: `git@untracked`, `git@uncommitted`, `git@modified` - ALL IMPLEMENTED
- **File Structure**: All actions in registry.go with proper interface implementation
- **Registration**: Explicit registration calls in New() function - IMPLEMENTED
- **Interface**: Complete Runner interface with Run, GetRepro, GetHelp, GetName methods - IMPLEMENTED
- **Reproduction**: Git command strings for manual debugging - IMPLEMENTED
- **Help**: Descriptive text for each action - IMPLEMENTED
- **Error Context**: Detailed error messages with file lists - IMPLEMENTED

## Critical Implementation Paths
**1. Configuration Loading Path - IMPLEMENTED:**
```
YAML File → Parser → Schema Validation → Variable Resolution → Action Creation
```
✅ All steps implemented with comprehensive validation and error handling

**2. DAG Execution Path - IMPLEMENTED:**
```
Actions → DAG Builder → Cycle Detection → Execution Plan → Parallel Scheduler → Results
```
✅ Complete implementation with topological sorting and cycle detection

**3. Git Hook Integration Path - PARTIALLY IMPLEMENTED:**
```
Git Push → Hook Invocation → Stdin Parsing → Branch/Tag Detection → Check Execution
```
✅ Branch/Tag detection implemented, hook invocation pending

**4. Error Handling Path - IMPLEMENTED:**
```
Step Failure → Error Policy Check → Continue/Stop Decision → Result Aggregation → Exit Code
```
✅ Complete error policy system with stop/warn behavior

**5. Self-Installation Path - PENDING:**
```
Binary Execution → Version Check → Hook Installation → Shim Creation → Verification
```
⏳ Install command not yet implemented