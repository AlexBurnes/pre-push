// Package version provides version detection and validation functionality.
package version

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"

    "github.com/AlexBurnes/version-go/pkg/version"
)

// VersionInfo contains version-related information from the version-go library
type VersionInfo struct {
    Version     string
    Project     string
    Module      string
    Modules     []string
    BuildType   string
    VersionType string
}

// GetVersionInfo retrieves version information using the version-go library
func GetVersionInfo(ctx context.Context) (*VersionInfo, error) {
    info := &VersionInfo{}
    
    // Get version from VERSION file or git
    if versionStr, err := getVersionFromLibrary(ctx); err == nil {
        info.Version = versionStr
        // Try to use version library's GetVersionType method if available
        if versionType, err := version.GetVersionType(versionStr); err == nil {
            info.VersionType = versionType
        } else {
            // Fallback to our own implementation
            info.VersionType = getVersionType(versionStr)
        }
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
    
    // Try to use version library's GetBuildType method if available
    // GetBuildType requires a version string parameter
    if info.Version != "" {
        if buildType, err := version.GetBuildType(info.Version); err == nil {
            info.BuildType = buildType
        } else {
            // Fallback to our own implementation
            info.BuildType = getBuildType(ctx)
        }
    } else {
        // Fallback to our own implementation if no version available
        info.BuildType = getBuildType(ctx)
    }
    
    return info, nil
}

// getVersionFromLibrary gets the current version using the version-go library
func getVersionFromLibrary(ctx context.Context) (string, error) {
    // Use version-go library to get version from git
    versionStr, err := version.GetVersion()
    if err != nil {
        return "", fmt.Errorf("failed to get version from git: %w", err)
    }
    
    if versionStr == "" {
        return "", fmt.Errorf("no version found")
    }
    
    return versionStr, nil
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

// getVersionFromGit gets version from git describe (deprecated, use getVersionFromLibrary instead)
func getVersionFromGit(ctx context.Context) (string, error) {
    // Deprecated: This function is no longer used
    // Version detection is now handled by version-go library
    return "", fmt.Errorf("git version detection deprecated")
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

// getVersionType determines the type of version string (fallback implementation)
func getVersionType(versionStr string) string {
    // This is a fallback implementation - the main code now uses version.GetVersionType()
    
    if versionStr == "" {
        return "unknown"
    }
    
    // Remove 'v' prefix if present
    versionStr = strings.TrimPrefix(versionStr, "v")
    
    // Check for semantic version pattern (X.Y.Z)
    if strings.Contains(versionStr, "-") {
        // Contains prerelease identifiers (e.g., "1.0.0-alpha", "1.0.0-beta.1")
        return "prerelease"
    }
    
    // Check if it's a snapshot build (contains commit hash or snapshot)
    if strings.Contains(versionStr, "SNAPSHOT") || strings.Contains(versionStr, "-") {
        return "snapshot"
    }
    
    // Check if it matches semantic version pattern (X.Y.Z)
    parts := strings.Split(versionStr, ".")
    if len(parts) >= 3 {
        // Basic semantic version validation
        return "semantic"
    }
    
    return "custom"
}

// getBuildType determines the build type (fallback implementation)
func getBuildType(ctx context.Context) string {
    // This is a fallback implementation - the main code now uses version.GetBuildType()
    
    // Check if we're in a git repository
    if !isGitRepository(ctx) {
        return "unknown"
    }
    
    // Check if there are uncommitted changes
    if hasUncommittedChanges(ctx) {
        return "debug"
    }
    
    // Check if we're on a release tag
    if isOnReleaseTag(ctx) {
        return "release"
    }
    
    // Default to snapshot for development builds
    return "snapshot"
}

// isGitRepository checks if the current directory is a git repository
func isGitRepository(ctx context.Context) bool {
    cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
    err := cmd.Run()
    return err == nil
}

// hasUncommittedChanges checks if there are uncommitted changes
func hasUncommittedChanges(ctx context.Context) bool {
    cmd := exec.CommandContext(ctx, "git", "diff", "--quiet")
    err := cmd.Run()
    return err != nil // If git diff --quiet fails, there are changes
}

// isOnReleaseTag checks if the current HEAD is on a release tag
func isOnReleaseTag(ctx context.Context) bool {
    cmd := exec.CommandContext(ctx, "git", "describe", "--exact-match", "--tags", "HEAD")
    err := cmd.Run()
    return err == nil
}