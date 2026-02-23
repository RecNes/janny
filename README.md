# Janny

Janny is a CLI tool designed to keep your home directory organized by moving files from source directories to specific target directories based on user-defined rules, and backing up important files.

## Features

- **Automated Organization**: Scans source directories and moves files based on extensions.
- **Smart Conflict Resolution**: Automatically renames files if a file with the same name exists in the target.
- **Pattern Matching**: Support for glob patterns, regex, and folder patterns for advanced file organization.
- **Auto-Clean**: Automatically delete old files from specific categories after a configurable number of days.
- **Cloud File Handling**: Automatically detects and skips iCloud placeholder files to prevent system hangs (macOS).
- **External Handlers**: Delegate unknown file types to external scripts for custom classification.
- **Backup System**: Integrated `rsync` wrapper to backup organized files to external storage.
- **Dry Run Mode**: Preview changes without actually moving files.

## Installation

```bash
go install github.com/evrenesat/janny@latest
```

## Configuration

Janny uses a TOML configuration file located at `~/.config/janny/config.toml`.

### Example Config

```toml
[general]
source_paths = ["~/Downloads", "~/Desktop"]

[storage]
documents = "~/Documents/Sorted"
images = "~/Pictures/Sorted"
archives = "~/Downloads/Archives"

[rules]
documents = "pdf,doc,docx,txt,md"
images = "jpg,jpeg,png,gif"
archives = "zip,tar,gz,7z"

# New: Advanced Pattern Matching
# Globs (implicit with *, ?, [], or explicit with glob: prefix)
reports = "*report*.pdf, glob:monthly_*.xls"

# Regex (must start with regex:)
scans = "regex:^scan_\\d{4}\\.pdf$"

# Note: Patterns are checked BEFORE simple extension rules.
# So a file "final_report.pdf" will match 'reports' (glob) instead of 'documents' (pdf).

# Folders (must start with folder:)
projects = "folder:project_*"
misc = "folder:*" # Catch-all for folders

[auto_clean]
# Automatically delete files older than X days in specific categories
# Note: This is a destructive action!
installers = 15
tmp = 1

[backup]
enabled = false
destination = "/Volumes/ExternalDisk/Backups"
exclude_file_types = ["tmp", "bak", "DS_Store"]
exclude_directories = ["temp", "cache", "node_modules"]

[advanced]
# Optional: Path to a script that takes a filename and returns a category
unknown_file_handler = "/path/to/script.sh"
```

### External Classifier Protocol

If `unknown_file_handler` is set, Janny will invoke the specified command for any file whose extension is not found in the configuration.

- **Input**:
  - **Argument**: The full path to the file is passed as the first argument (`$1`) to the script.
  - **Stdin**: The full Janny configuration (including storage paths and rules) is passed to the script's standard input (stdin) as a JSON object. This allows the script to make decisions based on existing categories.
- **Output**: The script must print the **category name** (e.g., `documents`, `images`) to standard output (stdout).
- **Behavior**:
  - If the script returns a category defined in your `[storage]` config, the file is moved there.
  - If the script returns a **new** category not in `[storage]`, Janny auto-creates a storage directory under `default_storage_path` (e.g., `~/Backup/torrents`) and moves the file there.
  - If `default_storage_path` is not set and the category is unknown, the file is skipped.
  - If the script returns an empty string, the file is skipped.
  - If the script exits with a non-zero status code, the error is logged to stderr and the file is skipped. **Janny proceeds to the next file; the process does NOT crash.**

#### Example Script

```bash
#!/bin/bash
FILE="$1"

# You can also read config from stdin if needed:
# CONFIG=$(cat)
# echo "Processing $FILE with config: $CONFIG" >> /tmp/janny_debug.log

MIME=$(file -b --mime-type "$FILE")

if [[ "$MIME" == "application/x-bittorrent" ]]; then
    # Can be an existing or new category
    echo "torrents"
fi
```

### Auto-Clean

The `auto_clean` feature automatically deletes old files from specific categories after they've been organized. This is useful for temporary directories or files you don't need to keep long-term.

**How it works:**

- Runs automatically after each organization (unless in dry-run mode)
- Checks the modification time of files in the specified categories
- Deletes files (and directories) older than the configured number of days
- Skips cloud-backed placeholder files to avoid triggering downloads

**Configuration:**

```toml
[auto_clean]
installers = 15  # Delete installer files older than 15 days
downloads = 30   # Delete downloads older than 30 days
temp = 1         # Delete temp files older than 1 day
```

**Important Notes:**

- This is a **destructive operation** - deleted files cannot be recovered
- Only applies to files in your configured storage directories
- Set backup = true if you want to keep copies before deletion
- Cloud-backed files are automatically skipped to prevent unwanted downloads

## Usage

```bash
# different flags
janny --help
janny --dry-run
janny --verbose
janny --config ~/.config/janny/myconfig.toml
```

### Learning Mode

Janny can learn file extension rules from your existing organized directories.

```bash
janny --learn
```

This will scan your configured storage directories, identify file extensions, and update your configuration file to map those extensions to the corresponding storage categories.

### Smart Learn Mode (Handler Protocol)

You can use the `--learn-from-handler` flag to let an external script (e.g. an LLM wrapper) propose rules for unknown extensions in bulk.

**Command:**

```bash
janny --learn-from-handler
```

**Input to Handler (Stdin):**
Janny sends two parts to your configured `unknown_file_handler` script's stdin, separated by a newline:

1.  The content of `smart_learn_prompt`.
2.  A compact JSON object with the unknown extensions and current rules.

Format:

```text
<PROMPT_STRING>
<JSON_PAYLOAD>
```

Example Input:

```text
You are a file organizer...
{"unknown_extensions":["xyz","abc"],"rules":{"documents":"txt,md"}}
```

**Expected Output from Handler (Stdout):**
The script must print a JSON object containing only `rules`:

```json
{
  "rules": {
    "documents": "txt,md,xyz",
    "foos": "foo,abc"
  }
}
```

The handler can update existing categories (e.g., adding `xyz` to `documents`) or propose new ones (e.g., `foos`). For new categories, Janny automatically creates a storage directory under `default_storage_path` (e.g., `~/Backup/foos`).

**Configuration:**

```toml
[advanced]
unknown_file_handler = "/path/to/my_llm_script.sh"
smart_learn_prompt = "You are a file organizer. Analyze these extensions..."
default_storage_path = "~/Backup"
```

## Terminal User Interface (TUI)

Janny includes a powerful Terminal User Interface for interactive management.

### Features
- **Dashboard**: Real-time system status and action output.
- **Interactive Config**: Full-form editing of your rules and paths.
- **File Browser**: Select directories directly from the terminal.
- **English/Turkish Support**: Fully localized interface.

### Build and Run TUI
To build the TUI separately:
```bash
go build -o bin/janny-tui ./tui
./bin/janny-tui
```

## Development

1. Clone the repository
2. Run `go mod tidy`
3. Build CLI: `go build -o bin/janny ./cmd/janny`
4. Build TUI: `go build -o bin/janny-tui ./tui`
