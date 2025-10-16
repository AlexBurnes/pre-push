# Pre-Push Variable Reference

## Overview
Pre-push provides comprehensive variable support through integration with buildfab v0.25.0. All buildfab variables are available, plus additional Git-specific variables for enhanced pre-push hook functionality.

## Variable Categories

### Platform Variables
Variables provided by buildfab's platform detection system:

- `${{ platform }}` - Operating system platform (e.g., "linux", "darwin", "windows")
- `${{ arch }}` - CPU architecture (e.g., "amd64", "arm64")
- `${{ os }}` - Operating system name (e.g., "ubuntu", "macos", "windows")
- `${{ os_version }}` - Operating system version (e.g., "22.04", "13.0")
- `${{ cpu }}` - Number of CPU cores (e.g., "20")

### Project Variables
Variables from project configuration and version detection:

- `${{ project }}` - Project name from .project.yml
- `${{ module }}` - Primary module name
- `${{ modules }}` - Comma-separated list of all modules

### Version Variables
Variables from version-go library integration:

- `${{ version }}` - Current version (e.g., "1.9.1")
- `${{ version.version }}` - Same as `version`
- `${{ version.project }}` - Project name from version detection
- `${{ version.module }}` - Module name from version detection
- `${{ version.modules }}` - Comma-separated list of modules
- `${{ version.build-type }}` - Build type (e.g., "Release", "Debug")
- `${{ version.version-type }}` - Version type (e.g., "release", "snapshot")
- `${{ version.tag }}` - Current Git tag
- `${{ version.branch }}` - Current Git branch
- `${{ version.major }}` - Major version number (via buildfab)
- `${{ version.minor }}` - Minor version number (via buildfab)
- `${{ version.patch }}` - Patch version number (via buildfab)
- `${{ version.commit }}` - Current Git commit hash (via buildfab)
- `${{ version.date }}` - Version date (via buildfab)
- `${{ version.rawversion }}` - Raw version string (via buildfab)
- `${{ version.type }}` - Version type (via buildfab)

### Git Push Variables (Pre-Push Specific)
Additional variables available during Git push operations:

- `${{ tag }}` - Single pushed tag (most common case)
- `${{ branch }}` - Single pushed branch (most common case)
- `${{ tags }}` - Comma-separated list of all pushed tags
- `${{ branches }}` - Comma-separated list of all pushed branches

### Environment Variables
All environment variables are available with the `env.` prefix:

- `${{ env.HOME }}` - User home directory
- `${{ env.PATH }}` - System PATH
- `${{ env.USER }}` - Current user
- `${{ env.SHELL }}` - Current shell
- `${{ env.TERM }}` - Terminal type
- `${{ env.LANG }}` - Language setting
- `${{ env.PWD }}` - Current working directory
- Any other environment variable: `${{ env.VARIABLE_NAME }}`

## Variable Sources

### Buildfab Variables
Pre-push inherits all variables from buildfab v0.25.0 through:
1. `buildfab.GetPlatformVariables()` - Platform detection
2. `buildfab.AddPlatformVariables()` - Additional platform variables including version.*

### Version Library Variables
Pre-push uses `github.com/AlexBurnes/version-go` for version detection:
1. Version information from git tags and .project.yml
2. Build type and version type detection
3. Module and project information

### Git-Specific Variables
Pre-push provides additional Git variables through:
1. Git push information parsing (stdin during pre-push hook)
2. Version detector for current repository state
3. Enhanced Git ref parsing for push operations

## Usage Examples

### Platform-Specific Actions
```yaml
actions:
  - name: linux-specific
    run: echo "Running on ${{ platform }} with ${{ arch }}"
    only: [linux]
```

### Version Validation
```yaml
actions:
  - name: version-check
    run: |
      echo "Checking version ${{ version }}"
      echo "Build type: ${{ version.build-type }}"
      echo "Version type: ${{ version.version-type }}"
```

### Git Push Information
```yaml
actions:
  - name: tag-validation
    run: |
      echo "Pushing tag: ${{ tag }}"
      echo "Current branch: ${{ version.branch }}"
      echo "Current tag: ${{ version.tag }}"
```

### Environment Configuration
```yaml
actions:
  - name: env-check
    run: |
      echo "User: ${{ env.USER }}"
      echo "Home: ${{ env.HOME }}"
      echo "Shell: ${{ env.SHELL }}"
```

## Variable Interpolation

All variables are interpolated before action execution using the `${{ variable_name }}` syntax. This interpolation happens in:
- Action `run:` commands
- Step conditions
- Any string value in configuration

## Backward Compatibility

Pre-push maintains backward compatibility with older variable naming conventions:
- Both `${{ version }}` and `${{ version.version }}` work
- Both `${{ project }}` and `${{ version.project }}` work
- Both `${{ module }}` and `${{ version.module }}` work
- Both `${{ modules }}` and `${{ version.modules }}` work

## Verification

To see all available variables in your environment, you can create a test action:

```yaml
actions:
  - name: list-vars
    run: |
      echo "Platform: ${{ platform }}"
      echo "Arch: ${{ arch }}"
      echo "OS: ${{ os }} ${{ os_version }}"
      echo "CPU: ${{ cpu }}"
      echo "Project: ${{ project }}"
      echo "Module: ${{ module }}"
      echo "Version: ${{ version }}"
      echo "Build Type: ${{ version.build-type }}"
      echo "Version Type: ${{ version.version-type }}"
      echo "Branch: ${{ version.branch }}"
      echo "Tag: ${{ version.tag }}"
```

## References

- [Buildfab Documentation](https://github.com/AlexBurnes/buildfab)
- [Version-Go Library](https://github.com/AlexBurnes/version-go)
- [Pre-Push Project Specification](Project-specification.md)

