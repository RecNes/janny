# Developer Log

## 2026-02-12

### Initial Setup

- Initialized `DEVLOG.md` to track progress.
- Starting implementation of `janny` based on `implementation_plan.md`.

### Component Implementation

- **Configuration**: Implemented TOML configuration loading using `github.com/pelletier/go-toml/v2`. Added logic to expand `~` to user home directory.
- **Organizer**: Implemented the core logic for scanning directories and moving files based on extension rules. Added conflict resolution (renaming) and dry-run support.
- **External Handler**: Created a handler to execute external scripts for unknown file types.
- **Backup**: Implemented an `rsync` wrapper to backup source directories to a configured destination, respecting exclusions.
- **CLI**: Created the main entry point `cmd/janny/main.go` tying all components together with flags for config path, dry-run, and verbose mode.

### Feature Update: Learn Mode & Defaults

- **Default Config**: `janny` now automatically generates a default configuration file if one is missing at the specified path.
- **Learn Mode**: Added `--learn` flag. This mode scans the target storage directories defined in the config. For every file found, if its extension is unknown to `janny`, it infers the rule (Extension -> Category) and updates the configuration file automatically.

### Issues

- Encountered permission errors when running `go mod tidy` in the environment. This prevented automatic dependency resolution and build verification. The user will need to run `go mod tidy` in their local environment to resolve dependencies.

### Bug Fixes

- **First Run Configuration**: Fixed a critical bug where the default configuration created on first run was not being processed to expand `~` in paths. This caused `lstat` errors like `lstat ~/Downloads: no such file or directory`. The fix ensures `cfg.process()` is called immediately after creating the default config.

### Refactoring

- **Organizer Output**: Refactored the organizer to split the process into planning and execution phases. This allows for a clean, tree-structured output in `dry-run` mode (and verbose mode), significantly reducing noise compared to the previous line-by-line logging.

### Feature Update: Directory Rules & Auto Clean

- **Directory Support**: Implemented `folder:` prefix for rules, enabling organization of entire directories based on patterns (e.g., `folder:project_*`).
- **Auto Clean**: Added `[auto_clean]` configuration section to automatically delete files in specific categories older than a configured number of days.
- **Organizer Logic**: Updated the organizer to handle directory entries and run the cleaning process after organization.
- **Verification**: Verified the new features using a reproduction script `temp/repro.go` in a controlled environment, confirming correct behavior for folder matching and file deletion.

### Bug Fix: Exclude Storage Paths

- **Issue**: Storage paths nested within source paths were being processed by the organizer, potentially causing infinite loops or errors when moving files into themselves.
- **Fix**: Updated `Organizer` to track configured storage paths and explicitly exclude them during the planning phase.
- **Verification**: Confirmed via `temp/repro_exclude.go` that nested storage directories are skipped during scanning.

### Improvement: UX & Stability

- **Graceful Shutdown**: Implemented signal handling (`SIGINT`/`SIGTERM`) to allow safe cancellation with Ctrl+C.
- **Progress Feedback**: Added in-place progress updates (`Scanning...`, `Moving...`) to provide visual feedback during long operations without flooding the terminal (when not in verbose mode).
- **Context Awareness**: Updated organizer loops to respect context cancellation immediately.

### Improvement: Cloud File Support

- **Detection**: Added logic to detect macOS-specific `SF_DATALESS` (0x40000000) flag on files, which indicates iCloud placeholders.
- **Cross-Platform**: Implemented platform-specific checks (`platform_darwin.go`, `platform_other.go`) to ensure Linux compatibility is maintained.
- **Always-On**: Cloud file detection is always enabled - no configuration required. This prevents accidental processing of placeholder files that would trigger downloads and cause system hangs.
- **Directory Safety**: The organizer checks contents of directories before moving them. If a directory contains any cloud placeholders, it is skipped entirely to prevent OS-level hangs during move operations.
- **Pattern-Matched Directories**: Cloud file checking is integrated into the pattern matching logic, so even directories matched by folder patterns (like `folder:*`) are checked for cloud files before moving.

### Bug Fix: Dry Run Consistency

- **Issue**: `dry-run` mode was not respected by `Learn` (which updated config) and `AutoCleanup` (which was skipped entirely or would have deleted files).
- **Fix**: Updated `Organizer.Learn` and `Organizer.Clean` to check `dryRun` flag. `Clean` now reports what it would delete, and `Learn` reports what it would learn, without side effects.
- **Verification**: Added `internal/organizer/dryrun_test.go` to verify these scenarios.

### Enhancement: Backup Exclusion Control

- **Granular Control**: Split the generic `exclude` backup configuration into `exclude_file_types` and `exclude_directories`.
- **Implementation**: Updated `internal/config` and `internal/backup` to handle these specific exclusions, ensuring correct argument generation for `rsync` (e.g., `*.ext` for types, `dir/` for directories).
- **Issues Fixed**:
  - **Backup Source**: Changed backup source from `source_paths` (input dirs) to `storage` (organized destination dirs), ensuring we backup the clean structure.
  - **Dry Run**: Improved `dry-run` behavior to execute `rsync` with `-n` flag instead of just printing the command, providing accurate preview of file operations.

### Feature Update: Learn From Handler

- **Bulk Learning**: Implemented `--learn-from-handler` flag.
- **Mechanism**: Scans both source and storage paths for unknown extensions. Sends the list (along with current rules) to the configured `unknown_file_handler` in a simplified JSON payload.
- **Protocol**:
  - Input (Stdin): `Prompt + \n + JSON({"unknown_extensions": [...], "rules": {...}})`
  - Output (Stdout): `{"rules": {...}}`
- **Config**:
  - `smart_learn_prompt`: Prompt for the external handler.
  - `default_storage_path`: Base path for automatically creating storage directories for new categories proposed by the handler.
- **Verification**: Updated `internal/organizer/learn_smart_test.go` to verify the simplified protocol and default storage path logic.

## 2026-02-13

### Documentation Fix: Learn-From-Handler Protocol

- **Issue**: `README.md` still documented the old handler protocol that included a `storage` field in the expected handler output. The code had already been updated to only expect `rules`.
- **Fix**: Updated the Smart Learn Mode section in `README.md` to remove `storage` from the expected output, document `default_storage_path` config variable, and clarify that Janny auto-creates storage directories for new categories.
