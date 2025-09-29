// Package ui provides user interface functionality for the pre-push tool.
package ui

import (
    "fmt"
    "io"
    "os"
    "strings"
    "time"

    "github.com/AlexBurnes/pre-push/pkg/prepush"
)

// UI handles user interface output and formatting
type UI struct {
    verboseLevel int
    debug        bool
    output       io.Writer
}

// New creates a new UI instance with boolean verbose (for backward compatibility)
func New(verbose, debug bool) *UI {
    verboseLevel := 0
    if verbose {
        verboseLevel = 1
    }
    return &UI{
        verboseLevel: verboseLevel,
        debug:        debug,
        output:       os.Stdout,
    }
}

// NewWithVerboseLevel creates a new UI instance with specific verbose level
func NewWithVerboseLevel(verboseLevel int, debug bool) *UI {
    return &UI{
        verboseLevel: verboseLevel,
        debug:        debug,
        output:       os.Stdout,
    }
}

// Printf prints formatted output
func (u *UI) Printf(format string, args ...interface{}) {
    fmt.Fprintf(u.output, format, args...)
}

// Println prints a line
func (u *UI) Println(args ...interface{}) {
    fmt.Fprintln(u.output, args...)
}

// PrintStepStatus prints the status of a step
func (u *UI) PrintStepStatus(stepName string, status prepush.Status, message string) {
    var icon string
    var color string
    
    switch status {
    case prepush.StatusOK:
        icon = "✔"
        color = "\033[32m" // Green
    case prepush.StatusWarn:
        icon = "⚠"
        color = "\033[33m" // Yellow
    case prepush.StatusError:
        icon = "✖"
        color = "\033[31m" // Red
    case prepush.StatusSkipped:
        icon = "⊘"
        color = "\033[90m" // Gray
    case prepush.StatusRunning:
        icon = "⟳"
        color = "\033[36m" // Cyan
    case prepush.StatusPending:
        icon = "○"
        color = "\033[37m" // White
    default:
        icon = "?"
        color = "\033[37m" // White
    }
    
    reset := "\033[0m"
    
    if message != "" {
        u.Printf("%s%s %s%s: %s\n", color, icon, reset, stepName, message)
    } else {
        u.Printf("%s%s %s%s\n", color, icon, reset, stepName)
    }
}

// PrintCLIHeader prints the CLI utility header with name and version
func (u *UI) PrintCLIHeader(name, version string) {
    u.Printf("\033[36m%s %s\033[0m\n", name, version)
}

// PrintProjectCheck prints the project name and version check message
func (u *UI) PrintProjectCheck(projectName, version string) {
    u.Printf("Checking \033[1m%s\033[0m (\033[1m%s\033[0m) before push\n", projectName, version)
}

// PrintStageHeader prints the header for a stage
func (u *UI) PrintStageHeader(stageName string) {
    u.Printf("\n\033[36m🚀 Running stage: %s\033[0m\n", stageName)
    u.Printf("\033[36m%s\033[0m\n", strings.Repeat("=", 20+len(stageName)))
}

// PrintStageResult prints the result of a stage
func (u *UI) PrintStageResult(stageName string, success bool, duration time.Duration) {
    if success {
        u.Printf("\n\033[32m✅ Stage '%s' completed successfully in %v\033[0m\n", stageName, duration)
    } else {
        u.Printf("\n\033[31m❌ Stage '%s' failed after %v\033[0m\n", stageName, duration)
    }
}

// PrintError prints an error message
func (u *UI) PrintError(err error) {
    u.Printf("❌ Error: %v\n", err)
}

// PrintWarning prints a warning message
func (u *UI) PrintWarning(message string) {
    u.Printf("⚠️  Warning: %s\n", message)
}

// PrintInfo prints an info message
func (u *UI) PrintInfo(message string) {
    u.Printf("ℹ️  Info: %s\n", message)
}

// PrintDebug prints a debug message (only if debug is enabled)
func (u *UI) PrintDebug(message string) {
    if u.debug {
        u.Printf("🐛 Debug: %s\n", message)
    }
}

// PrintVerbose prints a verbose message (only if verbose level > 0)
func (u *UI) PrintVerbose(message string) {
    if u.verboseLevel > 0 {
        u.Printf("📝 Verbose: %s\n", message)
    }
}

// PrintCommand prints a command that will be executed
func (u *UI) PrintCommand(command string) {
    if u.verboseLevel > 0 {
        u.Printf("🔧 Running: %s\n", command)
    }
}

// PrintCommandOutput prints the output of a command
func (u *UI) PrintCommandOutput(output string) {
    if u.verboseLevel > 0 && output != "" {
        u.Printf("📤 Output:\n%s\n", output)
    }
}

// PrintRepro prints reproduction instructions for a failed step
func (u *UI) PrintRepro(stepName, repro string) {
    u.Printf("\n🔍 To reproduce %s:\n", stepName)
    u.Printf("   %s\n", repro)
}

// PrintReproInline prints reproduction instructions inline with error message
func (u *UI) PrintReproInline(stepName, repro string) {
    u.Printf("   %s\n", repro)
}

// PrintSummary prints a summary of results
func (u *UI) PrintSummary(results []prepush.Result) {
    u.Printf("\n\033[36m📊 Summary:\033[0m\n")
    
    okCount := 0
    warnCount := 0
    errorCount := 0
    skippedCount := 0
    
    for _, result := range results {
        switch result.Status {
        case prepush.StatusOK:
            okCount++
        case prepush.StatusWarn:
            warnCount++
        case prepush.StatusError:
            errorCount++
        case prepush.StatusSkipped:
            skippedCount++
        }
    }
    
    // Only color items when count > 0, otherwise show in gray
    if okCount > 0 {
        u.Printf("   \033[32m✅ OK: %d\033[0m\n", okCount)
    } else {
        u.Printf("   \033[37m✅ OK: %d\033[0m\n", okCount)
    }
    
    if warnCount > 0 {
        u.Printf("   \033[33m⚠️  WARN: %d\033[0m\n", warnCount)
    } else {
        u.Printf("   \033[37m⚠️  WARN: %d\033[0m\n", warnCount)
    }
    
    if errorCount > 0 {
        u.Printf("   \033[31m❌ ERROR: %d\033[0m\n", errorCount)
    } else {
        u.Printf("   \033[37m❌ ERROR: %d\033[0m\n", errorCount)
    }
    
    if skippedCount > 0 {
        u.Printf("   \033[90m⊘ SKIPPED: %d\033[0m\n", skippedCount)
    } else {
        u.Printf("   \033[37m⊘ SKIPPED: %d\033[0m\n", skippedCount)
    }
    
    if errorCount > 0 {
        u.Printf("\n\033[31m❌ Some checks failed. Please fix the issues above before pushing.\033[0m\n")
    } else if warnCount > 0 {
        u.Printf("\n\033[33m⚠️  Some checks produced warnings. Review the output above.\033[0m\n")
    } else {
        u.Printf("\n\033[32m✅ All checks passed successfully!\033[0m\n")
    }
}

// IsVerbose returns true if verbose mode is enabled (for backward compatibility)
func (u *UI) IsVerbose() bool {
    return u.verboseLevel > 0
}

// GetVerboseLevel returns the current verbose level
func (u *UI) GetVerboseLevel() int {
    return u.verboseLevel
}

// IsDebug returns true if debug mode is enabled
func (u *UI) IsDebug() bool {
    return u.debug
}