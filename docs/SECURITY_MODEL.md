# Term-Agent Security Model & Execution Boundaries

## Overview
Term-Agent implements defense-in-depth to protect the host system and repository from unauthorized operations, path traversal, and destructive shell execution.

## Risk Classification Strategy (`CommandRiskLevel`)

All proposed operations (tool calls and shell execution) are classified into three risk tiers:

1. **`SAFE` (Auto-Executable)**: Read-only operations inside workspace (e.g., `read_file`, `search_workspace`, `git status`, `go test`).
2. **`REQUIRES_USER` (Human Gate Required)**: Stateful modifications within workspace bounds (e.g., file staging, `go build`, package installs).
3. **`BLOCKED` (Prohibited)**: Destructive commands, system escapes, or out-of-boundary access (e.g., `rm -rf /`, `sudo`, credential reads, external network pipes).

## Non-Naive Shell Classification Requirements (Phase 6)

The security classifier DOES NOT rely solely on simple substring checks like `strings.Contains(cmd, "rm -rf")`. Naive substring checks fail against obfuscation, command substitution, and shell operators.

Phase 6 command parsing enforces full POSIX shell AST parsing (or lexing) to evaluate:
- **Command Substitutions**: `$(...)` and `` `...` ``
- **Subshells & Pipes**: `(...)`, `|`, `&&`, `||`
- **Redirection Operators**: `>`, `>>`, `<`
- **Variable Expansion**: `$VAR` expansion evaluation
- **Interpreters & Scripts**: `bash -c "..."`, `python -c "..."`, `eval`

### Application vs Host Sandboxing
Application-level classification provides baseline security. When complete OS-level isolation is required (e.g. untrusted third-party tool execution), term-agent will interface with container sandboxes (Docker/Podman) or OS sandboxing mechanisms (`landlock`, `seccomp`).

## Workspace Boundary & Symlink Protections

1. **Lexical Validation (`ValidateWorkspacePath`)**: Uses `filepath.Rel` to ensure paths do not escape `workspaceRoot`. Prevents prefix overlap attacks (e.g. `/workspace` vs `/workspace-secret`).
2. **Symlink Escape Protection**: In Phase 3/6, runtime file operations evaluate symlinks using `filepath.EvalSymlinks` to prevent existing symlinks inside the workspace from resolving to locations outside the root boundary.
