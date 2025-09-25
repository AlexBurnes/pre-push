// Package exec provides execution functionality for pre-push checks using buildfab DAG executor.
package exec

import (
    "context"
    "fmt"
    "os"
    "os/exec"
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
    cliVersion string
}


// NewBuildfabExecutor creates a new buildfab-based executor
func NewBuildfabExecutor(config *prepush.Config, ui UI) *BuildfabExecutor {
    return &BuildfabExecutor{
        config: config,
        ui:     ui,
        versionDetector: version.New(),
        cliVersion: "unknown",
    }
}

// NewBuildfabExecutorWithCLIVersion creates a new buildfab-based executor with CLI version
func NewBuildfabExecutorWithCLIVersion(config *prepush.Config, ui UI, cliVersion string) *BuildfabExecutor {
    return &BuildfabExecutor{
        config: config,
        ui:     ui,
        versionDetector: version.New(),
        cliVersion: cliVersion,
    }
}

// RunStage executes a specific stage using buildfab SimpleRunner
func (e *BuildfabExecutor) RunStage(ctx context.Context, stageName string) error {
    _, exists := e.config.GetStage(stageName)
    if !exists {
        return fmt.Errorf("stage not found: %s", stageName)
    }

    // Print CLI header and project check first
    projectVersion := e.getVersion()
    cliVersion := e.getCLIVersion()
    e.ui.PrintCLIHeader("pre-push", cliVersion)
    e.ui.PrintProjectCheck(e.config.Project.Name, projectVersion)

    // Convert pre-push config to buildfab config
    buildfabConfig := e.convertToBuildfabConfig()
    
    // Debug output (remove in production)
    if e.ui.IsDebug() {
        fmt.Fprintf(os.Stderr, "DEBUG: UI IsVerbose=%v, IsDebug=%v\n", e.ui.IsVerbose(), e.ui.IsDebug())
    }
    
    // Create simple run options with verbose and debug settings
    opts := buildfab.DefaultSimpleRunOptions()
    opts.Verbose = e.ui.IsVerbose()  // Use UI verbose setting
    opts.Debug = e.ui.IsDebug()
    opts.WorkingDir = "."
    opts.Output = os.Stdout
    opts.ErrorOutput = os.Stderr
    
    // Create simple runner
    runner := buildfab.NewSimpleRunner(buildfabConfig, opts)
    
    // Execute the stage - buildfab handles all output automatically
    return runner.RunStage(ctx, stageName)
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
    
    // Convert actions (buildfab will handle variable interpolation)
    config.Actions = make([]buildfab.Action, len(e.config.Actions))
    for i, action := range e.config.Actions {
        config.Actions[i] = buildfab.Action{
            Name: action.Name,
            Run:  action.Run,
            Uses: action.Uses,
        }
    }
    
    // Convert stages (buildfab will handle variable interpolation)
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


// getVersion returns the current version using the version-go library
func (e *BuildfabExecutor) getVersion() string {
    // Use the version library to get project version
    if info, err := version.GetVersionInfo(context.Background()); err == nil && info != nil {
        return info.Version
    }
    
    // Fallback to git tag detection
    if version, err := e.versionDetector.DetectCurrentVersion(context.Background()); err == nil {
        return version
    }
    
    // Final fallback
    return "unknown"
}

// getCLIVersion returns the CLI version (compiled-in version)
func (e *BuildfabExecutor) getCLIVersion() string {
    return e.cliVersion
}

// GetAllVariables combines all available variables (Git, version, platform, environment)
func (e *BuildfabExecutor) GetAllVariables() map[string]string {
    variables := make(map[string]string)
    
    // Add buildfab platform variables with simple names
    platformVars := buildfab.GetPlatformVariables()
    if platformVars != nil {
        variables["platform"] = platformVars.Platform
        variables["arch"] = platformVars.Arch
        variables["os"] = platformVars.OS
        variables["os_version"] = platformVars.OSVersion
        variables["cpu"] = fmt.Sprintf("%d", platformVars.CPU)
    }
    
    // Add buildfab platform variables using the helper function (these will have platform. prefix)
    variables = buildfab.AddPlatformVariables(variables)
    
    // Add Git and version variables
    if gitVars, err := e.detectGitVariables(context.Background()); err == nil {
        for k, v := range gitVars {
            variables[k] = v
        }
    }
    
    // Add environment variables
    for _, env := range os.Environ() {
        if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
            key := strings.ToLower(parts[0])
            value := parts[1]
            variables["env."+key] = value
        }
    }
    
    return variables
}

// detectGitVariables detects Git-related variables
func (e *BuildfabExecutor) detectGitVariables(ctx context.Context) (map[string]string, error) {
    variables := make(map[string]string)
    
    // Detect current tag
    if tag, err := e.detectGitTag(ctx); err == nil {
        variables["tag"] = tag
    }
    
    // Detect current branch
    if branch, err := e.detectGitBranch(ctx); err == nil {
        variables["branch"] = branch
    }
    
    // Detect version library variables with simple names
    if versionInfo, err := version.GetVersionInfo(ctx); err == nil {
        if versionInfo.Version != "" {
            variables["version"] = versionInfo.Version
            variables["version.version"] = versionInfo.Version  // Keep both for compatibility
        }
        if versionInfo.Project != "" {
            variables["project"] = versionInfo.Project
            variables["version.project"] = versionInfo.Project  // Keep both for compatibility
        }
        if versionInfo.Module != "" {
            variables["module"] = versionInfo.Module
            variables["version.module"] = versionInfo.Module  // Keep both for compatibility
        }
        if len(versionInfo.Modules) > 0 {
            variables["modules"] = strings.Join(versionInfo.Modules, " ")
            variables["version.modules"] = strings.Join(versionInfo.Modules, " ")  // Keep both for compatibility
        }
    }
    
    return variables, nil
}

// detectGitTag detects the current Git tag
func (e *BuildfabExecutor) detectGitTag(ctx context.Context) (string, error) {
    cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// detectGitBranch detects the current Git branch
func (e *BuildfabExecutor) detectGitBranch(ctx context.Context) (string, error) {
    cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
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