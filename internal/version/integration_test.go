package version

import (
    "context"
    "testing"
)

func TestGetVersionInfo(t *testing.T) {
    ctx := context.Background()
    
    // Test getting version info
    info, err := GetVersionInfo(ctx)
    if err != nil {
        t.Logf("GetVersionInfo failed: %v", err)
        return
    }
    
    if info == nil {
        t.Fatal("GetVersionInfo returned nil info")
    }
    
    t.Logf("Version: '%s'", info.Version)
    t.Logf("Project: '%s'", info.Project)
    t.Logf("Module: '%s'", info.Module)
    t.Logf("Modules: %v", info.Modules)
    
    // Test individual functions
    if version, err := getVersionFromLibrary(ctx); err != nil {
        t.Logf("getVersionFromLibrary error: %v", err)
    } else {
        t.Logf("Direct version call: '%s'", version)
    }
}