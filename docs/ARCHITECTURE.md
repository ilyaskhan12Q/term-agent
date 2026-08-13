# Term-Agent Architecture & Design Principles

## Overview
Term-Agent is a local-first, terminal-native AI coding agent built with Go. Its primary architectural directive is:

> **"The model proposes. The runtime decides. The user approves. The transaction commits."**

## Subsystem Architecture & Dependency Hierarchy

To ensure security, structural integrity, and maintainability, subsystems are strictly layered in a one-way dependency flow:

```
TUI (Bubble Tea)
   │
   ▼
Application (Lifecycle / Config)
   │
   ▼
Agent Runtime & Scheduler (DAG / Worker Pool)
   │
   ▼
Tool Runtime & Security Engine (Policy / Path / Classifier)
   │
   ▼
Mutation Engine & Workspace (Locks / SHA-256 / File I/O)
   │
   ▼
Persistence (SQLite WAL / Repositories)
```

### Enforced Architectural Boundaries
1. **TUI Isolation**: The UI layer receives system events and emits user commands. It MUST NOT directly access SQLite repositories, execute shell commands, or write to the filesystem.
2. **Agent Isolation**: Agents propose tool calls and file edits. They CANNOT perform direct file I/O or shell execution. All requests must pass through the Security Classifier and Mutation Engine.
3. **Model Provider Isolation**: LLM provider implementations (OpenAI, Anthropic, Gemini) stream raw responses and tool invocations. They have zero knowledge of Bubble Tea, TUI models, or local filesystems.
4. **Mutation Engine Primacy**: All file edits (create, modify, delete) MUST go through the Mutation Engine via two-phase optimistic transactions.
5. **Persistence Funnel**: Direct database writes from worker goroutines are funneled through an asynchronous persistence channel (`AsyncWriter`) to prevent SQLite write lock contention.

## Core Component Contracts

### 1. Agent & Scheduler
- **`Agent`**: Evaluates model context and returns step decisions (`StepResult`).
- **`Scheduler`**: Bounded parallel worker pool executing task DAGs (default max concurrency = 5). Enforces cycle validation before task dispatch.

### 2. Security & Workspace
- **`ValidateWorkspacePath`**: Ensures target paths reside strictly inside the configured workspace root via lexical path normalization (`filepath.Rel`). (Full symlink resolution via `filepath.EvalSymlinks` is evaluated at runtime in Phase 3/6).
- **`LockManager`**: Fine-grained path mutexes preventing concurrent write collisions across task workers.

### 3. Mutation Engine State Machine
Transactions strictly follow an 11-stage state flow:
`PROPOSED` → `VALIDATED` → `WAITING_APPROVAL` → `APPROVED` → `COMMITTING` → `COMMITTED` (or `REJECTED`, `ROLLING_BACK`, `ROLLED_BACK`, `FAILED`, `CONFLICT`).

Optimistic concurrency validation checks `before_hash` against the live workspace file hash prior to committing.

### 4. Context & Tokenizer
- **`SimpleEstimator`**: A baseline heuristic estimator (~4 characters per token) used for initial foundation sizing. Provider-specific tokenizers (tiktoken, Anthropic count_tokens) will plug in during Phase 5.
