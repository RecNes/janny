# Architecture - janny

## Overview

`janny` is a Go-based CLI tool designed to keep the user's home directory organized by moving files from source directories (configured in `source_paths`) to specific target directories based on user-defined rules, and backing up important files.

## Core Components

### 1. Config

Handles loading and validation of TOML configuration.

- Location: `~/.config/janny/config.toml`
- **Logic**: Maps "categories" (e.g. `documents`) to lists of extensions (e.g. `pdf, doc`), and allows defining paths for those categories.

### 2. Organizer

The engine that traverses the source directories and decides where files should go.

- **Input**: List of Source Paths, Rules (Extension -> Category Map).
- **Process**:
  1. Iterate through each source path.
  2. For each file:
     - Determine extension.
     - Match extension to a Category.
     - Resolve Category to a Target Path.
     - If no match and external handler exists, query external handler.
     - Move file to target.

### 3. External Handler

Interface to communicate with an arbitrary external command to classify unknown files.

- **Protocol**: Pass filename/metadata as args or stdin, receive target category/path as stdout.

### 4. Backup

Wrapper around `rsync`.

- Handles exclusion lists.
- Syncs "safe" directories (not explicitly excluded) to external storage.

## Data Flow

1. User invokes `janny`.
2. App loads Config and builds the Extension -> Category lookup map.
3. App scans all `source_paths`.
4. For each file:
   - Check internal rules (O(1) lookup map).
   - If not found: Call External Handler.
   - Move file.
5. If backup enabled, run Rsync.

## Tech Stack

- Language: Go
- Config: TOML
- Backup: `rsync` (system command)
