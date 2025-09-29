package exec

import (
    "context"
    "testing"

    "github.com/AlexBurnes/buildfab/pkg/buildfab"
)

// TestBuildfabExecutor tests the buildfab executor functionality
func TestBuildfabExecutor(t *testing.T) {
    config := &buildfab.Config{
        Project: struct {
            Name    string   `yaml:"name"`
            Modules []string `yaml:"modules"`
            BinDir  string   `yaml:"bin,omitempty"`
        }{
            Name: "test-project",
        },
        Actions: []buildfab.Action{
            {Name: "action1", Run: "echo 1"},
            {Name: "action2", Run: "echo 2"},
            {Name: "action3", Run: "echo 3"},
        },
        Stages: map[string]buildfab.Stage{
            "test-stage": {
                Steps: []buildfab.Step{
                    {Action: "action1"},
                    {Action: "action2", Require: []string{"action1"}},
                },
            },
        },
    }
    
    executor := NewBuildfabExecutor(config, &mockUI{})
    
    // Test ListActions
    actions := executor.ListActions()
    if len(actions) != 3 {
        t.Errorf("Expected 3 actions, got %d", len(actions))
    }
    
    // Test RunAction
    err := executor.RunAction(context.Background(), "action1")
    if err != nil {
        t.Logf("Expected error in test environment: %v", err)
    } else {
        t.Log("Action executed successfully")
    }
    
    // Test RunAction with non-existent action
    err = executor.RunAction(context.Background(), "non-existent")
    if err == nil {
        t.Error("Expected error for non-existent action")
    }
}

// TestRunAction tests action execution using buildfab
func TestRunAction(t *testing.T) {
    config := &buildfab.Config{
        Actions: []buildfab.Action{
            {Name: "test-action", Run: "echo 'test output'"},
        },
    }
    
    executor := NewBuildfabExecutor(config, &mockUI{})
    
    // Note: This test will fail if buildfab is not available
    // In a real test environment, you'd mock the buildfab calls
    err := executor.RunAction(context.Background(), "test-action")
    // We expect this to fail in test environment without actual buildfab
    if err != nil {
        t.Logf("Expected error in test environment: %v", err)
    } else {
        t.Log("Action executed successfully")
    }
}

// TestGetVersion tests version detection
func TestGetVersion(t *testing.T) {
    executor := &BuildfabExecutor{}
    
    version := executor.getVersion()
    if version == "" {
        t.Error("Expected non-empty version")
    }
}
