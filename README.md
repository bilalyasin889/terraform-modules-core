# terraform-modules-core

**Central repository for reusable Terraform modules**

This repository contains Terraform modules maintained by DevOps for internal and external projects. Developers consume these modules via Terragrunt projects and **do not modify them directly**. All modules are versioned, tested, and follow best practices.

---

## Folder Structure

```
modules/        # All Terraform modules
templates/      # Module scaffolding templates
scripts/        # Helper scripts: generate_module.sh, release_modules.sh, generate_docs.sh
tests/          # Module-specific Terragrunt test folders
```

---

## Usage

### 1. Scaffold a New Module
Use the script:
```bash
scripts/generate_module.sh
```
- Prompts for module name (required) and description (optional).
- Generates module folder in modules/<module_name> with template files:
  - main.tf
  - variables.tf
  - outputs.tf
  - README.md (includes tables for inputs and outputs, and example usage with placeholders)
- Creates test folder inside the module with:
  - terragrunt.hcl (local backend, inputs placeholders)
  - test.sh (Linux/macOS)
  - test.ps1 (Windows PowerShell)
- All placeholders are prefixed with TODO_ and must be replaced by the user.

Prompts to optionally stage the module files in Git if inside a Git repository.
- Generates module folder with `main.tf`, `variables.tf`, `outputs.tf`, `README.md` templates
- Creates test folder in `tests/<module_name>` with `terragrunt.hcl`, `terraform.tfvars`, and `verify.sh`
- All placeholders are prefixed with `TODO` and must be replaced

### 2. Testing Modules
Navigate to:
```
tests/<module_name>
```
Run:
``` bash
./test.sh
```
This executes the full lifecycle:
- `init`
- `validate`
- `plan`
- `apply`
- verification
- `destroy`

All `TODO_` placeholders in tests must be replaced before release.

### 3. Release Modules
Run:
```bash
scripts/release_modules.sh <major|minor|patch>
```
- Executes all module tests
- Increments semantic version (`patch`, `minor`, `major`)

---

## Module Guidelines
- All inputs, outputs, and verification logic must include `TODO_` placeholders until implemented
- Developers **consume modules** and do **not** commit changes directly
- Follow the testing workflow under `tests/` to ensure module integrity

---

## Terraform / Terragrunt Notes
- `.terraform/` and `.terragrunt-cache/` are ignored
- Terraform state files (`*.tfstate`, `*.tfstate.*`) are ignored
- Tests use a **local backend** to avoid conflicts

---

## Contribution Policy
- Repository is **DevOps-owned and maintained**
- Public for showcase only — **no external contributions allowed**
- Only the owner can create, commit or merge changes
- CI/CD runs validation on module release

