# Janny

Janny is a CLI tool designed to keep your home directory organized by moving files from source directories to specific target directories based on user-defined rules, and backing up important files.

## Features

- **Automated Organization**: Scans source directories and moves files based on extensions.
- **Smart Conflict Resolution**: Automatically renames files if a file with the same name exists in the target.
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

[backup]
enabled = false
destination = "/Volumes/ExternalDisk/Backups"
exclude = [".DS_Store"]

[advanced]
# Optional: Path to a script that takes a filename and returns a category
unknown_file_handler = "/path/to/script.sh"
```

### External Classifier Protocol

If `unknown_file_handler` is set, Janny will invoke the specified command for any file whose extension is not found in the configuration.

- **Input**: The full path to the file is passed as the first argument to the script.
- **Output**: The script must print the **category name** (e.g., `documents`, `images`) to standard output (stdout).
- **Behavior**:
  - If the script returns a valid category defined in your `[storage]` config, the file is moved there.
  - If the script returns an empty string or an unknown category, the error is logged to stderr and existing rules continue processing.
  - If the script exits with a non-zero status code, the error is logged to stderr and the file is skipped. **Janny proceeds to the next file; the process does NOT crash.**

#### Example Script

```bash
#!/bin/bash
FILE="$1"
MIME=$(file -b --mime-type "$FILE")

if [[ "$MIME" == "application/x-bittorrent" ]]; then
    echo "torrents"
fi
```

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

## Development

1. Clone the repository
2. Run `go mod tidy`
3. Build with `go build ./cmd/janny`
