# Security Policy

## Overview

`term-agent` is designed with a defense-in-depth security model to ensure that autonomous terminal agent executions, external research retrievals, and workspace mutations operate within strictly bounded, verifiable safety guarantees.

## Threat Model

The primary security boundaries enforced by `term-agent` include:

1. **Untrusted External Data & Prompt Injection**: External web pages, arXiv PDFs, and search result content fetched during research workflows are treated as untrusted data inputs that could contain adversarial prompt injection instructions.
2. **Workspace Isolation**: Agents are strictly restricted to the authorized target workspace directory. Path traversal escapes or symlink redirection outside the workspace boundary are rejected.
3. **Command Execution Safeguards**: Commands executed via the terminal interface are parsed into POSIX AST trees, categorized by risk level, and gated by explicit human approval policies.
4. **Data Integrity & Concurrency**: Workspace modifications use 11-state transactional mutation control and optimistic concurrency control (OCC) to prevent unverified overwrites.
5. **Secret Redaction**: API tokens, private keys, and sensitive environment patterns are sanitized from application logs and persistent databases.

## Security Controls

### 1. Untrusted Content Envelopes (Prompt Injection Defense)

All external content fetched from academic feeds, web HTTP requests, or local file extracts is processed through `internal/security/isolation.go`.

- Control characters and known prompt injection signatures are stripped.
- Raw text is wrapped within an immutable XML safety envelope:
  `<untrusted_content source="..." hash="...">...</untrusted_content>`
- A SHA-256 hash digest is attached to each envelope to maintain integrity and prevent tampering during LLM context assembly.

### 2. Workspace Boundary Protection

- **Lexical Path Resolution**: All target file paths are sanitized via `filepath.Clean` and checked against the workspace root using `filepath.Rel`.
- **Runtime Symlink Evaluation**: Real file paths are evaluated at runtime via `filepath.EvalSymlinks` to prevent symlink traversal escapes to arbitrary filesystem paths.

### 3. POSIX AST Shell Command Risk Classifier

- Command strings executed by the agent runtime are parsed into Abstract Syntax Trees using `mvdan.cc/sh/v3`.
- Commands are classified into risk categories (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`) based on dangerous utility invocation, pipe chains, output redirections, and privilege escalation flags.
- High-risk operations require explicit human confirmation.

### 4. Secret Sanitization

- Sensitive patterns (e.g. `*_API_KEY`, `*_SECRET`, `PRIVATE_KEY`, tokens) are automatically redacted in application logs and SQLite message stores.

## Reporting a Vulnerability

If you discover a security vulnerability in `term-agent`, please report it responsibly:

- Email details to `security@term-agent.org` or open a confidential report.
- Include a summary of the vulnerability, steps to reproduce, and affected modules.
- Please do not open public GitHub issues for undisclosed security vulnerabilities.

Security patches will be published promptly alongside advisory notices in `CHANGELOG.md`.
