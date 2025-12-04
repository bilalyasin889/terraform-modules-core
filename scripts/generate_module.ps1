##############################
# Prompt for module info
##############################
$ModuleName = Read-Host "Enter module name (letters, numbers, underscores (_) or hyphens (-) only)"
if ($ModuleName -notmatch '^[a-zA-Z0-9_-]+$') {
    Write-Host "Error: Invalid module name. Use only letters, numbers, underscores (_) or hyphens (-)."
    exit 1
}

$ModuleDesc = Read-Host "Enter a short module description (press Enter to skip)"
# Set default description if skipped
if (-not $ModuleDesc) {
    $ModuleDesc = "TODO: Add module description"
}

##############################
# Set paths
##############################
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location (Resolve-Path "$ScriptDir\..")

# Template paths
$TemplateModulePath = "templates/module"
$TemplateTestPath   = "templates/test"

# Destination paths
$ModulePath     = "modules\$ModuleName"
$ModuleTestPath = "tests\$ModuleName"

##############################
# Validate module paths
##############################
Write-Host "Verifying module '$ModuleName' can be generated..."

if (Test-Path $ModulePath) {
    Write-Host "Error: Module '$ModuleName' already exists at $ModulePath"
    exit 1
}

if (Test-Path $ModuleTestPath) {
    Write-Host "Error: Module '$ModuleName' test folder already exists at $ModuleTestPath"
    exit 1
}

##############################
# Generate module source files
##############################
Write-Host "Generating module files at $ModulePath"

# Create module folder
New-Item -ItemType Directory -Force -Path "$ModulePath" | Out-Null

# Copy template files
Copy-Item "$TemplateModulePath\main.tf.template" "$ModulePath\main.tf" -Force
Copy-Item "$TemplateModulePath\variables.tf.template" "$ModulePath\variables.tf" -Force
Copy-Item "$TemplateModulePath\outputs.tf.template" "$ModulePath\outputs.tf" -Force
Copy-Item "$TemplateModulePath\README.md.template" "$ModulePath\README.md" -Force

# Replace placeholders
Get-ChildItem -Path $ModulePath -Recurse -File | ForEach-Object {
    $content = Get-Content $_.FullName
    $content = $content -replace "<module_name>", $ModuleName
    $content = $content -replace "<module_desc>", $ModuleDesc
    Set-Content -Path $_.FullName -Value $content
}

##############################
# Generate test files
##############################
Write-Host "Generating test file at $ModuleTestPath"

# Create test folder
New-Item -ItemType Directory -Force -Path "$ModuleTestPath" | Out-Null

# Copy template test file
Copy-Item "$TemplateTestPath\module_test.go.template" "$ModuleTestPath\${ModuleName}_test.go" -Force

# Replace placeholders in test files
Get-ChildItem -Path $ModuleTestPath -Recurse -File | ForEach-Object {
    $content = Get-Content $_.FullName
    $content = $content -replace "<module_name>", $ModuleName
    Set-Content -Path $_.FullName -Value $content
}

##############################
# Git staging (optional)
##############################
try {
    git rev-parse --is-inside-work-tree | Out-Null
    $isGitRepo = $true
} catch {
    $isGitRepo = $false
}

if ($isGitRepo) {
    $addGit = Read-Host "Do you want to stage the new module files in Git? (y/N) [Press Enter for Yes]"

    # Default to Yes if no input
    if (-not $addGit) { $addGit = "Y" }

    if ($addGit -match '^[Yy]$') {
        git add $ModulePath
        git add $ModuleTestPath
        Write-Host "Module files staged in Git. Review changes before committing."
    } else {
        Write-Host "Skipped staging module files in Git."
    }
} else {
    Write-Host "Not a Git repository. Skipping git add."
}

Write-Host "Module '$ModuleName' generated successfully at $ModulePath"
Write-Host "Next steps: edit files and implement your Terraform resources."
