// Package install provides Git hook installation functionality.
package install

import (
    "context"
    "crypto/md5"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

// Installer handles installation and management of Git hooks
type Installer struct {
    hookPath string
    binaryPath string
    binaryMD5 string
}

// New creates a new installer
func New() *Installer {
    return &Installer{
        hookPath: ".git/hooks/pre-push",
        binaryPath: "", // Will be set to the current binary path
        binaryMD5: "", // Will be set to the current binary MD5
    }
}

// Install installs or updates the pre-push hook
func (i *Installer) Install(ctx context.Context) error {
    // Check if we're in a git repository
    if !i.isGitRepository() {
        return fmt.Errorf("not in a git repository")
    }
    
    // Get the current binary path
    binaryPath, err := i.getCurrentBinaryPath()
    if err != nil {
        return fmt.Errorf("failed to get current binary path: %w", err)
    }
    i.binaryPath = binaryPath
    
    // Calculate MD5 of the current binary
    binaryMD5, err := i.calculateBinaryMD5()
    if err != nil {
        return fmt.Errorf("failed to calculate binary MD5: %w", err)
    }
    i.binaryMD5 = binaryMD5
    
    // Check if hook already exists and is up to date
    if i.isHookUpToDate() {
        fmt.Println("Pre-push hook is already up to date")
        return nil
    }
    
    // Create the hook directory if it doesn't exist
    hookDir := filepath.Dir(i.hookPath)
    if err := os.MkdirAll(hookDir, 0755); err != nil {
        return fmt.Errorf("failed to create hook directory: %w", err)
    }
    
    // Create the hook script
    if err := i.createHookScript(); err != nil {
        return fmt.Errorf("failed to create hook script: %w", err)
    }
    
    // Make the hook executable
    if err := os.Chmod(i.hookPath, 0755); err != nil {
        return fmt.Errorf("failed to make hook executable: %w", err)
    }
    
    fmt.Println("Pre-push hook installed successfully")
    return nil
}

// Uninstall removes the pre-push hook
func (i *Installer) Uninstall(ctx context.Context) error {
    if !i.isGitRepository() {
        return fmt.Errorf("not in a git repository")
    }
    
    if _, err := os.Stat(i.hookPath); os.IsNotExist(err) {
        fmt.Println("Pre-push hook is not installed")
        return nil
    }
    
    if err := os.Remove(i.hookPath); err != nil {
        return fmt.Errorf("failed to remove hook: %w", err)
    }
    
    fmt.Println("Pre-push hook uninstalled successfully")
    return nil
}

// isGitRepository checks if the current directory is a git repository
func (i *Installer) isGitRepository() bool {
    _, err := os.Stat(".git")
    return err == nil
}

// getCurrentBinaryPath gets the path to the current binary
func (i *Installer) getCurrentBinaryPath() (string, error) {
    // Get the absolute path to the current binary
    binaryPath, err := os.Executable()
    if err != nil {
        return "", err
    }
    
    // Resolve any symlinks
    resolvedPath, err := filepath.EvalSymlinks(binaryPath)
    if err != nil {
        return "", err
    }
    
    return resolvedPath, nil
}

// calculateBinaryMD5 calculates the MD5 hash of the current binary
func (i *Installer) calculateBinaryMD5() (string, error) {
    file, err := os.Open(i.binaryPath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    hash := md5.New()
    if _, err := io.Copy(hash, file); err != nil {
        return "", err
    }
    
    return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// isHookUpToDate checks if the hook is already installed and up to date
func (i *Installer) isHookUpToDate() bool {
    // Check if hook file exists
    if _, err := os.Stat(i.hookPath); os.IsNotExist(err) {
        return false
    }
    
    // Calculate MD5 of the hook file
    hookMD5, err := i.calculateHookMD5()
    if err != nil {
        return false
    }
    
    // Compare with current binary MD5
    return hookMD5 == i.binaryMD5
}

// calculateHookMD5 calculates the MD5 hash of the hook file
func (i *Installer) calculateHookMD5() (string, error) {
    file, err := os.Open(i.hookPath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    hash := md5.New()
    if _, err := io.Copy(hash, file); err != nil {
        return "", err
    }
    
    return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// createHookScript copies the current binary to the hook location
func (i *Installer) createHookScript() error {
    // Open the current binary
    sourceFile, err := os.Open(i.binaryPath)
    if err != nil {
        return fmt.Errorf("failed to open current binary: %w", err)
    }
    defer sourceFile.Close()
    
    // Create the hook file
    destFile, err := os.OpenFile(i.hookPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
    if err != nil {
        return fmt.Errorf("failed to create hook file: %w", err)
    }
    defer destFile.Close()
    
    // Copy the binary to the hook location
    if _, err := io.Copy(destFile, sourceFile); err != nil {
        return fmt.Errorf("failed to copy binary to hook: %w", err)
    }
    
    return nil
}