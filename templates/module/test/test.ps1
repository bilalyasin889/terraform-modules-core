Write-Host "===== Running tests for <module_name> ====="

Write-Host "Running terragrunt init..."
terragrunt init -no-color
if ($LASTEXITCODE -ne 0)
{
    Write-Error "terragrunt init failed."; exit $LASTEXITCODE
}

Write-Host "Running terragrunt validate..."
terragrunt validate -no-color
if ($LASTEXITCODE -ne 0)
{
    Write-Error "terragrunt validate failed."; exit $LASTEXITCODE
}

Write-Host "Running terragrunt plan..."
terragrunt plan -no-color -out tg.plan
if ($LASTEXITCODE -ne 0)
{
    Write-Error "terragrunt plan failed."; exit $LASTEXITCODE
}

Write-Host "Applying module..."
terragrunt apply -auto-approve -no-color
if ($LASTEXITCODE -ne 0)
{
    Write-Error "terragrunt apply failed."; exit $LASTEXITCODE
}

Write-Host "Capturing outputs..."
$OUTPUTS = terragrunt output -json
if ($LASTEXITCODE -ne 0)
{
    Write-Error "terragrunt output failed."; exit $LASTEXITCODE
}

Write-Host $OUTPUTS

# TODO: Add verification logic
# Example:
# if (($OUTPUTS | ConvertFrom-Json).output_1.value -ne "expected_value") { Write-Error "Output validation failed!"; exit 1 }

Write-Host "Destroying module..."
terragrunt destroy -auto-approve -no-color
if ($LASTEXITCODE -ne 0) {
    Write-Error "terragrunt destroy failed.";
    exit $LASTEXITCODE
}

Write-Host "===== Test completed successfully ====="
