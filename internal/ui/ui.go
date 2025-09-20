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
    verbose bool
    debug   bool
    output  io.Writer
}

// New creates a new UI instance
func New(verbose, debug bool) *UI {
    return &UI{
        verbose: verbose,
        debug:   debug,
        output:  os.Stdout,
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

// PrintStageHeader prints the header for a stage
func (u *UI) PrintStageHeader(stageName string) {
    u.Printf("\n🚀 Running stage: %s\n", stageName)
    u.Printf("%s\n", strings.Repeat("=", 20+len(stageName)))
}

// PrintStageResult prints the result of a stage
func (u *UI) PrintStageResult(stageName string, success bool, duration time.Duration) {
    if success {
        u.Printf("\n✅ Stage '%s' completed successfully in %v\n", stageName, duration)
    } else {
        u.Printf("\n❌ Stage '%s' failed after %v\n", stageName, duration)
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

// PrintVerbose prints a verbose message (only if verbose is enabled)
func (u *UI) PrintVerbose(message string) {
    if u.verbose {
        u.Printf("📝 Verbose: %s\n", message)
    }
}

// PrintCommand prints a command that will be executed
func (u *UI) PrintCommand(command string) {
    if u.verbose {
        u.Printf("🔧 Running: %s\n", command)
    }
}

// PrintCommandOutput prints the output of a command
func (u *UI) PrintCommandOutput(output string) {
    if u.verbose && output != "" {
        u.Printf("📤 Output:\n%s\n", output)
    }
}

// PrintRepro prints reproduction instructions for a failed step
func (u *UI) PrintRepro(stepName, repro string) {
    u.Printf("\n🔍 To reproduce %s:\n", stepName)
    u.Printf("   %s\n", repro)
}

// PrintSummary prints a summary of results
func (u *UI) PrintSummary(results []prepush.Result) {
    u.Printf("\n📊 Summary:\n")
    
    okCount := 0
    warnCount := 0
    errorCount := 0
    
    for _, result := range results {
        switch result.Status {
        case prepush.StatusOK:
            okCount++
        case prepush.StatusWarn:
            warnCount++
        case prepush.StatusError:
            errorCount++
        }
    }
    
    u.Printf("   ✅ OK: %d\n", okCount)
    u.Printf("   ⚠️  WARN: %d\n", warnCount)
    u.Printf("   ❌ ERROR: %d\n", errorCount)
    
    if errorCount > 0 {
        u.Printf("\n❌ Some checks failed. Please fix the issues above before pushing.\n")
    } else if warnCount > 0 {
        u.Printf("\n⚠️  Some checks produced warnings. Review the output above.\n")
    } else {
        u.Printf("\n✅ All checks passed successfully!\n")
    }
}

// IsVerbose returns true if verbose mode is enabled
func (u *UI) IsVerbose() bool {
    return u.verbose
}

// IsDebug returns true if debug mode is enabled
func (u *UI) IsDebug() bool {
    return u.debug
}