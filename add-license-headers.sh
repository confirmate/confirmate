#!/bin/bash
# Script to add/update license headers in all Go files

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to the repository root
cd "$SCRIPT_DIR"

# Apply headers to all .go files, excluding generated files
go run github.com/google/addlicense@latest \
    -f .addlicense-header.txt \
    -ignore "**/.*" \
    -ignore "**/*.pb.go" \
    -ignore "**/*.connect.go" \
    -ignore "**/node_modules/**" \
    -ignore "**/vendor/**" \
    -v \
    $(find . -name "*.go" -not -path "*/.*" -not -name "*.pb.go" -not -name "*.connect.go")

echo "✓ License headers updated!"
