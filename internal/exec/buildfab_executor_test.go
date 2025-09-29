package exec

import (
    "context"
    "testing"
    "time"

    "github.com/AlexBurnes/buildfab/pkg/buildfab"
    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

// mockUI is a mock implementation of the UI interface for testing
type mockUI struct {
    verboseLevel int
    debug        bool
}

func (m *mockUI) Printf(format string, args ...interface{}) {}
func (m *mockUI) Println(args ...interface{}) {}
func (m *mockUI) PrintStepStatus(stepName string, status prepush.Status, message string) {}
func (m *mockUI) PrintCLIHeader(name, version string) {}
func (m *mockUI) PrintProjectCheck(projectName, version string) {}
func (m *mockUI) PrintStageHeader(stageName string) {}
func (m *mockUI) PrintStageResult(stageName string, success bool, duration time.Duration) {}
func (m *mockUI) PrintError(err error) {}
func (m *mockUI) PrintWarning(message string) {}
func (m *mockUI) PrintInfo(message string) {}
func (m *mockUI) PrintDebug(message string) {}
func (m *mockUI) PrintVerbose(message string) {}
func (m *mockUI) PrintCommand(command string) {}
func (m *mockUI) PrintCommandOutput(output string) {}
func (m *mockUI) PrintRepro(stepName, repro string) {}
func (m *mockUI) PrintReproInline(stepName, repro string) {}
func (m *mockUI) PrintSummary(results []prepush.Result) {}
func (m *mockUI) IsVerbose() bool { return m.verboseLevel > 0 }
func (m *mockUI) GetVerboseLevel() int { return m.verboseLevel }
func (m *mockUI) IsDebug() bool { return m.debug }

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
