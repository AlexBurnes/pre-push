// Package config provides configuration loading and validation functionality.
package config

import (
    "context"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"

    "gopkg.in/yaml.v3"
    "github.com/AlexBurnes/buildfab/pkg/buildfab"
    "github.com/AlexBurnes/pre-push/internal/version"
    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

// Loader handles loading and parsing configuration files
type Loader struct {
    configPath string
}

// New creates a new configuration loader
func New(configPath string) *Loader {
    return &Loader{
        configPath: configPath,
    }
}

// Load loads configuration from the specified file
func Load(configPath string) (*prepush.Config, error) {
    loader := New(configPath)
    return loader.Load()
}

// LoadWithBuildfab loads configuration using buildfab's LoadConfig (supports includes)
func LoadWithBuildfab(configPath string) (*prepush.Config, error) {
    // Use buildfab's LoadConfig which handles includes
    buildfabConfig, err := buildfab.LoadConfig(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load configuration with buildfab: %w", err)
    }
    
    // Convert buildfab.Config to prepush.Config
    return convertBuildfabToPrepushConfig(buildfabConfig), nil
}

// convertBuildfabToPrepushConfig converts buildfab.Config to prepush.Config
func convertBuildfabToPrepushConfig(buildfabConfig *buildfab.Config) *prepush.Config {
    config := &prepush.Config{}
    
    // Convert project
    config.Project.Name = buildfabConfig.Project.Name
    config.Project.Modules = buildfabConfig.Project.Modules
    config.Project.BinDir = buildfabConfig.Project.BinDir
    
    // Convert actions
    config.Actions = make([]prepush.Action, len(buildfabConfig.Actions))
    for i, action := range buildfabConfig.Actions {
        config.Actions[i] = prepush.Action{
            Name: action.Name,
            Run:  action.Run,
            Uses: action.Uses,
        }
    }
    
    // Convert stages
    config.Stages = make(map[string]prepush.Stage)
    for name, stage := range buildfabConfig.Stages {
        prepushStage := prepush.Stage{
            Steps: make([]prepush.Step, len(stage.Steps)),
        }
        for i, step := range stage.Steps {
            prepushStage.Steps[i] = prepush.Step{
                Action:  step.Action,
                Require: step.Require,
                OnError: step.OnError,
                If:      step.If,
                Only:    step.Only,
            }
        }
        config.Stages[name] = prepushStage
    }
    
    return config
}

// Load loads and parses the configuration file
func (l *Loader) Load() (*prepush.Config, error) {
    // Check if file exists
    if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("configuration file not found: %s", l.configPath)
    }

    // Open file
    file, err := os.Open(l.configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open configuration file: %w", err)
    }
    defer file.Close()

    // Read file content
    content, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read configuration file: %w", err)
    }

    // Parse YAML
    var config prepush.Config
    if err := yaml.Unmarshal(content, &config); err != nil {
        return nil, fmt.Errorf("failed to parse YAML configuration: %w", err)
    }

    // Validate configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    return &config, nil
}

// LoadFromDir searches for configuration files in the specified directory
func LoadFromDir(dir string) (*prepush.Config, error) {
    // Common configuration file names
    configFiles := []string{
        ".project.yml",
        ".project.yaml",
        "project.yml",
        "project.yaml",
        "pre-push.yml",
        "pre-push.yaml",
    }

    for _, filename := range configFiles {
        configPath := filepath.Join(dir, filename)
        if _, err := os.Stat(configPath); err == nil {
            return Load(configPath)
        }
    }

    return nil, fmt.Errorf("no configuration file found in directory: %s", dir)
}

// LoadFromString loads configuration from a string
func LoadFromString(content string) (*prepush.Config, error) {
    var config prepush.Config
    if err := yaml.Unmarshal([]byte(content), &config); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }

    return &config, nil
}

// ResolveVariables resolves variable interpolation in configuration
func ResolveVariables(config *prepush.Config, variables map[string]string) error {
    // Resolve variables in action run commands
    for i := range config.Actions {
        if config.Actions[i].Run != "" {
            resolved, err := resolveString(config.Actions[i].Run, variables)
            if err != nil {
                return fmt.Errorf("failed to resolve variables in action %s: %w", config.Actions[i].Name, err)
            }
            config.Actions[i].Run = resolved
        }
    }

    return nil
}

// resolveString resolves variable interpolation in a string
func resolveString(s string, variables map[string]string) (string, error) {
    result := s
    
    // Find all variable references in the format ${{ variable }}
    for {
        start := strings.Index(result, "${{")
        if start == -1 {
            break
        }
        
        end := strings.Index(result[start:], "}}")
        if end == -1 {
            return "", fmt.Errorf("unclosed variable reference: %s", result[start:])
        }
        
        end += start + 2 // Adjust for the start position
        
        variableName := strings.TrimSpace(result[start+3 : end-2])
        
        value, exists := variables[variableName]
        if !exists {
            return "", fmt.Errorf("undefined variable: %s", variableName)
        }
        
        result = result[:start] + value + result[end:]
    }
    
    return result, nil
}

// GetDefaultVariables returns default variables available for interpolation
func GetDefaultVariables() map[string]string {
    return map[string]string{
        "tag":    "", // Will be set by version detection
        "branch": "", // Will be set by git detection
        // Version library variables
        "version.version":  "", // Will be set by version library
        "version.project":  "", // Will be set by version library
        "version.module":   "", // Will be set by version library
        "version.modules":  "", // Will be set by version library
        // Platform variables
        "platform.os":      "", // Will be set by platform detection
        "platform.arch":    "", // Will be set by platform detection
        "platform.goos":    "", // Will be set by platform detection
        "platform.goarch":   "", // Will be set by platform detection
        "platform.go":      "", // Will be set by platform detection
    }
}

// DetectAllVariables detects all available variables (Git, version, platform, environment)
func DetectAllVariables(ctx context.Context) (map[string]string, error) {
    variables := make(map[string]string)
    
    // Detect Git variables
    gitVars, err := DetectGitVariables(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to detect Git variables: %w", err)
    }
    for k, v := range gitVars {
        variables[k] = v
    }
    
    // Detect platform variables
    platformVars, err := DetectPlatformVariables(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to detect platform variables: %w", err)
    }
    for k, v := range platformVars {
        variables[k] = v
    }
    
    // Detect environment variables
    envVars := DetectEnvironmentVariables()
    for k, v := range envVars {
        variables[k] = v
    }
    
    return variables, nil
}

// DetectGitVariables detects Git-related variables from the current repository
func DetectGitVariables(ctx context.Context) (map[string]string, error) {
    variables := make(map[string]string)
    
    // Detect current tag
    if tag, err := detectGitTag(ctx); err == nil {
        variables["tag"] = tag
    }
    
    // Detect current branch
    if branch, err := detectGitBranch(ctx); err == nil {
        variables["branch"] = branch
    }
    
    // Detect version library variables
    if versionInfo, err := version.GetVersionInfo(ctx); err == nil {
        if versionInfo.Version != "" {
            variables["version.version"] = versionInfo.Version
        }
        if versionInfo.Project != "" {
            variables["version.project"] = versionInfo.Project
        }
        if versionInfo.Module != "" {
            variables["version.module"] = versionInfo.Module
        }
        if len(versionInfo.Modules) > 0 {
            variables["version.modules"] = strings.Join(versionInfo.Modules, " ")
        }
    }
    
    return variables, nil
}

// DetectPlatformVariables detects platform-specific variables
func DetectPlatformVariables(ctx context.Context) (map[string]string, error) {
    variables := make(map[string]string)
    
    // Basic platform information
    variables["platform.os"] = runtime.GOOS
    variables["platform.arch"] = runtime.GOARCH
    variables["platform.goos"] = runtime.GOOS
    variables["platform.goarch"] = runtime.GOARCH
    variables["platform.go"] = runtime.Version()
    
    // Detect operating system details
    if osName, err := detectOSName(); err == nil {
        variables["platform.os.name"] = osName
    }
    
    // Detect architecture details
    if archName, err := detectArchName(); err == nil {
        variables["platform.arch.name"] = archName
    }
    
    // Detect shell information
    if shell, err := detectShell(); err == nil {
        variables["platform.shell"] = shell
    }
    
    return variables, nil
}

// DetectEnvironmentVariables detects environment variables for substitution
func DetectEnvironmentVariables() map[string]string {
    variables := make(map[string]string)
    
    // Common environment variables
    envVars := []string{
        "HOME", "USER", "USERNAME", "PATH", "PWD", "SHELL",
        "GOPATH", "GOROOT", "GOBIN", "GOCACHE", "GOMODCACHE",
        "CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL",
        "BUILD_NUMBER", "BUILD_ID", "JOB_NAME", "WORKSPACE",
    }
    
    for _, envVar := range envVars {
        if value := os.Getenv(envVar); value != "" {
            variables["env."+strings.ToLower(envVar)] = value
        }
    }
    
    // Add all environment variables with env. prefix
    for _, env := range os.Environ() {
        if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
            key := strings.ToLower(parts[0])
            value := parts[1]
            variables["env."+key] = value
        }
    }
    
    return variables
}

// detectGitTag detects the current Git tag
func detectGitTag(ctx context.Context) (string, error) {
    cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// detectGitBranch detects the current Git branch
func detectGitBranch(ctx context.Context) (string, error) {
    cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// detectOSName detects the operating system name
func detectOSName() (string, error) {
    switch runtime.GOOS {
    case "linux":
        // Try to detect Linux distribution
        if data, err := os.ReadFile("/etc/os-release"); err == nil {
            lines := strings.Split(string(data), "\n")
            for _, line := range lines {
                if strings.HasPrefix(line, "PRETTY_NAME=") {
                    name := strings.TrimPrefix(line, "PRETTY_NAME=")
                    name = strings.Trim(name, "\"")
                    return name, nil
                }
            }
        }
        return "Linux", nil
    case "darwin":
        return "macOS", nil
    case "windows":
        return "Windows", nil
    default:
        return runtime.GOOS, nil
    }
}

// detectArchName detects the architecture name
func detectArchName() (string, error) {
    switch runtime.GOARCH {
    case "amd64":
        return "x86_64", nil
    case "arm64":
        return "aarch64", nil
    case "386":
        return "i386", nil
    case "arm":
        return "arm", nil
    default:
        return runtime.GOARCH, nil
    }
}

// detectShell detects the current shell
func detectShell() (string, error) {
    if shell := os.Getenv("SHELL"); shell != "" {
        return filepath.Base(shell), nil
    }
    
    // Fallback detection
    switch runtime.GOOS {
    case "windows":
        return "cmd", nil
    case "darwin":
        return "zsh", nil
    default:
        return "bash", nil
    }
}