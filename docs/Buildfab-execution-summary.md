# Buildfab Execution Path Summary

## Overview

Pre-push integrates with buildfab v0.29.1 library as its core DAG execution engine. This document summarizes the execution flow and lists all buildfab library methods called.

## High-Level Execution Flow

```
User Action (git push / pre-push test)
    ↓
main() - Entry Point
    ↓
Route to Handler (runGitHook / runTest / runInstall)
    ↓
buildfab.LoadConfig() - Load YAML configuration
    ↓
BuildfabExecutor - Pre-push wrapper
    ↓
buildfab.NewSimpleRunner() - Create runner
    ↓
runner.RunStage() - Execute stage
    ↓
Results - Success/Failure
```

## Buildfab Library Methods Called

### Configuration Methods

#### 1. `buildfab.LoadConfig(path string) (*buildfab.Config, error)`
**Location**: `cmd/pre-push/main.go` (lines 366, 676)
**Purpose**: Load and parse YAML configuration file with include support
**Usage**:
```go
buildfabConfig, err := buildfab.LoadConfig(".project.yml")
```

**What it does**:
- Reads YAML configuration file
- Processes include directives recursively
- Validates configuration structure
- Returns buildfab.Config struct

**Called in**:
- `runGitHook()` - When running as Git hook
- `runTest()` - When running test command

---

#### 2. `(*buildfab.Config).GetStage(name string) (buildfab.Stage, bool)`
**Location**: `internal/exec/buildfab_executor.go` (line 152, 221)
**Purpose**: Retrieve stage configuration by name
**Usage**:
```go
stage, exists := e.config.GetStage(stageName)
```

**What it does**:
- Looks up stage by name in configuration
- Returns stage struct and existence flag
- Used for validation before execution

**Called in**:
- `BuildfabExecutor.RunStage()` - Before stage execution

---

#### 3. `(*buildfab.Config).GetAction(name string) (buildfab.Action, bool)`
**Location**: `internal/exec/buildfab_executor.go` (line 225)
**Purpose**: Retrieve action configuration by name
**Usage**:
```go
action, exists := e.config.GetAction(step.Action)
```

**What it does**:
- Looks up action by name in configuration
- Returns action struct and existence flag
- Used for debug output and validation

**Called in**:
- `BuildfabExecutor.RunStage()` - For debug logging

---

### Variable Methods

#### 4. `buildfab.GetPlatformVariables() *buildfab.PlatformVariables`
**Location**: `internal/exec/buildfab_executor.go` (line 330)
**Purpose**: Get platform-specific variables
**Usage**:
```go
platformVars := buildfab.GetPlatformVariables()
variables["platform"] = platformVars.Platform
variables["arch"] = platformVars.Arch
variables["os"] = platformVars.OS
variables["os_version"] = platformVars.OSVersion
variables["cpu"] = fmt.Sprintf("%d", platformVars.CPU)
```

**Returns**:
```go
type PlatformVariables struct {
    Platform  string  // e.g., "linux-amd64", "darwin-arm64"
    Arch      string  // e.g., "amd64", "arm64"
    OS        string  // e.g., "linux", "darwin", "windows"
    OSVersion string  // OS version string
    CPU       int     // Number of CPU cores
}
```

**Called in**:
- `BuildfabExecutor.GetAllVariables()` - Variable collection

---

#### 5. `buildfab.AddPlatformVariables(vars map[string]string) map[string]string`
**Location**: `internal/exec/buildfab_executor.go` (line 415)
**Purpose**: Add platform variables with "platform." prefix
**Usage**:
```go
variables = buildfab.AddPlatformVariables(variables)
```

**What it does**:
- Adds variables like `platform.os`, `platform.arch`, `platform.cpu`
- Provides backward compatibility with prefixed names
- Supplements simple variable names

**Called in**:
- `BuildfabExecutor.GetAllVariables()` - After collecting all other variables

---

#### 6. `buildfab.AddVersionVariables(vars map[string]string) map[string]string`
**Location**: `internal/exec/buildfab_executor.go` (line 418)
**Purpose**: Add detailed version variables
**Usage**:
```go
variables = buildfab.AddVersionVariables(variables)
```

**Adds variables**:
- `version.rawversion` - Raw version string
- `version.major` - Major version number
- `version.minor` - Minor version number
- `version.patch` - Patch version number
- `version.commit` - Git commit hash
- `version.date` - Build date
- `version.type` - Version type (release, prerelease, etc.)

**Called in**:
- `BuildfabExecutor.GetAllVariables()` - After platform variables

---

### Execution Methods

#### 7. `buildfab.DefaultSimpleRunOptions() *buildfab.SimpleRunOptions`
**Location**: `internal/exec/buildfab_executor.go` (line 170, 266)
**Purpose**: Create default execution options
**Usage**:
```go
opts := buildfab.DefaultSimpleRunOptions()
opts.VerboseLevel = e.ui.GetVerboseLevel()
opts.Debug = e.ui.IsDebug()
opts.WorkingDir = "."
opts.Output = os.Stdout
opts.ErrorOutput = os.Stderr
opts.BuildfabBinaryPath = buildfabPath
opts.Variables = variables
```

**Options struct**:
```go
type SimpleRunOptions struct {
    VerboseLevel        int                // 0=quiet, 1=basic, 2=detailed, 3=maximum
    Debug               bool               // Enable debug output
    WorkingDir          string             // Working directory for execution
    Output              io.Writer          // Standard output stream
    ErrorOutput         io.Writer          // Error output stream
    BuildfabBinaryPath  string             // Path to buildfab binary (for containers)
    Variables           map[string]string  // Variables for interpolation
}
```

**Called in**:
- `BuildfabExecutor.RunStage()` - Before creating runner
- `BuildfabExecutor.RunAction()` - Before creating runner

---

#### 8. `buildfab.NewSimpleRunner(config *buildfab.Config, opts *buildfab.SimpleRunOptions) *buildfab.SimpleRunner`
**Location**: `internal/exec/buildfab_executor.go` (line 241, 276)
**Purpose**: Create buildfab simple runner for execution
**Usage**:
```go
runner := buildfab.NewSimpleRunner(e.config, opts)
```

**What it does**:
- Creates runner instance with configuration
- Applies execution options
- Prepares for stage/action execution
- Handles DAG building and scheduling

**Called in**:
- `BuildfabExecutor.RunStage()` - Before stage execution
- `BuildfabExecutor.RunAction()` - Before action execution

---

#### 9. `(*buildfab.SimpleRunner).RunStage(ctx context.Context, stageName string) error`
**Location**: `internal/exec/buildfab_executor.go` (line 249)
**Purpose**: Execute a complete stage with all its steps
**Usage**:
```go
err = runner.RunStage(ctx, stageName)
```

**What it does**:
- Builds DAG from stage steps
- Resolves dependencies between steps
- Detects circular dependencies
- Executes steps in parallel (respecting dependencies)
- Handles errors according to `onerror` policy
- Displays output in declaration order
- Streams results as they complete
- Returns error if any step fails with `onerror: stop`

**Called in**:
- `BuildfabExecutor.RunStage()` - Main execution method

---

#### 10. `(*buildfab.SimpleRunner).RunAction(ctx context.Context, actionName string) error`
**Location**: `internal/exec/buildfab_executor.go` (line 279)
**Purpose**: Execute a single action directly
**Usage**:
```go
err := runner.RunAction(ctx, actionName)
```

**What it does**:
- Executes single action without DAG
- Interpolates variables in run command
- Executes shell command or container action
- Captures output and exit code
- Returns error on failure

**Called in**:
- `BuildfabExecutor.RunAction()` - Direct action execution

---

## Complete Execution Path Details

### Path 1: Git Hook Execution

```
1. Git calls pre-push hook with refs on stdin
   ↓
2. main() detects Git hook mode via isGitHook()
   ↓
3. runGitHook() is called
   ↓
4. Parse Git refs from stdin (readGitRefs)
   ↓
5. Check delete operation → Skip if true
   ↓
6. Validate tag semantics
   ↓
7. Check if should skip (shouldSkipPrePushStage)
   ↓
8. buildfab.LoadConfig(".project.yml")
   │  └─> Loads YAML, processes includes
   ↓
9. Create UI with verbose/debug settings
   ↓
10. Create BuildfabExecutor
    ↓
11. Set Git push info for variables
    ↓
12. BuildfabExecutor.RunStage(ctx, "pre-push")
    ↓
13. GetAllVariables()
    │  ├─> buildfab.GetPlatformVariables()
    │  ├─> buildfab.AddPlatformVariables()
    │  └─> buildfab.AddVersionVariables()
    ↓
14. buildfab.DefaultSimpleRunOptions()
    ↓
15. Set options (VerboseLevel, Debug, Variables, etc.)
    ↓
16. buildfab.NewSimpleRunner(config, opts)
    ↓
17. runner.RunStage(ctx, "pre-push")
    │  ├─> Build DAG from steps
    │  ├─> Resolve dependencies
    │  ├─> Execute steps in parallel
    │  ├─> Display ordered output
    │  └─> Stream results
    ↓
18. Return results (exit 0 or 1)
```

### Path 2: Test Command Execution

```
1. User runs: pre-push test
   ↓
2. main() routes to runTest()
   ↓
3. buildfab.LoadConfig(".project.yml")
   │  └─> Loads YAML, processes includes
   ↓
4. Get verbose/debug settings
   │  ├─> Check PRE_PUSH_VERBOSE env
   │  └─> Check CLI flags
   ↓
5. Create UI
   ↓
6. Create BuildfabExecutor
   ↓
7. BuildfabExecutor.RunStage(ctx, "pre-push")
   ↓
8. [Same as steps 13-18 from Git Hook path]
```

### Path 3: Variable Resolution

```
BuildfabExecutor.GetAllVariables()
├─> Initialize variables map
├─> buildfab.GetPlatformVariables()
│   └─> Returns: platform, arch, os, os_version, cpu
├─> version.GetVersionInfo()
│   └─> Returns: version, project, module, build-type, version-type
├─> Add Git variables
│   ├─> From push info: tag, tags, branch, branches
│   └─> From detector: version.tag, version.branch
├─> Add environment variables
│   └─> All env.* prefixed
├─> buildfab.AddPlatformVariables(variables)
│   └─> Adds: platform.os, platform.arch, etc.
└─> buildfab.AddVersionVariables(variables)
    └─> Adds: version.major, version.minor, version.patch, etc.
```

## Method Call Sequence Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        main()                                │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │ runGitHook()  │
                    │  or runTest() │
                    └───────┬───────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ buildfab.LoadConfig()         │
            │ Returns: *buildfab.Config     │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ BuildfabExecutor created      │
            │ with buildfab.Config          │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ executor.RunStage()           │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ config.GetStage()             │
            │ Validate stage exists         │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ GetAllVariables()             │
            ├───────────────────────────────┤
            │ ├─> buildfab.GetPlatform...() │
            │ ├─> buildfab.AddPlatform...() │
            │ └─> buildfab.AddVersion...()  │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ buildfab.DefaultSimple...()   │
            │ Create execution options      │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ buildfab.NewSimpleRunner()    │
            │ Create runner with config     │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ runner.RunStage()             │
            ├───────────────────────────────┤
            │ • Build DAG                   │
            │ • Resolve dependencies        │
            │ • Execute steps in parallel   │
            │ • Display ordered output      │
            │ • Stream results              │
            └───────────────┬───────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │    Results    │
                    │ Success/Fail  │
                    └───────────────┘
```

## Buildfab Library Methods Summary Table

| # | Method | File | Line | Purpose |
|---|--------|------|------|---------|
| 1 | `buildfab.LoadConfig()` | main.go | 366, 676 | Load YAML configuration |
| 2 | `config.GetStage()` | buildfab_executor.go | 152, 221 | Get stage by name |
| 3 | `config.GetAction()` | buildfab_executor.go | 225 | Get action by name |
| 4 | `buildfab.GetPlatformVariables()` | buildfab_executor.go | 330 | Get platform variables |
| 5 | `buildfab.AddPlatformVariables()` | buildfab_executor.go | 415 | Add platform vars with prefix |
| 6 | `buildfab.AddVersionVariables()` | buildfab_executor.go | 418 | Add version variables |
| 7 | `buildfab.DefaultSimpleRunOptions()` | buildfab_executor.go | 170, 266 | Create default options |
| 8 | `buildfab.NewSimpleRunner()` | buildfab_executor.go | 241, 276 | Create runner instance |
| 9 | `runner.RunStage()` | buildfab_executor.go | 249 | Execute stage |
| 10 | `runner.RunAction()` | buildfab_executor.go | 279 | Execute action |

## Key Integration Points

### 1. Configuration Loading
- **Method**: `buildfab.LoadConfig()`
- **When**: At start of Git hook or test command
- **Why**: Buildfab handles YAML parsing and include directives
- **Benefit**: Pre-push doesn't need custom YAML parser

### 2. Variable Collection
- **Methods**: `GetPlatformVariables()`, `AddPlatformVariables()`, `AddVersionVariables()`
- **When**: Before stage execution
- **Why**: Pre-push enhances with Git-specific variables
- **Benefit**: 40+ variables available for interpolation

### 3. Execution Orchestration
- **Methods**: `NewSimpleRunner()`, `runner.RunStage()`
- **When**: After configuration and variable setup
- **Why**: Buildfab handles DAG execution, parallelization, output
- **Benefit**: Pre-push focuses on Git logic, buildfab handles workflow

## Architecture Benefits

### Separation of Concerns
- **Pre-push**: Git-specific logic (hooks, refs, validation)
- **Buildfab**: Workflow execution (DAG, parallel, output)

### Code Reuse
- Pre-push leverages buildfab's proven DAG engine
- No need to reimplement parallel execution
- Benefits from buildfab improvements automatically

### Maintainability
- Clear boundaries between components
- Buildfab updates don't require pre-push changes
- Pre-push focuses on its unique value (Git integration)

### Extensibility
- New buildfab features automatically available
- Pre-push can add more Git-specific enhancements
- Variable system easily extensible

## Performance Characteristics

### Buildfab Contributions
1. **Parallel Execution**: Independent steps run simultaneously
2. **Ordered Display**: Steps shown in declaration order
3. **Streaming Output**: Results displayed as they complete
4. **Smart Scheduling**: Dependency-aware execution
5. **Container Support**: Optional container action execution

### Pre-push Enhancements
1. **Smart Skipping**: Skip checks for delete operations
2. **Branch Detection**: Skip non-current tags/branches
3. **Tag Validation**: Semantic version checking
4. **Git Variables**: Enhanced interpolation support

## Conclusion

Pre-push uses **10 core buildfab library methods** to achieve its functionality:

**Configuration** (3 methods):
- LoadConfig, GetStage, GetAction

**Variables** (3 methods):
- GetPlatformVariables, AddPlatformVariables, AddVersionVariables

**Execution** (4 methods):
- DefaultSimpleRunOptions, NewSimpleRunner, RunStage, RunAction

This clean integration allows pre-push to focus on Git-specific functionality while leveraging buildfab's powerful DAG execution engine for workflow automation.

