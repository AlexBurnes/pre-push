// Package uses provides built-in action implementations.
package uses

import (
    "context"
    "fmt"
    "os/exec"
    "strings"

    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

// Runner defines the interface for action runners
type Runner interface {
    Run(ctx context.Context) (prepush.Result, error)
    GetRepro() string
    GetHelp() string
    GetName() string
}

// Registry manages built-in action runners
type Registry struct {
    runners map[string]Runner
}

// New creates a new registry
func New() *Registry {
    registry := &Registry{
        runners: make(map[string]Runner),
    }
    
    // Register built-in actions
    registry.Register("git@untracked", &GitUntrackedRunner{})
    registry.Register("git@uncommitted", &GitUncommittedRunner{})
    registry.Register("git@modified", &GitModifiedRunner{})
    
    return registry
}

// Register registers a runner for a specific action
func (r *Registry) Register(name string, runner Runner) {
    r.runners[name] = runner
}

// GetRunner returns the runner for a specific action
func (r *Registry) GetRunner(name string) (Runner, bool) {
    runner, exists := r.runners[name]
    return runner, exists
}

// ListRunners returns all registered runners
func (r *Registry) ListRunners() map[string]Runner {
    return r.runners
}

// GitUntrackedRunner checks for untracked files
type GitUntrackedRunner struct{}

// Run executes the git untracked check
func (r *GitUntrackedRunner) Run(ctx context.Context) (prepush.Result, error) {
    // Check for untracked files using git status --porcelain
    cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
    output, err := cmd.Output()
    if err != nil {
        return prepush.Result{
            Status: prepush.StatusError,
            Message: fmt.Sprintf("failed to check git status: %v", err),
        }, fmt.Errorf("failed to check git status: %w", err)
    }
    
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    var untrackedFiles []string
    
    for _, line := range lines {
        if len(line) >= 2 && line[0:2] == "??" {
            untrackedFiles = append(untrackedFiles, strings.TrimSpace(line[2:]))
        }
    }
    
    if len(untrackedFiles) > 0 {
        return prepush.Result{
            Status: prepush.StatusError,
            Message: "untracked files found, to manually check run:\n     git status",
        }, fmt.Errorf("untracked files found")
    }
    
    return prepush.Result{
        Status: prepush.StatusOK,
        Message: "no untracked files found",
    }, nil
}

// GetRepro returns the reproduction command for this check
func (r *GitUntrackedRunner) GetRepro() string {
    return "git status --porcelain"
}

// GetHelp returns help text for this action
func (r *GitUntrackedRunner) GetHelp() string {
    return "Check for untracked files in the working directory"
}

// GetName returns the name of this action
func (r *GitUntrackedRunner) GetName() string {
    return "git@untracked"
}

// GitUncommittedRunner checks for uncommitted changes
type GitUncommittedRunner struct{}

// Run executes the git uncommitted check
func (r *GitUncommittedRunner) Run(ctx context.Context) (prepush.Result, error) {
    // Check for uncommitted changes using git status --porcelain
    cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
    output, err := cmd.Output()
    if err != nil {
        return prepush.Result{
            Status: prepush.StatusError,
            Message: fmt.Sprintf("failed to check git status: %v", err),
        }, fmt.Errorf("failed to check git status: %w", err)
    }
    
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    var uncommittedFiles []string
    
    for _, line := range lines {
        if len(line) >= 2 {
            status := line[0:2]
            // Check for staged (M, A, D, R, C) or unstaged (M, D) changes
            if (status[0] != ' ' && status[0] != '?') || (status[1] != ' ' && status[1] != '?') {
                uncommittedFiles = append(uncommittedFiles, strings.TrimSpace(line[2:]))
            }
        }
    }
    
    if len(uncommittedFiles) > 0 {
        return prepush.Result{
            Status: prepush.StatusError,
            Message: "uncommitted changes found, to manually check run:\n     git status",
        }, fmt.Errorf("uncommitted changes found")
    }
    
    return prepush.Result{
        Status: prepush.StatusOK,
        Message: "no uncommitted changes found",
    }, nil
}

// GetRepro returns the reproduction command for this check
func (r *GitUncommittedRunner) GetRepro() string {
    return "git status --porcelain"
}

// GetHelp returns help text for this action
func (r *GitUncommittedRunner) GetHelp() string {
    return "Check for uncommitted changes in the working directory"
}

// GetName returns the name of this action
func (r *GitUncommittedRunner) GetName() string {
    return "git@uncommitted"
}

// GitModifiedRunner checks for modified files
type GitModifiedRunner struct{}

// Run executes the git modified check
func (r *GitModifiedRunner) Run(ctx context.Context) (prepush.Result, error) {
    // Check if working tree differs from HEAD
    cmd := exec.CommandContext(ctx, "git", "diff", "--quiet", "HEAD")
    err := cmd.Run()
    
    if err != nil {
        // Command failed, which means there are differences
        // Get the actual diff to show what's modified
        diffCmd := exec.CommandContext(ctx, "git", "diff", "--name-only", "HEAD")
        diffOutput, diffErr := diffCmd.Output()
        if diffErr != nil {
            return prepush.Result{
                Status: prepush.StatusError,
                Message: "working tree differs from HEAD (unable to get file list)",
            }, fmt.Errorf("working tree differs from HEAD: %w", err)
        }
        
        modifiedFiles := strings.Split(strings.TrimSpace(string(diffOutput)), "\n")
        if len(modifiedFiles) == 1 && modifiedFiles[0] == "" {
            modifiedFiles = []string{}
        }
        
        return prepush.Result{
            Status: prepush.StatusError,
            Message: "working tree differs from HEAD, to manually check run:\n     git diff",
        }, fmt.Errorf("working tree differs from HEAD")
    }
    
    return prepush.Result{
        Status: prepush.StatusOK,
        Message: "working tree matches HEAD",
    }, nil
}

// GetRepro returns the reproduction command for this check
func (r *GitModifiedRunner) GetRepro() string {
    return "git diff --quiet HEAD"
}

// GetHelp returns help text for this action
func (r *GitModifiedRunner) GetHelp() string {
    return "Check if working tree differs from HEAD"
}

// GetName returns the name of this action
func (r *GitModifiedRunner) GetName() string {
    return "git@modified"
}
