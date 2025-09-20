// Package version provides version detection and validation functionality.
package version

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// Detector handles version detection and validation
type Detector struct{}

// New creates a new version detector
func New() *Detector {
    return &Detector{}
}

// DetectCurrentVersion detects the current version from git tags
func (d *Detector) DetectCurrentVersion(ctx context.Context) (string, error) {
    // Try to get the latest tag
    cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to get git tag: %w", err)
    }
    
    version := strings.TrimSpace(string(output))
    if version == "" {
        return "", fmt.Errorf("no version tags found")
    }
    
    return version, nil
}

// DetectCurrentBranch detects the current branch
func (d *Detector) DetectCurrentBranch(ctx context.Context) (string, error) {
    cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to get current branch: %w", err)
    }
    
    branch := strings.TrimSpace(string(output))
    if branch == "" {
        return "", fmt.Errorf("no current branch found")
    }
    
    return branch, nil
}

// ValidateVersion validates a version string
func (d *Detector) ValidateVersion(version string) error {
    if version == "" {
        return fmt.Errorf("version cannot be empty")
    }
    
    // Basic validation - should start with 'v' and contain numbers
    if !strings.HasPrefix(version, "v") {
        return fmt.Errorf("version should start with 'v': %s", version)
    }
    
    // Check if it contains at least one digit
    hasDigit := false
    for _, r := range version {
        if r >= '0' && r <= '9' {
            hasDigit = true
            break
        }
    }
    
    if !hasDigit {
        return fmt.Errorf("version should contain at least one digit: %s", version)
    }
    
    return nil
}

// CompareVersions compares two version strings
func (d *Detector) CompareVersions(v1, v2 string) (int, error) {
    // Remove 'v' prefix for comparison
    v1 = strings.TrimPrefix(v1, "v")
    v2 = strings.TrimPrefix(v2, "v")
    
    // Simple string comparison for now
    // TODO: Implement proper semantic version comparison
    if v1 == v2 {
        return 0, nil
    } else if v1 < v2 {
        return -1, nil
    } else {
        return 1, nil
    }
}

// IsVersionGreatest checks if the given version is the greatest among all tags
func (d *Detector) IsVersionGreatest(ctx context.Context, version string) (bool, error) {
    // Get all tags
    cmd := exec.CommandContext(ctx, "git", "tag", "--sort=-version:refname")
    output, err := cmd.Output()
    if err != nil {
        return false, fmt.Errorf("failed to get git tags: %w", err)
    }
    
    tags := strings.Split(strings.TrimSpace(string(output)), "\n")
    if len(tags) == 0 {
        return false, fmt.Errorf("no tags found")
    }
    
    // The first tag should be the greatest
    greatest := strings.TrimSpace(tags[0])
    return version == greatest, nil
}