// Package main implements the pre-push CLI tool.
// It provides a cross-platform, configurable Git pre-push hook runner
// with DAG executor, built-in checks, and Buildfab integration.
package main

import (
    "bufio"
    "context"
    "crypto/md5"
    "fmt"
    "io"
    "os"
    "os/signal"
    "path/filepath"
    "strconv"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/AlexBurnes/buildfab/pkg/buildfab"
    "github.com/AlexBurnes/pre-push/internal/exec"
    "github.com/AlexBurnes/pre-push/internal/install"
    "github.com/AlexBurnes/pre-push/internal/ui"
    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

const (
    appName = "pre-push"
)

// appVersion is set at build time via ldflags
var appVersion = "unknown"

// getVersion returns the compiled-in version
func getVersion() string {
    return appVersion
}

// isVerboseEnabled checks if verbose mode should be enabled for Git hooks (for backward compatibility)
func isVerboseEnabled() bool {
    return getVerboseLevel() > 0
}

// getVerboseLevel gets the verbose level from environment variable or CLI flag
func getVerboseLevel() int {
    // Check environment variable first
    if envVerbose := os.Getenv("PRE_PUSH_VERBOSE"); envVerbose != "" {
        // Parse as integer, default to 0 if invalid
        if level, err := strconv.Atoi(envVerbose); err == nil && level >= 0 {
            return level
        }
        // If invalid, default to 0 (quiet mode)
        return 0
    }
    
    // Check if CLI verbose flag is set (for manual testing)
    if verbose {
        return 1
    }
    
    return 0
}

// isDebugEnabled checks if debug mode should be enabled for Git hooks
func isDebugEnabled() bool {
    // Check environment variable first
    if os.Getenv("PRE_PUSH_DEBUG") == "1" {
        return true
    }
    
    // Check if CLI debug flag is set (for manual testing)
    return debug
}


// getCurrentBinaryPath returns the path to the current running binary
func getCurrentBinaryPath() (string, error) {
    return os.Executable()
}

// getGitHookPath returns the path to the Git pre-push hook
func getGitHookPath() (string, error) {
    // Find .git directory
    cwd, err := os.Getwd()
    if err != nil {
        return "", fmt.Errorf("failed to get current directory: %w", err)
    }
    
    // Walk up the directory tree to find .git
    for {
        gitDir := filepath.Join(cwd, ".git")
        if _, err := os.Stat(gitDir); err == nil {
            return filepath.Join(gitDir, "hooks", "pre-push"), nil
        }
        
        parent := filepath.Dir(cwd)
        if parent == cwd {
            break // Reached root directory
        }
        cwd = parent
    }
    
    return "", fmt.Errorf("not in a git repository")
}

// calculateFileMD5 calculates the MD5 hash of a file
func calculateFileMD5(filePath string) (string, error) {
    file, err := os.Open(filePath)
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

// isBinaryDifferent checks if the Git hook binary is different from the current binary
func isBinaryDifferent() (bool, error) {
    currentBinary, err := getCurrentBinaryPath()
    if err != nil {
        return false, fmt.Errorf("failed to get current binary path: %w", err)
    }
    
    hookPath, err := getGitHookPath()
    if err != nil {
        return false, fmt.Errorf("failed to get git hook path: %w", err)
    }
    
    // Check if hook file exists
    if _, err := os.Stat(hookPath); os.IsNotExist(err) {
        return true, nil // Hook doesn't exist, so it's different
    }
    
    // Calculate MD5 hashes
    currentHash, err := calculateFileMD5(currentBinary)
    if err != nil {
        return false, fmt.Errorf("failed to calculate current binary hash: %w", err)
    }
    
    hookHash, err := calculateFileMD5(hookPath)
    if err != nil {
        return false, fmt.Errorf("failed to calculate hook binary hash: %w", err)
    }
    
    return currentHash != hookHash, nil
}

// updateGitHook copies the current binary to the Git hook location
func updateGitHook() error {
    currentBinary, err := getCurrentBinaryPath()
    if err != nil {
        return fmt.Errorf("failed to get current binary path: %w", err)
    }
    
    hookPath, err := getGitHookPath()
    if err != nil {
        return fmt.Errorf("failed to get git hook path: %w", err)
    }
    
    // Ensure hooks directory exists
    hooksDir := filepath.Dir(hookPath)
    if err := os.MkdirAll(hooksDir, 0755); err != nil {
        return fmt.Errorf("failed to create hooks directory: %w", err)
    }
    
    // Copy current binary to hook location
    sourceFile, err := os.Open(currentBinary)
    if err != nil {
        return fmt.Errorf("failed to open current binary: %w", err)
    }
    defer sourceFile.Close()
    
    destFile, err := os.OpenFile(hookPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
    if err != nil {
        return fmt.Errorf("failed to create hook file: %w", err)
    }
    defer destFile.Close()
    
    if _, err := io.Copy(destFile, sourceFile); err != nil {
        return fmt.Errorf("failed to copy binary to hook: %w", err)
    }
    
    return nil
}

// checkAndUpdateGitHook checks if the Git hook needs updating and updates it if necessary
func checkAndUpdateGitHook() error {
    different, err := isBinaryDifferent()
    if err != nil {
        return fmt.Errorf("failed to check if binary is different: %w", err)
    }
    
    if different {
        if err := updateGitHook(); err != nil {
            return fmt.Errorf("failed to update git hook: %w", err)
        }
        fmt.Printf("Updated Git pre-push hook with current binary\n")
    }
    
    return nil
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

When invoked without arguments, it checks and installs or updates itself as a Git
pre-push hook with MD5 verification. When invoked by Git as a hook, it reads the 
standard pre-push input and runs configured checks.

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
        // When called by Git, run the hook directly (no update checks)
        if err := runGitHook(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        return
    }
    
    // When running as CLI, check and update Git hook if needed
    if err := checkAndUpdateGitHook(); err != nil {
        // Don't fail the CLI command if hook update fails, just warn
        fmt.Fprintf(os.Stderr, "Warning: failed to update Git hook: %v\n", err)
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
    
    // Load configuration using buildfab (supports includes)
    buildfabConfig, err := buildfab.LoadConfig(".project.yml")
    if err != nil {
        return fmt.Errorf("failed to load configuration: %w", err)
    }
    
    // Variables will be resolved by buildfab automatically
    
    // Determine verbose and debug modes for Git hooks
    hookVerboseLevel := getVerboseLevel()
    hookDebug := isDebugEnabled()
    
    // Debug output (only when debug mode is enabled)
    if hookDebug {
        fmt.Fprintf(os.Stderr, "DEBUG: hookVerboseLevel=%d, hookDebug=%v\n", hookVerboseLevel, hookDebug)
        fmt.Fprintf(os.Stderr, "DEBUG: env PRE_PUSH_VERBOSE=%s\n", os.Getenv("PRE_PUSH_VERBOSE"))
        fmt.Fprintf(os.Stderr, "DEBUG: getVerboseLevel()=%d\n", getVerboseLevel())
    }
    
    // Create UI with Git hook specific settings
    ui := ui.NewWithVerboseLevel(hookVerboseLevel, hookDebug)
    
    // Create buildfab executor with CLI version
    executor := exec.BuildfabExecutorWithCLIVersion(buildfabConfig, ui, getVersion())
    
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

// runRoot handles the root command (check and install hook)
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


// runTest runs all checks in dry-run mode
func runTest(cmd *cobra.Command, args []string) error {
    // Create context with cancellation
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    // Load configuration using buildfab (supports includes)
    buildfabConfig, err := buildfab.LoadConfig(".project.yml")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Determine verbose and debug modes for Git hooks
    hookVerboseLevel := getVerboseLevel()
    hookDebug := isDebugEnabled()
    
    // Debug output (remove in production)
    if hookDebug {
        fmt.Fprintf(os.Stderr, "DEBUG: hookVerboseLevel=%d, hookDebug=%v\n", hookVerboseLevel, hookDebug)
        fmt.Fprintf(os.Stderr, "DEBUG: env PRE_PUSH_VERBOSE=%s\n", os.Getenv("PRE_PUSH_VERBOSE"))
        fmt.Fprintf(os.Stderr, "DEBUG: getVerboseLevel()=%d\n", getVerboseLevel())
    }
    
    // Variables will be resolved by buildfab automatically
    
    // Create UI with detected verbose and debug modes
    ui := ui.NewWithVerboseLevel(hookVerboseLevel, hookDebug)
    
    // Create buildfab executor with CLI version
    executor := exec.BuildfabExecutorWithCLIVersion(buildfabConfig, ui, getVersion())
    
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