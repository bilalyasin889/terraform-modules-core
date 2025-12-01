#!/usr/bin/env bash
set -e

# Prompt for module name
read -p "Enter module name (letters, numbers, underscores (_) or hyphens (-) only): " MODULE_NAME

# Validate module name
if [[ ! "$MODULE_NAME" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    echo "Error: Invalid module name. Use only letters, numbers, underscores (_) or hyphens (-)."
    exit 1
fi

# Move to repo root (assumes script is in scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

MODULE_PATH="modules/$MODULE_NAME"
TEMPLATE_PATH="templates/module"

# Check if module already exists
if [ -d "$MODULE_PATH" ]; then
    echo "Error: Module '$MODULE_NAME' already exists at $MODULE_PATH"
    exit 1
fi

echo "Generating module: $MODULE_NAME in $MODULE_PATH"

# Create folder structure
mkdir -p "$MODULE_PATH/test"

# Copy template files
cp "$TEMPLATE_PATH/main.tf" "$MODULE_PATH/main.tf"
cp "$TEMPLATE_PATH/variables.tf" "$MODULE_PATH/variables.tf"
cp "$TEMPLATE_PATH/outputs.tf" "$MODULE_PATH/outputs.tf"
cp "$TEMPLATE_PATH/README.md" "$MODULE_PATH/README.md"
cp "$TEMPLATE_PATH/test/terragrunt.hcl" "$MODULE_PATH/test/terragrunt.hcl"
cp "$TEMPLATE_PATH/test/test.sh" "$MODULE_PATH/test/test.sh"
cp "$TEMPLATE_PATH/test/test.ps1" "$MODULE_PATH/test/test.ps1"

# Replace <module_name> placeholder
find "$MODULE_PATH" -type f -exec sed -i '' "s/<module_name>/$MODULE_NAME/g" {} +

# Prompt for module description (optional)
read -p "Enter a short module description (press Enter to skip): " MODULE_DESC
if [ -z "$MODULE_DESC" ]; then
    MODULE_DESC="TODO: Add module description"
fi

# Replace <module_desc> placeholder
find "$MODULE_PATH" -type f -exec sed -i '' "s/<module_desc>/$MODULE_DESC/g" {} +

# Git staging
if git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    read -p "Do you want to stage the new module files in Git? (y/N) [Press Enter for Yes]: " ADD_GIT
    if [ -z "$ADD_GIT" ]; then
        ADD_GIT="Y"
    fi

    if [[ "$ADD_GIT" =~ ^[Yy]$ ]]; then
        git add "$MODULE_PATH"
        echo "Module files staged in Git. Review changes before committing."
    else
        echo "Skipped staging module files in Git."
    fi
else
    echo "Not a Git repository. Skipping git add."
fi

echo "Module $MODULE_NAME generated successfully at $MODULE_PATH"
echo "Next steps: edit files and implement your Terraform resources."
