# terraform-version-detector

A CLI tool to automatically detect Terraform version and pass it to GitHub Actions' `hashicorp/setup-terraform` action.

## Installation

### Install via Go

```bash
go install github.com/mrtc0/terraform-version-detector@latest
```

### Download Binary

Download the latest version from the [Releases](https://github.com/mrtc0/terraform-version-detector/releases) page.

## Usage

### Basic Usage

```bash
terraform-version-detector
```

By default, it detects the version from `.terraform-version` file or `.tf` files in the current directory.

### Configuration via Environment Variables

| Environment Variable | Description                                                         | Default Value        |
| -------------------- | ------------------------------------------------------------------- | -------------------- |
| `TARGET_DIR`         | Path to the directory to scan                                       | `.`                  |
| `VERSION_FILE`       | Name of the version file                                            | `.terraform-version` |
| `FALLBACK_VERSION`   | Fallback version if none found                                      | none                 |
| `GITHUB_OUTPUT`      | Output file path for GitHub Actions (if not set, outputs to stdout) | stdout               |

## Version Detection Logic

### 1. Detection from `.terraform-version` File

Reads the file content and extracts the version string.

**Example .terraform-version:**

```
1.5.0
```

or

```
v1.5.0
```

### 2. Detection from `required_version`

Extracts `required_version` from the `terraform` block in `.tf` files.

**Example main.tf:**

```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
```

This tool extracts the specific version number (`1.5.0`) from the version constraint (e.g., `>= 1.5.0`).

**Notes:**

- If `required_version` is defined in multiple `.tf` files, an error will be returned
- Uses the first version number found in the version constraint

### 3. Fallback Version

If the version is not found using the above methods, the version specified in `FALLBACK_VERSION` will be used.

## Output Format

The tool outputs in the following format:

```
version=1.5.0
source=terraform-version-file
```

When `GITHUB_OUTPUT` environment variable is set, the output will be appended to the specified file (typically used in GitHub Actions). Otherwise, it will be written to stdout.

### Output Fields

- `version`: Detected Terraform version
- `source`: Source of version detection
  - `terraform-version-file`: Detected from `.terraform-version` file
  - `required-version`: Detected from `required_version` in `.tf` files
  - `fallback`: Used fallback version
  - `not-found`: Version not found (on error)

## GitHub Actions Integration

This tool is designed to work seamlessly with GitHub Actions. The `GITHUB_OUTPUT` environment variable is automatically set in GitHub Actions workflows.

### Example Workflow

```yaml
name: Terraform Setup

on: [push]

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Detect Terraform Version
        id: detect-version
        run: |
          terraform-version-detector

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ steps.detect-version.outputs.version }}

      - name: Show Terraform Version
        run: terraform version
```

### With Custom Configuration

```yaml
- name: Detect Terraform Version
  id: detect-version
  env:
    TARGET_DIR: ./infrastructure
    VERSION_FILE: .tfversion
    FALLBACK_VERSION: 1.5.0
  run: |
    terraform-version-detector
```
