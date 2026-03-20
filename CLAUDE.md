# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build System

The project uses CMake + Conan as its build system, with a pre-configured `.build/` directory already present.

```bash
# Configure (first time or after CMakeLists.txt changes)
cmake -S . -B .build

# Build for current platform
cmake --build .build --target pre-push

# Run tests
cmake --build .build --target test
# or directly:
go test ./... -race

# Run a single test package
go test ./internal/exec/...

# Run a specific test
go test -run TestName ./internal/exec/...

# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run
```

Build output goes to `bin/`. The version is injected at build time via ldflags (`-X main.appVersion=...`), sourced from the `VERSION` file or git tags via `scripts/version`.

## Architecture

This tool is a Git `pre-push` hook runner. The binary serves a dual role: **it is both the CLI tool and the hook itself** — `pre-push install` copies the binary to `.git/hooks/pre-push`. When Git invokes the hook, the binary detects it's running as a hook (by checking whether stdin is a terminal) and switches behavior automatically.

### Execution flow

When invoked as a Git hook:
1. Reads ref information from stdin (format: `<local_ref> <local_sha> <remote_ref> <remote_sha>`)
2. Skips if it's a delete operation or if pushing a branch/tag not on the current HEAD
3. Validates pushed tags against semantic versioning
4. Loads `.project.yml` via `buildfab.LoadConfig()` (supports YAML includes)
5. Creates a `BuildfabExecutor` and calls `RunStage(ctx, "pre-push")`

### Package structure

- **`cmd/pre-push/main.go`** — Single-file CLI entry point. Contains all Git hook detection logic, ref parsing, tag validation, and branch detection. Wires together the packages below.
- **`pkg/prepush/prepush.go`** — Public API types (`Config`, `Action`, `Stage`, `Step`, `Result`, `Status`, `Executor` interface). Also contains `Validate()` logic and the built-in action registry map.
- **`internal/exec/buildfab_executor.go`** — Core executor. Wraps `buildfab.SimpleRunner` for DAG-based parallel execution. Builds the variable map (`GetAllVariables()`) combining platform, version-go library, Git push info, and environment variables. Variable syntax is `${{ variable }}`.
- **`internal/config/config.go`** — Config loading (YAML parsing, variable resolution). The active code path uses `buildfab.LoadConfig()` directly; the internal loader exists for tests/alternate use.
- **`internal/uses/registry.go`** — Built-in `uses:` action implementations (`git@untracked`, `git@uncommitted`, `git@modified`). Each runner implements the `Runner` interface with `Run()`, `GetRepro()`, `GetHelp()`, `GetName()`.
- **`internal/version/`** — Version detection via the `version-go` library and git tag fallback.
- **`internal/ui/ui.go`** — Terminal output formatting (colored status indicators, verbose/debug modes).
- **`internal/install/installer.go`** — Hook installation helpers.

### Key design points

- **Hook detection**: `isGitHook()` in `main.go` checks if stdin is not a character device (i.e., it's a pipe from Git). Known subcommand names are explicitly excluded from hook mode.
- **Variable interpolation**: `BuildfabExecutor.GetAllVariables()` builds a flat `map[string]string` passed to buildfab. Variables from push info (tags, branches) take precedence over git-detected fallbacks.
- **Dependency on buildfab**: The actual DAG execution, container support, and output streaming are all delegated to `github.com/AlexBurnes/buildfab`. The `buildfab` binary is searched at runtime for container actions.
- **Version baked in at build**: `appVersion` is set via `-X main.appVersion=...` ldflags. The `version-go` library (`github.com/AlexBurnes/version-go`) provides additional version metadata at runtime from the module.

### Configuration file (`.project.yml`)

Loaded by `buildfab.LoadConfig()`, not the internal `config` package. The `internal/config` package is used for parsing/validation in tests. Actions either specify `run:` (shell command) or `uses:` (built-in action name like `git@untracked`). Steps in a stage can declare `require:` dependencies for DAG ordering.

## Environment Variables

- `PRE_PUSH_VERBOSE` — Sets verbose level (integer) when running as a Git hook
- `PRE_PUSH_DEBUG` — Set to `1` to enable debug output when running as a Git hook

## Go Coding Standards

- CGO disabled (`CGO_ENABLED=0`) for reproducible static builds.
- `gofmt -s` and `goimports` are mandatory; run `golangci-lint` before committing.
- Build flags: `-trimpath -ldflags="-s -w ..."` for releases.
- Imports ordered: standard library, third-party, local application.
- Exported identifiers must have doc comments (`// Foo …`).
- Functions should stay under ~60 lines; extract helpers when logic branches.
- No `util`/`helpers` package names — use specific names (`parse`, `exec`, `git`).

### Error handling

- No panics in normal flow; use `panic` only for truly unrecoverable init failures.
- Always wrap errors with context: `fmt.Errorf("opening %q: %w", path, err)`.
- User-facing CLI errors go to **stderr**; machine output goes to **stdout**.

### Context & concurrency

- Every long-running/IO function accepts `context.Context` as the first parameter.
- Never store contexts in structs; propagate deadlines.
- Top-level `main` derives a root context cancelled on signals.
- Prefer `errgroup` over manual `WaitGroup` + channels.

### Testing

- Table-driven tests; no sleep/timing-based flakiness.
- Always run with `-race`: `go test ./... -race`.
- Golden tests for stable textual outputs; keep fixtures in `testdata/`.
- Do not rely on network/filesystem in unit tests unless marked as integration.

## Version Management

The `VERSION` file is the single source of truth for the version number.

### Before making changes

Check if `VERSION` matches the current git tag. If they match, bump the version first:

```bash
# Check status
CURRENT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
cat VERSION

# Bump version (updates VERSION file + packaging files)
./scripts/version-bump patch   # bug fixes
./scripts/version-bump minor   # new features
./scripts/version-bump major   # breaking changes
```

### Files that must be updated on every version bump

1. `VERSION` — updated automatically by `./scripts/version-bump`
2. `packaging/macos/version.rb` — Homebrew formula
3. `packaging/windows/scoop-bucket/version.json` — Scoop manifest
4. `CHANGELOG.md` — new version entry

### Rebuild after version bump

```bash
go build -ldflags "-X main.appVersion=$(cat VERSION)" -o bin/pre-push ./cmd/pre-push
```

### Version format

- `VERSION` file and git tags: `vX.Y.Z` (with `v` prefix)
- URL format for releases: `https://github.com/AlexBurnes/pre-push/releases/download/vX.Y.Z/pre-push_X.Y.Z_...`

## Git & Commit Rules

### Commit message format

All commits must follow Conventional Commits with the version included:

```
<type>(<scope>): <vX.Y.Z>, <description>

<detailed description as bullet points>

<footer: Closes #N, BREAKING CHANGE: ...>
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`, `revert`

**Common scopes**: `ui`, `config`, `exec`, `tests`, `docs`, `build`, `ci`

**Examples**:
```
feat(exec): v1.2.0, add container support via buildfab

- Implement container action execution using buildfab SimpleRunner
- Add BuildfabBinaryPath search for container support
- Update executor to pass container config through opts

Closes #42
```
```
fix(git): v1.1.1, fix branch detection for detached HEAD

- Use rev-parse --abbrev-ref instead of symbolic-ref
- Handle detached HEAD state gracefully
```

### Tagging and pushing

Always push commits and tags together:

```bash
git tag vX.Y.Z
git push origin master --tags
```

### Retagging (when push fails or tag is missing on remote)

**Do not bump the version** — retag instead:

```bash
git tag -d vX.Y.Z                          # delete local tag
git add . && git commit -m "chore(retag): retag vX.Y.Z after push failure

- Remove and recreate tag due to push failure
- No version bump - using existing version number"
git tag vX.Y.Z
git push origin master --tags
```

Decision: bump version when `VERSION != latest remote tag`; retag when they match but tag is missing/wrong on remote.

### Pre-push checklist

```bash
go test ./... -race          # tests must pass
golangci-lint run            # linting must pass
git diff --cached            # review staged changes
cat VERSION                  # verify version is incremented
head -20 CHANGELOG.md        # verify changelog is updated
```

## Changelog Rules

`CHANGELOG.md` **must** be updated for every code or documentation change.

- Use `### Added`, `### Fixed`, `### Changed`, `### Documentation` sections.
- Version format: `## [X.Y.Z] - YYYY-MM-DD` (newest first).
- **Dates**: for existing tags use `git log -1 --format="%ai" <tag>`; for current work use `date +%Y-%m-%d`.
- Never use placeholder dates.

```bash
# Get all version dates
git tag --sort=-version:refname | xargs -I {} git log -1 --format="%ai {}" {}

# Get date for specific tag
git log -1 --format="%ai" v1.1.0
```

## Release Workflow ("complete changes")

When the user says "complete changes", "complete", "finish changes", or similar, execute this full workflow:

1. `./scripts/check-version-status` — check current version
2. `./scripts/version-bump <patch|minor|major>` — bump version
3. `go build -ldflags "-X main.appVersion=$(cat VERSION)" -o bin/pre-push ./cmd/pre-push` — rebuild binary
4. `date +%Y-%m-%d` — get current date for changelog
5. Update `CHANGELOG.md` and `README.md`
6. `git add . && git commit -m "<type>(<scope>): vX.Y.Z, <description>"` — commit
7. `git tag vX.Y.Z` — tag
8. `git push origin master --tags` — push (**never use `--no-verify`**)
9. If push fails: fix issues, commit, `git tag -d vX.Y.Z`, retag, push again

## Documentation Conventions

- All docs except `README.md` and `CHANGELOG.md` go in `docs/`.
- Filename format: `First-word-rest-of-name.md` (e.g., `Project-specification.md`, `Developer-workflow.md`).
- Abbreviations stay uppercase: `CI-cd.md`, `BNF-grammar.md`.
- `docs/Project-specification.md` is the single source of truth for the project spec — update it when behavior changes.
- Internal links use relative paths with forward slashes.
