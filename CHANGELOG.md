# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Enhanced UI coloring**: Improved stage headers, stage results, and summary with conditional coloring
- **Better visual feedback**: Items with count > 0 get colored, items with count = 0 show in gray
- **Professional appearance**: Stage headers in cyan, success in green, errors in red, warnings in yellow

### Changed
- **UI system**: Enhanced colored output with conditional highlighting for better user experience

### Added
- **Project initialization**: Complete project structure following Go conventions
- **Memory bank documentation**: Comprehensive project documentation system
- **Go module setup**: go.mod and go.sum files for dependency management
- **CLI structure**: cmd/pre-push/main.go with test, list-uses, and install commands
- **Library API**: pkg/prepush/ with core functionality and interfaces
- **Internal packages**: config, exec, uses, version, ui, install packages
- **Configuration system**: YAML-based configuration with variable interpolation
- **DAG executor**: Parallel execution engine with dependency management
- **Built-in actions**: git@untracked, git@uncommitted, git@modified runners
- **Version detection**: Git tag and branch detection functionality
- **Complete core implementation**: All essential functionality implemented and working
- **Comprehensive testing**: Unit tests, integration tests, and E2E tests with race detection
- **Custom action execution**: Full shell command execution for run: actions
- **Variable interpolation**: Complete ${{ }} syntax with Git tag/branch detection
- **UI system**: Complete colored output, status reporting, and progress display
- **Error handling**: Complete error policies with stop/warn behavior
- **Git integration**: Real Git repository operations for status checking

### Fixed
- **GitHub Actions workflows**: Updated CI and release workflows to use 'pre-push' instead of 'version'
- **DRY principle violation**: Replaced inline code in CI workflow with proper buildtools script calls
- **Workflow consistency**: Both CI and release workflows now use the same buildtools scripts
- **Version management**: Fixed hardcoded version in check-version-status script to use current version from utility
- **Build system**: Updated all buildtools scripts to use 'pre-push' instead of 'version' references
- **CMakeLists.txt**: Removed broken bootstrap-version target and corrected dependencies
- **GoReleaser configuration**: Fixed project name, binary paths, URLs, and test commands for pre-push project

### Changed
- **CI workflow structure**: Replaced 70+ lines of inline code with proper buildtools script calls
- **Workflow maintainability**: Improved code organization and reduced duplication
- **Project references**: Updated all build scripts and configuration files to reference 'pre-push' instead of 'version'
- **Archive naming**: Changed archive naming patterns from 'version_*' to 'pre-push_*'
- **Binary naming**: Updated binary names from 'version' to 'pre-push' in all build scripts
- **Package manager integration**: Updated Homebrew formula and Scoop manifest for pre-push project

### Documentation
- **Memory bank updates**: Updated all memory bank documents to reflect current implementation status
- **README updates**: Added status section, package manager installation instructions, CI/CD pipeline documentation, and current build information
- **Build documentation**: Updated build instructions with GoReleaser dry-run and release commands
- **Workflow documentation**: Added comprehensive CI/CD pipeline documentation to README

## [0.1.0] - 2025-01-27

### Added
- **Initial release**: Basic project structure and documentation
- **Go module**: Initial go.mod with required dependencies
- **CLI framework**: Command-line interface using Cobra library
- **Configuration parser**: YAML configuration loading and validation
- **DAG executor**: Basic execution engine with dependency management
- **Built-in actions**: Placeholder implementations for Git checks with clarified requirements
- **Version detection**: Basic version and branch detection
- **Hook installation**: Basic Git hook installation functionality
- **Build system**: CMake + Conan + GoReleaser build configuration
- **Cross-platform support**: Linux, Windows, macOS (amd64/arm64)