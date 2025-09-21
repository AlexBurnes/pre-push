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
    "strings"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/AlexBurnes/pre-push/internal/config"
    "github.com/AlexBurnes/pre-push/internal/exec"
    "github.com/AlexBurnes/pre-push/internal/install"
    "github.com/AlexBurnes/pre-push/internal/ui"
    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

const (
    appName = "pre-push"
)

// getVersion reads the version from the VERSION file
func getVersion() string {
    data, err := os.ReadFile("VERSION")
    if err != nil {
        return "unknown"
    }
    
    version := strings.TrimSpace(string(data))
    if version == "" {
        return "unknown"
    }
    
    return version
}

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
    
    // Add version flags
    rootCmd.Flags().BoolP("version", "", false, "print version and module name")
    rootCmd.Flags().BoolP("version-only", "V", false, "print version only")
    
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
    // Git calls hooks with arguments (like "origin master --tags") and passes ref info via stdin
    // We detect this by checking if stdin is not a terminal (pipe or file)
    // and we're not being called with known pre-push subcommands
    if len(os.Args) > 1 {
        // Check if the first argument is a known pre-push subcommand
        firstArg := os.Args[1]
        if firstArg == "test" || firstArg == "list-uses" || 
           firstArg == "-h" || firstArg == "--help" ||
           firstArg == "-v" || firstArg == "--version" ||
           firstArg == "-d" || firstArg == "--debug" {
            return false
        }
    }
    
    // When Git calls the hook, it passes ref info via stdin
    // Check if stdin is not a terminal (pipe or file)
    stat, err := os.Stdin.Stat()
    if err != nil {
        return false
    }
    
    // If stdin is not a character device, it's likely a pipe or file
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
    
    // Create buildfab executor
    executor := exec.NewBuildfabExecutor(cfg, ui)
    
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
    // Check if version flags were set
    if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
        fmt.Printf("%s version %s\n", appName, getVersion())
        return nil
    }
    if versionOnlyFlag, _ := cmd.Flags().GetBool("version-only"); versionOnlyFlag {
        fmt.Printf("%s\n", getVersion())
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
    // Check if version flags were set
    if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
        fmt.Printf("%s version %s\n", appName, getVersion())
        return nil
    }
    if versionOnlyFlag, _ := cmd.Flags().GetBool("version-only"); versionOnlyFlag {
        fmt.Printf("%s\n", getVersion())
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
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Detect Git variables
    gitVars, err := config.DetectGitVariables(ctx)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Resolve variables in configuration
    if err := config.ResolveVariables(cfg, gitVars); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Create UI
    ui := ui.New(verbose, debug)
    
    // Create buildfab executor
    executor := exec.NewBuildfabExecutor(cfg, ui)
    
    // Run pre-push stage
    if err := executor.RunStage(ctx, "pre-push"); err != nil {
        os.Exit(1)
    }
    
    return nil
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