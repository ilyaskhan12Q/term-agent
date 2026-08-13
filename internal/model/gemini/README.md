# Gemini Model Provider Package

This directory is reserved for the Google Gemini `ModelProvider` implementation (Phase 7).

Responsibilities:
- Transform term-agent `CompletionRequest` into Gemini REST/gRPC API payloads.
- Handle Gemini Function Calling declarations and response parsing.
- Enforce API key configuration via environment variable (`GEMINI_API_KEY`).
