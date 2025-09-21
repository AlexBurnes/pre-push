# System Patterns: pre-push

## System Architecture
**Layered Architecture with Buildfab Integration**

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
│  │   buildfab  │ version-go  │   YAML     │   Context   │  │
│  │   v0.7.2    │  library    │  parsing   │  handling   │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Key Technical Decisions
- **Buildfab Integration**: Powered by [buildfab](https://github.com/AlexBurnes/buildfab) v0.7.2 for advanced DAG execution and workflow automation
- **DAG-based execution**: Parallel execution of independent checks with explicit dependencies via buildfab
- **YAML configuration**: GitHub Actions-inspired config format for familiarity and flexibility
- **Context-aware execution**: All operations respect context cancellation and timeouts
- **Error policy system**: Configurable stop/warn behavior per check step
- **Variable interpolation**: GitHub-style `${{ }}` syntax for dynamic configuration
- **Self-installing hooks**: Automatic installation and update of Git pre-push hooks
- **Static binary distribution**: CGO-disabled builds for maximum compatibility
- **PRD Requirements**: Enhanced parallel execution with ordered output and failure resilience
  - Steps displayed in project.yml declaration order (not execution order)
  - Independent steps continue running even when other steps fail
  - Dependent steps marked as SKIPPED when their dependencies fail

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
- **version/**: Integrates with `github.com/AlexBurnes/version-go/pkg/version` v0.8.22 for semantic versioning
- **ui/**: Handles colored output, status reporting, error formatting
- **install/**: Manages Git hook installation and updates

## Built-in Actions Architecture
**Single Group Design - SIMPLIFIED:**
- **Group**: `git@` - Git repository checks (IMPLEMENTED)
- **Actions**: 
  - `git@untracked`, `git@uncommitted`, `git@modified` - IMPLEMENTED
- **File Structure**: All actions in registry.go with proper interface implementation
- **Registration**: Explicit registration calls in New() function - IMPLEMENTED
- **Interface**: Complete Runner interface with Run, GetRepro, GetHelp, GetName methods - IMPLEMENTED
- **Reproduction**: Command strings for manual debugging - IMPLEMENTED
- **Help**: Descriptive text for each action - IMPLEMENTED
- **Error Context**: Clear error messages with reproduction commands - IMPLEMENTED
- **Version Validation**: Handled via flexible shell script approach using project configuration

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

**6. Test Execution Path - IMPLEMENTED:**
```
Test Files → Mock Setup → Interface Implementation → Test Execution → Race Detection → Results
```
✅ Complete test suite with race detection and comprehensive coverage

**7. Version Module Validation Path - IMPLEMENTED:**
```
Project Config → scripts/version modules → Shell Script → Module -V Check → Version Comparison
```
✅ Flexible shell script approach using project-specific configuration

**8. Enhanced Version Flag Path - IMPLEMENTED:**
```
Command Line → Flag Detection → Version Source (VERSION file) → Format Output → Display
```
✅ -V outputs only version, --version outputs full module info

**9. Version Library Integration Path - IMPLEMENTED:**
```
VERSION File → version.Parse() → Version Validation → Variable Detection → Interpolation
```
✅ Complete integration with `github.com/AlexBurnes/version-go/pkg/version` v0.8.22

**10. Buildfab v0.7.2 Integration Path - IMPLEMENTED:**
```
.project.yml → buildfab.SimpleRunner → Step Execution → Status Output → Summary
```
✅ Enhanced step output with buildfab v0.7.2 SimpleRunner interface
✅ Clean, professional step-by-step execution with proper status icons
✅ Single stage execution as requested (not individual action execution)
✅ Fixed version format mismatch and template variable resolution issues