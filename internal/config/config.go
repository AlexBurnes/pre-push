// Package config provides configuration loading and validation functionality.
package config

import (
    "context"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v3"
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
    }
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