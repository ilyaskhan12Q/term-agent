# OpenAI Model Provider Package

This directory is reserved for the OpenAI `ModelProvider` implementation (Phase 7).

Responsibilities:
- Transform term-agent `CompletionRequest` into OpenAI Chat Completion API payloads.
- Handle OpenAI tool calling schema and streaming response parsing.
- Enforce API key configuration via environment variable (`OPENAI_API_KEY`).
