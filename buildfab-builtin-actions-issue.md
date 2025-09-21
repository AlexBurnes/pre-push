# Buildfab Built-in Actions Issue Report

## Problem
The buildfab library executor does not support built-in actions (like `git@untracked`, `git@uncommitted`, `git@modified`) in its public API. When using `uses:` field in actions, the executor fails with "built-in action not supported in public API" error.

## Error Details
```
Error: step git-untracked failed: built-in action git@untracked not supported in public API
```

## Failing Code Example
The issue occurs when using built-in actions in the configuration:

```yaml
actions:
  - name: git-untracked
    uses: git@untracked  # This fails in buildfab executor

  - name: git-uncommitted
    uses: git@uncommitted  # This fails in buildfab executor

  - name: git-modified
    uses: git@modified  # This fails in buildfab executor
```

## Root Cause
The buildfab library's `Runner.RunStage()` method only supports actions with `run:` field, but does not support the `uses:` field for built-in actions. The built-in action registry and execution logic is not integrated into the buildfab library.

## Expected Behavior
The buildfab executor should support both:
- `run:` field for custom shell commands
- `uses:` field for built-in actions (git@untracked, git@uncommitted, git@modified, etc.)

## Current Workaround
Convert built-in actions to shell commands:

```yaml
actions:
  - name: git-untracked
    run: |
      untracked=$(git status --porcelain | grep '^??' | wc -l)
      if [ "$untracked" -gt 0 ]; then
        echo "❌ Found $untracked untracked files:"
        git status --porcelain | grep '^??' | sed 's/^?? /  /'
        echo "To fix: git add <files> or git clean -fd"
        exit 1
      else
        echo "✅ No untracked files"
      fi
```

## Files Affected
- `internal/exec/buildfab_executor.go` - Uses buildfab library
- `.project.yml` - Configuration with built-in actions
- `internal/uses/` - Built-in action implementations (not accessible to buildfab)

## Test Case
```bash
cd /home/burnes/projects/burnes/go/pre-push
./bin/pre-push test
```

**Configuration that fails:**
```yaml
actions:
  - name: git-untracked
    uses: git@untracked
```

**Error:** `built-in action git@untracked not supported in public API`

## Minimal Reproduction
```yaml
project:
  name: "test-project"

actions:
  - name: git-untracked
    uses: git@untracked

stages:
  test:
    steps:
      - action: git-untracked
```

**Expected:** Built-in action should execute
**Actual:** Error about built-in action not supported

## Buildfab Library Integration
The project is trying to integrate with:
- `github.com/AlexBurnes/buildfab` v0.5.1
- Using `buildfab.NewRunner()` and `runner.RunStage()`
- Expecting support for `uses:` field in actions

## Request
Please add support for built-in actions in the buildfab library:

1. **Add built-in action registry** to the buildfab library
2. **Support `uses:` field** in action configuration
3. **Implement built-in action execution** (git@untracked, git@uncommitted, git@modified, etc.)
4. **Maintain compatibility** with existing `run:` field

## Additional Context
The pre-push tool has a comprehensive built-in action system in `internal/uses/` package that provides:
- `git@untracked` - Check for untracked files
- `git@uncommitted` - Check for uncommitted changes  
- `git@modified` - Check for modified files

These actions are essential for Git pre-push hooks and should be available in the buildfab library for proper integration.

## Impact
This limitation prevents full integration with buildfab library and forces users to:
- Convert all built-in actions to shell commands
- Lose the benefits of the built-in action system
- Maintain duplicate logic for Git operations