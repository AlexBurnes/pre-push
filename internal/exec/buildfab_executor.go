// Package exec provides execution functionality for pre-push checks using buildfab DAG executor.
package exec

import (
    "context"
    "fmt"
    "os"
    "strings"

    "github.com/AlexBurnes/buildfab/pkg/buildfab"
    "github.com/AlexBurnes/pre-push/internal/version"
    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

// BuildfabExecutor handles execution of pre-push stages and actions using buildfab DAG executor
type BuildfabExecutor struct {
    config *prepush.Config
    ui     UI
    versionDetector *version.Detector
}


// NewBuildfabExecutor creates a new buildfab-based executor
func NewBuildfabExecutor(config *prepush.Config, ui UI) *BuildfabExecutor {
    return &BuildfabExecutor{
        config: config,
        ui:     ui,
        versionDetector: version.New(),
    }
}

// RunStage executes a specific stage using buildfab SimpleRunner
func (e *BuildfabExecutor) RunStage(ctx context.Context, stageName string) error {
    _, exists := e.config.GetStage(stageName)
    if !exists {
        return fmt.Errorf("stage not found: %s", stageName)
    }

    // Print CLI header and project check
    version := e.getVersion()
    e.ui.PrintCLIHeader("pre-push", version)
    e.ui.PrintProjectCheck(e.config.Project.Name, version)

    // Use buildfab RunStageSimple to execute the entire stage
    // This handles all output internally, so we don't need to add our own
    return buildfab.RunStageSimple(ctx, ".project.yml", stageName, e.ui.IsVerbose())
}

// RunAction executes a specific action using buildfab SimpleRunner
func (e *BuildfabExecutor) RunAction(ctx context.Context, actionName string) error {
    // Convert pre-push config to buildfab config
    buildfabConfig := e.convertToBuildfabConfig()
    
    // Create simple run options
    opts := buildfab.DefaultSimpleRunOptions()
    opts.Verbose = e.ui.IsVerbose()
    opts.Debug = e.ui.IsDebug()
    opts.WorkingDir = "."
    
    // Create simple runner (handles all output internally)
    runner := buildfab.NewSimpleRunner(buildfabConfig, opts)

    // Execute using buildfab SimpleRunner
    err := runner.RunAction(ctx, actionName)
    if err != nil {
        e.ui.PrintStepStatus(actionName, prepush.StatusError, err.Error())
        return err
    }
    
    e.ui.PrintStepStatus(actionName, prepush.StatusOK, "executed successfully")
    return nil
}

// ListActions returns all available actions
func (e *BuildfabExecutor) ListActions() []prepush.Action {
    return e.config.Actions
}

// executeActionWithBuildfab executes a single action using buildfab
func (e *BuildfabExecutor) executeActionWithBuildfab(ctx context.Context, action prepush.Action) prepush.Result {
    // Convert pre-push config to buildfab config
    buildfabConfig := e.convertToBuildfabConfig()
    
    // Create simple run options
    opts := buildfab.DefaultSimpleRunOptions()
    opts.Verbose = e.ui.IsVerbose()
    opts.Debug = e.ui.IsDebug()
    opts.WorkingDir = "."
    
    // Create simple runner
    runner := buildfab.NewSimpleRunner(buildfabConfig, opts)

    // Execute the action
    err := runner.RunAction(ctx, action.Name)
    
    if err != nil {
        return prepush.Result{
            Status:  prepush.StatusError,
            Message: err.Error(),
            Error:   err,
        }
    }
    
    return prepush.Result{
        Status:  prepush.StatusOK,
        Message: "executed successfully",
    }
}

// convertToBuildfabConfig converts pre-push config to buildfab config
func (e *BuildfabExecutor) convertToBuildfabConfig() *buildfab.Config {
    config := &buildfab.Config{}
    
    // Convert project
    config.Project.Name = e.config.Project.Name
    config.Project.Modules = e.config.Project.Modules
    config.Project.BinDir = e.config.Project.BinDir
    
    // Convert actions
    config.Actions = make([]buildfab.Action, len(e.config.Actions))
    for i, action := range e.config.Actions {
        config.Actions[i] = buildfab.Action{
            Name: action.Name,
            Run:  action.Run,
            Uses: action.Uses,
        }
    }
    
    // Convert stages
    config.Stages = make(map[string]buildfab.Stage)
    for name, stage := range e.config.Stages {
        buildfabStage := buildfab.Stage{
            Steps: make([]buildfab.Step, len(stage.Steps)),
        }
        for i, step := range stage.Steps {
            buildfabStage.Steps[i] = buildfab.Step{
                Action:  step.Action,
                Require: step.Require,
                OnError: step.OnError,
                If:      step.If,
                Only:    step.Only,
            }
        }
        config.Stages[name] = buildfabStage
    }
    
    return config
}


// getVersion returns the current version from the VERSION file
func (e *BuildfabExecutor) getVersion() string {
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
func (e *BuildfabExecutor) readVersionFile() (string, error) {
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