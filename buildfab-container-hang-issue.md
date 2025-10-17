# Container Hang Issue with Verbose Mode

## ✅ RESOLVED in buildfab v0.26.0

**Solution:** buildfab v0.26.0 added `BuildfabBinaryPath` option to `SimpleRunOptions`, which allows pre-push to specify the buildfab binary path for container preparation.

**Implementation:** pre-push now searches for buildfab binary in system directories and provides it to SimpleRunner, enabling container actions with `run_action` to work properly.

**Status:** Containers now work perfectly in all verbose modes (0-3) ✓

---

## Problem (Historical)
When running `pre-push test` with verbose mode enabled (PRE_PUSH_VERBOSE=1-3), container actions hang indefinitely. At verbose level 0, containers fail with exit code 1. The same configuration works fine when run directly with `buildfab pre-push`.

### Behavior by Verbosity Level:
- **PRE_PUSH_VERBOSE=0**: Container exits with code 1 (fast failure) ✓
- **PRE_PUSH_VERBOSE=1-3**: Container HANGS indefinitely (requires Ctrl+C) ✗
- **buildfab pre-push** (direct): Works correctly at all verbose levels ✓

## Error Behavior

### Verbose Level 0 (Fast Failure)
```bash
$ PRE_PUSH_VERBOSE=0 ./bin/pre-push test
pre-push v1.9.2
Checking buildfab-pre-push (1.9.2) before push
▶️  Running stage: pre-push

  ✓ version-check executed successfully - in '0.006s'
  ✓ version-greatest executed successfully - in '0.008s'
  ✓ version-module executed successfully - in '0.010s'
  ✓ test-variables executed successfully - in '0.001s'
  ✓ run-tests executed successfully - in '0.187s'
  ✓ platform-view executed successfully - in '0.001s'
  ✗ container-platform-view container exited with code 1 - in '0.695s'
```

### Verbose Level 3 (Hangs Indefinitely)
```bash
$ PRE_PUSH_VERBOSE=3 ./bin/pre-push test
pre-push v1.9.2
Checking buildfab-pre-push (1.9.2) before push
▶️  Running stage: pre-push

  # ... other steps pass with verbose output ...
  
  💻 platform-view
    === Platform Detection Test ===
    Platform: linux
    # ... platform output ...
  ✓ platform-view executed successfully - in '0.001s'
  
  💻 container-platform-view
  [HANGS HERE INDEFINITELY - NO OUTPUT]
^C  # User forced to interrupt
```

But works fine when run directly:
```bash
$ buildfab pre-push
# ... all steps pass including container-platform-view ...
✓ container-platform-view executed successfully - in '0.793s'
```

## Root Cause Analysis

**Status: CONFIRMED - buildfab SimpleRunner API limitation with container stdin/tty handling**

### Investigation Results:

**Debug output reveals the exact issue:**
```
[DEBUG] checkAndShowNextStep: showing next step: container-platform-view
  💻 container-platform-view
[HANGS FOREVER - NO OUTPUT, NO COMPLETION]
```

The container action:
1. ✓ **Starts** successfully (shown in output)
2. ✗ **Never completes** - no output buffered, no completion status
3. ✗ **Hangs indefinitely** - requires Ctrl+C to stop

**Root Cause:**
- `run_action` requires buildfab CLI's container preparation (copying buildfab binary into container)
- buildfab CLI does this preparation automatically before executing containers
- SimpleRunner API doesn't expose container preparation functionality
- When `run_action: platform-view` tries to execute `buildfab action platform-view` inside container:
  - Container doesn't have buildfab binary (no preparation was done)
  - Command fails with "sh: buildfab: not found"
  - At verbose level 0: Fails fast with error message
  - At verbose level 1-3: Hangs trying to execute non-existent command interactively

### Confirmed Facts:
1. **buildfab CLI works** → Does container preparation (copies buildfab binary) AND proper stdin/tty handling
2. **SimpleRunner hangs with ANY container action** → Whether using `run_action:` OR `run:`, containers hang in verbose mode
3. **Not a pre-push configuration issue** → SimpleRunner API has fundamental container execution bug
4. **Two separate issues:**
   - `run_action` needs buildfab binary in container (preparation issue)
   - **All containers hang in verbose mode** (stdin/tty issue) ← **PRIMARY ISSUE**

### Tests Attempted:
- ✗ Changed `run_action` to `run:` with simple shell commands → **STILL HANGS**
- ✗ Removed `opts.Output`/`opts.ErrorOutput` → Still hangs
- ✗ Added `interactive: false` to container config → Field not supported
- ✓ Comprehensive debug logging → Confirmed container starts but never completes

**Important:** Even with `run:` (no buildfab dependency), containers still hang. This confirms the issue is with SimpleRunner's general container execution, not just `run_action`.

### Next Steps for Diagnosis:
1. **Review buildfab CLI source code** - See how it creates SimpleRunner
2. **Add debug logging** - Log what Docker commands are being executed
3. **Test minimal reproduction** - Create simple Go program using SimpleRunner with containers
4. **Contact buildfab maintainer** - Ask if there's known limitations with programmatic container usage
5. **Check Docker logs** - See what's happening at the Docker level when it hangs

## Configuration
The failing action:
```yaml
actions:
  - name: container-platform-view
    description: Test platform detection in ${{ matrix.image }}
    container:
      image:
        from: ${{ matrix.image }}
      run_action: platform-view

stages:
  pre-push:
    steps:
      - action: container-platform-view
        variables:
          matrix.image: "alpine:latest"
```

## Code That Triggers the Issue
In `internal/exec/buildfab_executor.go`:
```go
func (e *BuildfabExecutor) RunStage(ctx context.Context, stageName string) error {
    // Create simple run options
    opts := buildfab.DefaultSimpleRunOptions()
    opts.VerboseLevel = e.ui.GetVerboseLevel()
    opts.Debug = e.ui.IsDebug()
    opts.WorkingDir = "."
    opts.Output = os.Stdout
    opts.ErrorOutput = os.Stderr
    // NOTE: No stdin configuration available
    
    // Create simple runner
    runner := buildfab.NewSimpleRunner(e.config, opts)
    
    // Execute the stage - containers hang here
    return runner.RunStage(ctx, stageName)
}
```

## Attempted Solutions

### 1. Add stdin to SimpleRunOptions
**Status:** NOT POSSIBLE - SimpleRunOptions doesn't have an Input field

### 2. Add `interactive: false` to container config
**Status:** NOT SUPPORTED - ContainerConfig doesn't support this field
```yaml
container:
  interactive: false  # ERROR: field not found in type container.ContainerConfig
```

## Expected Behavior
Container actions should run in non-interactive mode when executed via SimpleRunner, just as they do when run via the buildfab CLI.

## Requested Fix
Add one of the following to the buildfab library:

### Option 1: Add Input field to SimpleRunOptions
```go
type SimpleRunOptions struct {
    // ... existing fields ...
    Input       io.Reader         // Input reader (default: nil for non-interactive)
    Output      io.Writer         // Output writer (default: os.Stdout)
    ErrorOutput io.Writer         // Error output writer (default: os.Stderr)
}
```

### Option 2: Add interactive field to ContainerConfig
```yaml
actions:
  - name: container-platform-view
    container:
      image:
        from: alpine:latest
      interactive: false  # Disable stdin attachment
      run_action: platform-view
```

### Option 3: Auto-detect non-interactive mode
When SimpleRunner is used programmatically, automatically detect that stdin is not available and run containers with `--interactive=false` flag.

## Solution Implemented

### buildfab v0.26.0 Changes

**New Option:** `BuildfabBinaryPath` in `SimpleRunOptions`
```go
type SimpleRunOptions struct {
    // ... other fields
    BuildfabBinaryPath string  // Path to buildfab binary for run_action/run_stage
}
```

### pre-push Implementation

**Binary Search Strategy:**
```go
func findBuildfabBinary() (string, error) {
    searchPaths := []string{
        "/usr/local/bin/buildfab",
        "/usr/bin/buildfab",
        "$HOME/bin/buildfab",
        "<current-binary-dir>/buildfab",
        "./scripts/buildfab",
    }
    // Also searches PATH using exec.LookPath
}
```

**Usage in pre-push:**
```go
opts := buildfab.DefaultSimpleRunOptions()
buildfabPath, err := findBuildfabBinary()
if err == nil {
    opts.BuildfabBinaryPath = buildfabPath
}
runner := buildfab.NewSimpleRunner(e.config, opts)
```

### Working Configuration

```yaml
- name: container-platform-view
  container:
    image:
      from: alpine:latest
    run_action: platform-view  # ✓ Now works with BuildfabBinaryPath!
```

**Result:** Container completes successfully in 0.6-0.8s at all verbose levels (0-3)

### Error Handling

**If buildfab binary is not found:**
- pre-push logs a debug warning (only in debug mode)
- Container actions using `run_action` will fail with "buildfab: not found"
- Container actions using `run:` with shell commands will still work

**To install buildfab:**
```bash
# Install to /usr/local/bin (recommended)
sudo cp buildfab /usr/local/bin/

# Or install to user directory
cp buildfab $HOME/bin/

# Or add to PATH
export PATH="$PATH:/path/to/buildfab"
```

## Legacy Workarounds (No longer needed with v0.26.0)

### Option 1: Run with Verbose Level 0
```bash
# Container fails fast with error:
PRE_PUSH_VERBOSE=0 ./bin/pre-push test
```

### Option 2: Use buildfab CLI directly
```bash
# Works at all verbose levels (has container preparation):
buildfab pre-push
buildfab -vvv pre-push
```

### Option 3: Disable container actions temporarily
Comment out container actions in `.project.yml`

## Impact
This issue blocks pre-push from being used as a standalone CLI tool when container actions are present in the configuration. Users are forced to use the buildfab CLI directly, defeating the purpose of the pre-push wrapper.

## Files Affected
- `internal/exec/buildfab_executor.go` - Uses SimpleRunner
- `.project.yml` - Contains container actions
- All workflows that use `pre-push test` command

## Buildfab Version
- Tested: buildfab v0.25.0, v0.25.2
- Issue affects: All versions with container support
- **Update (v0.25.2)**: Partial fix - container no longer hangs at verbose level 0, but STILL HANGS at verbose levels 1-3

## Related Issues
- This may be related to how buildfab's container executor handles stdin/tty configuration
- The buildfab CLI likely has special logic to handle interactive vs non-interactive mode that is not exposed via SimpleRunner API

