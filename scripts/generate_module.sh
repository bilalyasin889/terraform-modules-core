#!/usr/bin/env bash
set -e
##############################
# Prompt for module info
##############################
read -p "Enter module name (letters, numbers, underscores (_) or hyphens (-) only): " MODULE_NAME
if [[ ! "$MODULE_NAME" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    echo "Error: Invalid module name. Use only letters, numbers, underscores (_) or hyphens (-)."
    exit 1
fi

read -p "Enter a short module description (press Enter to skip): " MODULE_DESC
if [ -z "$MODULE_DESC" ]; then
    MODULE_DESC="TODO: Add module description"
fi

##############################
# Set paths
##############################
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Template paths
TEMPLATE_MODULE_PATH="templates/module"
TEMPLATE_TEST_PATH="templates/test"

# Destination paths
MODULE_PATH="modules/$MODULE_NAME"
MODULE_TEST_PATH="tests/$MODULE_NAME"

##############################
# Validate module paths
##############################
echo "Verifying module '$MODULE_NAME' can be generated..."

if [ -d "$MODULE_PATH" ]; then
    echo "Error: Module '$MODULE_NAME' already exists at $MODULE_PATH"
    exit 1
fi

if [ -d "$MODULE_TEST_PATH" ]; then
    echo "Error: Module '$MODULE_NAME' test folder already exists at $MODULE_TEST_PATH"
    exit 1
fi

##############################
# Generate module source files
##############################
echo "Generating module files at $MODULE_PATH"

# Create module folder
mkdir -p "$MODULE_PATH"

# Copy template files
cp "$TEMPLATE_MODULE_PATH/main.tf.template" "$MODULE_PATH/main.tf"
cp "$TEMPLATE_MODULE_PATH/variables.tf.template" "$MODULE_PATH/variables.tf"
cp "$TEMPLATE_MODULE_PATH/outputs.tf.template" "$MODULE_PATH/outputs.tf"
cp "$TEMPLATE_MODULE_PATH/README.md.template" "$MODULE_PATH/README.md"

# Replace placeholders in module files
find "$MODULE_PATH" -type f -exec sed -i '' "s/<module_name>/$MODULE_NAME/g" {} +
find "$MODULE_PATH" -type f -exec sed -i '' "s/<module_desc>/$MODULE_DESC/g" {} +

##############################
# Generate test files
##############################
echo "Generating test file at $MODULE_TEST_PATH"

# Create test folder
mkdir -p "$MODULE_TEST_PATH"

# Copy template test file
cp "$TEMPLATE_TEST_PATH/module_test.go.template" "$MODULE_TEST_PATH/${MODULE_NAME}_test.go"

# Replace placeholders in test files
find "$MODULE_TEST_PATH" -type f -exec sed -i '' "s/<module_name>/$MODULE_NAME/g" {} +

##############################
# Git staging (optional)
##############################
if git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    read -p "Do you want to stage the new module files in Git? (y/N) [Press Enter for Yes]: " ADD_GIT
    if [ -z "$ADD_GIT" ]; then
        ADD_GIT="Y"
    fi

    if [[ "$ADD_GIT" =~ ^[Yy]$ ]]; then
        git add "$MODULE_PATH" "$MODULE_TEST_PATH"
        echo "Module files staged in Git. Review changes before committing."
    else
        echo "Skipped staging module files in Git."
    fi
else
    echo "Not a Git repository. Skipping git add."
fi

echo "Module $MODULE_NAME generated successfully at $MODULE_PATH"
echo "Next steps: edit files and implement your Terraform resources."
