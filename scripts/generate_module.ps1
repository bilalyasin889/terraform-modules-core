# Prompt for module name
$ModuleName = Read-Host "Enter module name (letters, numbers, underscores (_) or hyphens (-) only)"

# Validate module name
if ($ModuleName -notmatch '^[a-zA-Z0-9_-]+$') {
    Write-Host "Error: Invalid module name. Use only letters, numbers, underscores (_) or hyphens (-)."
    exit 1
}

# Move to repo root (assumes script is in scripts/)
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location (Resolve-Path "$ScriptDir\..")

$ModulePath = "modules\$ModuleName"
$TemplatePath = "templates/module"

# Check if module already exists
if (Test-Path $ModulePath) {
    Write-Host "Error: Module '$ModuleName' already exists at $ModulePath"
    exit 1
}

Write-Host "Generating module: $ModuleName in $ModulePath"

# Create folder structure
New-Item -ItemType Directory -Force -Path "$ModulePath\test" | Out-Null

# Copy template files
Copy-Item "$TemplatePath\main.tf" "$ModulePath\main.tf" -Force
Copy-Item "$TemplatePath\variables.tf" "$ModulePath\variables.tf" -Force
Copy-Item "$TemplatePath\outputs.tf" "$ModulePath\outputs.tf" -Force
Copy-Item "$TemplatePath\README.md" "$ModulePath\README.md" -Force
Copy-Item "$TemplatePath\test\terragrunt.hcl" "$ModulePath\test\terragrunt.hcl" -Force
Copy-Item "$TemplatePath\test\test.sh" "$ModulePath\test\test.sh" -Force
Copy-Item "$TemplatePath\test\test.ps1" "$ModulePath\test\test.ps1" -Force

# Replace placeholders in all files
Get-ChildItem -Path $ModulePath -Recurse -File | ForEach-Object {
    (Get-Content $_.FullName) -replace "<module_name>", $ModuleName | Set-Content $_.FullName
}

# Prompt for module description (optional)
$ModuleDesc = Read-Host "Enter a short module description (press Enter to skip)"
if (-not $ModuleDesc) {
    $ModuleDesc = "TODO: Add module description"
}

# Replace <module_desc> placeholder in all module files
Get-ChildItem -Path $ModulePath -Recurse -File | ForEach-Object {
    (Get-Content $_.FullName) -replace "<module_desc>", $ModuleDesc | Set-Content $_.FullName
}

# Git staging
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
        Write-Host "Module files staged in Git. Review changes before committing."
    } else {
        Write-Host "Skipped staging module files in Git."
    }
} else {
    Write-Host "Not a Git repository. Skipping git add."
}

Write-Host "Module $ModuleName generated successfully at $ModulePath"
Write-Host "Next steps: edit files and implement your Terraform resources."
