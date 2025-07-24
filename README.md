
# Hooks for Infrastructure as Code (IaC) tools

This repository contains a collection of hooks for Infrastructure as Code (IaC) tools. The hooks are designed to be used with [pre-commit](https://pre-commit.com/), a framework for managing and maintaining multi-language pre-commit hooks.

## Available Hooks

### tofu-fmt

#### Formats OpenTofu configuration files

Runs `tofu fmt` to rewrite your OpenTofu (`.tf`, `.tofu`, `.tfvars`) files to a canonical format and style. This helps ensure consistency and readability across your infrastructure codebase. It will not modify files in `.terraform/` directories.

### tofu-validate

#### Validates OpenTofu configuration files

Runs `tofu validate` to check your configuration for syntax errors and internal consistency, without accessing remote services or APIs. This helps catch mistakes before applying changes. It will not validate files in `.terraform/` directories.

---

## Usage

To use these hooks, add them to your `.pre-commit-config.yaml` file. Below are example configurations for the `tofufmt` and `tofuvalidate` hooks.

### Example: `tofu-fmt`

Formats your OpenTofu configuration files to a canonical format and style.

```yaml
- repo: https://github.com/osinfra-io/pre-commit-hooks
 rev: <release-or-commit-sha>
 hooks:
  - id: tofu-fmt
   # Optional: pass additional args to tofu fmt
   # args: ["-diff"]
```

### Example: `tofu-validate`

Validates your OpenTofu configuration files for syntax and internal consistency.

```yaml
- repo: https://github.com/osinfra-io/pre-commit-hooks
 rev: <release-or-commit-sha>
 hooks:
  - id: tofu-validate
   # Optional: pass additional args to tofu validate
   # args: ["-no-color"]
```

Replace `<release-or-commit-sha>` with the desired version or commit hash.

For more details, see the `.pre-commit-hooks.yaml` in this repository.
