# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.0] - 2025-09-21

### Added
- **Version Library Integration**: Integrated `github.com/AlexBurnes/version-go/pkg/version` v0.8.22 library
  - Replaced CLI utility calls with direct Go library integration
  - Added support for `${{ version.version }}`, `${{ version.project }}`, `${{ version.module }}`, `${{ version.modules }}` variables
  - Enhanced variable detection with proper version parsing and validation
  - Improved version detection reliability and cross-platform compatibility
- **Enhanced Variable Support**: Extended variable interpolation system
  - Added `version.version` variable for current version (e.g., "v1.3.0")
  - Added `version.project` variable for project name (e.g., "buildfab-pre-push")
  - Added `version.module` variable for module name (e.g., "pre-push")
  - Added `version.modules` variable for comma-separated modules list
- **Version Integration Module**: Created `internal/version/integration.go`
  - Direct integration with version-go library for semantic versioning
  - Automatic version detection from VERSION file
  - Project and module detection from go.mod file
  - Comprehensive error handling and validation

### Changed
- **Variable Detection**: Enhanced `DetectGitVariables()` function
  - Added version library variable detection alongside existing Git variables
  - Improved variable resolution with proper error handling
  - Updated configuration system to support all version library variables
- **Project Configuration**: Updated .project.yml to use version library variables
  - Replaced basic `${{ tag }}` with comprehensive `${{ version.version }}` support
  - Enhanced variable interpolation for better project configuration flexibility

### Documentation
- **Memory Bank Updates**: Updated all memory bank files with version library integration
  - Updated `activeContext.md` with recent changes and current work focus
  - Updated `progress.md` to mark version library integration as completed
  - Updated `systemPatterns.md` with version library integration path
  - Updated `techContext.md` with version library dependency information
- **Technical Documentation**: Enhanced system architecture documentation
  - Added version library integration to system patterns
  - Updated dependency information in technical context
  - Documented new variable support and integration patterns

## [Unreleased]

### Added
- **Buildfab Integration**: Integrated [buildfab](https://github.com/AlexBurnes/buildfab) v0.5.0 as the core DAG execution engine
  - Replaced custom DAG implementation with buildfab's proven execution engine
  - Enhanced parallel execution with advanced dependency resolution
  - Improved error handling with comprehensive error types
  - Added support for buildfab's advanced workflow features
- **BuildfabExecutor**: New executor implementation using buildfab API
  - Simplified executor architecture leveraging buildfab's capabilities
  - Maintained compatibility with existing pre-push configuration format
  - Enhanced performance and reliability through buildfab's optimized execution
- **Documentation**: Added comprehensive buildfab integration documentation
  - Created `docs/Buildfab-integration.md` with detailed integration guide
  - Updated README and project specification to reflect buildfab integration
  - Documented API usage and configuration compatibility

### Changed
- **Dependencies**: Updated go.mod to include buildfab v0.5.0 dependency
- **Executor Architecture**: Refactored executor to use buildfab's DAG execution engine
- **Error Handling**: Enhanced error handling through buildfab's comprehensive error types
- **Performance**: Improved execution performance through buildfab's optimized parallel execution

### Technical Details
- **API Integration**: Uses buildfab.RunStage() and buildfab.RunAction() for execution
- **Configuration**: Maintains full compatibility with existing .project.yml format
- **Error Types**: Leverages buildfab's ConfigurationError, ExecutionError, DependencyError, and VariableError
- **Testing**: Updated test suite to work with buildfab integration

## [1.2.0] - 2025-01-27

### Added
- **PRD for parallel execution improvements**: Created comprehensive Product Requirements Document for enhanced parallel execution
  - Ordered output display (steps shown in project.yml declaration order)
  - Continue running independent steps on failure
  - Add SKIPPED status for dependent steps
  - Enhanced error reporting and summary
- **PRD parallel execution improvements implementation**: Implemented all PRD requirements for better user experience
  - Steps now display in project.yml declaration order instead of execution order
  - Independent steps continue running even when other steps fail
  - Added SKIPPED status for steps that can't run due to failed dependencies
  - Enhanced summary includes SKIPPED count and improved error reporting
  - Streaming output displays results as soon as they complete while maintaining declaration order
  - Dependency-aware display ensures dependent steps wait for their requirements before showing results

### Changed
- **Project settings integration**: Updated .project.yml to use scripts/version utility for version-check and version-greatest actions
- **Version validation**: version-check action now uses scripts/version-check command, version-greatest uses scripts/version greatest command
- **Executor behavior**: Implemented streaming output that displays results as they complete while maintaining declaration order
- **Error handling**: Removed early termination on step failure to allow independent steps to continue running
- **UI enhancements**: Added SKIPPED status support with gray color and ⊘ icon
- **Summary reporting**: Enhanced summary to include SKIPPED count alongside OK, WARN, and ERROR counts
- **Parallel execution**: Implemented true parallel execution with dependency-aware result display
- **Streaming display**: Results are shown immediately when steps complete, but dependent steps wait for their requirements

## [1.1.0] - 2025-01-27

### Added
- **Enhanced version flag handling**: -V now outputs only version, --version outputs full module info
- **Bin directory support**: Project configuration now supports custom module locations via `bin` field
- **Flexible version validation**: Shell script approach using `scripts/version modules` for project-specific validation
- **Language-agnostic test execution**: Support for any test command via `run:` actions with output suppression
- **Conditional execution support**: Added `only` and `if` conditions for step execution based on version type
- **Project rules documentation**: Comprehensive documentation for pre-push validation requirements

### Changed
- **Custom action execution**: `run:` actions now suppress output by default, showing only OK/ERROR status
- **Configuration schema**: Extended to support bin directory and conditional execution fields (`only`, `if`)
- **Error handling**: Improved error messages with exact reproduction commands for custom actions
- **Version handling**: Binary now reads version from VERSION file instead of hardcoded value

### Removed
- **version@check-modules-version built-in action**: Replaced with flexible shell script approach

### Documentation
- **Pre-push rules**: Created comprehensive documentation for validation requirements
- **Project specification**: Updated to include enhanced configuration options
- **Memory bank**: Updated with new architecture decisions and implementation patterns

## [1.0.3] - 2025-09-20

### Fixed
- **Test compilation errors**: Fixed missing UI interface methods in mockUI test struct
- **Test coverage**: All tests now pass with race detection enabled
- **Mock implementation**: Added all required UI interface methods to mockUI for proper test execution

## [1.0.2] - 2025-09-20

### Added
- **Enhanced pre-push hook output**: CLI utility name, version, and project information now displayed at start
- **Better error message formatting**: Built-in action error messages now include inline reproduction commands
- **Improved user experience**: Error messages show "to manually check run:" with properly indented git commands

### Fixed
- **Cleaner error output**: Removed unwanted Usage information when built-in actions fail
- **Consistent error formatting**: All built-in actions now use 4-space indentation and simple git commands

### Changed
- **Error message format**: Built-in actions now show inline reproduction commands instead of separate sections
- **UI improvements**: Enhanced error message display with better formatting and alignment

## [1.0.1] - 2025-09-20

### Added
- **Enhanced UI coloring**: Improved stage headers, stage results, and summary with conditional coloring
- **Better visual feedback**: Items with count > 0 get colored, items with count = 0 show in gray
- **Professional appearance**: Stage headers in cyan, success in green, errors in red, warnings in yellow

### Fixed
- **Git hook argument parsing**: Fixed argument parsing to ignore Git arguments and read from stdin when called by Git
- **Hook detection**: Enhanced detection logic to distinguish between manual commands and Git hook calls
- **Command line parsing**: Pre-push now correctly ignores Git arguments (like "origin master --tags") when called as a hook

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

## [0.1.0] - 2025-09-20

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