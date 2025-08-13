#!/bin/bash

# Install Git hooks for pkm-sync development
# This script copies hooks from scripts/hooks/ to .git/hooks/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_SOURCE="$SCRIPT_DIR/hooks"
HOOKS_TARGET="$REPO_ROOT/.git/hooks"

echo "Installing Git hooks for pkm-sync..."

# Check if we're in a git repository
if [ ! -d "$REPO_ROOT/.git" ]; then
    echo "Error: Not in a Git repository"
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_TARGET"

# Install each hook
for hook in "$HOOKS_SOURCE"/*; do
    if [ -f "$hook" ]; then
        hook_name=$(basename "$hook")
        echo "Installing $hook_name hook..."
        
        # Copy the hook
        cp "$hook" "$HOOKS_TARGET/$hook_name"
        
        # Make it executable
        chmod +x "$HOOKS_TARGET/$hook_name"
        
        echo "✓ $hook_name hook installed"
    fi
done

echo ""
echo "Git hooks installed successfully!"
echo ""
echo "Available hooks:"
ls -la "$HOOKS_TARGET" | grep -v "\.sample$" | grep -E "^-.*x.*" | awk '{print "  - " $9}'
echo ""
echo "The pre-commit hook will automatically:"
echo "  - Run 'go fmt' on staged Go files"
echo "  - Execute 'make ci' (lint, test, build) to ensure code quality"
echo "  - Prevent commits if any checks fail"