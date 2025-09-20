# Implementation Summary

This document provides a high-level overview of the pre-push implementation.

## Architecture Overview

The pre-push tool follows a layered architecture with clear separation of concerns:

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

## Core Components

### 1. CLI Layer (`cmd/pre-push`)

The command-line interface provides three main commands:

- **`pre-push`** - Install/update the Git pre-push hook
- **`pre-push test`** - Run all checks in dry-run mode
- **`pre-push list-uses`** - List available built-in actions

### 2. Library API (`pkg/prepush`)

The library provides the core data structures and interfaces:

- **`Config`** - Configuration structure with validation
- **`Action`** - Individual action definition
- **`Stage`** - Collection of steps to execute
- **`Result`** - Execution result with status
- **`Executor`** - Interface for executing checks

### 3. Internal Packages

#### Configuration (`internal/config`)

- **YAML parsing** - Loads and validates `.project.yml` files
- **Variable interpolation** - Resolves `${{ variable }}` syntax
- **Schema validation** - Ensures configuration correctness

#### Execution (`internal/exec`)

- **DAG builder** - Constructs execution dependency graph
- **Parallel scheduler** - Executes independent steps in parallel
- **Error handling** - Manages stop/warn policies per step
- **Result aggregation** - Collects and reports execution results

#### Built-in Actions (`internal/uses`)

- **Registry** - Manages available built-in actions with explicit registration
- **Single group** - Only `git@` group with its actions
- **File structure** - Each action in separate source file
- **Git checks** - `git@untracked`, `git@uncommitted`, `git@modified`
- **Simple interface** - Run, GetRepro, GetHelp, GetName methods
- **Reproduction hints** - Simple command strings for manual debugging
- **No auto-discovery** - Developer manually registers actions

#### Version Detection (`internal/version`)

- **Git integration** - Detects current tag and branch
- **Version validation** - Validates version format
- **Comparison** - Compares versions for greatest check

#### User Interface (`internal/ui`)

- **Colored output** - Status indicators with colors
- **Progress reporting** - Real-time execution feedback
- **Error formatting** - Clear error messages and reproduction instructions

#### Hook Installation (`internal/install`)

- **Git hook management** - Installs and updates pre-push hooks
- **Version checking** - Ensures hook is up to date
- **Cross-platform** - Works on all supported platforms

## Execution Flow

### 1. Configuration Loading

```
YAML File → Parser → Schema Validation → Variable Resolution → Action Creation
```

### 2. DAG Construction

```
Actions → DAG Builder → Cycle Detection → Execution Plan → Parallel Scheduler
```

### 3. Execution

```
Ready Steps → Parallel Execution → Result Collection → Dependency Unlocking → Continue
```

### 4. Error Handling

```
Step Failure → Error Policy Check → Continue/Stop Decision → Result Aggregation
```

## Key Features

### DAG-based Execution

The tool uses a Directed Acyclic Graph (DAG) to manage dependencies between steps:

- **Parallel execution** - Independent steps run simultaneously
- **Dependency management** - Steps wait for their dependencies
- **Cycle detection** - Prevents infinite loops
- **Error propagation** - Failed steps can stop dependent steps

### Configuration System

YAML-based configuration with GitHub Actions-inspired syntax:

```yaml
project:
  name: "my-project"
  modules: ["pre-push"]

actions:
  - name: version-check
    run: |
      version check ${{tag}}

  - name: git-untracked
    uses: git@untracked

stages:
  pre-push:
    steps:
      - action: version-check
      - action: git-untracked
        require: [version-check]
        onerror: warn
```

### Variable Interpolation

Supports GitHub-style variable interpolation:

- `${{ tag }}` - Current git tag
- `${{ branch }}` - Current git branch
- `${{ version.version }}` - Version from version-go library

### Error Policies

Configurable error handling per step:

- **`stop`** (default) - Stop execution on error
- **`warn`** - Continue execution with warning

## Build System

### CMake + Conan

- **Cross-platform builds** - Linux, Windows, macOS
- **Dependency management** - Go toolchain via Conan
- **Static binaries** - CGO disabled for maximum compatibility

### GoReleaser

- **Automated releases** - GitHub releases with binaries
- **Cross-compilation** - All platforms in one build
- **Package managers** - Scoop, Homebrew integration

## Testing Strategy

### Unit Tests

- **Table-driven tests** - Comprehensive test coverage
- **Mock interfaces** - Isolated component testing
- **Error conditions** - Test failure scenarios

### Integration Tests

- **End-to-end testing** - Full workflow validation
- **Temporary repositories** - Git state testing
- **Configuration validation** - YAML parsing and validation

### Performance Testing

- **Parallel execution** - Verify DAG performance
- **Large repositories** - Test with many files
- **Memory usage** - Monitor resource consumption

## Security Considerations

- **Input validation** - All inputs are validated
- **Path sanitization** - Prevent directory traversal
- **Command execution** - Safe execution of external commands
- **File permissions** - Proper file mode handling

## Future Enhancements

### Planned Features

- **Custom action execution** - Support for `run:` commands
- **Matrix builds** - Parallel execution with different parameters
- **Conditional execution** - `if` conditions for steps
- **Environment variables** - `env` configuration
- **Reusable actions** - External action files

### Performance Improvements

- **Caching** - Cache results between runs
- **Incremental execution** - Only run changed steps
- **Parallel file operations** - Concurrent Git operations

### Usability Improvements

- **Interactive mode** - Step-by-step execution
- **Progress bars** - Visual progress indication
- **Configuration validation** - Real-time config checking
- **Better error messages** - More helpful error descriptions