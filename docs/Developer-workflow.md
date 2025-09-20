# Developer Workflow

This document describes the development workflow for the pre-push project.

## Prerequisites

- Go 1.22 or later
- CMake 3.16 or later
- Conan 2.x
- Git with version tags
- Make (for build automation)

## Development Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/AlexBurnes/pre-push.git
   cd pre-push
   ```

2. **Install dependencies**:
   ```bash
   # Install Conan dependencies
   conan install . --build=missing
   
   # Or use the build script
   ./buildtools/build-conan.sh deps
   ```

3. **Build the project**:
   ```bash
   # Using CMake
   cmake --build .build --target pre-push
   
   # Or using the build script
   ./buildtools/build-conan.sh build
   ```

4. **Run tests**:
   ```bash
   go test ./...
   ```

## Development Workflow

### 1. Feature Development

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write code following Go conventions
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**:
   ```bash
   # Run tests
   go test ./...
   
   # Run with race detection
   go test ./... -race
   
   # Run with coverage
   go test ./... -coverprofile=coverage.out
   go tool cover -html=coverage.out -o coverage.html
   ```

4. **Format and lint**:
   ```bash
   # Format code
   go fmt ./...
   
   # Lint code
   golangci-lint run
   ```

5. **Build and test**:
   ```bash
   # Build the project
   ./buildtools/build-conan.sh build
   
   # Test the CLI
   ./bin/pre-push --help
   ./bin/pre-push test
   ```

### 2. Pull Request Process

1. **Push your changes**:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   git push origin feature/your-feature-name
   ```

2. **Create a pull request**:
   - Use a descriptive title
   - Include a detailed description
   - Reference any related issues
   - Ensure all checks pass

3. **Code review**:
   - Address review feedback
   - Update documentation if needed
   - Ensure tests pass

### 3. Release Process

1. **Update version**:
   ```bash
   # Update VERSION file
   echo "v1.2.3" > VERSION
   
   # Update CHANGELOG.md
   # Update README.md if needed
   ```

2. **Create release**:
   ```bash
   # Tag the release
   git tag v1.2.3
   git push origin v1.2.3
   
   # Build and release
   ./buildtools/build-goreleaser.sh release
   ```

## Code Standards

### Go Code Style

- Use 4 spaces for indentation
- Maximum line length of 120 characters
- Follow Go naming conventions
- Add docstrings for all exported functions and types
- Use descriptive variable and function names

### Error Handling

- Use explicit error handling with context wrapping
- Use `fmt.Errorf` for error wrapping
- Return errors, don't panic in normal flow
- Provide helpful error messages

### Testing

- Write unit tests for all public functions
- Use table-driven tests where appropriate
- Test error conditions
- Use `t.Run` for subtests
- Aim for high test coverage

### Documentation

- Update README.md for user-facing changes
- Update CHANGELOG.md for all changes
- Update memory bank files for architectural changes
- Add GoDoc comments for all exported symbols

## Build System

### CMake + Conan

The project uses CMake with Conan for dependency management:

```bash
# Install dependencies
conan install . --build=missing

# Configure
cmake -B .build

# Build
cmake --build .build --target pre-push
```

### GoReleaser

For releases, the project uses GoReleaser:

```bash
# Setup
./buildtools/build-goreleaser.sh setup

# Snapshot build
./buildtools/build-goreleaser.sh snapshot

# Release build
./buildtools/build-goreleaser.sh release
```

## Memory Bank

The project uses a memory bank system for documentation:

- `projectbrief.md` - Project overview and requirements
- `productContext.md` - User experience and success metrics
- `activeContext.md` - Current work focus and decisions
- `systemPatterns.md` - Architecture and design patterns
- `techContext.md` - Technologies and constraints
- `progress.md` - Current status and what's left to build

Update these files when making significant changes to the project.

## Troubleshooting

### Common Issues

1. **Conan not found**:
   ```bash
   pip install conan
   ```

2. **Go version too old**:
   ```bash
   # Install Go 1.22 or later
   # Update PATH to include new Go installation
   ```

3. **Build failures**:
   ```bash
   # Clean and rebuild
   rm -rf .build bin/
   ./buildtools/build-conan.sh clean
   ./buildtools/build-conan.sh deps
   ./buildtools/build-conan.sh build
   ```

4. **Test failures**:
   ```bash
   # Run tests with verbose output
   go test ./... -v
   
   # Run specific test
   go test ./internal/config -v
   ```

### Getting Help

- Check the [GitHub Issues](https://github.com/AlexBurnes/pre-push/issues)
- Join the [GitHub Discussions](https://github.com/AlexBurnes/pre-push/discussions)
- Review the memory bank files for project context