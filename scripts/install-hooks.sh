#!/bin/bash

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Copy the pre-commit hook
cp .git/hooks/pre-commit .git/hooks/pre-commit

# Make it executable
chmod +x .git/hooks/pre-commit

echo "Pre-commit hook installed successfully!" 