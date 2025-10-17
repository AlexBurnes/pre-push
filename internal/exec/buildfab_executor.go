// Package exec provides execution functionality for pre-push checks using buildfab DAG executor.
package exec

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "github.com/AlexBurnes/buildfab/pkg/buildfab"
    "github.com/AlexBurnes/pre-push/internal/version"
    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

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
    GetVerboseLevel() int
    IsDebug() bool
}

// GitPushInfo contains information about the Git push operation
type GitPushInfo struct {
    RemoteName string
    RemoteURL  string
    Refs       []GitRef
    Tags       []string
    Branches   []string
    IsDelete   bool
}

// GitRef represents a Git reference being pushed
type GitRef struct {
    LocalRef  string
    LocalSHA  string
    RemoteRef string
    RemoteSHA string
    IsDelete  bool
    IsTag     bool
    IsBranch  bool
}

// BuildfabExecutor handles execution of pre-push stages and actions using buildfab DAG executor
type BuildfabExecutor struct {
    config *buildfab.Config
    ui     UI
    versionDetector *version.Detector
    cliVersion string
    gitPushInfo *GitPushInfo
}


// NewBuildfabExecutor creates a new buildfab-based executor
func NewBuildfabExecutor(config *buildfab.Config, ui UI) *BuildfabExecutor {
    return &BuildfabExecutor{
        config: config,
        ui:     ui,
        versionDetector: version.New(),
        cliVersion: "unknown",
    }
}

// BuildfabExecutorWithCLIVersion creates a new buildfab-based executor with CLI version
func BuildfabExecutorWithCLIVersion(config *buildfab.Config, ui UI, cliVersion string) *BuildfabExecutor {
    return &BuildfabExecutor{
        config: config,
        ui:     ui,
        versionDetector: version.New(),
        cliVersion: cliVersion,
        gitPushInfo: nil,
    }
}

// SetGitPushInfo sets the Git push information for enhanced variable interpolation
func (e *BuildfabExecutor) SetGitPushInfo(pushInfo *GitPushInfo) {
    e.gitPushInfo = pushInfo
}

// findBuildfabBinary searches for buildfab binary in system directories
func findBuildfabBinary() (string, error) {
    // Get HOME directory
    homeDir, err := os.UserHomeDir()
    if err != nil {
        homeDir = ""
    }
    
    // Get current executable directory
    execPath, err := os.Executable()
    if err != nil {
        execPath = ""
    }
    execDir := filepath.Dir(execPath)
    
    // Define search paths in order of preference
    searchPaths := []string{
        "/usr/local/bin/buildfab",
        "/usr/bin/buildfab",
    }
    
    // Add HOME/bin if HOME is available
    if homeDir != "" {
        searchPaths = append(searchPaths, filepath.Join(homeDir, "bin", "buildfab"))
    }
    
    // Add current executable directory
    if execDir != "" {
        searchPaths = append(searchPaths, filepath.Join(execDir, "buildfab"))
    }
    
    // Add ./scripts directory (relative to working directory)
    searchPaths = append(searchPaths, "./scripts/buildfab")
    searchPaths = append(searchPaths, "scripts/buildfab")
    
    // Search for buildfab in each path
    for _, path := range searchPaths {
        if _, err := os.Stat(path); err == nil {
            // Check if file is executable
            if info, err := os.Stat(path); err == nil && !info.IsDir() {
                // Get absolute path
                absPath, err := filepath.Abs(path)
                if err == nil {
                    return absPath, nil
                }
                return path, nil
            }
        }
    }
    
    // Try using exec.LookPath as fallback (searches PATH)
    if path, err := exec.LookPath("buildfab"); err == nil {
        return path, nil
    }
    
    return "", fmt.Errorf("buildfab binary not found in system directories")
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
    
    // Debug output
    if e.ui.IsDebug() {
        fmt.Fprintf(os.Stderr, "DEBUG: UI IsVerbose=%v, IsDebug=%v\n", e.ui.IsVerbose(), e.ui.IsDebug())
        fmt.Fprintf(os.Stderr, "DEBUG: UI VerboseLevel=%d\n", e.ui.GetVerboseLevel())
    }
    
    // Create simple run options with verbose and debug settings
    opts := buildfab.DefaultSimpleRunOptions()
    opts.VerboseLevel = e.ui.GetVerboseLevel()  // Use UI verbose level directly
    opts.Debug = e.ui.IsDebug()
    opts.WorkingDir = "."
    opts.Output = os.Stdout
    opts.ErrorOutput = os.Stderr
    
    // Search for buildfab binary for container support
    buildfabPath, err := findBuildfabBinary()
    if err != nil {
        // Log warning but continue - containers will fail if they need buildfab
        if e.ui.IsDebug() {
            fmt.Fprintf(os.Stderr, "DEBUG: Could not find buildfab binary: %v\n", err)
            fmt.Fprintf(os.Stderr, "DEBUG: Container actions using run_action will not work\n")
        }
    } else {
        opts.BuildfabBinaryPath = buildfabPath
        if e.ui.IsDebug() {
            fmt.Fprintf(os.Stderr, "DEBUG: Found buildfab binary: %s\n", buildfabPath)
        }
    }
    
    // Debug: Log SimpleRunOptions configuration
    if e.ui.IsDebug() {
        fmt.Fprintf(os.Stderr, "DEBUG: SimpleRunOptions:\n")
        fmt.Fprintf(os.Stderr, "  VerboseLevel: %d\n", opts.VerboseLevel)
        fmt.Fprintf(os.Stderr, "  Debug: %v\n", opts.Debug)
        fmt.Fprintf(os.Stderr, "  WorkingDir: %s\n", opts.WorkingDir)
        fmt.Fprintf(os.Stderr, "  Output: %T\n", opts.Output)
        fmt.Fprintf(os.Stderr, "  ErrorOutput: %T\n", opts.ErrorOutput)
        fmt.Fprintf(os.Stderr, "  BuildfabBinaryPath: %s\n", opts.BuildfabBinaryPath)
    }
    
    // Pass variables to buildfab for interpolation
    variables := e.GetAllVariables()
    opts.Variables = variables
    
    // Debug: Log variables
    if e.ui.IsDebug() {
        fmt.Fprintf(os.Stderr, "DEBUG: Variables passed to buildfab:\n")
        for k, v := range variables {
            if len(v) > 50 {
                fmt.Fprintf(os.Stderr, "  %s = %s... (truncated)\n", k, v[:50])
            } else {
                fmt.Fprintf(os.Stderr, "  %s = %s\n", k, v)
            }
        }
    }
    
    // Debug: Log stage configuration
    if e.ui.IsDebug() {
        stage, _ := e.config.GetStage(stageName)
        fmt.Fprintf(os.Stderr, "DEBUG: Stage '%s' has %d steps\n", stageName, len(stage.Steps))
        for i, step := range stage.Steps {
            fmt.Fprintf(os.Stderr, "DEBUG: Step %d: action=%s\n", i+1, step.Action)
            if action, exists := e.config.GetAction(step.Action); exists {
                fmt.Fprintf(os.Stderr, "DEBUG:   Action type: ")
                if action.Run != "" {
                    fmt.Fprintf(os.Stderr, "run command\n")
                } else if action.Uses != "" {
                    fmt.Fprintf(os.Stderr, "uses %s\n", action.Uses)
                } else if action.Container != nil {
                    fmt.Fprintf(os.Stderr, "container action\n")
                    fmt.Fprintf(os.Stderr, "DEBUG:   Container image: %s\n", action.Container.Image.From)
                    fmt.Fprintf(os.Stderr, "DEBUG:   Container run_action: %s\n", action.Container.RunAction)
                }
            }
        }
    }
    
    // Create simple runner
    runner := buildfab.NewSimpleRunner(e.config, opts)
    
    // Debug: Log before execution
    if e.ui.IsDebug() {
        fmt.Fprintf(os.Stderr, "DEBUG: Starting stage execution via SimpleRunner\n")
    }
    
    // Execute the stage - buildfab handles all output automatically
    err = runner.RunStage(ctx, stageName)
    
    // Debug: Log after execution
    if e.ui.IsDebug() {
        if err != nil {
            fmt.Fprintf(os.Stderr, "DEBUG: Stage execution failed: %v\n", err)
        } else {
            fmt.Fprintf(os.Stderr, "DEBUG: Stage execution completed successfully\n")
        }
    }
    
    return err
}

// RunAction executes a specific action using buildfab SimpleRunner
func (e *BuildfabExecutor) RunAction(ctx context.Context, actionName string) error {
    // Create simple run options
    opts := buildfab.DefaultSimpleRunOptions()
    opts.VerboseLevel = e.ui.GetVerboseLevel()  // Use UI verbose level directly
    opts.Debug = e.ui.IsDebug()
    opts.WorkingDir = "."
    
    // Pass variables to buildfab for interpolation
    variables := e.GetAllVariables()
    opts.Variables = variables
    
    // Create simple runner (handles all output internally)
    runner := buildfab.NewSimpleRunner(e.config, opts)

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
    actions := make([]prepush.Action, len(e.config.Actions))
    for i, action := range e.config.Actions {
        actions[i] = prepush.Action{
            Name: action.Name,
            Run:  action.Run,
            Uses: action.Uses,
        }
    }
    return actions
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
    
    // Add version library variables
    if versionInfo, err := version.GetVersionInfo(context.Background()); err == nil && versionInfo != nil {
        // Standard version variables
        if versionInfo.Version != "" {
            variables["version"] = versionInfo.Version
            variables["version.version"] = versionInfo.Version
        }
        if versionInfo.Project != "" {
            variables["project"] = versionInfo.Project
            variables["version.project"] = versionInfo.Project
        }
        if versionInfo.Module != "" {
            variables["module"] = versionInfo.Module
            variables["version.module"] = versionInfo.Module
        }
        if len(versionInfo.Modules) > 0 {
            modulesStr := strings.Join(versionInfo.Modules, ",")
            variables["modules"] = modulesStr
            variables["version.modules"] = modulesStr
        }
        // New build-type and version-type variables
        if versionInfo.BuildType != "" {
            variables["version.build-type"] = versionInfo.BuildType
        }
        if versionInfo.VersionType != "" {
            variables["version.version-type"] = versionInfo.VersionType
        }
    }
    
    // Add enhanced Git variables from push information
    if e.gitPushInfo != nil {
        // Current push information
        if len(e.gitPushInfo.Tags) > 0 {
            // Single tag (most common case)
            if len(e.gitPushInfo.Tags) == 1 {
                variables["tag"] = e.gitPushInfo.Tags[0]
            }
            // Multiple tags as comma-separated list
            variables["tags"] = strings.Join(e.gitPushInfo.Tags, ",")
        }
        
        if len(e.gitPushInfo.Branches) > 0 {
            // Single branch (most common case)
            if len(e.gitPushInfo.Branches) == 1 {
                variables["branch"] = e.gitPushInfo.Branches[0]
            }
            // Multiple branches as comma-separated list
            variables["branches"] = strings.Join(e.gitPushInfo.Branches, ",")
        }
        
        // Version-specific variables for current branch and tag
        if currentBranch, err := e.versionDetector.DetectCurrentBranch(context.Background()); err == nil && currentBranch != "" {
            variables["version.branch"] = currentBranch
        }
        if currentTag, err := e.versionDetector.DetectCurrentVersion(context.Background()); err == nil && currentTag != "" {
            variables["version.tag"] = currentTag
        }
    } else {
        // Fallback to legacy Git detection if no push info available
        if tag, err := e.versionDetector.DetectCurrentVersion(context.Background()); err == nil && tag != "" {
            variables["tag"] = tag
        }
        if branch, err := e.versionDetector.DetectCurrentBranch(context.Background()); err == nil && branch != "" {
            variables["branch"] = branch
        }
    }
    
    // Add environment variables
    for _, env := range os.Environ() {
        parts := strings.SplitN(env, "=", 2)
        if len(parts) == 2 {
            variables["env."+parts[0]] = parts[1]
        }
    }
    
    // Add buildfab platform variables using the helper function (these will have platform. prefix)
    variables = buildfab.AddPlatformVariables(variables)
    
    // Add buildfab version variables (including version.rawversion, version.major, version.minor, version.patch, etc.)
    variables = buildfab.AddVersionVariables(variables)
    
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