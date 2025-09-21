# Buildfab Library Step Output Issue Report

## Problem
The buildfab library does not provide step-by-step output or progress callbacks during stage execution. When using `runner.RunStage()`, users only see the final result without visibility into individual step execution progress.

## Current Behavior
```go
// Current usage
runner := buildfab.NewRunner(config, opts)
err := runner.RunStage(ctx, "my-stage")
// Only shows final success/failure, no step-by-step output
```

**Output:**
```
🚀 Running stage: my-stage
============================

✅ Stage 'my-stage' completed successfully in 1.298148178s
```

## Expected Behavior
The buildfab library should provide:
1. **Step-by-step progress callbacks** during execution
2. **Individual step status updates** (running, success, failure)
3. **Step output streaming** for verbose mode
4. **Progress reporting** for long-running stages

**Expected Output:**
```
🚀 Running stage: my-stage
============================

✔ version-check: executed successfully
✔ version-greatest: executed successfully  
✔ version-module: executed successfully
✔ git-untracked: no untracked files
✔ git-uncommitted: no uncommitted changes
⚠ git-modified: 2 modified files (warning)

✅ Stage 'my-stage' completed successfully in 1.298148178s
```

## Use Cases
This is essential for:
- **CLI tools** that need to show progress to users
- **CI/CD systems** that need step-by-step logging
- **Debugging** long-running stages
- **User experience** - users want to see what's happening

## Proposed Solution

### Option 1: Callback Interface
```go
type StepCallback interface {
    OnStepStart(stepName string)
    OnStepComplete(stepName string, status StepStatus, message string)
    OnStepOutput(stepName string, output string)
}

type RunOptions struct {
    // ... existing fields
    StepCallback StepCallback
}
```

### Option 2: Event Channel
```go
type StepEvent struct {
    Type      string // "start", "complete", "output"
    StepName  string
    Status    StepStatus
    Message   string
    Output    string
}

type RunOptions struct {
    // ... existing fields
    EventChannel chan StepEvent
}
```

### Option 3: Context with Values
```go
type StepContext struct {
    StepName string
    Status   StepStatus
    Message  string
}

// Pass context with step information
ctx = context.WithValue(ctx, "step", &StepContext{...})
```

## Implementation Details

### Step Status Types
```go
type StepStatus int

const (
    StepStatusPending StepStatus = iota
    StepStatusRunning
    StepStatusSuccess
    StepStatusWarning
    StepStatusError
    StepStatusSkipped
)
```

### Callback Interface
```go
type StepCallback interface {
    // Called when a step starts execution
    OnStepStart(ctx context.Context, stepName string)
    
    // Called when a step completes (success, warning, or error)
    OnStepComplete(ctx context.Context, stepName string, status StepStatus, message string, duration time.Duration)
    
    // Called for step output (when verbose mode is enabled)
    OnStepOutput(ctx context.Context, stepName string, output string)
    
    // Called for step errors
    OnStepError(ctx context.Context, stepName string, err error)
}
```

### Updated RunStage Method
```go
func (r *Runner) RunStage(ctx context.Context, stageName string) error {
    // ... existing validation
    
    // Get step callback from options
    callback := r.opts.StepCallback
    
    for _, step := range stage.Steps {
        if callback != nil {
            callback.OnStepStart(ctx, step.Name)
        }
        
        start := time.Now()
        err := r.executeStep(ctx, step)
        duration := time.Since(start)
        
        if callback != nil {
            status := StepStatusSuccess
            message := "executed successfully"
            
            if err != nil {
                status = StepStatusError
                message = err.Error()
                callback.OnStepError(ctx, step.Name, err)
            }
            
            callback.OnStepComplete(ctx, step.Name, status, message, duration)
        }
    }
    
    return nil
}
```

## Backward Compatibility
- **No breaking changes** - callback is optional
- **Default behavior unchanged** - existing code continues to work
- **Progressive enhancement** - users can opt into step callbacks

## Example Usage

### Basic Usage (No Callbacks)
```go
runner := buildfab.NewRunner(config, opts)
err := runner.RunStage(ctx, "my-stage")
// Works exactly as before
```

### With Step Callbacks
```go
opts := buildfab.DefaultRunOptions()
opts.StepCallback = &MyStepCallback{}

runner := buildfab.NewRunner(config, opts)
err := runner.RunStage(ctx, "my-stage")
```

### Custom Step Callback Implementation
```go
type MyStepCallback struct{}

func (c *MyStepCallback) OnStepStart(ctx context.Context, stepName string) {
    fmt.Printf("🔄 Running step: %s\n", stepName)
}

func (c *MyStepCallback) OnStepComplete(ctx context.Context, stepName string, status StepStatus, message string, duration time.Duration) {
    var icon string
    switch status {
    case StepStatusSuccess:
        icon = "✔"
    case StepStatusWarning:
        icon = "⚠"
    case StepStatusError:
        icon = "✖"
    }
    
    fmt.Printf("%s %s: %s (%v)\n", icon, stepName, message, duration)
}

func (c *MyStepCallback) OnStepOutput(ctx context.Context, stepName string, output string) {
    if output != "" {
        fmt.Printf("📤 %s output:\n%s\n", stepName, output)
    }
}

func (c *MyStepCallback) OnStepError(ctx context.Context, stepName string, err error) {
    fmt.Printf("❌ %s failed: %v\n", stepName, err)
}
```

## Benefits
1. **Better user experience** - users see progress in real-time
2. **Improved debugging** - step-by-step visibility
3. **Flexible integration** - callbacks can be customized
4. **Backward compatible** - no breaking changes
5. **Performance** - optional feature with no overhead when not used

## Alternative: Event-Driven Approach
Instead of callbacks, could use an event-driven approach:

```go
type StepEvent struct {
    Type      string
    StepName  string
    Status    StepStatus
    Message   string
    Output    string
    Duration  time.Duration
    Timestamp time.Time
}

// Usage
events := make(chan StepEvent, 100)
opts.EventChannel = events

go func() {
    for event := range events {
        // Handle step events
        fmt.Printf("[%s] %s: %s\n", event.Type, event.StepName, event.Message)
    }
}()

runner.RunStage(ctx, "my-stage")
close(events)
```

## Impact
This enhancement would significantly improve the usability of buildfab for:
- CLI tools and applications
- CI/CD systems
- Development tools
- Any application that needs step-by-step progress reporting

## Request
Please add step output and progress callback capabilities to the buildfab library to enable real-time step execution visibility.

## Additional Context
This issue was discovered while integrating buildfab into a pre-push CLI tool that needs to show users what steps are being executed and their individual results. The current implementation only provides final stage results, which is insufficient for user-facing CLI applications.