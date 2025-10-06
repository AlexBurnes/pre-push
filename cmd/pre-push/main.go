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
    "os/exec"
    "os/signal"
    "path/filepath"
    "strconv"
    "strings"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/AlexBurnes/buildfab/pkg/buildfab"
    preexec "github.com/AlexBurnes/pre-push/internal/exec"
    "github.com/AlexBurnes/pre-push/internal/ui"
    "github.com/AlexBurnes/pre-push/internal/version"
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
func checkAndUpdateGitHook() (bool, error) {
    different, err := isBinaryDifferent()
    if err != nil {
        return false, fmt.Errorf("failed to check if binary is different: %w", err)
    }
    
    if different {
        if err := updateGitHook(); err != nil {
            return false, fmt.Errorf("failed to update git hook: %w", err)
        }
        fmt.Printf("Updated Git pre-push hook with current binary\n")
        return true, nil
    } else {
        fmt.Printf("Git pre-push hook is already up to date (version %s)\n", getVersion())
        return false, nil
    }
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

// installCmd represents the install command
var installCmd = &cobra.Command{
    Use:   "install",
    Short: "Install or update Git pre-push hook",
    Long: `Install or update the Git pre-push hook with the current binary.
This command checks if the Git hook needs updating and updates it if necessary.
The hook is installed in .git/hooks/pre-push and will be called automatically
by Git before pushing changes.`,
    RunE: runInstall,
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
    
    // Add global flags
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
    rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug output")
    
    // Add version flags
    rootCmd.Flags().BoolP("version", "", false, "print version and module name")
    rootCmd.Flags().BoolP("version-only", "V", false, "print version only")
    
    // Add subcommands
    rootCmd.AddCommand(testCmd)
    rootCmd.AddCommand(listUsesCmd)
    rootCmd.AddCommand(installCmd)
    
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
        if firstArg == "test" || firstArg == "list-uses" || firstArg == "install" ||
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
    
    // Read and parse ref information from stdin
    refs, err := readGitRefs()
    if err != nil {
        return fmt.Errorf("failed to read Git refs: %w", err)
    }
    
    // If no refs to push, exit successfully
    if len(refs) == 0 {
        return nil
    }
    
    // Parse Git push information
    pushInfo, err := parseGitPushInfo(refs)
    if err != nil {
        return fmt.Errorf("failed to parse Git push info: %w", err)
    }
    
    // 1. Check if this is a delete operation - if so, skip all checks
    if pushInfo.IsDelete {
        fmt.Fprintf(os.Stderr, "Delete operation detected, skipping pre-push checks\n")
        return nil
    }
    
    // 2. Validate pushed tags for semantic versioning
    if len(pushInfo.Tags) > 0 {
        for _, tag := range pushInfo.Tags {
            if err := validateTagSemantics(tag); err != nil {
                return fmt.Errorf("invalid tag semantics for %s: %w", tag, err)
            }
        }
    }
    
    // 3. Check if pushing tag/branch that is not current - if so, skip pre-push stage
    if shouldSkipPrePushStage(pushInfo) {
        fmt.Fprintf(os.Stderr, "Pushing tag/branch that is not current, skipping pre-push stage\n")
        return nil
    }
    
    // Load configuration using buildfab (supports includes)
    buildfabConfig, err := buildfab.LoadConfig(".project.yml")
    if err != nil {
        return fmt.Errorf("failed to load configuration: %w", err)
    }
    
    // Determine verbose and debug modes for Git hooks
    hookVerboseLevel := getVerboseLevel()
    hookDebug := isDebugEnabled()
    
    // Debug output (only when debug mode is enabled)
    if hookDebug {
        fmt.Fprintf(os.Stderr, "DEBUG: hookVerboseLevel=%d, hookDebug=%v\n", hookVerboseLevel, hookDebug)
        fmt.Fprintf(os.Stderr, "DEBUG: env PRE_PUSH_VERBOSE=%s\n", os.Getenv("PRE_PUSH_VERBOSE"))
        fmt.Fprintf(os.Stderr, "DEBUG: getVerboseLevel()=%d\n", getVerboseLevel())
        fmt.Fprintf(os.Stderr, "DEBUG: pushInfo=%+v\n", pushInfo)
    }
    
    // Create UI with Git hook specific settings
    ui := ui.NewWithVerboseLevel(hookVerboseLevel, hookDebug)
    
    // Create buildfab executor with CLI version and enhanced Git variables
    executor := preexec.BuildfabExecutorWithCLIVersion(buildfabConfig, ui, getVersion())
    
    // Enhance executor with Git push information for variable interpolation
    executor.SetGitPushInfo(&preexec.GitPushInfo{
        RemoteName: pushInfo.RemoteName,
        RemoteURL:  pushInfo.RemoteURL,
        Refs:       convertGitRefs(pushInfo.Refs),
        Tags:       pushInfo.Tags,
        Branches:   pushInfo.Branches,
        IsDelete:   pushInfo.IsDelete,
    })
    
    // Run pre-push stage
    return executor.RunStage(ctx, "pre-push")
}

// GitRef represents a Git reference being pushed
type GitRef struct {
    LocalRef  string
    LocalSHA  string
    RemoteRef string
    RemoteSHA string
    IsDelete  bool
    IsTag     bool
    IsBranch  bool
}

// GitPushInfo contains information about the Git push operation
type GitPushInfo struct {
    RemoteName string
    RemoteURL  string
    Refs       []GitRef
    Tags       []string
    Branches   []string
    IsDelete   bool
}

// readGitRefs reads Git ref information from stdin and parses it
func readGitRefs() ([]GitRef, error) {
    var refs []GitRef
    scanner := bufio.NewScanner(os.Stdin)
    
    for scanner.Scan() {
        line := scanner.Text()
        if line != "" {
            ref, err := parseGitRef(line)
            if err != nil {
                return nil, fmt.Errorf("failed to parse Git ref line %q: %w", line, err)
            }
            refs = append(refs, ref)
        }
    }
    
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    
    return refs, nil
}

// parseGitRef parses a single Git ref line from stdin
// Format: <local_ref> <local_sha> <remote_ref> <remote_sha>
// For delete operations: <local_ref> <zero_sha> <remote_ref> <remote_sha>
func parseGitRef(line string) (GitRef, error) {
    parts := strings.Fields(line)
    if len(parts) != 4 {
        return GitRef{}, fmt.Errorf("invalid Git ref format, expected 4 fields, got %d", len(parts))
    }
    
    localRef := parts[0]
    localSHA := parts[1]
    remoteRef := parts[2]
    remoteSHA := parts[3]
    
    // Detect delete operation (local SHA is all zeros)
    isDelete := localSHA == "0000000000000000000000000000000000000000"
    
    // Detect if it's a tag or branch
    isTag := strings.HasPrefix(localRef, "refs/tags/")
    isBranch := strings.HasPrefix(localRef, "refs/heads/")
    
    return GitRef{
        LocalRef:  localRef,
        LocalSHA:  localSHA,
        RemoteRef: remoteRef,
        RemoteSHA: remoteSHA,
        IsDelete:  isDelete,
        IsTag:     isTag,
        IsBranch:  isBranch,
    }, nil
}

// parseGitPushInfo parses Git push information from refs and command line arguments
func parseGitPushInfo(refs []GitRef) (*GitPushInfo, error) {
    if len(os.Args) < 2 {
        return nil, fmt.Errorf("insufficient arguments for Git hook")
    }
    
    remoteName := os.Args[1]
    remoteURL := ""
    if len(os.Args) > 2 {
        remoteURL = os.Args[2]
    }
    
    var tags []string
    var branches []string
    isDelete := false
    
    for _, ref := range refs {
        if ref.IsDelete {
            isDelete = true
            // Extract tag/branch name from local ref
            if ref.IsTag {
                tagName := strings.TrimPrefix(ref.LocalRef, "refs/tags/")
                tags = append(tags, tagName)
            } else if ref.IsBranch {
                branchName := strings.TrimPrefix(ref.LocalRef, "refs/heads/")
                branches = append(branches, branchName)
            }
        } else {
            // Extract tag/branch name from local ref
            if ref.IsTag {
                tagName := strings.TrimPrefix(ref.LocalRef, "refs/tags/")
                tags = append(tags, tagName)
            } else if ref.IsBranch {
                branchName := strings.TrimPrefix(ref.LocalRef, "refs/heads/")
                branches = append(branches, branchName)
            }
        }
    }
    
    return &GitPushInfo{
        RemoteName: remoteName,
        RemoteURL:  remoteURL,
        Refs:       refs,
        Tags:       tags,
        Branches:   branches,
        IsDelete:   isDelete,
    }, nil
}

// validateTagSemantics validates that a tag follows semantic versioning
func validateTagSemantics(tag string) error {
    // Use version library to validate tag semantics
    detector := version.New()
    
    // Check if tag is a valid semantic version
    if err := detector.ValidateVersion(tag); err != nil {
        return fmt.Errorf("tag %s is not a valid semantic version: %w", tag, err)
    }
    
    return nil
}

// shouldSkipPrePushStage determines if we should skip the pre-push stage
func shouldSkipPrePushStage(pushInfo *GitPushInfo) bool {
    // Get current branch and tag
    currentBranch, err := getCurrentBranch()
    if err != nil {
        // If we can't get current branch, don't skip
        return false
    }
    
    currentTag, err := getCurrentTag()
    if err != nil {
        // If we can't get current tag, don't skip
        return false
    }
    
    // Check if we're pushing tags that are not the current tag
    if len(pushInfo.Tags) > 0 {
        for _, tag := range pushInfo.Tags {
            if tag != currentTag {
                return true
            }
        }
    }
    
    // Check if we're pushing branches that are not the current branch
    if len(pushInfo.Branches) > 0 {
        for _, branch := range pushInfo.Branches {
            if branch != currentBranch {
                return true
            }
        }
    }
    
    return false
}

// getCurrentBranch gets the current Git branch
func getCurrentBranch() (string, error) {
    cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// getCurrentTag gets the current Git tag
func getCurrentTag() (string, error) {
    cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// convertGitRefs converts GitRef to preexec.GitRef
func convertGitRefs(refs []GitRef) []preexec.GitRef {
    result := make([]preexec.GitRef, len(refs))
    for i, ref := range refs {
        result[i] = preexec.GitRef{
            LocalRef:  ref.LocalRef,
            LocalSHA:  ref.LocalSHA,
            RemoteRef: ref.RemoteRef,
            RemoteSHA: ref.RemoteSHA,
            IsDelete:  ref.IsDelete,
            IsTag:     ref.IsTag,
            IsBranch:  ref.IsBranch,
        }
    }
    return result
}

// runRoot handles the root command (print usage)
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
    
    // Print usage when no command is provided
    return cmd.Usage()
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
    executor := preexec.BuildfabExecutorWithCLIVersion(buildfabConfig, ui, getVersion())
    
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

// runInstall handles the install command (check and update Git hook)
func runInstall(cmd *cobra.Command, args []string) error {
    // Check and update Git hook if needed
    updated, err := checkAndUpdateGitHook()
    if err != nil {
        return fmt.Errorf("failed to update Git hook: %w", err)
    }
    
    // Only print success message if we actually updated the hook
    if updated {
        fmt.Printf("Git pre-push hook installed successfully (version %s)\n", getVersion())
    }
    
    return nil
}