// Package exec provides execution functionality for pre-push checks.
package exec

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"

    "github.com/AlexBurnes/pre-push/internal/uses"
    "github.com/AlexBurnes/pre-push/internal/version"
    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

// Executor handles execution of pre-push stages and actions
type Executor struct {
    config *prepush.Config
    ui     UI
    registry *uses.Registry
    versionDetector *version.Detector
}

// UI defines the interface for user interface operations
type UI interface {
    PrintCLIHeader(name, version string)
    PrintProjectCheck(projectName, version string)
    PrintStepStatus(stepName string, status prepush.Status, message string)
    PrintStageHeader(stageName string)
    PrintStageResult(stageName string, success bool, duration time.Duration)
    PrintCommand(command string)
    PrintCommandOutput(output string)
    PrintRepro(stepName, repro string)
    PrintReproInline(stepName, repro string)
    PrintSummary(results []prepush.Result)
    IsVerbose() bool
    IsDebug() bool
}

// New creates a new executor
func New(config *prepush.Config, ui UI) *Executor {
    return &Executor{
        config: config,
        ui:     ui,
        registry: uses.New(),
        versionDetector: version.New(),
    }
}

// RunStage executes a specific stage
func (e *Executor) RunStage(ctx context.Context, stageName string) error {
    stage, exists := e.config.GetStage(stageName)
    if !exists {
        return fmt.Errorf("stage not found: %s", stageName)
    }

    // Print CLI header and project check
    version := e.getVersion()
    e.ui.PrintCLIHeader("pre-push", version)
    e.ui.PrintProjectCheck(e.config.Project.Name, version)

    e.ui.PrintStageHeader(stageName)
    start := time.Now()

    // Build execution DAG
    dag, err := e.buildDAG(stage.Steps)
    if err != nil {
        return fmt.Errorf("failed to build execution DAG: %w", err)
    }

    // Execute DAG
    results, err := e.executeDAG(ctx, dag)
    
    duration := time.Since(start)
    success := err == nil && !hasErrors(results)
    
    e.ui.PrintStageResult(stageName, success, duration)
    e.ui.PrintSummary(results)

    return err
}

// RunAction executes a specific action
func (e *Executor) RunAction(ctx context.Context, actionName string) error {
    action, exists := e.config.GetAction(actionName)
    if !exists {
        return fmt.Errorf("action not found: %s", actionName)
    }

    result, _ := e.executeAction(ctx, action)
    e.ui.PrintStepStatus(actionName, result.Status, result.Message)
    
    if result.Error != nil {
        return result.Error
    }
    
    return nil
}

// ListActions returns all available actions
func (e *Executor) ListActions() []prepush.Action {
    return e.config.Actions
}

// DAGNode represents a node in the execution DAG
type DAGNode struct {
    Step     prepush.Step
    Action   prepush.Action
    Dependencies []string
    Dependents   []string
}

// buildDAG builds the execution DAG from stage steps
func (e *Executor) buildDAG(steps []prepush.Step) (map[string]*DAGNode, error) {
    dag := make(map[string]*DAGNode)
    
    // Create nodes for each step
    for _, step := range steps {
        action, exists := e.config.GetAction(step.Action)
        if !exists {
            return nil, fmt.Errorf("action not found: %s", step.Action)
        }
        
        node := &DAGNode{
            Step:   step,
            Action: action,
            Dependencies: step.Require,
            Dependents:   []string{},
        }
        
        dag[step.Action] = node
    }
    
    // Build dependency relationships
    for _, node := range dag {
        for _, dep := range node.Dependencies {
            if depNode, exists := dag[dep]; exists {
                depNode.Dependents = append(depNode.Dependents, node.Step.Action)
            } else {
                return nil, fmt.Errorf("dependency not found: %s", dep)
            }
        }
    }
    
    // Check for cycles
    if err := e.detectCycles(dag); err != nil {
        return nil, fmt.Errorf("circular dependency detected: %w", err)
    }
    
    return dag, nil
}

// detectCycles detects cycles in the DAG using DFS
func (e *Executor) detectCycles(dag map[string]*DAGNode) error {
    visited := make(map[string]bool)
    recStack := make(map[string]bool)
    
    var dfs func(string) error
    dfs = func(nodeName string) error {
        if recStack[nodeName] {
            return fmt.Errorf("cycle detected involving node: %s", nodeName)
        }
        if visited[nodeName] {
            return nil
        }
        
        visited[nodeName] = true
        recStack[nodeName] = true
        defer func() { recStack[nodeName] = false }()
        
        node := dag[nodeName]
        for _, dep := range node.Dependencies {
            if err := dfs(dep); err != nil {
                return err
            }
        }
        
        return nil
    }
    
    for nodeName := range dag {
        if !visited[nodeName] {
            if err := dfs(nodeName); err != nil {
                return err
            }
        }
    }
    
    return nil
}

// executeDAG executes the DAG in topological order
func (e *Executor) executeDAG(ctx context.Context, dag map[string]*DAGNode) ([]prepush.Result, error) {
    var results []prepush.Result
    completed := make(map[string]bool)
    failed := make(map[string]bool)
    
    // Get topological order
    order, err := e.topologicalSort(dag)
    if err != nil {
        return results, err
    }
    
    // Execute nodes in topological order
    for _, nodeName := range order {
        node := dag[nodeName]
        
        // Check if all dependencies are completed
        if !e.allDependenciesCompleted(node, completed) {
            continue
        }
        
        // Skip if already failed and this node requires it
        if e.hasFailedDependency(node, failed) {
            continue
        }
        
        // Execute the node
        result, _ := e.executeAction(ctx, node.Action)
        result.Name = nodeName
        results = append(results, result)
        
        completed[nodeName] = true
        
        if result.Status == prepush.StatusError {
            failed[nodeName] = true
            e.ui.PrintStepStatus(nodeName, result.Status, result.Message)
            
            // Check if we should stop on error
            if node.Step.OnError != "warn" {
                return results, fmt.Errorf("step %s failed", nodeName)
            }
        } else {
            e.ui.PrintStepStatus(nodeName, result.Status, result.Message)
        }
    }
    
    return results, nil
}

// topologicalSort returns nodes in topological order
func (e *Executor) topologicalSort(dag map[string]*DAGNode) ([]string, error) {
    var result []string
    visited := make(map[string]bool)
    
    var dfs func(string) error
    dfs = func(nodeName string) error {
        if visited[nodeName] {
            return nil
        }
        
        visited[nodeName] = true
        node := dag[nodeName]
        
        for _, dep := range node.Dependencies {
            if err := dfs(dep); err != nil {
                return err
            }
        }
        
        result = append(result, nodeName)
        return nil
    }
    
    for nodeName := range dag {
        if !visited[nodeName] {
            if err := dfs(nodeName); err != nil {
                return nil, err
            }
        }
    }
    
    return result, nil
}

// allDependenciesCompleted checks if all dependencies are completed
func (e *Executor) allDependenciesCompleted(node *DAGNode, completed map[string]bool) bool {
    for _, dep := range node.Dependencies {
        if !completed[dep] {
            return false
        }
    }
    return true
}

// hasFailedDependency checks if any required dependency has failed
func (e *Executor) hasFailedDependency(node *DAGNode, failed map[string]bool) bool {
    for _, dep := range node.Dependencies {
        if failed[dep] {
            return true
        }
    }
    return false
}

// executeAction executes a single action
func (e *Executor) executeAction(ctx context.Context, action prepush.Action) (prepush.Result, error) {
    if action.Uses != "" {
        return e.executeBuiltInAction(ctx, action)
    }
    
    return e.executeCustomAction(ctx, action)
}

// executeBuiltInAction executes a built-in action
func (e *Executor) executeBuiltInAction(ctx context.Context, action prepush.Action) (prepush.Result, error) {
    runner, exists := e.registry.GetRunner(action.Uses)
    if !exists {
        return prepush.Result{
            Status: prepush.StatusError,
            Message: fmt.Sprintf("unknown built-in action: %s", action.Uses),
        }, fmt.Errorf("unknown built-in action: %s", action.Uses)
    }
    
    return runner.Run(ctx)
}

// executeCustomAction executes a custom action with run command
func (e *Executor) executeCustomAction(ctx context.Context, action prepush.Action) (prepush.Result, error) {
    if action.Run == "" {
        return prepush.Result{
            Status: prepush.StatusError,
            Message: "no run command specified",
        }, fmt.Errorf("no run command specified for action %s", action.Name)
    }
    
    // Print command if verbose mode is enabled
    if e.ui.IsVerbose() {
        e.ui.PrintCommand(action.Run)
    }
    
    // Execute the command
    cmd := exec.CommandContext(ctx, "sh", "-c", action.Run)
    output, err := cmd.CombinedOutput()
    
    if e.ui.IsVerbose() {
        e.ui.PrintCommandOutput(string(output))
    }
    
    if err != nil {
        return prepush.Result{
            Status: prepush.StatusError,
            Message: fmt.Sprintf("command failed: %v", err),
        }, fmt.Errorf("command failed: %w", err)
    }
    
    return prepush.Result{
        Status: prepush.StatusOK,
        Message: "command executed successfully",
    }, nil
}

// getVersion returns the current version from the VERSION file
func (e *Executor) getVersion() string {
    // Try to read from VERSION file first
    if version, err := e.readVersionFile(); err == nil {
        return version
    }
    
    // Fallback to git tag detection
    if version, err := e.versionDetector.DetectCurrentVersion(context.Background()); err == nil {
        return version
    }
    
    // Final fallback
    return "unknown"
}

// readVersionFile reads the version from the VERSION file
func (e *Executor) readVersionFile() (string, error) {
    data, err := os.ReadFile("VERSION")
    if err != nil {
        return "", err
    }
    
    version := strings.TrimSpace(string(data))
    if version == "" {
        return "", fmt.Errorf("VERSION file is empty")
    }
    
    return version, nil
}

// hasErrors checks if any results have errors
func hasErrors(results []prepush.Result) bool {
    for _, result := range results {
        if result.Status == prepush.StatusError {
            return true
        }
    }
    return false
}