# Pre-push Rules and Validation

## Overview
This document defines the mandatory pre-push validation rules for the pre-push project. These rules ensure code quality, build integrity, and proper version management before code is pushed to the repository.

## Core Validation Rules

### 1. Build Validation
**Rule**: The project module must be buildable and output the correct version.

**Implementation**: 
- Built-in action: `build@validate`
- Validates that `go build` succeeds for the main module
- Verifies that the built binary outputs a valid version string
- Version must start with 'v' and contain at least one digit

**Configuration Example**:
```yaml
actions:
  - name: build-validation
    uses: build@validate

stages:
  pre-push:
    steps:
      - action: build-validation
```

### 2. Test Execution
**Rule**: All tests must pass before pushing to repository with release tags.

**Implementation**:
- Built-in action: `test@run`
- Executes `go test ./... -v -race` with race detection
- Validates that tests actually run and produce results
- Fails if any test fails or no test results are found

**Configuration Example**:
```yaml
actions:
  - name: test-execution
    uses: test@run

stages:
  pre-push:
    steps:
      - action: test-execution
        only: [release]  # Only run for release versions
```

## Conditional Execution

### Version-Based Conditions
The pre-push utility supports conditional execution based on version type:

**Supported Conditions**:
- `release`: Standard release versions (e.g., v1.0.0, v2.1.3)
- `prerelease`: Pre-release versions (e.g., v1.0.0-alpha, v2.1.3-beta.1)
- `patch`: Patch releases (e.g., v1.0.1, v2.1.4)
- `minor`: Minor releases (e.g., v1.1.0, v2.2.0)
- `major`: Major releases (e.g., v2.0.0, v3.0.0)

**Configuration Examples**:

```yaml
# Run tests only for release versions
- action: test-execution
  only: [release]

# Run build validation for all versions
- action: build-validation

# Run additional checks for prerelease versions
- action: prerelease-checks
  only: [prerelease]

# Run major release checks
- action: major-release-checks
  only: [major]
```

### If Conditions
Support for conditional execution using `if` statements:

```yaml
# Run only if version type is release
- action: release-checks
  if: "version.type == 'release'"

# Run only if version type is prerelease
- action: prerelease-checks
  if: "version.type == 'prerelease'"
```

## Complete Configuration Example

```yaml
project:
  name: "pre-push"
  modules: ["pre-push"]

actions:
  - name: build-validation
    uses: build@validate

  - name: test-execution
    uses: test@run

  - name: git-untracked
    uses: git@untracked

  - name: git-uncommitted
    uses: git@uncommitted

  - name: git-modified
    uses: git@modified

stages:
  pre-push:
    steps:
      # Always run build validation
      - action: build-validation
      
      # Run tests for all versions
      - action: test-execution
        require: [build-validation]
      
      # Run Git checks for all versions
      - action: git-untracked
        require: [test-execution]
      
      - action: git-uncommitted
        require: [git-untracked]
      
      - action: git-modified
        require: [git-uncommitted]
        onerror: warn
```

## Error Handling

### Error Policies
- `stop` (default): Stop execution on error, fail the pre-push
- `warn`: Continue execution, mark as warning

### Exit Codes
- `0`: All checks passed or only warnings occurred
- `1`: At least one check failed with error status

## Integration with Git Hooks

### Automatic Installation
The pre-push utility automatically installs itself as a Git pre-push hook:

```bash
# Install the hook
./pre-push

# Test the configuration
./pre-push test

# List available actions
./pre-push list-uses
```

### Hook Behavior
- Reads Git ref information from stdin
- Detects current branch and tag information
- Executes configured stages based on conditions
- Provides detailed output with colored status indicators

## Best Practices

### 1. Dependency Management
- Always specify `require` dependencies for proper execution order
- Build validation should run before test execution
- Git checks should run after build and test validation

### 2. Conditional Execution
- Use `only` conditions for version-specific checks
- Use `if` conditions for complex logic
- Test both release and prerelease scenarios

### 3. Error Handling
- Use `onerror: warn` for non-critical checks
- Use `onerror: stop` (default) for critical checks
- Provide clear error messages and reproduction steps

### 4. Performance
- Keep build and test execution fast
- Use parallel execution where possible
- Avoid unnecessary checks for development branches

## Troubleshooting

### Common Issues

1. **Build Validation Fails**
   - Check that `go build` works manually
   - Verify VERSION file exists and is valid
   - Ensure all dependencies are available

2. **Test Execution Fails**
   - Run `go test ./... -v -race` manually
   - Check for race conditions in tests
   - Verify test data and fixtures

3. **Conditional Execution Not Working**
   - Check version detection logic
   - Verify `only` and `if` conditions
   - Test with different version types

### Debug Mode
Use debug mode for detailed troubleshooting:

```bash
./pre-push test --debug
```

This provides:
- Detailed execution logs
- Variable expansion information
- Timing information
- Environment details

## Migration from Bash Scripts

### Legacy Script Compatibility
The pre-push utility maintains compatibility with existing bash-based pre-push scripts:

1. **Configuration Migration**: Convert bash logic to YAML configuration
2. **Action Mapping**: Map bash commands to built-in actions or `run` commands
3. **Dependency Management**: Convert sequential execution to DAG-based execution
4. **Error Handling**: Convert bash error handling to pre-push error policies

### Example Migration

**Before (Bash)**:
```bash
#!/bin/bash
go build ./cmd/pre-push
go test ./... -race
git status --porcelain
```

**After (YAML)**:
```yaml
actions:
  - name: build-validation
    uses: build@validate
  - name: test-execution
    uses: test@run
  - name: git-untracked
    uses: git@untracked

stages:
  pre-push:
    steps:
      - action: build-validation
      - action: test-execution
        require: [build-validation]
      - action: git-untracked
        require: [test-execution]
```

## Future Enhancements

### Planned Features
- Environment variable support
- Matrix execution for multiple Go versions
- Custom action definitions
- Integration with external CI/CD systems
- Advanced version type detection
- Performance metrics and reporting

### Extension Points
- Custom built-in actions
- External action providers
- Custom condition evaluators
- Custom UI renderers
- Custom error handlers