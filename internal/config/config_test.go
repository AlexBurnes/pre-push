package config

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"

    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

func TestLoad(t *testing.T) {
    // Create a temporary configuration file
    tempDir, err := os.MkdirTemp("", "pre-push-config-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)
    
    configPath := filepath.Join(tempDir, "test.yml")
    configContent := `
project:
  name: "test-project"
  modules: ["test"]

actions:
  - name: test-action
    run: "echo test"

stages:
  test:
    steps:
      - action: test-action
`
    
    if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
        t.Fatalf("Failed to write config file: %v", err)
    }
    
    // Test loading configuration
    config, err := Load(configPath)
    if err != nil {
        t.Fatalf("Failed to load configuration: %v", err)
    }
    
    if config.Project.Name != "test-project" {
        t.Errorf("Expected project name 'test-project', got '%s'", config.Project.Name)
    }
    
    if len(config.Actions) != 1 {
        t.Errorf("Expected 1 action, got %d", len(config.Actions))
    }
    
    if config.Actions[0].Name != "test-action" {
        t.Errorf("Expected action name 'test-action', got '%s'", config.Actions[0].Name)
    }
    
    if len(config.Stages) != 1 {
        t.Errorf("Expected 1 stage, got %d", len(config.Stages))
    }
}

func TestLoadFromDir(t *testing.T) {
    // Create a temporary directory
    tempDir, err := os.MkdirTemp("", "pre-push-config-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)
    
    configContent := `
project:
  name: "test-project"
  modules: ["test"]

actions:
  - name: test-action
    run: "echo test"

stages:
  test:
    steps:
      - action: test-action
`
    
    // Test with .project.yml
    configPath := filepath.Join(tempDir, ".project.yml")
    if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
        t.Fatalf("Failed to write config file: %v", err)
    }
    
    config, err := LoadFromDir(tempDir)
    if err != nil {
        t.Fatalf("Failed to load configuration from dir: %v", err)
    }
    
    if config.Project.Name != "test-project" {
        t.Errorf("Expected project name 'test-project', got '%s'", config.Project.Name)
    }
}

func TestResolveVariables(t *testing.T) {
    config := &prepush.Config{
        Actions: []prepush.Action{
            {
                Name: "test-action",
                Run:  "echo ${{ tag }} and ${{ branch }}",
            },
        },
    }
    
    variables := map[string]string{
        "tag":    "v1.0.0",
        "branch": "main",
    }
    
    err := ResolveVariables(config, variables)
    if err != nil {
        t.Fatalf("Failed to resolve variables: %v", err)
    }
    
    expected := "echo v1.0.0 and main"
    if config.Actions[0].Run != expected {
        t.Errorf("Expected resolved command '%s', got '%s'", expected, config.Actions[0].Run)
    }
}

func TestResolveVariablesError(t *testing.T) {
    config := &prepush.Config{
        Actions: []prepush.Action{
            {
                Name: "test-action",
                Run:  "echo ${{ undefined }}",
            },
        },
    }
    
    variables := map[string]string{
        "tag": "v1.0.0",
    }
    
    err := ResolveVariables(config, variables)
    if err == nil {
        t.Error("Expected error for undefined variable")
    }
}

func TestDetectGitVariables(t *testing.T) {
    // Create a temporary git repository
    tempDir, err := os.MkdirTemp("", "pre-push-git-test")
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
    if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    
    if err := exec.Command("git", "add", "test.txt").Run(); err != nil {
        t.Fatalf("Failed to add file: %v", err)
    }
    
    if err := exec.Command("git", "commit", "-m", "Initial commit").Run(); err != nil {
        t.Fatalf("Failed to commit: %v", err)
    }
    
    // Create a tag
    if err := exec.Command("git", "tag", "v1.0.0").Run(); err != nil {
        t.Fatalf("Failed to create tag: %v", err)
    }
    
    // Test detecting variables
    variables, err := DetectGitVariables(context.Background())
    if err != nil {
        t.Fatalf("Failed to detect git variables: %v", err)
    }
    
    if variables["tag"] != "v1.0.0" {
        t.Errorf("Expected tag 'v1.0.0', got '%s'", variables["tag"])
    }
    
    if variables["branch"] != "main" && variables["branch"] != "master" {
        t.Errorf("Expected branch 'main' or 'master', got '%s'", variables["branch"])
    }
}

func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  prepush.Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: prepush.Config{
                Project: struct {
                    Name    string   `yaml:"name"`
                    Modules []string `yaml:"modules"`
                    BinDir  string   `yaml:"bin,omitempty"`
                }{
                    Name:    "test",
                    Modules: []string{"test"},
                },
                Actions: []prepush.Action{
                    {
                        Name: "test-action",
                        Run:  "echo test",
                    },
                },
                Stages: map[string]prepush.Stage{
                    "test": {
                        Steps: []prepush.Step{
                            {
                                Action: "test-action",
                            },
                        },
                    },
                },
            },
            wantErr: false,
        },
        {
            name: "missing project name",
            config: prepush.Config{
                Actions: []prepush.Action{
                    {
                        Name: "test-action",
                        Run:  "echo test",
                    },
                },
            },
            wantErr: true,
        },
        {
            name: "no actions",
            config: prepush.Config{
                Project: struct {
                    Name    string   `yaml:"name"`
                    Modules []string `yaml:"modules"`
                    BinDir  string   `yaml:"bin,omitempty"`
                }{
                    Name: "test",
                },
            },
            wantErr: true,
        },
        {
            name: "action without run or uses",
            config: prepush.Config{
                Project: struct {
                    Name    string   `yaml:"name"`
                    Modules []string `yaml:"modules"`
                    BinDir  string   `yaml:"bin,omitempty"`
                }{
                    Name: "test",
                },
                Actions: []prepush.Action{
                    {
                        Name: "test-action",
                    },
                },
            },
            wantErr: true,
        },
        {
            name: "action with both run and uses",
            config: prepush.Config{
                Project: struct {
                    Name    string   `yaml:"name"`
                    Modules []string `yaml:"modules"`
                    BinDir  string   `yaml:"bin,omitempty"`
                }{
                    Name: "test",
                },
                Actions: []prepush.Action{
                    {
                        Name: "test-action",
                        Run:  "echo test",
                        Uses: "git@untracked",
                    },
                },
            },
            wantErr: true,
        },
        {
            name: "duplicate action name",
            config: prepush.Config{
                Project: struct {
                    Name    string   `yaml:"name"`
                    Modules []string `yaml:"modules"`
                    BinDir  string   `yaml:"bin,omitempty"`
                }{
                    Name: "test",
                },
                Actions: []prepush.Action{
                    {
                        Name: "test-action",
                        Run:  "echo test",
                    },
                    {
                        Name: "test-action",
                        Run:  "echo test2",
                    },
                },
            },
            wantErr: true,
        },
        {
            name: "step with unknown action",
            config: prepush.Config{
                Project: struct {
                    Name    string   `yaml:"name"`
                    Modules []string `yaml:"modules"`
                    BinDir  string   `yaml:"bin,omitempty"`
                }{
                    Name: "test",
                },
                Actions: []prepush.Action{
                    {
                        Name: "test-action",
                        Run:  "echo test",
                    },
                },
                Stages: map[string]prepush.Stage{
                    "test": {
                        Steps: []prepush.Step{
                            {
                                Action: "unknown-action",
                            },
                        },
                    },
                },
            },
            wantErr: true,
        },
        {
            name: "invalid onerror value",
            config: prepush.Config{
                Project: struct {
                    Name    string   `yaml:"name"`
                    Modules []string `yaml:"modules"`
                    BinDir  string   `yaml:"bin,omitempty"`
                }{
                    Name: "test",
                },
                Actions: []prepush.Action{
                    {
                        Name: "test-action",
                        Run:  "echo test",
                    },
                },
                Stages: map[string]prepush.Stage{
                    "test": {
                        Steps: []prepush.Step{
                            {
                                Action:  "test-action",
                                OnError: "invalid",
                            },
                        },
                    },
                },
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}