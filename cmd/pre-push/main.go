// Package main implements the pre-push CLI tool.
// It provides a cross-platform, configurable Git pre-push hook runner
// with DAG executor, built-in checks, and Buildfab integration.
package main

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/AlexBurnes/pre-push/internal/config"
    "github.com/AlexBurnes/pre-push/internal/exec"
    "github.com/AlexBurnes/pre-push/internal/install"
    "github.com/AlexBurnes/pre-push/internal/ui"
    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

const (
    version = "0.1.0"
    appName = "pre-push"
)

// Global flags
var (
    verbose bool
    debug   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "pre-push",
    Short: "Git pre-push hook runner with DAG executor",
    Long: `pre-push is a cross-platform, configurable Git pre-push hook runner that
provides built-in checks and supports custom actions via YAML configuration.

When invoked without arguments, it installs or updates itself as a Git
pre-push hook. When invoked by Git as a hook, it reads the standard
pre-push input and runs configured checks.

Configuration is provided via .project.yml file in the repository root.`,
    RunE: runRoot,
}

// testCmd represents the test command
var testCmd = &cobra.Command{
    Use:   "test",
    Short: "Run all checks in dry-run mode",
    Long: `Run all configured checks in dry-run mode without installing or updating
the pre-push hook. This is useful for testing your configuration before
pushing changes.`,
    RunE: runTest,
}

// listUsesCmd represents the list-uses command
var listUsesCmd = &cobra.Command{
    Use:   "list-uses",
    Short: "List available built-in actions",
    Long: `List all available built-in actions that can be used in the 'uses' field
of your configuration. Each action includes a brief description of what it does.`,
    RunE: runListUses,
}

func main() {
    // Check if we're being called by Git as a hook
    if isGitHook() {
        // When called by Git, run the hook directly
        if err := runGitHook(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        return
    }
    
    // Add global flags
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
    rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug output")
    
    // Add version flag
    rootCmd.Flags().BoolP("version", "V", false, "print version and exit")
    
    // Add subcommands
    rootCmd.AddCommand(testCmd)
    rootCmd.AddCommand(listUsesCmd)
    
    // Execute the root command
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// isGitHook determines if we're being called by Git as a hook
func isGitHook() bool {
    // Git calls hooks with no arguments and passes ref info via stdin
    // We can detect this by checking if we have no args and stdin has data
    if len(os.Args) > 1 {
        return false
    }
    
    // Check if stdin has data (Git passes ref info via stdin)
    stat, err := os.Stdin.Stat()
    if err != nil {
        return false
    }
    
    // If stdin has data, we're likely being called by Git
    return (stat.Mode() & os.ModeCharDevice) == 0
}

// runGitHook runs the pre-push hook when called by Git
func runGitHook() error {
    // Create context with cancellation
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()
    
    // Read ref information from stdin
    refs, err := readGitRefs()
    if err != nil {
        return fmt.Errorf("failed to read Git refs: %w", err)
    }
    
    // If no refs to push, exit successfully
    if len(refs) == 0 {
        return nil
    }
    
    // Load configuration
    cfg, err := config.Load(".project.yml")
    if err != nil {
        return fmt.Errorf("failed to load configuration: %w", err)
    }
    
    // Detect Git variables
    gitVars, err := config.DetectGitVariables(ctx)
    if err != nil {
        return fmt.Errorf("failed to detect Git variables: %w", err)
    }
    
    // Resolve variables in configuration
    if err := config.ResolveVariables(cfg, gitVars); err != nil {
        return fmt.Errorf("failed to resolve variables: %w", err)
    }
    
    // Create UI
    ui := ui.New(verbose, debug)
    
    // Create executor
    executor := exec.New(cfg, ui)
    
    // Run pre-push stage
    return executor.RunStage(ctx, "pre-push")
}

// readGitRefs reads Git ref information from stdin
func readGitRefs() ([]string, error) {
    var refs []string
    scanner := bufio.NewScanner(os.Stdin)
    
    for scanner.Scan() {
        line := scanner.Text()
        if line != "" {
            refs = append(refs, line)
        }
    }
    
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    
    return refs, nil
}

// runRoot handles the root command (install or update hook)
func runRoot(cmd *cobra.Command, args []string) error {
    // Check if version flag was set
    if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
        fmt.Printf("%s version %s\n", appName, version)
        return nil
    }
    
    // Create context with cancellation
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()
    
    installer := install.New()
    return installer.Install(ctx)
}

// runInstall installs or updates the pre-push hook
func runInstall(cmd *cobra.Command, args []string) error {
    // Check if version flag was set
    if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
        fmt.Printf("%s version %s\n", appName, version)
        return nil
    }
    
    // Create context with cancellation
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()
    
    installer := install.New()
    return installer.Install(ctx)
}

// runTest runs all checks in dry-run mode
func runTest(cmd *cobra.Command, args []string) error {
    // Create context with cancellation
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()
    
    // Load configuration
    cfg, err := config.Load(".project.yml")
    if err != nil {
        return fmt.Errorf("failed to load configuration: %w", err)
    }
    
    // Detect Git variables
    gitVars, err := config.DetectGitVariables(ctx)
    if err != nil {
        return fmt.Errorf("failed to detect Git variables: %w", err)
    }
    
    // Resolve variables in configuration
    if err := config.ResolveVariables(cfg, gitVars); err != nil {
        return fmt.Errorf("failed to resolve variables: %w", err)
    }
    
    // Create UI
    ui := ui.New(verbose, debug)
    
    // Create executor
    executor := exec.New(cfg, ui)
    
    // Run pre-push stage
    return executor.RunStage(ctx, "pre-push")
}

// runListUses lists all available built-in actions
func runListUses(cmd *cobra.Command, args []string) error {
    uses := prepush.ListBuiltInActions()
    
    fmt.Println("Available built-in actions:")
    fmt.Println()
    
    for name, description := range uses {
        fmt.Printf("  %-20s %s\n", name, description)
    }
    
    return nil
}