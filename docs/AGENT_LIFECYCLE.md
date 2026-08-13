# Term-Agent Runtime Lifecycle

## Overview
The Term-Agent runtime manages prompt execution, task planning, tool evaluation, human approval, and state persistence.

## 6-Stage Execution Pipeline

```
1. Model Proposal
   ↓
2. Security Classification & Validation
   ↓
3. Human Approval Gate (TUI Overlay)
   ↓
4. Transaction Staging (Mutation Engine / Subprocess)
   ↓
5. Optimistic Concurrency Check & Commit
   ↓
6. Asynchronous Event & Persistence Logging
```

### Stage 1: Model Proposal
The Orchestrator sends prompt context to the configured LLM provider. The model returns structured tool invocation proposals or text responses.

### Stage 2: Security Classification & Validation
The proposed action passes through `SecurityClassifier`:
- Path boundaries are validated via `ValidateWorkspacePath`.
- Shell commands are evaluated for risk (`SAFE`, `REQUIRES_USER`, `BLOCKED`).

### Stage 3: Human Approval Gate
If the operation risk is `REQUIRES_USER`:
- Execution pauses.
- The TUI renders an interactive diff or command preview modal.
- User selects `[a] Approve` or `[r] Reject`.

### Stage 4: Transaction Staging
- File edits are staged in a `Transaction` via `MutationEngine.StageMutation`.
- Original file contents are backed up into a `FileSnapshot`.

### Stage 5: Optimistic Concurrency & Commit
Before writing changes to disk:
- `MutationEngine` recalculates live file hashes.
- If live hash matches `before_hash`, transaction is committed atomically.
- If live hash differs, state transitions to `CONFLICT` and rolls back.

### Stage 6: Persistence & Event Broadcast
- Operation results emit an `Event` over `EventBus`.
- History and usage metrics are enqueued to `AsyncWriter` for SQLite persistence.
