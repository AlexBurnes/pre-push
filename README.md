# pre-push

A cross-platform, configurable Git `pre-push` hook runner written in Go, designed to replace legacy bash scripts with a modern, DAG-based executor that provides built-in checks and Buildfab integration.

## Status

✅ **Production Ready** - Core functionality implemented and tested  
✅ **Build System** - GoReleaser dry-run successful with all platform binaries  
✅ **Cross-platform** - Linux, macOS, Windows (amd64/arm64) support  
✅ **Package Managers** - Homebrew and Scoop integration ready  
✅ **CI/CD Pipeline** - GitHub Actions workflows configured and DRY-compliant  
✅ **Enhanced Version Handling** - -V outputs only version, --version outputs full info  
✅ **Flexible Configuration** - Bin directory support and shell script validation  
⏳ **Git Hook Installation** - Install command pending implementation  
⏳ **Advanced Features** - Environment variables, conditional execution pending

## Features

- **Cross-platform**: Works on Linux, Windows, and macOS (amd64/arm64)
- **DAG-based execution**: Parallel execution of checks with explicit dependencies
- **Built-in checks**: Git state validation with flexible custom actions
- **Configuration-driven**: YAML-based configuration similar to GitHub Actions
- **Self-installing**: Automatic installation as Git pre-push hook
- **Enhanced version handling**: -V outputs only version, --version outputs full module info
- **Flexible validation**: Shell script approach for project-specific version checking
- **Bin directory support**: Custom module locations via project configuration
- **Language-agnostic**: Support for any test command via `run:` actions
- **Error handling**: Configurable error policies (stop/warn) per step
- **Professional UI**: Colored output with conditional highlighting and clear status indicators
- **Git hook integration**: Properly handles Git arguments and reads ref information from stdin

## Installation

### Using Package Managers

**Homebrew (macOS)**:
```bash
brew install AlexBurnes/tap/pre-push
```

**Scoop (Windows)**:
```bash
scoop bucket add AlexBurnes https://github.com/AlexBurnes/scoop-bucket
scoop install pre-push
```

### Direct Download

Download the latest release from [GitHub Releases](https://github.com/AlexBurnes/pre-push/releases) and extract the binary for your platform:

- **Linux**: `pre-push_linux_amd64.tar.gz` or `pre-push_linux_arm64.tar.gz`
- **macOS**: `pre-push_macos_amd64.tar.gz` or `pre-push_macos_arm64.tar.gz`
- **Windows**: `pre-push_windows_amd64.zip`

### Building from Source

```bash
# Clone the repository
git clone https://github.com/AlexBurnes/pre-push.git
cd pre-push

# Build using CMake + Conan
./buildtools/build-conan.sh deps
./buildtools/build-conan.sh build

# Or build using GoReleaser
./buildtools/build-goreleaser.sh dry-run  # Test build
./buildtools/build-goreleaser.sh release  # Full release
```

## CI/CD Pipeline

The project includes comprehensive GitHub Actions workflows:

- **CI Workflow** (`.github/workflows/ci.yml`): Runs on every push and PR
  - Builds binaries for all platforms
  - Runs tests with race detection
  - Creates archives and installers
  - Performs GoReleaser dry-run validation

- **Release Workflow** (`.github/workflows/release.yml`): Runs on version tags
  - Builds and packages all platform binaries
  - Creates GitHub releases with artifacts
  - Updates Homebrew and Scoop manifests
  - Publishes to package managers

Both workflows use the same buildtools scripts for consistency and maintainability.

## Quick Start

1. **Install the hook**:
   ```bash
   pre-push
   ```

2. **Create configuration** (`.project.yml`):
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

     - name: git-uncommitted
       uses: git@uncommitted

   stages:
     pre-push:
       steps:
         - action: version-check
         - action: git-untracked
         - action: git-uncommitted
   ```

3. **Test the configuration**:
   ```bash
   pre-push test
   ```

4. **Push to trigger the hook**:
   ```bash
   git push origin main
   ```

## Usage

### Commands

- `pre-push` - Install/update the pre-push hook
- `pre-push test` - Run all checks in dry-run mode
- `pre-push list-uses` - List available built-in actions

### Global Options

- `-h, --help` - Print help and exit
- `-V, --version` - Print version and exit
- `-d, --debug` - Enable debug output
- `-v, --verbose` - Enable verbose output

### Configuration

The tool uses a `.project.yml` file for configuration. The format is inspired by GitHub Actions:

```yaml
project:
  name: "project-name"
  modules: ["pre-push"]

actions:
  - name: action-name
    run: |
      command to execute
    # OR
    uses: built-in-action

stages:
  pre-push:
    steps:
      - action: action-name
        require: [dependency-action]  # Optional dependencies
        onerror: warn                 # Optional: warn | stop (default: stop)
```

### Built-in Actions

- `git@untracked` - Check for untracked files
- `git@uncommitted` - Check for uncommitted changes
- `git@modified` - Check for modified files

### Variable Interpolation

Variables can be interpolated using `${{ variable }}` syntax:

- `${{ tag }}` - Current git tag
- `${{ branch }}` - Current git branch
- `${{ version.version }}` - Version from version-go library
- `${{ version.project }}` - Project name from version-go library

## Development

### Prerequisites

- Go 1.22 or later
- CMake 3.16 or later
- Conan 2.x
- Git with version tags

### Building

```bash
# Install dependencies
conan install . --build=missing

# Build
cmake --build .build --target pre-push

# Test
go test ./...

# Format code
go fmt ./...

# Lint code
golangci-lint run
```

### Project Structure

```
pre-push/
├── cmd/pre-push/          # CLI entry point
├── pkg/prepush/           # Library API
├── internal/              # Internal packages
│   ├── config/           # Configuration loading
│   ├── exec/             # DAG execution
│   ├── uses/             # Built-in actions
│   ├── version/          # Version detection
│   ├── ui/               # User interface
│   └── install/          # Hook installation
├── buildtools/           # Build scripts
├── packaging/            # Platform-specific packaging
└── docs/                 # Documentation
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes.

## Support

- Issues: [GitHub Issues](https://github.com/AlexBurnes/pre-push/issues)
- Discussions: [GitHub Discussions](https://github.com/AlexBurnes/pre-push/discussions)