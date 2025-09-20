// Package prepush provides the core library API for the pre-push tool.
// It exposes the main functionality for running pre-push checks and
// managing Git hooks in a reusable way.
package prepush

import (
    "context"
    "fmt"
)

// Config represents the pre-push configuration
type Config struct {
    Project struct {
        Name    string   `yaml:"name"`
        Modules []string `yaml:"modules"`
        BinDir  string   `yaml:"bin,omitempty"`
    } `yaml:"project"`
    
    Actions []Action `yaml:"actions"`
    
    Stages map[string]Stage `yaml:"stages"`
}

// Action represents a single action that can be executed
type Action struct {
    Name string `yaml:"name"`
    Run  string `yaml:"run,omitempty"`
    Uses string `yaml:"uses,omitempty"`
}

// Stage represents a collection of steps to execute
type Stage struct {
    Steps []Step `yaml:"steps"`
}

// Step represents a single step in a stage
type Step struct {
    Action  string   `yaml:"action"`
    Require []string `yaml:"require,omitempty"`
    OnError string   `yaml:"onerror,omitempty"`
    If      string   `yaml:"if,omitempty"`
    Only    []string `yaml:"only,omitempty"`
}

// Result represents the result of executing a step
type Result struct {
    Name    string
    Status  Status
    Message string
    Error   error
}

// Status represents the execution status of a step
type Status int

const (
    StatusPending Status = iota
    StatusRunning
    StatusOK
    StatusWarn
    StatusError
    StatusSkipped
)

// String returns the string representation of the status
func (s Status) String() string {
    switch s {
    case StatusPending:
        return "PENDING"
    case StatusRunning:
        return "RUNNING"
    case StatusOK:
        return "OK"
    case StatusWarn:
        return "WARN"
    case StatusError:
        return "ERROR"
    case StatusSkipped:
        return "SKIPPED"
    default:
        return "UNKNOWN"
    }
}

// Executor defines the interface for executing pre-push checks
type Executor interface {
    // RunStage executes a specific stage
    RunStage(ctx context.Context, stageName string) error
    
    // RunAction executes a specific action
    RunAction(ctx context.Context, actionName string) error
    
    // ListActions returns all available actions
    ListActions() []Action
}

// Built-in action registry
var builtInActions = map[string]string{
    "git@untracked":   "Check for untracked files",
    "git@uncommitted": "Check for uncommitted changes",
    "git@modified":    "Check for modified files",
}

// ListBuiltInActions returns a map of built-in action names to their descriptions
func ListBuiltInActions() map[string]string {
    return builtInActions
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
    return &Config{
        Stages: make(map[string]Stage),
    }
}

// Validate validates the configuration
func (c *Config) Validate() error {
    if c.Project.Name == "" {
        return fmt.Errorf("project name is required")
    }
    
    if len(c.Actions) == 0 {
        return fmt.Errorf("at least one action is required")
    }
    
    // Validate actions
    actionNames := make(map[string]bool)
    for _, action := range c.Actions {
        if action.Name == "" {
            return fmt.Errorf("action name is required")
        }
        
        if action.Run == "" && action.Uses == "" {
            return fmt.Errorf("action %s must have either 'run' or 'uses'", action.Name)
        }
        
        if action.Run != "" && action.Uses != "" {
            return fmt.Errorf("action %s cannot have both 'run' and 'uses'", action.Name)
        }
        
        if actionNames[action.Name] {
            return fmt.Errorf("duplicate action name: %s", action.Name)
        }
        actionNames[action.Name] = true
    }
    
    // Validate stages
    for stageName, stage := range c.Stages {
        if len(stage.Steps) == 0 {
            return fmt.Errorf("stage %s must have at least one step", stageName)
        }
        
        for i, step := range stage.Steps {
            if step.Action == "" {
                return fmt.Errorf("step %d in stage %s must have an action", i+1, stageName)
            }
            
            if !actionNames[step.Action] {
                return fmt.Errorf("step %d in stage %s references unknown action: %s", i+1, stageName, step.Action)
            }
            
            if step.OnError != "" && step.OnError != "stop" && step.OnError != "warn" {
                return fmt.Errorf("step %d in stage %s has invalid onerror value: %s (must be 'stop' or 'warn')", i+1, stageName, step.OnError)
            }
            
            // Validate only field contains valid values
            for _, onlyValue := range step.Only {
                if onlyValue != "release" && onlyValue != "prerelease" && onlyValue != "patch" && onlyValue != "minor" && onlyValue != "major" {
                    return fmt.Errorf("step %d in stage %s has invalid only value: %s (must be 'release', 'prerelease', 'patch', 'minor', or 'major')", i+1, stageName, onlyValue)
                }
            }
        }
    }
    
    return nil
}

// GetAction returns the action with the specified name
func (c *Config) GetAction(name string) (Action, bool) {
    for _, action := range c.Actions {
        if action.Name == name {
            return action, true
        }
    }
    return Action{}, false
}

// GetStage returns the stage with the specified name
func (c *Config) GetStage(name string) (Stage, bool) {
    stage, exists := c.Stages[name]
    return stage, exists
}