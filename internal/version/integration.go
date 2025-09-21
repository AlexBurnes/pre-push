// Package version provides version detection and validation functionality.
package version

import (
    "context"
    "fmt"
    "os"
    "strings"

    "github.com/AlexBurnes/version-go/pkg/version"
)

// VersionInfo contains version-related information from the version-go library
type VersionInfo struct {
    Version  string
    Project  string
    Module   string
    Modules  []string
}

// GetVersionInfo retrieves version information using the version-go library
func GetVersionInfo(ctx context.Context) (*VersionInfo, error) {
    info := &VersionInfo{}
    
    // Get version from VERSION file or git
    if version, err := getVersionFromLibrary(ctx); err == nil {
        info.Version = version
    }
    
    // Get project name from .project.yml or git
    if project, err := getProjectFromLibrary(ctx); err == nil {
        info.Project = project
    }
    
    // Get module name from .project.yml or git
    if module, err := getModuleFromLibrary(ctx); err == nil {
        info.Module = module
    }
    
    // Get modules list from .project.yml or git
    if modules, err := getModulesFromLibrary(ctx); err == nil {
        info.Modules = modules
    }
    
    return info, nil
}

// getVersionFromLibrary gets the current version using the version-go library
func getVersionFromLibrary(ctx context.Context) (string, error) {
    // First try to read from VERSION file
    if versionStr, err := readVersionFile(); err == nil {
        // Validate the version using the library
        if err := version.Validate(versionStr); err == nil {
            return versionStr, nil
        }
    }
    
    // Fallback to git describe
    return getVersionFromGit(ctx)
}

// readVersionFile reads version from VERSION file
func readVersionFile() (string, error) {
    // Try common VERSION file locations
    versionFiles := []string{
        "VERSION",
        "./VERSION",
        "../VERSION",
        "../../VERSION",
    }
    
    for _, file := range versionFiles {
        if content, err := os.ReadFile(file); err == nil {
            versionStr := strings.TrimSpace(string(content))
            if versionStr != "" {
                return versionStr, nil
            }
        }
    }
    
    return "", fmt.Errorf("VERSION file not found")
}

// getVersionFromGit gets version from git describe
func getVersionFromGit(ctx context.Context) (string, error) {
    // This would require implementing git integration
    // For now, return empty to indicate no version found
    return "", fmt.Errorf("git version detection not implemented")
}

// getProjectFromLibrary gets the project name using the version-go library
func getProjectFromLibrary(ctx context.Context) (string, error) {
    // Create a configuration provider
    cp := version.NewConfigProvider()
    
    // Load configuration from .project.yml
    config, err := cp.LoadProjectConfig()
    if err != nil {
        return "", fmt.Errorf("failed to load project config: %w", err)
    }
    
    if config != nil && config.Project.Name != "" {
        return config.Project.Name, nil
    }
    
    // Fallback to git-based detection
    return getProjectFromGit(ctx)
}

// getProjectFromGit gets project name from git remote
func getProjectFromGit(ctx context.Context) (string, error) {
    // This would require implementing git integration
    // For now, return empty to indicate no project found
    return "", fmt.Errorf("git project detection not implemented")
}

// getModuleFromLibrary gets the module name using the version-go library
func getModuleFromLibrary(ctx context.Context) (string, error) {
    // Create a configuration provider
    cp := version.NewConfigProvider()
    
    // Load configuration from .project.yml
    config, err := cp.LoadProjectConfig()
    if err != nil {
        return "", fmt.Errorf("failed to load project config: %w", err)
    }
    
    if config != nil && len(config.Project.Modules) > 0 {
        return config.Project.Modules[0], nil // First module is primary
    }
    
    // Fallback to git-based detection
    return getModuleFromGit(ctx)
}

// getModuleFromGit gets module name from git remote
func getModuleFromGit(ctx context.Context) (string, error) {
    // This would require implementing git integration
    // For now, return empty to indicate no module found
    return "", fmt.Errorf("git module detection not implemented")
}

// getModulesFromLibrary gets the modules list using the version-go library
func getModulesFromLibrary(ctx context.Context) ([]string, error) {
    // Create a configuration provider
    cp := version.NewConfigProvider()
    
    // Load configuration from .project.yml
    config, err := cp.LoadProjectConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to load project config: %w", err)
    }
    
    if config != nil && len(config.Project.Modules) > 0 {
        return config.Project.Modules, nil
    }
    
    // Fallback to git-based detection
    return getModulesFromGit(ctx)
}

// getModulesFromGit gets modules list from git remote
func getModulesFromGit(ctx context.Context) ([]string, error) {
    // This would require implementing git integration
    // For now, return empty to indicate no modules found
    return []string{}, fmt.Errorf("git modules detection not implemented")
}