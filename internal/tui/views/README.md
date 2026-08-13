# TUI Views Package

This directory is reserved for primary view screens in Bubble Tea (Phase 2):
- `agent_view.go`: Live agent thoughts, streaming output, and tool call status.
- `diff_view.go`: Side-by-side or unified interactive diff view for transaction approval.
- `plan_view.go`: Task DAG tree visualization and progress tracker.
- `log_view.go`: Event bus log inspector.

CRITICAL RULE:
TUI views are strictly presentation components. They MUST NOT execute shell commands, modify files, call SQLite directly, or invoke model providers.
