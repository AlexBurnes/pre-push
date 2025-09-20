package uses

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"

    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

func TestGitUntrackedRunner(t *testing.T) {
    runner := &GitUntrackedRunner{}
    
    // Test basic interface methods
    if runner.GetName() != "git@untracked" {
        t.Errorf("Expected name 'git@untracked', got '%s'", runner.GetName())
    }
    
    if runner.GetHelp() == "" {
        t.Error("Expected non-empty help text")
    }
    
    if runner.GetRepro() == "" {
        t.Error("Expected non-empty repro command")
    }
    
    // Test in a temporary directory
    tempDir, err := os.MkdirTemp("", "pre-push-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)
    
    // Change to temp directory
    oldDir, err := os.Getwd()
    if err != nil {
        t.Fatalf("Failed to get current dir: %v", err)
    }
    defer os.Chdir(oldDir)
    
    if err := os.Chdir(tempDir); err != nil {
        t.Fatalf("Failed to change to temp dir: %v", err)
    }
    
    // Initialize git repository
    if err := exec.Command("git", "init").Run(); err != nil {
        t.Fatalf("Failed to init git repo: %v", err)
    }
    
    // Test with no untracked files
    result, err := runner.Run(context.Background())
    if err != nil {
        t.Errorf("Expected no error with clean repo, got: %v", err)
    }
    if result.Status != prepush.StatusOK {
        t.Errorf("Expected StatusOK, got %v", result.Status)
    }
    
    // Create an untracked file
    untrackedFile := filepath.Join(tempDir, "untracked.txt")
    if err := os.WriteFile(untrackedFile, []byte("test"), 0644); err != nil {
        t.Fatalf("Failed to create untracked file: %v", err)
    }
    
    // Test with untracked files
    result, err = runner.Run(context.Background())
    if err == nil {
        t.Error("Expected error with untracked files")
    }
    if result.Status != prepush.StatusError {
        t.Errorf("Expected StatusError, got %v", result.Status)
    }
}

func TestGitUncommittedRunner(t *testing.T) {
    runner := &GitUncommittedRunner{}
    
    // Test basic interface methods
    if runner.GetName() != "git@uncommitted" {
        t.Errorf("Expected name 'git@uncommitted', got '%s'", runner.GetName())
    }
    
    if runner.GetHelp() == "" {
        t.Error("Expected non-empty help text")
    }
    
    if runner.GetRepro() == "" {
        t.Error("Expected non-empty repro command")
    }
    
    // Test in a temporary directory
    tempDir, err := os.MkdirTemp("", "pre-push-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)
    
    // Change to temp directory
    oldDir, err := os.Getwd()
    if err != nil {
        t.Fatalf("Failed to get current dir: %v", err)
    }
    defer os.Chdir(oldDir)
    
    if err := os.Chdir(tempDir); err != nil {
        t.Fatalf("Failed to change to temp dir: %v", err)
    }
    
    // Initialize git repository
    if err := exec.Command("git", "init").Run(); err != nil {
        t.Fatalf("Failed to init git repo: %v", err)
    }
    
    // Configure git user
    exec.Command("git", "config", "user.email", "test@example.com").Run()
    exec.Command("git", "config", "user.name", "Test User").Run()
    
    // Test with no uncommitted changes
    result, err := runner.Run(context.Background())
    if err != nil {
        t.Errorf("Expected no error with clean repo, got: %v", err)
    }
    if result.Status != prepush.StatusOK {
        t.Errorf("Expected StatusOK, got %v", result.Status)
    }
    
    // Create a file and add it
    testFile := filepath.Join(tempDir, "test.txt")
    if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    
    if err := exec.Command("git", "add", "test.txt").Run(); err != nil {
        t.Fatalf("Failed to add file: %v", err)
    }
    
    // Test with staged changes
    result, err = runner.Run(context.Background())
    if err == nil {
        t.Error("Expected error with staged changes")
    }
    if result.Status != prepush.StatusError {
        t.Errorf("Expected StatusError, got %v", result.Status)
    }
}

func TestGitModifiedRunner(t *testing.T) {
    runner := &GitModifiedRunner{}
    
    // Test basic interface methods
    if runner.GetName() != "git@modified" {
        t.Errorf("Expected name 'git@modified', got '%s'", runner.GetName())
    }
    
    if runner.GetHelp() == "" {
        t.Error("Expected non-empty help text")
    }
    
    if runner.GetRepro() == "" {
        t.Error("Expected non-empty repro command")
    }
    
    // Test in a temporary directory
    tempDir, err := os.MkdirTemp("", "pre-push-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)
    
    // Change to temp directory
    oldDir, err := os.Getwd()
    if err != nil {
        t.Fatalf("Failed to get current dir: %v", err)
    }
    defer os.Chdir(oldDir)
    
    if err := os.Chdir(tempDir); err != nil {
        t.Fatalf("Failed to change to temp dir: %v", err)
    }
    
    // Initialize git repository
    if err := exec.Command("git", "init").Run(); err != nil {
        t.Fatalf("Failed to init git repo: %v", err)
    }
    
    // Configure git user
    exec.Command("git", "config", "user.email", "test@example.com").Run()
    exec.Command("git", "config", "user.name", "Test User").Run()
    
    // Create initial commit
    testFile := filepath.Join(tempDir, "test.txt")
    if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    
    if err := exec.Command("git", "add", "test.txt").Run(); err != nil {
        t.Fatalf("Failed to add file: %v", err)
    }
    
    if err := exec.Command("git", "commit", "-m", "Initial commit").Run(); err != nil {
        t.Fatalf("Failed to commit: %v", err)
    }
    
    // Test with clean working tree
    result, err := runner.Run(context.Background())
    if err != nil {
        t.Errorf("Expected no error with clean working tree, got: %v", err)
    }
    if result.Status != prepush.StatusOK {
        t.Errorf("Expected StatusOK, got %v", result.Status)
    }
    
    // Modify the file
    if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
        t.Fatalf("Failed to modify file: %v", err)
    }
    
    // Test with modified working tree
    result, err = runner.Run(context.Background())
    if err == nil {
        t.Error("Expected error with modified working tree")
    }
    if result.Status != prepush.StatusError {
        t.Errorf("Expected StatusError, got %v", result.Status)
    }
}

func TestRegistry(t *testing.T) {
    registry := New()
    
    // Test that all expected runners are registered
    expectedRunners := []string{"git@untracked", "git@uncommitted", "git@modified"}
    
    for _, name := range expectedRunners {
        runner, exists := registry.GetRunner(name)
        if !exists {
            t.Errorf("Expected runner %s to be registered", name)
        }
        if runner == nil {
            t.Errorf("Expected runner %s to not be nil", name)
        }
    }
    
    // Test unknown runner
    _, exists := registry.GetRunner("unknown@action")
    if exists {
        t.Error("Expected unknown runner to not exist")
    }
    
    // Test list runners
    runners := registry.ListRunners()
    if len(runners) != len(expectedRunners) {
        t.Errorf("Expected %d runners, got %d", len(expectedRunners), len(runners))
    }
}