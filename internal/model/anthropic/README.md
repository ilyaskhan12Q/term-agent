# Anthropic Model Provider Package

This directory is reserved for the Anthropic `ModelProvider` implementation (Phase 7).

Responsibilities:
- Transform term-agent `CompletionRequest` into Anthropic Messages API payloads.
- Handle Claude tool use JSON schema formatting and system prompt rules.
- Enforce API key configuration via environment variable (`ANTHROPIC_API_KEY`).
