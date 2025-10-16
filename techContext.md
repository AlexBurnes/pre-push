# Tech Context: pre-push

## Technologies Used
- **Go 1.22+**: Primary language with CGO disabled for static binaries - IMPLEMENTED
- **Cobra**: CLI framework for command-line interface - IMPLEMENTED
- **YAML**: Configuration format using gopkg.in/yaml.v3 - IMPLEMENTED
- **Git**: Version control integration via os/exec (not go-git) - IMPLEMENTED
- **CMake**: Cross-platform build system configuration - CONFIGURED
- **Conan**: C++ package manager for Go toolchain management - CONFIGURED
- **GoReleaser**: Automated release and packaging tool - CONFIGURED
- **GitHub Actions**: CI/CD pipeline for automated testing and releases - ✅ CONFIGURED

## Development Setup
**Prerequisites:**
- Go 1.22 or later
- CMake 3.16 or later
- Conan 2.x
- Git with version tags
- Make (for build automation)

**Local Development:**
```bash
# Clone and setup
git clone <repository>
cd pre-push

# Install dependencies
conan install . --build=missing

# Build and test
cmake --build .build --target pre-push
go test ./... -race

# Run locally
./bin/pre-push test
./bin/pre-push list-uses
./bin/pre-push --version
```

**IDE Configuration:**
- VS Code with Go extension
- Go modules enabled
- Format on save with gofmt
- Linting with golangci-lint

## Technical Constraints
- **CGO disabled**: All builds must be static binaries for maximum compatibility
- **Go 1.22 minimum**: Leveraging modern Go features (generics, improved error handling)
- **Cross-platform**: Must work on Linux, Windows, macOS (amd64/arm64)
- **Minimal dependencies**: Prefer standard library, avoid heavy frameworks
- **Reproducible builds**: Deterministic builds with -trimpath and -ldflags
- **Memory efficient**: Handle large repositories without excessive memory usage
- **Fast execution**: Pre-push checks must complete quickly to avoid developer friction

## Dependencies
**Core Dependencies:**
- `gopkg.in/yaml.v3`: YAML configuration parsing
- `github.com/AlexBurnes/version-go/pkg/version`: Version parsing and validation library v1.5.0
- `github.com/AlexBurnes/buildfab`: DAG execution engine v0.25.0 with enhanced variable system
- `golang.org/x/sync/errgroup`: Parallel execution management
- `github.com/spf13/cobra`: CLI framework (optional, may use standard flag package)

**Build Dependencies:**
- `conanfile-golang.py`: Go toolchain via Conan
- `CMakeLists.txt`: Cross-platform build configuration
- `.goreleaser.yml`: Release automation configuration

**Development Dependencies:**
- `golangci-lint`: Code linting and quality checks
- `go test`: Unit and integration testing
- `gofmt`: Code formatting
- `goimports`: Import organization

## Tool Usage Patterns
**Build System:**
- **CMake**: Primary build configuration with platform detection
- **Conan**: Manages Go toolchain and build dependencies
- **GoReleaser**: Handles cross-compilation and release packaging
- **Make**: Build automation and development workflows

**Development Workflow:**
- **go mod**: Dependency management and versioning
- **go test**: Comprehensive testing with race detection
- **golangci-lint**: Code quality and style enforcement
- **git hooks**: Pre-commit validation and pre-push testing

**Release Process:**
- **Version tagging**: Semantic versioning with git tags
- **GoReleaser**: Automated builds for all platforms
- **GitHub Releases**: Binary distribution and changelog management
- **Scoop/Homebrew**: Package manager integration for easy installation

**Testing Strategy:**
- **Unit tests**: Individual component testing with table-driven tests
- **Integration tests**: End-to-end testing with temporary Git repositories
- **Race detection**: `go test -race` for concurrency validation
- **Coverage reporting**: `go test -coverprofile` for coverage analysis