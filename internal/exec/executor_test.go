package exec

import (
    "context"
    "testing"
    "time"

    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

// mockUI is a mock implementation of the UI interface for testing
type mockUI struct {
    verbose bool
    debug   bool
}

func (m *mockUI) PrintStepStatus(stepName string, status prepush.Status, message string) {}
func (m *mockUI) PrintStageHeader(stageName string) {}
func (m *mockUI) PrintStageResult(stageName string, success bool, duration time.Duration) {}
func (m *mockUI) PrintCommand(command string) {}
func (m *mockUI) PrintCommandOutput(output string) {}
func (m *mockUI) PrintRepro(stepName, repro string) {}
func (m *mockUI) PrintSummary(results []prepush.Result) {}
func (m *mockUI) IsVerbose() bool { return m.verbose }
func (m *mockUI) IsDebug() bool { return m.debug }

func TestBuildDAG(t *testing.T) {
    config := &prepush.Config{
        Actions: []prepush.Action{
            {Name: "action1", Run: "echo 1"},
            {Name: "action2", Run: "echo 2"},
            {Name: "action3", Run: "echo 3"},
        },
    }
    
    steps := []prepush.Step{
        {Action: "action1"},
        {Action: "action2", Require: []string{"action1"}},
        {Action: "action3", Require: []string{"action2"}},
    }
    
    executor := &Executor{
        config: config,
        ui:     &mockUI{},
    }
    
    dag, err := executor.buildDAG(steps)
    if err != nil {
        t.Fatalf("Failed to build DAG: %v", err)
    }
    
    if len(dag) != 3 {
        t.Errorf("Expected 3 nodes in DAG, got %d", len(dag))
    }
    
    // Check dependencies
    if len(dag["action1"].Dependencies) != 0 {
        t.Errorf("Expected action1 to have no dependencies, got %d", len(dag["action1"].Dependencies))
    }
    
    if len(dag["action2"].Dependencies) != 1 || dag["action2"].Dependencies[0] != "action1" {
        t.Errorf("Expected action2 to depend on action1")
    }
    
    if len(dag["action3"].Dependencies) != 1 || dag["action3"].Dependencies[0] != "action2" {
        t.Errorf("Expected action3 to depend on action2")
    }
}

func TestDetectCycles(t *testing.T) {
    config := &prepush.Config{
        Actions: []prepush.Action{
            {Name: "action1", Run: "echo 1"},
            {Name: "action2", Run: "echo 2"},
            {Name: "action3", Run: "echo 3"},
        },
    }
    
    // Test with cycle
    steps := []prepush.Step{
        {Action: "action1", Require: []string{"action3"}},
        {Action: "action2", Require: []string{"action1"}},
        {Action: "action3", Require: []string{"action2"}},
    }
    
    executor := &Executor{
        config: config,
        ui:     &mockUI{},
    }
    
    dag, err := executor.buildDAG(steps)
    if err == nil {
        t.Error("Expected error for cyclic dependencies")
    }
    if dag != nil {
        t.Error("Expected nil DAG for cyclic dependencies")
    }
}

func TestTopologicalSort(t *testing.T) {
    config := &prepush.Config{
        Actions: []prepush.Action{
            {Name: "action1", Run: "echo 1"},
            {Name: "action2", Run: "echo 2"},
            {Name: "action3", Run: "echo 3"},
        },
    }
    
    steps := []prepush.Step{
        {Action: "action1"},
        {Action: "action2", Require: []string{"action1"}},
        {Action: "action3", Require: []string{"action1"}},
    }
    
    executor := &Executor{
        config: config,
        ui:     &mockUI{},
    }
    
    dag, err := executor.buildDAG(steps)
    if err != nil {
        t.Fatalf("Failed to build DAG: %v", err)
    }
    
    order, err := executor.topologicalSort(dag)
    if err != nil {
        t.Fatalf("Failed to sort DAG: %v", err)
    }
    
    if len(order) != 3 {
        t.Errorf("Expected 3 nodes in order, got %d", len(order))
    }
    
    // action1 should come before action2 and action3
    action1Index := -1
    action2Index := -1
    action3Index := -1
    
    for i, action := range order {
        switch action {
        case "action1":
            action1Index = i
        case "action2":
            action2Index = i
        case "action3":
            action3Index = i
        }
    }
    
    if action1Index == -1 || action2Index == -1 || action3Index == -1 {
        t.Error("Expected all actions to be in topological order")
    }
    
    if action1Index >= action2Index || action1Index >= action3Index {
        t.Error("Expected action1 to come before action2 and action3")
    }
}

func TestAllDependenciesCompleted(t *testing.T) {
    executor := &Executor{}
    
    node := &DAGNode{
        Dependencies: []string{"dep1", "dep2"},
    }
    
    completed := map[string]bool{
        "dep1": true,
        "dep2": true,
    }
    
    if !executor.allDependenciesCompleted(node, completed) {
        t.Error("Expected all dependencies to be completed")
    }
    
    completed["dep2"] = false
    
    if executor.allDependenciesCompleted(node, completed) {
        t.Error("Expected not all dependencies to be completed")
    }
}

func TestHasFailedDependency(t *testing.T) {
    executor := &Executor{}
    
    node := &DAGNode{
        Dependencies: []string{"dep1", "dep2"},
    }
    
    failed := map[string]bool{
        "dep1": false,
        "dep2": false,
    }
    
    if executor.hasFailedDependency(node, failed) {
        t.Error("Expected no failed dependencies")
    }
    
    failed["dep1"] = true
    
    if !executor.hasFailedDependency(node, failed) {
        t.Error("Expected failed dependency")
    }
}

func TestExecuteCustomAction(t *testing.T) {
    executor := &Executor{
        ui: &mockUI{verbose: true},
    }
    
    action := prepush.Action{
        Name: "test-action",
        Run:  "echo 'test output'",
    }
    
    result, err := executor.executeCustomAction(context.Background(), action)
    if err != nil {
        t.Fatalf("Failed to execute custom action: %v", err)
    }
    
    if result.Status != prepush.StatusOK {
        t.Errorf("Expected StatusOK, got %v", result.Status)
    }
}

func TestExecuteCustomActionError(t *testing.T) {
    executor := &Executor{
        ui: &mockUI{},
    }
    
    action := prepush.Action{
        Name: "test-action",
        Run:  "exit 1",
    }
    
    result, err := executor.executeCustomAction(context.Background(), action)
    if err == nil {
        t.Error("Expected error for failing command")
    }
    
    if result.Status != prepush.StatusError {
        t.Errorf("Expected StatusError, got %v", result.Status)
    }
}

func TestExecuteCustomActionNoRun(t *testing.T) {
    executor := &Executor{
        ui: &mockUI{},
    }
    
    action := prepush.Action{
        Name: "test-action",
    }
    
    result, err := executor.executeCustomAction(context.Background(), action)
    if err == nil {
        t.Error("Expected error for action without run command")
    }
    
    if result.Status != prepush.StatusError {
        t.Errorf("Expected StatusError, got %v", result.Status)
    }
}