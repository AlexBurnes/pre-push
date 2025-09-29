## [1.6.9] - 2025-09-29

### Fixed
- **Shell Option Support**: Fixed shell option not being respected in action configuration
  - **Buildfab Integration**: Updated pre-push to use buildfab configuration directly instead of custom conversion
    - Modified BuildfabExecutor to accept buildfab.Config directly instead of prepush.Config
    - Removed custom configuration conversion that was losing shell field information
    - Updated main.go to use buildfab.LoadConfig() directly for proper shell option handling
    - Shell option now properly respected for actions with `shell: bash` configuration
    - Resolves issue where test_bash action was not using bash shell as specified
    - Ensures consistent behavior between pre-push and buildfab commands

## [1.6.8] - 2025-09-29

### Fixed
- **Static Builds for Linux and Darwin**: Fixed static builds for cross-platform compatibility
  - **GoReleaser Configuration**: Updated .goreleaser.yml with proper static build settings
    - Added CGO_ENABLED=0 to environment variables to disable CGO and enable static builds
    - Added -extldflags "-static" to ldflags to tell the linker to create static binaries
    - Ensured cross-platform static binaries work correctly on Linux and Darwin platforms
    - Resolves issues with dynamic linking dependencies in static environments
    - Improves binary portability and reduces external dependencies

## [1.6.7] - 2025-09-25

### Fixed
- **GitHub Actions CI Workflow**: Fixed CI workflow artifact upload issue by using proper CMake install-current target
  - **CI Workflow**: Updated .github/workflows/ci.yml to use CMake install-current target
    - Fixed "No files were found with the provided path: bin/pre-push" error
    - Replaced manual file copying with proper CMake target usage
    - Added verification step to ensure binary installation succeeds
    - Enhanced error reporting with directory listing for debugging
    - CI workflow now follows project configuration patterns from .project.yml
    - Ensures bin/pre-push binary is properly created for artifact upload

## [1.6.6] - 2025-09-25

### Fixed
- **Git Hook Binary Installation**: Fixed Git hook installation to copy binary directly instead of shell script wrapper
  - **Install Package**: Updated to copy binary directly to .git/hooks/pre-push
    - Removed shell script wrapper approach that was creating bash scripts instead of binaries
    - Updated isHookUpToDate() method to compare binary MD5 hashes directly
    - Added calculateHookMD5() method for proper binary comparison
    - Removed unused strings import from install package
    - Verified binary and Git hook now have identical MD5 hashes
    - Git hook is now a proper ELF binary instead of shell script

## [1.6.5] - 2025-09-25

### Changed
- **Buildfab Library Update**: Updated buildfab library to latest stable release
  - **Buildfab Library**: Updated from v0.15.1 to v0.16.0
    - Latest features and improvements from buildfab project
    - Enhanced execution capabilities and performance optimizations
    - Improved error handling and output formatting
    - Maintained full compatibility with existing configuration
    - All tests passing with new buildfab version

## [1.6.4] - 2025-01-27

### Changed
- **Version Bump**: Version incremented for buildfab library update preparation

## [1.6.3] - 2025-01-27

### Fixed
- **Platform Naming Consistency**: Fixed darwin/macos naming inconsistency in build system
  - Removed darwin->macos replacement in GoReleaser name_template
  - Updated CMakeLists.txt platform detection to use darwin instead of macos
  - Fixed buildtools scripts to use darwin platform naming consistently
  - Updated documentation to reference darwin binaries instead of macos
  - Ensured all installer scripts and URLs use consistent darwin naming
  - Fixed README.md download links to use pre-push_darwin_* archives
  - Updated packaging/macos/README.md to reference darwin binary names

### Added
- **Smart Hook Update System**: Elegant self-updating Git hook system
  - Pre-push now manages its own updates using MD5 hash comparison
  - CLI mode automatically checks and updates Git hook if binary differs
  - Git hook mode executes without update checks for optimal performance
  - No complex shell scripts needed - all logic contained in pre-push binary

## [1.6.2] - 2025-01-27

### Fixed
- **Buildfab Build System**: Fixed buildfab build error and improved CMake configuration
  - Removed unnecessary config/version.yaml copying from CMake install targets
  - Fixed CMakeLists.txt working directories from cmd/version to cmd/pre-push
  - Removed broken bootstrap-version dependency
  - Updated package name from "version" to "pre-push"
  - Fixed test commands to use ./... for comprehensive testing
  - CMake install target now uses detected platform/architecture automatically
  - Buildfab build system works correctly with platform detection

## [1.6.1] - 2025-01-27

### Fixed
- **Include Functionality**: Fixed buildfab include functionality not working in pre-push tool
  - Updated pre-push tool to use buildfab.LoadConfig() which properly handles include statements
  - Created LoadWithBuildfab() function in config package to convert buildfab.Config to prepush.Config
  - Removed manual variable resolution since buildfab handles it automatically
  - check-pre-push-releases action now properly loaded from config/update-checking-actions.yml
  - All tests passing with include functionality working correctly

### Changed
- **Library Updates**: Updated core dependencies to latest stable releases
  - **Buildfab Library**: Updated from v0.10.0 to v0.15.1
    - Latest features and improvements from buildfab project
    - Enhanced execution capabilities and performance optimizations
    - Improved error handling and output formatting
    - Maintained full compatibility with existing configuration
  - **Version Library**: Updated from v1.1.1 to v1.2.2
    - Enhanced version detection and parsing capabilities
    - Improved stability and cross-platform compatibility
    - Better error handling for version operations
    - Latest features from version-go project

## [1.6.0] - 2025-01-27

### Added
- **Platform Variable Detection**: Comprehensive platform variable detection and variable substitution
  - Built-in platform variables: `platform`, `arch`, `os`, `os_version`, `cpu` (simple names)
  - Environment variable support: All environment variables available as `env.*` variables
  - Git variables: `tag`, `branch` for current repository state
  - Version variables: `version`, `project`, `module`, `modules` (simple names)
  - Variable interpolation in all action `run:` commands and step conditions
  - Cross-platform variable detection for Linux, macOS, and Windows
- **Buildfab Library Upgrade**: Updated from v0.9.0 to v0.10.0
  - Latest buildfab features with built-in platform variable support
  - Enhanced variable interpolation capabilities
  - Improved platform detection across all supported operating systems
  - Maintained full compatibility with existing configuration
- **Version Library Upgrade**: Updated from v0.8.22 to v1.1.1
  - Latest version detection capabilities
  - Enhanced version parsing and validation
  - Improved cross-platform version detection

### Changed
- **Enhanced Variable Substitution**: All actions and steps now support variable interpolation
  - Actions can use `${{ os }}`, `${{ version }}`, `${{ tag }}`, `${{ branch }}` variables (simple names)
  - Step conditions support platform-specific logic using variables
  - Environment variables accessible as `${{ env.VARIABLE_NAME }}` in all contexts
  - Automatic variable detection and substitution during configuration loading

## [1.5.0] - 2025-01-27

### Added
- **Verbose Mode Support**: Comprehensive verbose and debug mode implementation for enhanced user experience
  - Environment variable support: `PRE_PUSH_VERBOSE=1` enables verbose mode
  - Environment variable support: `PRE_PUSH_DEBUG=1` enables debug mode
  - Project.yml configuration support: `verbose: true` and `debug: true` in pre-push stage
  - Enhanced buildfab integration with detailed command execution output
  - Debug output control (only shown when debug mode is enabled)
  - Works correctly in both CLI commands and Git hooks
  - Clean output for normal usage with detailed verbose output when requested
- **Buildfab Library Upgrade**: Updated from v0.8.18 to v0.9.0
  - Latest buildfab features and improvements
  - Enhanced verbose output capabilities
  - Improved command execution details display
  - Maintained full compatibility with existing configuration

### Changed
- **Enhanced Output Formatting**: Verbose mode now shows detailed command execution
  - Command execution details: `💻 action-name` for each step
  - Full command output display with execution details
  - Enhanced visual feedback for better debugging experience
  - Maintained clean output for normal usage

## [1.4.9] - 2025-01-27

### Added
- **Version v1.4.9**: New release with version display and retrieval improvements

## [1.4.8] - 2025-09-23

### Added
- **Version v1.4.8**: New release with buildfab library upgrade and version display fix

### Changed
- **Buildfab Library Upgrade**: Updated from v0.8.11 to v0.8.18
  - Enhanced output formatting and execution capabilities
  - Improved step-by-step execution with latest buildfab features
  - Maintained full compatibility with existing .project.yml configuration format
  - All tests passing with race detection enabled
  - Updated documentation to reflect new buildfab version

### Fixed
- **Version Display and Retrieval**: Fixed pre-push utility version display and improved version retrieval system
  - Pre-push now shows "pre-push v1.4.8" at first line like buildfab project
  - Enhanced UI output to ensure custom header is displayed before buildfab execution
  - Separated CLI version (compiled-in) from project version (version-go library)
  - CLI version now uses compiled-in version from ldflags (build time)
  - Project version now uses github.com/AlexBurnes/version-go library (runtime)
  - Added NewBuildfabExecutorWithCLIVersion constructor to pass CLI version
  - Updated getVersion() method to use version-go library instead of VERSION file
  - Version display now consistent with expected format and proper source separation

## [1.4.7] - 2025-09-23

### Added
- **Version v1.4.7**: New release with improvements and bug fixes


### Changed
- **Buildfab Library Upgrade**: Updated from v0.8.7 to v0.8.11
  - Enhanced output formatting and execution capabilities
  - Improved step-by-step execution with latest buildfab features
  - Maintained full compatibility with existing .project.yml configuration format
  - All tests passing with race detection enabled
  - Updated documentation to reflect new buildfab version

## [1.4.6] - 2025-09-23

### Fixed
- **Version Compilation**: Fixed version handling to use compiled-in version instead of reading VERSION file at runtime
  - Updated main.go to use appVersion variable set via ldflags from GoReleaser
  - Removed unused strings import after removing VERSION file reading
  - GoReleaser automatically detects version from Git tags and compiles it correctly
  - Both -V and --version flags now show correct compiled-in version
  - Version shows as "1.4.6-SNAPSHOT-<commit>" for snapshot builds and "1.4.6" for releases
  - VERSION file is still used for development process but not read at runtime

## [1.4.5] - 2025-09-23

### Added
- **Version v1.4.5**: New release with buildfab library upgrade

### Changed
- **Buildfab Library Upgrade**: Updated from v0.7.3 to v0.8.7
  - Enhanced output formatting and execution capabilities
  - Improved step-by-step execution with latest buildfab features
  - Maintained full compatibility with existing .project.yml configuration format
  - All tests passing with race detection enabled


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