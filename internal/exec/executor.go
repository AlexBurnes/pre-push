// Package exec provides execution functionality for pre-push checks.
package exec

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "sync"
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

    // Execute DAG with streaming output in declaration order
    results, err := e.executeDAGWithStreaming(ctx, dag, stage.Steps)
    
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
    
    // Get topological order for execution
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
            // Mark as skipped due to failed dependency
            result := prepush.Result{
                Name: nodeName,
                Status: prepush.StatusSkipped,
                Message: "skipped (dependency failed)",
            }
            results = append(results, result)
            completed[nodeName] = true
            continue
        }
        
        // Check if step should be executed based on conditions
        if !e.shouldExecuteStep(ctx, node) {
            result := prepush.Result{
                Name: nodeName,
                Status: prepush.StatusOK,
                Message: "skipped (condition not met)",
            }
            results = append(results, result)
            completed[nodeName] = true
            continue
        }
        
        // Execute the node
        result, _ := e.executeAction(ctx, node.Action)
        result.Name = nodeName
        results = append(results, result)
        
        completed[nodeName] = true
        
        if result.Status == prepush.StatusError {
            failed[nodeName] = true
        }
        
        // Note: We no longer return early on error - continue executing independent steps
    }
    
    return results, nil
}

// executeDAGWithStreaming executes the DAG with streaming output in declaration order
func (e *Executor) executeDAGWithStreaming(ctx context.Context, dag map[string]*DAGNode, steps []prepush.Step) ([]prepush.Result, error) {
    var results []prepush.Result
    completed := make(map[string]bool)
    failed := make(map[string]bool)
    displayed := make(map[string]bool)
    executing := make(map[string]bool)
    
    // Create a map of results by step name for quick lookup
    resultMap := make(map[string]prepush.Result)
    
    // Create channels for communication
    resultChan := make(chan prepush.Result, len(dag))
    done := make(chan bool)
    
    // Mutex for thread-safe access to shared state
    var mu sync.Mutex
    
    // Start a goroutine to handle results and display them
    go func() {
        defer close(done)
        for result := range resultChan {
            mu.Lock()
            results = append(results, result)
            resultMap[result.Name] = result
            completed[result.Name] = true
            executing[result.Name] = false
            
            if result.Status == prepush.StatusError {
                failed[result.Name] = true
            }
            mu.Unlock()
            
            // Display immediately if it's ready in declaration order
            e.displayStepImmediately(result.Name, steps, resultMap, displayed, completed)
        }
    }()
    
    // Execute all ready steps in parallel
    for {
        mu.Lock()
        readySteps := e.getReadyStepsLocked(dag, completed, failed, executing)
        mu.Unlock()
        
        if len(readySteps) == 0 {
            break
        }
        
        // Execute ready steps in parallel
        var wg sync.WaitGroup
        for _, nodeName := range readySteps {
            node := dag[nodeName]
            
            mu.Lock()
            executing[nodeName] = true
            mu.Unlock()
            
            // Skip if already failed and this node requires it
            if e.hasFailedDependency(node, failed) {
                result := prepush.Result{
                    Name: nodeName,
                    Status: prepush.StatusSkipped,
                    Message: "skipped (dependency failed)",
                }
                resultChan <- result
                continue
            }
            
            // Check if step should be executed based on conditions
            if !e.shouldExecuteStep(ctx, node) {
                result := prepush.Result{
                    Name: nodeName,
                    Status: prepush.StatusOK,
                    Message: "skipped (condition not met)",
                }
                resultChan <- result
                continue
            }
            
            // Execute the node in parallel
            wg.Add(1)
            go func(nodeName string, node *DAGNode) {
                defer wg.Done()
                result, _ := e.executeAction(ctx, node.Action)
                result.Name = nodeName
                resultChan <- result
            }(nodeName, node)
        }
        
        wg.Wait()
    }
    
    close(resultChan)
    <-done
    
    // Display any remaining steps that weren't displayed yet
    e.displayRemainingSteps(steps, resultMap, displayed)
    
    return results, nil
}

// displayStepIfReady displays a step if it's ready in declaration order
func (e *Executor) displayStepIfReady(stepName string, steps []prepush.Step, resultMap map[string]prepush.Result, displayed map[string]bool) {
    // Find the step in declaration order
    for _, step := range steps {
        if step.Action == stepName {
            // Check if all previous steps in declaration order have been displayed
            if e.canDisplayStep(step, steps, displayed) {
                if result, exists := resultMap[stepName]; exists {
                    e.ui.PrintStepStatus(stepName, result.Status, result.Message)
                    displayed[stepName] = true
                }
            }
            break
        }
    }
    
    // Try to display any buffered steps that are now ready
    e.displayBufferedSteps(steps, resultMap, displayed)
}

// displayStepImmediately displays a step immediately if it can be shown (streaming)
func (e *Executor) displayStepImmediately(stepName string, steps []prepush.Step, resultMap map[string]prepush.Result, displayed map[string]bool, completed map[string]bool) {
    // Find the step in declaration order
    for _, step := range steps {
        if step.Action == stepName {
            // Check if all previous steps in declaration order have been completed
            if e.canDisplayStepImmediately(step, steps, displayed, completed) {
                if result, exists := resultMap[stepName]; exists {
                    e.ui.PrintStepStatus(stepName, result.Status, result.Message)
                    displayed[stepName] = true
                }
            }
            break
        }
    }
}

// displayBufferedSteps displays any steps that are now ready to be displayed
func (e *Executor) displayBufferedSteps(steps []prepush.Step, resultMap map[string]prepush.Result, displayed map[string]bool) {
    for _, step := range steps {
        if !displayed[step.Action] {
            if result, exists := resultMap[step.Action]; exists {
                if e.canDisplayStep(step, steps, displayed) {
                    e.ui.PrintStepStatus(step.Action, result.Status, result.Message)
                    displayed[step.Action] = true
                }
            }
        }
    }
}

// canDisplayStep checks if a step can be displayed based on declaration order
func (e *Executor) canDisplayStep(step prepush.Step, steps []prepush.Step, displayed map[string]bool) bool {
    // Find the position of this step in the declaration order
    stepIndex := -1
    for i, s := range steps {
        if s.Action == step.Action {
            stepIndex = i
            break
        }
    }
    
    if stepIndex == -1 {
        return false
    }
    
    // Check if all previous steps in declaration order have been displayed
    for i := 0; i < stepIndex; i++ {
        if !displayed[steps[i].Action] {
            return false
        }
    }
    
    return true
}

// canDisplayStepImmediately checks if a step can be displayed immediately (streaming)
func (e *Executor) canDisplayStepImmediately(step prepush.Step, steps []prepush.Step, displayed map[string]bool, completed map[string]bool) bool {
    // Find the position of this step in the declaration order
    stepIndex := -1
    for i, s := range steps {
        if s.Action == step.Action {
            stepIndex = i
            break
        }
    }
    
    if stepIndex == -1 {
        return false
    }
    
    // Check if all previous steps in declaration order have been completed (not just displayed)
    for i := 0; i < stepIndex; i++ {
        if !completed[steps[i].Action] {
            return false
        }
    }
    
    return true
}

// getReadySteps returns steps that are ready to execute (dependencies completed)
func (e *Executor) getReadySteps(dag map[string]*DAGNode, completed map[string]bool, failed map[string]bool) []string {
    var ready []string
    
    for nodeName, node := range dag {
        if completed[nodeName] {
            continue
        }
        
        // Check if all dependencies are completed
        if e.allDependenciesCompleted(node, completed) {
            ready = append(ready, nodeName)
        }
    }
    
    return ready
}

// getReadyStepsLocked returns steps that are ready to execute (thread-safe version)
func (e *Executor) getReadyStepsLocked(dag map[string]*DAGNode, completed map[string]bool, failed map[string]bool, executing map[string]bool) []string {
    var ready []string
    
    for nodeName, node := range dag {
        if completed[nodeName] || executing[nodeName] {
            continue
        }
        
        // Check if all dependencies are completed
        if e.allDependenciesCompleted(node, completed) {
            ready = append(ready, nodeName)
        }
    }
    
    return ready
}

// displayRemainingSteps displays any steps that weren't displayed yet
func (e *Executor) displayRemainingSteps(steps []prepush.Step, resultMap map[string]prepush.Result, displayed map[string]bool) {
    for _, step := range steps {
        if !displayed[step.Action] {
            if result, exists := resultMap[step.Action]; exists {
                e.ui.PrintStepStatus(step.Action, result.Status, result.Message)
                displayed[step.Action] = true
            }
        }
    }
}

// displayResultsInOrder displays results in the order they appear in the YAML configuration
func (e *Executor) displayResultsInOrder(steps []prepush.Step, results []prepush.Result) {
    // Create a map of results by step name for quick lookup
    resultMap := make(map[string]prepush.Result)
    for _, result := range results {
        resultMap[result.Name] = result
    }
    
    // Display results in the order they appear in the YAML configuration
    for _, step := range steps {
        if result, exists := resultMap[step.Action]; exists {
            e.ui.PrintStepStatus(step.Action, result.Status, result.Message)
        }
    }
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
    
    // Only show output in verbose mode for custom actions
    if e.ui.IsVerbose() {
        e.ui.PrintCommandOutput(string(output))
    }
    
    if err != nil {
        // For custom actions, provide the exact command to run manually
        return prepush.Result{
            Status: prepush.StatusError,
            Message: fmt.Sprintf("Command failed. To debug run: %s", action.Run),
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

// shouldExecuteStep determines if a step should be executed based on conditions
func (e *Executor) shouldExecuteStep(ctx context.Context, node *DAGNode) bool {
    // Check 'if' condition
    if node.Step.If != "" {
        if !e.evaluateCondition(ctx, node.Step.If) {
            return false
        }
    }
    
    // Check 'only' conditions
    if len(node.Step.Only) > 0 {
        if !e.matchesOnlyConditions(ctx, node.Step.Only) {
            return false
        }
    }
    
    return true
}

// evaluateCondition evaluates a condition string
func (e *Executor) evaluateCondition(ctx context.Context, condition string) bool {
    // For now, support simple conditions like "version.type == 'release'"
    // This is a simplified implementation - in a real system, you'd want a proper expression parser
    
    if condition == "version.type == 'release'" {
        versionType, err := e.getVersionType(ctx)
        return err == nil && versionType == "release"
    }
    
    if condition == "version.type == 'prerelease'" {
        versionType, err := e.getVersionType(ctx)
        return err == nil && versionType == "prerelease"
    }
    
    // Default to true for unknown conditions
    return true
}

// matchesOnlyConditions checks if current context matches any of the 'only' conditions
func (e *Executor) matchesOnlyConditions(ctx context.Context, onlyConditions []string) bool {
    versionType, err := e.getVersionType(ctx)
    if err != nil {
        // If we can't determine version type, don't execute steps with 'only' conditions
        return false
    }
    
    for _, condition := range onlyConditions {
        if condition == versionType {
            return true
        }
    }
    
    return false
}

// getVersionType determines the type of version (release, prerelease, patch, minor, major)
func (e *Executor) getVersionType(ctx context.Context) (string, error) {
    version, err := e.versionDetector.DetectCurrentVersion(ctx)
    if err != nil {
        return "", err
    }
    
    // Simple heuristic: if version contains "alpha", "beta", "rc", it's prerelease
    versionLower := strings.ToLower(version)
    if strings.Contains(versionLower, "alpha") || 
       strings.Contains(versionLower, "beta") || 
       strings.Contains(versionLower, "rc") {
        return "prerelease", nil
    }
    
    // For now, assume all other versions are releases
    // In a more sophisticated implementation, you'd analyze the version string
    // to determine if it's a patch, minor, or major release
    return "release", nil
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