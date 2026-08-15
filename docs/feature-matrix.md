# PRD Requirement Traceability Matrix — `term-agent`

This document maintains full traceability between requirements specified in `PRD.md`, their targeted phase, assigned tasks, implementation files, test coverage, and current status.

## Status Legend
- `NOT_STARTED`: Requirement not yet implemented.
- `FOUNDATION_ONLY`: Interfaces, schemas, or placeholders established in Phase 0 foundation.
- `IN_PROGRESS`: Currently under active development.
- `IMPLEMENTED`: Implementation complete but pending full verification.
- `VERIFIED`: Complete, verified via unit/integration/security tests, static analysis, and race detector.
- `DEFERRED`: Planned for post-MVP / future phase.
- `BLOCKED`: Development currently blocked.

---

## 1. Functional Requirements (FR)

| Requirement | Phase | Task | Implementation | Tests | Status |
|---|---|---|---|---|---|
| **FR-01**: Local-First Architecture | PHASE_0 | TASK-0.1 | `internal/config/config.go`, `internal/persistence/database.go` | `tests/unit/config_test.go`, `tests/integration/foundation_test.go` | `VERIFIED` |
| **FR-02**: Multi-Provider Interface Abstraction | PHASE_0 | TASK-0.2 | `internal/model/provider.go`, `internal/model/registry.go` | `tests/unit/model_test.go` | `FOUNDATION_ONLY` |
| **FR-03**: Human-in-the-Loop Diff Approval Gate | PHASE_0 | TASK-0.3 | `internal/security/policy.go`, `internal/mutation/engine_impl.go` | `tests/unit/mutation_test.go` | `VERIFIED` |
| **FR-04**: Transactional Mutation Engine (11-State Machine) | PHASE_0 | TASK-0.4 | `internal/mutation/engine_impl.go`, `internal/mutation/transaction.go` | `tests/unit/mutation_test.go`, `tests/integration/mutation_integration_test.go` | `VERIFIED` |
| **FR-05**: Bounded Concurrent Scheduler (Default 5 Workers) | PHASE_0 | TASK-0.5 | `internal/scheduler/scheduler.go`, `internal/scheduler/dependency.go` | `tests/unit/scheduler_test.go` | `VERIFIED` |
| **FR-06**: Context Budgeting & Compaction | PHASE_0 | TASK-0.6 | `internal/context/manager.go`, `internal/context/tokenizer.go` | `tests/unit/context_test.go` | `FOUNDATION_ONLY` |
| **FR-07**: Session Persistence & Recovery | PHASE_1 | TASK-1.1 | `internal/app/lifecycle.go`, `internal/persistence/repository/` | `tests/integration/session_test.go` | `VERIFIED` |
| **FR-08**: Workspace Discovery & Metadata Extraction | PHASE_3 | TASK-3.1 | `internal/workspace/discovery.go` | `tests/unit/workspace_test.go` | `VERIFIED` |
| **FR-09**: Read-Only Workspace Search & Inspections | PHASE_3 | TASK-3.2 | `internal/tools/search.go`, `internal/tools/read.go`, `internal/tools/list.go` | `tests/unit/workspace_test.go` | `VERIFIED` |
| **FR-10**: Optimistic Concurrency Control (`before_hash`) | PHASE_4 | TASK-4.1 | `internal/mutation/engine_impl.go`, `internal/workspace/hashing.go` | `tests/unit/mutation_test.go`, `tests/integration/mutation_integration_test.go` | `VERIFIED` |
| **FR-11**: Unified Diff Generation & Rendering | PHASE_4 | TASK-4.2 | `internal/diff/engine.go`, `internal/diff/renderer.go` | `tests/unit/diff_test.go` | `VERIFIED` |
| **FR-12**: Single-Agent Execution Loop | PHASE_7 | TASK-7.1 | `internal/agent/orchestrator.go`, `internal/agent/worker.go` | `tests/integration/agent_test.go` | `VERIFIED` |
| **FR-13**: Hierarchical Multi-Agent Task Orchestration | PHASE_10 | TASK-10.1 | `internal/agent/planner.go`, `internal/agent/orchestrator.go` | `tests/integration/multi_agent_test.go` | `NOT_STARTED` |
| **FR-14**: OpenAI Provider Integration | PHASE_11 | TASK-11.1 | `internal/model/openai/client.go` | `tests/integration/openai_test.go` | `NOT_STARTED` |
| **FR-15**: Anthropic Provider Integration | PHASE_11 | TASK-11.2 | `internal/model/anthropic/client.go` | `tests/integration/anthropic_test.go` | `NOT_STARTED` |
| **FR-16**: Gemini Provider Integration | PHASE_11 | TASK-11.3 | `internal/model/gemini/client.go` | `tests/integration/gemini_test.go` | `NOT_STARTED` |
| **FR-17**: Real Research Source Tools Ingestion | PHASE_7 | TASK-7.2 | `internal/tools/research/` | `tests/unit/research_tools_test.go` | `VERIFIED` |
| **FR-18**: Research Provenance Tracking & DAG Decomposition | PHASE_8 | TASK-8.1 | `internal/workflows/research/provenance/`, `internal/workflows/research/agents/planner.go` | `tests/unit/research_tools_test.go` | `VERIFIED` |

---

## 2. Non-Functional Requirements (NFR)

| Requirement | Phase | Task | Implementation | Tests | Status |
|---|---|---|---|---|---|
| **NFR-01**: Cold Startup Latency (< 150ms) | PHASE_1 | TASK-1.2 | `cmd/termagent/main.go`, `internal/app/app.go` | `tests/integration/startup_test.go` | `VERIFIED` |
| **NFR-02**: Zero Uncontrolled Workspace Mutations | PHASE_4 | TASK-4.3 | `internal/mutation/engine_impl.go`, `internal/tools/write.go`, `internal/tools/edit.go` | `tests/unit/mutation_test.go`, `tests/integration/mutation_integration_test.go` | `VERIFIED` |
| **NFR-03**: Memory Bounded File Streaming | PHASE_3 | TASK-3.3 | `internal/tools/read.go`, `internal/workspace/reader.go` | `tests/unit/workspace_test.go` | `VERIFIED` |
| **NFR-04**: Decoupled Event-Driven UI Architecture | PHASE_0 | TASK-0.7 | `internal/events/bus.go`, `internal/events/memory_bus.go` | `tests/unit/eventbus_test.go` | `VERIFIED` |
| **NFR-05**: Thread-Safe Event Bus with Panic Recovery | PHASE_0 | TASK-0.8 | `internal/events/memory_bus.go` | `tests/unit/eventbus_test.go` | `VERIFIED` |

---

## 3. Security Requirements (SEC)

| Requirement | Phase | Task | Implementation | Tests | Status |
|---|---|---|---|---|---|
| **SEC-01**: Lexical Path Traversal Prevention | PHASE_0 | TASK-0.9 | `internal/security/path.go` | `tests/security/boundary_test.go` | `VERIFIED` |
| **SEC-02**: Runtime Symlink Escape Resolution | PHASE_3 | TASK-3.4 | `internal/security/path.go` (`filepath.EvalSymlinks`) | `tests/security/boundary_test.go` | `VERIFIED` |
| **SEC-03**: Non-Naive Shell Risk Classifier (POSIX AST) | PHASE_5 | TASK-5.1 | `internal/security/classifier.go` | `tests/unit/classifier_test.go`, `tests/security/boundary_test.go` | `VERIFIED` |
| **SEC-04**: Sensitive Path Check (`.env`, `~/.ssh/`) | PHASE_5 | TASK-5.1 | `internal/security/policy.go`, `internal/security/permissions.go` | `tests/security/policy_test.go` | `VERIFIED` |
| **SEC-05**: Command Timeout & Execution Limits | PHASE_5 | TASK-5.1 | `internal/tools/shell.go` | `tests/unit/shell_test.go` | `VERIFIED` |
| **SEC-06**: Secret Redaction in Logs & SQLite Messages | PHASE_1 | TASK-1.3 | `internal/config/logging.go`, `internal/persistence/writer.go` | `tests/unit/logger_test.go` | `VERIFIED` |
| **SEC-07**: Prompt Injection Defense (Data Isolation) | PHASE_6 | TASK-6.1 | `internal/security/isolation.go` | `tests/security/isolation_test.go` | `VERIFIED` |

---

## 4. Database & Persistence Requirements (DB)

| Requirement | Phase | Task | Implementation | Tests | Status |
|---|---|---|---|---|---|
| **DB-01**: SQLite Storage with WAL Mode & Foreign Keys | PHASE_0 | TASK-0.10 | `migrations/000001_initial_schema.sql`, `internal/persistence/database.go` | `tests/integration/foundation_test.go` | `VERIFIED` |
| **DB-02**: 11-Table Foundational Schema DDL | PHASE_0 | TASK-0.11 | `migrations/000001_initial_schema.sql` | `tests/integration/foundation_test.go` | `VERIFIED` |
| **DB-03**: Serialized Async Writer Channel (`AsyncWriter`) | PHASE_1 | TASK-1.4 | `internal/persistence/writer.go` | `tests/integration/writer_test.go` | `VERIFIED` |
| **DB-04**: Session State & Pending Mutation Crash Recovery | PHASE_4 | TASK-4.4 | `internal/mutation/rollback.go`, `internal/persistence/repository/` | `tests/integration/recovery_test.go` | `NOT_STARTED` |

---

## 5. CLI & TUI Requirements (CLI/UI)

| Requirement | Phase | Task | Implementation | Tests | Status |
|---|---|---|---|---|---|
| **CLI-01**: Layered Configuration Precedence (CLI > Env > File > Defaults) | PHASE_0 | TASK-0.12 | `internal/config/config.go` | `tests/unit/config_test.go` | `VERIFIED` |
| **CLI-02**: Standard Flag Set (`--workspace`, `--model`, `--session`, `--dry-run`) | PHASE_1 | TASK-1.5 | `cmd/termagent/main.go` | `tests/unit/cli_test.go` | `VERIFIED` |
| **UI-01**: Bubble Tea Elm Architecture Separation | PHASE_2 | TASK-2.1 | `internal/tui/model.go`, `internal/tui/update.go` | `tests/unit/tui_test.go` | `NOT_STARTED` |
| **UI-02**: Lip Gloss Styling & Theme Tokens | PHASE_0 | TASK-0.13 | `internal/tui/styles/styles.go` | `tests/unit/tui_test.go` | `FOUNDATION_ONLY` |
| **UI-03**: Vim Keybinding Mappings | PHASE_0 | TASK-0.14 | `internal/tui/keymap/keymap.go` | `tests/unit/tui_test.go` | `FOUNDATION_ONLY` |
| **UI-04**: Interactive Unified Diff Review View | PHASE_2 | TASK-2.2 | `internal/tui/views/diff.go` | `tests/unit/tui_diff_test.go` | `NOT_STARTED` |
| **UI-05**: Dynamic Resizing & Terminal Layout Adaptation | PHASE_2 | TASK-2.3 | `internal/tui/update.go` | `tests/unit/tui_resize_test.go` | `NOT_STARTED` |
