Write-Host "===== Running tests for <module_name> ====="

# Cleanup previous runs
if (Test-Path ".terragrunt-cache") {
    Remove-Item -Recurse -Force .terragrunt-cache
}

Write-Host "Running terragrunt init..."
terragrunt init -no-color

Write-Host "Running terragrunt validate..."
terragrunt validate -no-color

Write-Host "Running terragrunt plan..."
terragrunt plan -no-color -out tg.plan

Write-Host "Applying module..."
terragrunt apply -auto-approve -no-color

Write-Host "Capturing outputs..."
$OUTPUTS = terragrunt output -json
Write-Host $OUTPUTS

# TODO: Add verification logic
# Example:
# if (($OUTPUTS | ConvertFrom-Json).output_1.value -ne "expected_value") { Write-Error "Output validation failed!"; exit 1 }

Write-Host "Destroying module..."
terragrunt destroy -auto-approve -no-color

Write-Host "===== Test completed successfully ====="
