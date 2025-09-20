# Project Brief: pre-push

## Overview
A cross-platform, configurable Git `pre-push` hook runner written in Go, designed to replace legacy bash scripts with a modern, DAG-based executor that provides built-in checks and Buildfab integration.

## Core Requirements
- **Cross-platform compatibility**: Linux, Windows, macOS (amd64/arm64)
- **DAG-based execution**: Parallel execution of checks with explicit dependencies
- **Built-in checks**: Git state validation, version checking, and custom actions
- **Configuration-driven**: YAML-based configuration similar to GitHub Actions
- **Self-installing**: Automatic installation as Git pre-push hook
- **Version management**: Integration with version-go library for semantic versioning
- **Error handling**: Configurable error policies (stop/warn) per step
- **Reproducible builds**: Static binaries with deterministic builds

## Goals
- **Replace bash scripts**: Modernize the existing bash-based pre-push implementation
- **Improve performance**: Parallel execution of independent checks
- **Enhance maintainability**: Structured Go codebase with clear separation of concerns
- **Provide flexibility**: Configurable checks and actions via YAML
- **Ensure reliability**: Comprehensive error handling and validation
- **Enable CI/CD integration**: Seamless integration with existing build pipelines

## Project Scope
**In Scope:**
- CLI tool with `test`, `list-uses` commands - ✅ IMPLEMENTED
- YAML configuration parser with variable interpolation - ✅ IMPLEMENTED
- DAG executor with parallel step execution - ✅ IMPLEMENTED
- Built-in Git checks (untracked, uncommitted, modified files) - ✅ IMPLEMENTED
- Version validation and checking - ✅ IMPLEMENTED
- Self-installation as Git hook - ⏳ PENDING
- Test suite with race detection - ✅ IMPLEMENTED
- Cross-platform build system (CMake + Conan + GoReleaser) - ✅ CONFIGURED
- Comprehensive documentation and testing - ✅ IMPLEMENTED

**Out of Scope:**
- GUI interface (CLI only)
- Real-time monitoring dashboard
- Cloud-based configuration management
- Integration with non-Git version control systems
- Advanced workflow orchestration beyond pre-push hooks