# Buildfab Execution Paths Documentation

## Overview

This document provides a comprehensive analysis of the buildfab integration and execution paths in the pre-push tool. The pre-push tool uses buildfab v0.29.1 as its core DAG execution engine, providing advanced workflow automation, parallel execution, and variable interpolation capabilities.

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│                         CLI Layer                               │
│                    (cmd/pre-push/main.go)                      │
├────────────────────────────────────────────────────────────────┤
│              Configuration Loading Layer                        │
│           buildfab.LoadConfig(".project.yml")                  │
├────────────────────────────────────────────────────────────────┤
│                   Execution Layer                              │
│          BuildfabExecutor (internal/exec/)                     │
├────────────────────────────────────────────────────────────────┤
│                   Buildfab Core                                │
│     buildfab.NewSimpleRunner() + runner.RunStage()            │
├────────────────────────────────────────────────────────────────┤
│              Variable Resolution Layer                          │
│   Platform + Git + Version + Environment Variables            │
└────────────────────────────────────────────────────────────────┘
```

## Main Execution Paths

### 1. CLI Entry Point Path

**Entry Function**: `main()` in `cmd/pre-push/main.go`

```
main()
├─> Check if called by Git (isGitHook())
│   ├─> YES → runGitHook()
│   └─> NO  → rootCmd.Execute()
│       ├─> test command → runTest()
│       ├─> list-uses command → runListUses()
│       ├─> install command → runInstall()
│       └─> no command → print usage
```

**Decision Logic**:
- **Git Hook Mode**: Detected by checking stdin characteristics (pipe vs terminal)
- **CLI Mode**: When explicit commands (test, install, list-uses) are provided

### 2. Git Hook Execution Path

**Entry Function**: `runGitHook()` in `cmd/pre-push/main.go`

```
runGitHook()
├─> Create context with cancellation
├─> Read Git refs from stdin (readGitRefs())
│   └─> Parse ref lines: <local_ref> <local_sha> <remote_ref> <remote_sha>
├─> Parse push information (parseGitPushInfo())
│   ├─> Extract remote name and URL from args
│   ├─> Identify tags and branches being pushed
│   └─> Detect delete operations (SHA = 0000...0000)
├─> Check if delete operation → Skip if yes
├─> Validate tag semantics (validateTagSemantics())
│   └─> Use version library to validate semantic versioning
├─> Check if should skip (shouldSkipPrePushStage())
│   ├─> Get current branch
│   ├─> Check if pushing current branch → Don't skip
│   ├─> Check if tag is on current branch → Don't skip
│   └─> Otherwise → Skip
├─> Load buildfab config (buildfab.LoadConfig())
├─> Create UI with verbose/debug settings
├─> Create BuildfabExecutor
├─> Set Git push information for variables
└─> Execute pre-push stage (executor.RunStage())
```

**Key Decision Points**:

1. **Delete Operation Detection** (Line 345-348):
   ```go
   if pushInfo.IsDelete {
       fmt.Fprintf(os.Stderr, "Delete operation detected, skipping pre-push checks\n")
       return nil
   }
   ```

2. **Tag Semantic Validation** (Line 351-357):
   ```go
   if len(pushInfo.Tags) > 0 {
       for _, tag := range pushInfo.Tags {
           if err := validateTagSemantics(tag); err != nil {
               return fmt.Errorf("invalid tag semantics for %s: %w", tag, err)
           }
       }
   }
   ```

3. **Smart Pre-Push Skipping** (Line 359-363):
   ```go
   if shouldSkipPrePushStage(pushInfo) {
       fmt.Fprintf(os.Stderr, "Pushing tag/branch that is not current, skipping pre-push stage\n")
       return nil
   }
   ```

### 3. Test Command Execution Path

**Entry Function**: `runTest()` in `cmd/pre-push/main.go`

```
runTest()
├─> Create context with cancellation
├─> Load buildfab config (buildfab.LoadConfig())
├─> Get verbose/debug settings
│   ├─> Check PRE_PUSH_VERBOSE env var
│   └─> Check CLI flags (-v, -d)
├─> Create UI with settings
├─> Create BuildfabExecutor with CLI version
└─> Execute pre-push stage (executor.RunStage())
```

**Verbose Level Handling** (Line 45-63):
```go
func getVerboseLevel() int {
    if envVerbose := os.Getenv("PRE_PUSH_VERBOSE"); envVerbose != "" {
        if level, err := strconv.Atoi(envVerbose); err == nil && level >= 0 {
            return level
        }
        return 0
    }
    if verbose {
        return 1
    }
    return 0
}
```

### 4. BuildfabExecutor.RunStage() Execution Path

**Entry Function**: `RunStage()` in `internal/exec/buildfab_executor.go`

```
BuildfabExecutor.RunStage(ctx, stageName)
├─> Verify stage exists in config
├─> Get project version (getVersion())
│   ├─> Try version library (version.GetVersionInfo())
│   ├─> Fallback to detector.DetectCurrentVersion()
│   └─> Final fallback to "unknown"
├─> Print CLI header and project check
├─> Create SimpleRunOptions
│   ├─> Set VerboseLevel from UI
│   ├─> Set Debug from UI
│   ├─> Set WorkingDir = "."
│   ├─> Set Output and ErrorOutput streams
│   └─> Find buildfab binary for container support
├─> Get all variables (GetAllVariables())
│   ├─> Platform variables (buildfab.GetPlatformVariables())
│   ├─> Version library variables (version.GetVersionInfo())
│   ├─> Git variables from push info or detector
│   └─> Environment variables (os.Environ())
├─> Pass variables to SimpleRunOptions
├─> Create SimpleRunner (buildfab.NewSimpleRunner())
└─> Execute stage (runner.RunStage(ctx, stageName))
```

**Key Components**:

1. **Version Detection** (Line 304-318):
   ```go
   func (e *BuildfabExecutor) getVersion() string {
       if info, err := version.GetVersionInfo(context.Background()); err == nil && info != nil {
           return info.Version
       }
       if version, err := e.versionDetector.DetectCurrentVersion(context.Background()); err == nil {
           return version
       }
       return "unknown"
   }
   ```

2. **Buildfab Binary Discovery** (Line 92-148):
   - Searches: `/usr/local/bin`, `/usr/bin`, `$HOME/bin`, executable dir, `./scripts`
   - Falls back to `exec.LookPath("buildfab")`
   - Required for container actions using `run_action`

### 5. Variable Resolution Path

**Entry Function**: `GetAllVariables()` in `internal/exec/buildfab_executor.go`

```
GetAllVariables()
├─> Initialize variables map
├─> Add buildfab platform variables (buildfab.GetPlatformVariables())
│   ├─> platform: "linux-amd64", "darwin-arm64", etc.
│   ├─> arch: "amd64", "arm64"
│   ├─> os: "linux", "darwin", "windows"
│   ├─> os_version: OS version string
│   └─> cpu: CPU count
├─> Add version library variables (version.GetVersionInfo())
│   ├─> version: "v1.10.2"
│   ├─> project: "pre-push"
│   ├─> module: "github.com/AlexBurnes/pre-push"
│   ├─> modules: comma-separated module list
│   ├─> version.build-type: "Release" or "Debug"
│   └─> version.version-type: "release", "prerelease", etc.
├─> Add Git variables
│   ├─> From push info (if available):
│   │   ├─> tag: single pushed tag (most common)
│   │   ├─> tags: comma-separated pushed tags
│   │   ├─> branch: single pushed branch
│   │   ├─> branches: comma-separated pushed branches
│   │   ├─> version.branch: current repository branch
│   │   └─> version.tag: current repository tag
│   └─> Fallback to detector:
│       ├─> tag: detector.DetectCurrentVersion()
│       └─> branch: detector.DetectCurrentBranch()
├─> Add environment variables (os.Environ())
│   └─> All env vars with "env." prefix
├─> Add buildfab platform variables with prefix
│   └─> buildfab.AddPlatformVariables()
└─> Add buildfab version variables
    └─> buildfab.AddVersionVariables()
        ├─> version.rawversion
        ├─> version.major
        ├─> version.minor
        ├─> version.patch
        ├─> version.commit
        ├─> version.date
        └─> version.type
```

**Variable Categories**:

1. **Platform Variables** (Simple names):
   - `platform`: "linux-amd64", "darwin-arm64", "windows-amd64"
   - `arch`: "amd64", "arm64"
   - `os`: "linux", "darwin", "windows"
   - `os_version`: Detected OS version string
   - `cpu`: Number of CPU cores as string

2. **Version Variables** (From version-go library):
   - `version`: Current version (e.g., "v1.10.2")
   - `project`: Project name (e.g., "pre-push")
   - `module`: Go module path
   - `modules`: Comma-separated list of all modules
   - `version.build-type`: "Release" or "Debug"
   - `version.version-type`: "release", "prerelease", "alpha", "beta"

3. **Git Variables** (Enhanced with push info):
   - `tag`: Single pushed tag (most common case)
   - `tags`: Comma-separated list of pushed tags
   - `branch`: Single pushed branch (most common case)
   - `branches`: Comma-separated list of pushed branches
   - `version.tag`: Current repository tag (from git describe)
   - `version.branch`: Current repository branch (from git rev-parse)

4. **Environment Variables** (All with "env." prefix):
   - `env.HOME`, `env.USER`, `env.PATH`, etc.
   - All environment variables accessible as `env.VARIABLE_NAME`

5. **Buildfab Version Variables** (Detailed version info):
   - `version.rawversion`: Raw version string
   - `version.major`: Major version number
   - `version.minor`: Minor version number
   - `version.patch`: Patch version number
   - `version.commit`: Git commit hash
   - `version.date`: Build date
   - `version.type`: Version type (release, prerelease, etc.)

### 6. Configuration Loading Path

**Entry Function**: `buildfab.LoadConfig()` (buildfab library)

```
buildfab.LoadConfig(".project.yml")
├─> Read YAML file
├─> Parse project section
│   ├─> name: Project name
│   ├─> modules: List of Go modules
│   └─> bin: Binary directory (optional)
├─> Parse actions section
│   ├─> name: Action name
│   ├─> run: Shell command (optional)
│   ├─> uses: Built-in action reference (optional)
│   ├─> shell: Shell type (bash, sh, zsh, etc.)
│   └─> container: Container configuration (optional)
├─> Parse stages section
│   ├─> steps: List of steps
│   │   ├─> action: Action name to execute
│   │   ├─> name: Custom step name (optional, v0.28.0+)
│   │   ├─> stage: Stage reference (optional, v0.27.0+)
│   │   ├─> require: Dependencies
│   │   ├─> onerror: Error policy (stop/warn)
│   │   ├─> if: Conditional expression
│   │   └─> only: Version filter (release/prerelease/patch/minor/major)
│   ├─> verbose: Enable verbose output
│   └─> debug: Enable debug output
└─> Process include directives (recursive)
    ├─> Load referenced files
    ├─> Merge configurations
    └─> Resolve all includes
```

**Include Support** (Buildfab v0.10.0+):
- Pre-push uses `buildfab.LoadConfig()` directly for include support
- Includes are processed recursively
- Example: `include: config/update-checking-actions.yml`

### 7. Git Push Information Parsing Path

**Entry Function**: `parseGitPushInfo()` in `cmd/pre-push/main.go`

```
parseGitPushInfo(refs)
├─> Get remote name from args[1]
├─> Get remote URL from args[2] (if available)
├─> Initialize tags and branches slices
├─> Process each ref:
│   ├─> Check if delete operation (SHA = 0000...0000)
│   ├─> Extract tag name (refs/tags/*)
│   ├─> Extract branch name (refs/heads/*)
│   ├─> Add to appropriate list
│   └─> Set isDelete flag if any ref is deleted
└─> Return GitPushInfo struct
    ├─> RemoteName: "origin"
    ├─> RemoteURL: "https://github.com/..."
    ├─> Refs: List of GitRef structs
    ├─> Tags: ["v1.10.2", "v1.10.3"]
    ├─> Branches: ["master", "develop"]
    └─> IsDelete: true/false
```

**Git Ref Format**:
```
<local_ref> <local_sha> <remote_ref> <remote_sha>
refs/heads/master 1234abc refs/heads/master 5678def
refs/tags/v1.0.0 9abcdef refs/tags/v1.0.0 0000000
```

### 8. Smart Pre-Push Skipping Path

**Entry Function**: `shouldSkipPrePushStage()` in `cmd/pre-push/main.go`

```
shouldSkipPrePushStage(pushInfo)
├─> Get current branch (getCurrentBranch())
├─> If pushing branches:
│   ├─> Check if any pushed branch matches current branch
│   ├─> If match found → Don't skip (return false)
│   └─> If no match → Skip (return true)
├─> If pushing tags:
│   ├─> For each tag:
│   │   └─> Check if tag is on current branch (isTagOnCurrentBranch())
│   │       ├─> Get tag commit SHA (git rev-parse)
│   │       └─> Check if ancestor of HEAD (git merge-base --is-ancestor)
│   ├─> If any tag on current branch → Don't skip (return false)
│   └─> If no tags on current branch → Skip (return true)
└─> Default → Don't skip (return false)
```

**Tag-on-Branch Detection Logic** (Line 595-613):
```go
func isTagOnCurrentBranch(tag, currentBranch string) bool {
    // Get the commit SHA of the tag
    cmd := exec.Command("git", "rev-parse", tag+"^{commit}")
    tagCommit, err := cmd.Output()
    if err != nil {
        return false // Tag doesn't exist or can't resolve
    }
    tagCommitStr := strings.TrimSpace(string(tagCommit))
    
    // Check if tag commit is reachable from current branch HEAD
    cmd = exec.Command("git", "merge-base", "--is-ancestor", tagCommitStr, "HEAD")
    err = cmd.Run()
    return err == nil // Exit code 0 = tag is ancestor of HEAD
}
```

**Scenarios**:
1. **Pushing current branch** → Don't skip (our branch)
2. **Pushing tag on current branch** → Don't skip (our tag)
3. **Pushing tag NOT on current branch** → Skip (not our tag)
4. **Pushing different branch** → Skip (not our branch)

### 9. UI Output and Display Path

**Entry Function**: `PrintCLIHeader()`, `PrintProjectCheck()` in `internal/ui/ui.go`

```
UI Output Flow
├─> Print CLI Header
│   └─> "pre-push v1.10.2"
├─> Print Project Check
│   └─> "Checking pre-push (version v1.10.2)"
├─> Buildfab SimpleRunner handles step output
│   ├─> Step execution with status icons
│   ├─> Ordered display (declaration order)
│   ├─> Streaming output (results as they complete)
│   └─> Summary with counts
└─> Print Summary (if needed)
    ├─> Total steps
    ├─> Successful steps (green)
    ├─> Failed steps (red)
    ├─> Skipped steps (yellow)
    └─> Execution time
```

**Verbose Levels**:
- **Level 0** (Quiet): Minimal output, only errors
- **Level 1** (Basic): Step status with icons (💻), basic command output
- **Level 2** (Detailed): Detailed command output with full logs
- **Level 3** (Maximum): Step-by-step execution with reproduction commands

### 10. Container Action Execution Path

**Entry Function**: Buildfab SimpleRunner (when action has container config)

```
Container Action Execution
├─> Check if buildfab binary is available
│   └─> Required for container actions using run_action
├─> Buildfab processes container configuration:
│   ├─> image.from: Container image name
│   ├─> image.pull: Always, never, missing
│   ├─> run_action: Action to run inside container
│   ├─> volumes: Volume mounts
│   └─> environment: Environment variables
├─> Buildfab runs action inside container
│   ├─> Pull image if needed
│   ├─> Create container with mounts
│   ├─> Execute action inside container
│   └─> Capture output and exit code
└─> Return results to pre-push
```

**Container Support**:
- Requires `buildfab` binary in PATH or system directories
- Supports Docker and Podman
- Volume mounting for workspace access
- Environment variable passing
- Action execution inside containers

## Data Flow Diagrams

### Overall Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                          Git Hook                               │
│                  (stdin: ref information)                       │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Pre-Push Tool                              │
│                   (cmd/pre-push/main.go)                        │
├─────────────────────────────────────────────────────────────────┤
│ 1. Parse Git refs and push information                         │
│ 2. Validate operations (delete, tag semantics, skip logic)     │
│ 3. Load buildfab configuration (.project.yml)                  │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   BuildfabExecutor                              │
│             (internal/exec/buildfab_executor.go)                │
├─────────────────────────────────────────────────────────────────┤
│ 1. Get project version (version library)                       │
│ 2. Collect all variables (platform, version, git, env)         │
│ 3. Create SimpleRunOptions with verbose/debug settings         │
│ 4. Find buildfab binary for container support                  │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Buildfab SimpleRunner                         │
│              (github.com/AlexBurnes/buildfab)                   │
├─────────────────────────────────────────────────────────────────┤
│ 1. Build DAG from stage steps                                  │
│ 2. Resolve dependencies and detect cycles                      │
│ 3. Execute steps in parallel (respecting dependencies)         │
│ 4. Handle errors according to onerror policy                   │
│ 5. Display output in declaration order                         │
│ 6. Stream results as they complete                             │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Action Execution                           │
│                  (Shell commands or containers)                 │
├─────────────────────────────────────────────────────────────────┤
│ 1. Interpolate variables in run command                        │
│ 2. Execute shell command or container action                   │
│ 3. Capture stdout/stderr                                       │
│ 4. Return exit code and output                                 │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Results                                  │
│                  (Exit code and summary)                        │
├─────────────────────────────────────────────────────────────────┤
│ - Exit 0: All checks passed                                    │
│ - Exit 1: One or more checks failed                            │
│ - Summary: Counts of OK, ERROR, WARN, SKIPPED steps           │
└─────────────────────────────────────────────────────────────────┘
```

### Variable Resolution Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Variable Sources                             │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  Platform    │  │   Version    │  │     Git      │         │
│  │  Variables   │  │   Library    │  │  Variables   │         │
│  │              │  │              │  │              │         │
│  │ • platform   │  │ • version    │  │ • tag        │         │
│  │ • arch       │  │ • project    │  │ • branch     │         │
│  │ • os         │  │ • module     │  │ • tags       │         │
│  │ • os_version │  │ • build-type │  │ • branches   │         │
│  │ • cpu        │  │ • ver-type   │  │ • ver.tag    │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│         │                 │                 │                  │
│         └─────────────────┼─────────────────┘                  │
│                           │                                    │
│                           ▼                                    │
│                ┌─────────────────────┐                         │
│                │  GetAllVariables()  │                         │
│                │   (merge all vars)  │                         │
│                └──────────┬──────────┘                         │
│                           │                                    │
│  ┌────────────────────────┼────────────────────────┐          │
│  │                        │                        │          │
│  ▼                        ▼                        ▼          │
│ ┌────────┐         ┌────────────┐         ┌────────────┐     │
│ │  Env   │         │  Buildfab  │         │  Buildfab  │     │
│ │  Vars  │         │  Platform  │         │   Version  │     │
│ │        │         │  Variables │         │  Variables │     │
│ │ • All  │         │            │         │            │     │
│ │  env.* │         │ • platform │         │ • version  │     │
│ │  vars  │         │  prefixed  │         │  detailed  │     │
│ └────┬───┘         └─────┬──────┘         └─────┬──────┘     │
│      │                   │                      │            │
│      └───────────────────┼──────────────────────┘            │
│                          │                                   │
└──────────────────────────┼───────────────────────────────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │  Complete Variables  │
                │  Map[string]string   │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │  SimpleRunOptions    │
                │  opts.Variables      │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │  Buildfab Runner     │
                │  Variable Interp.    │
                └──────────────────────┘
```

## Key Design Patterns

### 1. Delegation Pattern
Pre-push delegates core execution to buildfab, focusing on:
- Git-specific logic (hooks, refs, push info)
- Variable enhancement (Git variables, version info)
- UI customization (headers, project info)
- Workflow orchestration (when to skip, validate, execute)

### 2. Strategy Pattern
Different execution strategies:
- **Git Hook Mode**: Full validation and smart skipping
- **Test Mode**: Dry-run without Git context
- **Install Mode**: Binary installation and update

### 3. Observer Pattern
BuildfabExecutor observes buildfab execution:
- Receives stage/action results
- Formats output via UI interface
- Handles errors according to policy

### 4. Builder Pattern
Configuration building:
- Buildfab loads and parses YAML
- Pre-push enhances with Git variables
- SimpleRunOptions built with all settings
- SimpleRunner created with complete config

### 5. Template Method Pattern
Execution template:
1. Load configuration
2. Detect environment (Git hook vs CLI)
3. Collect variables
4. Create executor
5. Run stage/action
6. Handle results

## Performance Characteristics

### Parallel Execution
- Independent steps run in parallel
- Dependent steps wait for requirements
- DAG ensures correct execution order
- Streaming output shows results immediately

### Ordered Display
- Steps displayed in declaration order (not execution order)
- Improves user experience and predictability
- Dependency-aware display waits for requirements
- Maintains visual consistency across runs

### Smart Skipping
- Delete operations skip immediately (no checks)
- Non-current tags/branches skip pre-push stage
- Reduces unnecessary validation overhead
- Improves push performance for non-current refs

### Container Support
- Optional buildfab binary for containers
- Lazy loading (only if container actions used)
- Efficient image management (pull policies)
- Volume mounting for workspace access

## Error Handling Strategy

### Error Policies
- **stop**: Fail immediately, mark dependents as SKIPPED
- **warn**: Continue execution, show warning
- **default**: Same as "stop" if not specified

### Error Flow
```
Step Execution
├─> Success (exit 0)
│   ├─> Mark step as OK
│   └─> Continue with dependents
├─> Failure (exit != 0) with onerror=warn
│   ├─> Mark step as WARN
│   └─> Continue with dependents
└─> Failure (exit != 0) with onerror=stop
    ├─> Mark step as ERROR
    ├─> Mark all dependents as SKIPPED
    └─> Continue with independent steps
```

### Resilience Features
- Independent steps continue on failure
- Dependent steps marked as SKIPPED
- Clear error reporting with context
- Reproduction commands for manual debugging

## Configuration Examples

### Basic Pre-Push Configuration

```yaml
project:
  name: my-project
  modules:
    - go.mod

actions:
  - name: check-version
    run: scripts/version check
  
  - name: run-tests
    run: go test ./... -v

stages:
  pre-push:
    steps:
      - action: check-version
        onerror: stop
      
      - action: run-tests
        require: [check-version]
        onerror: stop
```

### Advanced Configuration with Variables

```yaml
project:
  name: my-project

actions:
  - name: version-check
    run: |
      echo "Current version: ${{ version }}"
      echo "Platform: ${{ platform }}"
      echo "Branch: ${{ branch }}"
  
  - name: tag-specific
    run: echo "Tag: ${{ tag }}"
    if: "${{ tag != '' }}"
  
  - name: build-release
    run: ./scripts/build-release.sh ${{ version }}
    only: [release]

stages:
  pre-push:
    verbose: false
    steps:
      - action: version-check
      
      - action: tag-specific
      
      - action: build-release
        require: [version-check]
```

### Container Action Example

```yaml
actions:
  - name: test-in-container
    container:
      image:
        from: golang:1.22
        pull: missing
      run_action: run-tests
      volumes:
        - $PWD:/workspace
      environment:
        - GOMODCACHE=/workspace/.cache/go-mod
  
  - name: run-tests
    run: go test ./... -v

stages:
  pre-push:
    steps:
      - action: test-in-container
```

## Debugging and Troubleshooting

### Enable Debug Mode

```bash
# Via environment variable
export PRE_PUSH_DEBUG=1
git push origin master

# Via CLI flag
pre-push test -d
```

### Debug Output Includes
- UI verbose/debug settings
- Variable values (truncated if > 50 chars)
- Stage configuration details
- Step counts and action types
- Execution flow and timing

### Common Issues and Solutions

1. **Buildfab binary not found**
   - **Issue**: Container actions fail
   - **Solution**: Install buildfab or add to PATH
   - **Debug**: Check `DEBUG: Could not find buildfab binary`

2. **Variable not resolved**
   - **Issue**: Variable shows as `${{ varname }}`
   - **Solution**: Check variable name and availability
   - **Debug**: Enable debug to see all variables

3. **Steps not running in expected order**
   - **Issue**: Display order confusion
   - **Solution**: Check dependencies in require field
   - **Debug**: Review DAG structure in debug output

4. **Pre-push skipping unexpectedly**
   - **Issue**: Smart skipping activates
   - **Solution**: Check branch/tag being pushed
   - **Debug**: Review skip logic decision output

## Future Enhancements

### Planned Features
1. **Matrix builds**: Multiple platform/version combinations
2. **Caching**: Cache intermediate results
3. **Artifacts**: Store and share build artifacts
4. **Notifications**: Slack/email notifications
5. **Web UI**: Visual DAG and execution monitoring

### Extension Points
1. **Custom runners**: Implement custom execution logic
2. **Variable providers**: Add custom variable sources
3. **Output formatters**: Custom output formats (JSON, XML)
4. **Action plugins**: Reusable action libraries

## References

- **Buildfab Documentation**: [github.com/AlexBurnes/buildfab](https://github.com/AlexBurnes/buildfab)
- **Version Library**: [github.com/AlexBurnes/version-go](https://github.com/AlexBurnes/version-go)
- **Project Specification**: `docs/Project-specification.md`
- **Variable Reference**: `docs/Variable-reference.md`
- **Developer Workflow**: `docs/Developer-workflow.md`

## Conclusion

The pre-push tool leverages buildfab as a powerful DAG execution engine while adding Git-specific functionality and enhanced variable support. The architecture is clean, modular, and extensible, with clear separation of concerns between Git hook management, variable collection, and execution orchestration.

The execution paths are well-defined and follow consistent patterns throughout the codebase. The integration with buildfab provides advanced features like parallel execution, ordered display, streaming output, and container support, while maintaining simplicity and ease of use for end users.

