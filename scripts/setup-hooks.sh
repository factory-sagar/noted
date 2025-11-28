#!/bin/bash

# =============================================================================
# Noted - Git Hooks Setup
# =============================================================================
# Run this script to configure git hooks for this repository
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Setting up git hooks for Noted..."
echo ""

# Configure git to use .githooks directory
cd "$REPO_ROOT"
git config core.hooksPath .githooks

# Make hooks executable
chmod +x .githooks/pre-commit
chmod +x .githooks/post-commit
chmod +x .githooks/commit-msg

echo "✓ Git hooks configured successfully!"
echo ""
echo "Hooks installed:"
echo "  • pre-commit  - Code quality and security checks"
echo "  • commit-msg  - Conventional commit message validation"
echo "  • post-commit - Dashcode tracking (localhost:3001)"
echo ""
echo "To skip hooks (not recommended):"
echo "  git commit --no-verify"
echo ""
