# Technical Specification (TS)

**Project:** `buildfab-pre-push`
**Language:** Go
**Scope:** Cross-platform, configurable Git `pre-push` hook runner with DAG executor, built-in checks, and Buildfab integration.

## 1) CLI & Behavior

**Syntax:** `pre-push [options] [command]`
**Options:** `--version`, `-V`, `-h|--help`, `-d|--debug`, `-v|--verbose` (unchanged)
**Commands:**

* `test` — dry-run all steps (no self-install).
* `list-uses` — print registered built-ins (names + short help).
* *(no command)* — install/update itself at `.git/hooks/pre-push` if missing or version differs; exit 0.

When invoked by Git as a hook, read the standard `pre-push` stdin/argv (remote, refs) exactly as in the Bash version; detect current branch and tag before running stages. (Reference layout for actions/stages in your current YAML: actions, `uses` entries, and `stages.pre-push.steps`.) &#x20;

## 2) Config File: `project.yml`

Format mirrors GitHub Actions concepts but tailored to local pre-push checks.

### Structure (example with fixes & `require`)

```yaml
project:
  name: "buildfab-pre-push"
  modules: ["pre-push"]

actions:
  - name: version-check
    run: |
      version check ${{ tag }}

  - name: version-greatest
    run: |
      version check-greates ${{ tag }}  # external 'version' tool

  - name: git-untracked
    uses: git@untracked

  - name: git-uncommitted
    uses: git@uncommitted

  - name: git-modified
    uses: git@modified

stages:
  pre-push:
    steps:
      - action: version-check
      - action: version-greatest
      - action: git-untracked
        require: [version-check]          # runs after version-check completes
      - action: git-uncommitted
        require: [git-untracked]
      - action: git-modified
        require: [git-uncommitted]
        onerror: warn                     # warn | stop (default: stop)
```

Notes:

* Your current sample shows these same blocks and `uses` targets; names `version-greates` & `version-gretest` look like typos; the TS uses `version-greatest`. &#x20;
* Variables available for interpolation (GitHub-style):
  `${{ tag }}`, `${{ branch }}` (current repo state), plus the **version-go** library: `${{ version.version }}`, `${{ version.project }}`, `${{ version.module }}`, `${{ version.modules }}`.
* `onerror`: `stop` (default) or `warn` per step.
* Future: `env`, `if`, matrix, and reusable action files (out of scope for v1).

## 3) Executor (DAG Planner)

**Goal:** Run actions in parallel where possible; enforce explicit dependencies via `require`.

### 3.1 DAG Construction

* For each step in `stages.pre-push.steps`:

  * Node = referenced `action`.
  * Edges = `require:` list (string or list).
* Validate:

  * All required actions exist in `actions`.
  * No cycles (topological sort); fail fast with a readable error.
* Execution plan:

  * **Wave scheduling**: Execute all “ready” nodes (no unmet deps) **in parallel**.
  * When a node finishes, unblock its dependents; continue until DAG is complete.

### 3.2 Concurrency & Determinism

* Parallelism limited by `--max-parallel` (hidden flag; default = CPU count).
* Log stream per node; merge with step-scoped prefixes.
* Deterministic summary order: topological order for final report.

### 3.3 Error Policy

* Step failure:

  * If step has `onerror: warn` → mark **WARN**, continue unlocking dependents.
  * Else (`stop`) → mark **ERROR** and **do not** unlock dependents; executor continues only steps whose deps are already satisfied and don’t rely on the failed node (i.e., partial progress where allowed by the DAG).
* Final exit code:

  * Any `ERROR` → exit 1.
  * Only `WARN`/`OK` → exit 0.

## 4) Actions

Two kinds:

### 4.1 `uses:` (built-ins)

Registry keys like:

* `git@untracked` — fail if untracked files present.
* `git@uncommitted` — fail if staged/unstaged changes present.
* `git@modified` — fail if working tree differs from HEAD.
  (These mirror your current YAML `uses` entries.)&#x20;

**Repro hints (for errors):** built-ins expose a `--repro` string in code (e.g., `git status --porcelain=v1` or a small sequence) to print when a step fails (see §6).

**`list-uses` command:** prints name and 1-line help for each built-in.

### 4.2 `run:` (external)

Execute shell commands (multi-line allowed). If a `run:` step fails, the repro hint is simply the exact command block to run manually (printed verbatim).

## 5) Output & UX

### 5.1 Status lines

* Per step:

  * **OK**: `✔` in green with step name.
  * **WARN**: `⚠` in yellow.
  * **ERROR**: `✖` in red with short reason.
    Use same ANSI coloring convention as your Bash helpers (green/red/yellow).&#x20;

### 5.2 Module banner

When stage `pre-push` starts, print module banner line with **green check** on success or **red error** on failure (visual style akin to your second Bash variant; color policy per above).

### 5.3 Verbose / Debug

* `-v|--verbose`: show each concrete command **before** execution and its **captured stdout/stderr** after it completes (step-scoped, truncated with `--verbose-limit`).
* `-d|--debug`: includes environment details, variable expansions, working dir, and timing.

### 5.4 Repro tips on failure

* For **built-ins**: print “To reproduce:” followed by the built-in’s `--repro` string(s).
* For **run** steps: print “To reproduce:” plus the exact `run` block (indented, ready to paste).

## 6) Variable Interpolation

* Resolve `${{ ... }}` placeholders in `run:` blocks and action inputs:

  * Repo state: `tag`, `branch` (current).
  * `version-go` library values: `version.version`, `version.project`, `version.module`, `version.modules`.
* Fail with clear message for unresolved variables; suggest available keys.

## 7) Self-Install Logic

* On user invocation without command: install/update as `.git/hooks/pre-push`; preserve executable bit; write a tiny shim that execs the managed binary with pass-through of Git’s stdin/argv.
* Version check compares embedded version string to the installed shim’s metadata (or a marker file).

## 8) Build & Release Pipeline

* **Build**: integrate with **Conan + CMake** via `buildtools/build-conan.sh` (already present; adapt module naming and targets).
* **Package/Publish**: use **goreleaser** via `buildtools/build-and-package.sh` (already present; adapt module name); supports cross-platform builds and GitHub Releases.
* **CI/CD**: adapt existing **GitHub Actions** workflow files to this module (mostly renaming project/module).
  (Your YAML already declares the `pre-push` stage/steps; this TS focuses the executor and runtime; CI remains a thin wrapper.)&#x20;

## 9) Program Structure

* `cmd/pre-push/main.go` — CLI & top-level.
* `internal/install/` — hook install/update.
* `internal/config/` — YAML parsing, schema validation, variable expansion.
* `internal/exec/` — DAG builder, scheduler, runners, logging.
* `internal/uses/` — registry + implementations (`git_untracked.go`, `git_uncommitted.go`, `git_modified.go`, …).
* `internal/version/` — adapter for `version-go` library.
* `internal/ui/` — colored/status rendering, truncation, repro print.

## 10) Validation & Tests

* Unit tests for: YAML schema, variable resolver, DAG cycle detection, parallel scheduling, onerror semantics.
* E2E tests with temp Git repos: untracked/modified/uncommitted scenarios; tag/branch detection; verbose output snapshots.

## 11) Migration Notes

* Preserve action names from your current `project.yml`; fix typos in configs (`version-greates` → `version-greatest`; `version-gretest` → `version-greatest`). &#x20;
* Keep `uses: git@…` shortcuts as-is; provide `list-uses` to discoverability.&#x20;