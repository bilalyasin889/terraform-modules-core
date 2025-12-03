#!/usr/bin/env bash
set -e

echo "===== Running tests for <module_name> ====="

echo "Running terragrunt init..."
terragrunt init -no-color

echo "Running terragrunt validate..."
terragrunt validate -no-color

echo "Running terragrunt plan..."
terragrunt plan -no-color -out tg.plan

echo "Applying module..."
terragrunt apply -auto-approve -no-color

echo "Capturing outputs..."
OUTPUTS=$(terragrunt output -json)
echo "$OUTPUTS"

# TODO: Add verification logic
# Example:
# echo "$OUTPUTS" | jq -e '.output_1.value == "expected_value"' \
#   || { echo "Output validation failed!"; exit 1; }

echo "Destroying module..."
terragrunt destroy -auto-approve -no-color

echo "===== Test completed successfully ====="
