# Buildfab Integration

The pre-push utility is powered by [buildfab](https://github.com/AlexBurnes/buildfab) v0.5.0, a comprehensive Go library for building and running project automation stages and actions defined in YAML configuration.

## Overview

Buildfab provides the core DAG execution engine for pre-push, handling:

- **DAG Construction**: Building execution graphs from YAML configuration
- **Cycle Detection**: Preventing circular dependencies
- **Topological Sorting**: Determining optimal execution order
- **Parallel Execution**: Running independent steps concurrently
- **Dependency Management**: Ensuring proper step ordering
- **Error Handling**: Managing failures and skipping dependent steps

## Integration Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    pre-push CLI                             │
├─────────────────────────────────────────────────────────────┤
│                BuildfabExecutor                             │
│  ┌─────────────────┬─────────────────┬─────────────────┐    │
│  │   UI Layer      │  Config Layer   │  Version Layer  │    │
│  └─────────────────┴─────────────────┴─────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                    buildfab v0.5.0                         │
│  ┌─────────────────┬─────────────────┬─────────────────┐    │
│  │   DAG Engine    │  Action Runner  │  Result Handler │    │
│  └─────────────────┴─────────────────┴─────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## API Usage

The pre-push utility uses buildfab's main API functions:

### Stage Execution

```go
opts := &buildfab.RunOptions{
    ConfigPath: ".project.yml",
    Verbose:    true,
    Debug:      false,
    WorkingDir: ".",
}

err := buildfab.RunStage(ctx, "pre-push", opts)
```

### Action Execution

```go
err := buildfab.RunAction(ctx, "version-check", opts)
```

### Step Execution

```go
// Run a single step with dependencies
opts.WithRequires = true
err := buildfab.RunStageStep(ctx, "pre-push", "version-check", opts)
```

## Configuration Compatibility

The pre-push `.project.yml` format is fully compatible with buildfab's configuration format:

```yaml
project:
  name: "buildfab-pre-push"
  modules: ["pre-push"]

actions:
  - name: version-check
    run: scripts/version check

  - name: git-untracked
    uses: git@untracked

stages:
  pre-push:
    steps:
      - action: version-check
      - action: git-untracked
        require: [version-check]
```

## Error Handling

Buildfab provides comprehensive error types that pre-push leverages:

- **ConfigurationError**: Invalid YAML or configuration issues
- **ExecutionError**: Step execution failures
- **DependencyError**: Circular dependencies or missing dependencies
- **VariableError**: Variable interpolation issues

## Benefits of Buildfab Integration

### 1. **Advanced DAG Execution**
- Sophisticated dependency resolution
- Cycle detection and prevention
- Optimal parallel execution scheduling

### 2. **Robust Error Handling**
- Comprehensive error types and messages
- Graceful failure handling
- Detailed execution context

### 3. **Performance Optimization**
- Efficient parallel execution
- Memory-optimized DAG traversal
- Minimal overhead for simple workflows

### 4. **Extensibility**
- Easy to add new action types
- Flexible configuration options
- Plugin architecture support

### 5. **Maintainability**
- Well-tested core execution engine
- Clear separation of concerns
- Comprehensive documentation

## Migration from Custom DAG

The integration replaces the previous custom DAG implementation with buildfab's proven execution engine:

### Before (Custom DAG)
- Custom DAG construction and cycle detection
- Manual topological sorting
- Custom parallel execution logic
- Basic error handling

### After (Buildfab Integration)
- Professional DAG execution engine
- Advanced dependency resolution
- Optimized parallel execution
- Comprehensive error handling

## Future Enhancements

With buildfab as the execution engine, pre-push can easily support:

- **Advanced Workflow Features**: Conditional execution, retries, timeouts
- **Plugin System**: Custom action types and extensions
- **Monitoring**: Execution metrics and performance tracking
- **Distributed Execution**: Remote step execution capabilities

## Dependencies

The buildfab integration adds the following dependency:

```go
require github.com/burnes/buildfab v0.5.0
```

This dependency provides:
- Core DAG execution engine
- Action runner framework
- Result handling and reporting
- Error management system

## Testing

The buildfab integration is tested through:

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end workflow testing
- **Error Handling Tests**: Failure scenario validation
- **Performance Tests**: Execution efficiency verification

## Documentation

For detailed buildfab API documentation, see:
- [buildfab Library API Reference](https://github.com/AlexBurnes/buildfab/blob/master/docs/Library.md)
- [buildfab Project Documentation](https://github.com/AlexBurnes/buildfab)

## Support

For issues related to:
- **pre-push functionality**: Create an issue in this repository
- **buildfab DAG execution**: Create an issue in the [buildfab repository](https://github.com/AlexBurnes/buildfab)
- **Integration problems**: Check the compatibility matrix and version requirements