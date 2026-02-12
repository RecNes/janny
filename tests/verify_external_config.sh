#!/bin/bash
set -e

# Build janny - SKIPPED to avoid permission issues
export GOTMPDIR=$(pwd)/.gotmp
mkdir -p .gotmp
# rm -f janny
# go build -v -o janny ./cmd/janny

# Create temp directories
mkdir -p temp/source
mkdir -p temp/docs
touch temp/source/unknown.xyz

# Create temp config
cat <<EOF > temp/config.toml
[general]
source_paths = ["$(pwd)/temp/source"]

[storage]
documents = "$(pwd)/temp/docs"

[rules]
documents = "txt"

[advanced]
unknown_file_handler = "$(pwd)/tests/repro_external_config.py"
EOF

# Run janny using go run
go run ./cmd/janny --config temp/config.toml --dry-run

# Check log
if grep -q "Successfully parsed JSON config" external_handler.log; then
    echo "Verification PASSED: JSON config received and parsed."
else
    echo "Verification FAILED: JSON config not found in log."
    cat external_handler.log
    exit 1
fi

# Cleanup
rm -f janny temp/config.toml temp/source/unknown.xyz
rm -rf temp/source temp/docs
rm -f external_handler.log
