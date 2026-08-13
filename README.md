# term-agent — Local-First Terminal AI Coding Agent

`term-agent` is a production-grade, local-first terminal AI coding agent built in Go. It empowers software engineers to inspect workspaces, generate execution plans, propose atomic code mutations, visually review diffs, execute policy-controlled shell tasks, and maintain full transactional control over codebase modifications.

> **Core Architectural Principle:**
> *The model proposes. The runtime decides. The user approves. The transaction commits.*

---

## 📚 Project Documentation & Governance

- **[PRD.md](PRD.md)** — Production-Grade Product Requirements Document (Source of Truth).
- **[Requirement Traceability Matrix](docs/feature-matrix.md)** — Complete requirement mapping (FR, NFR, SEC, DB, CLI) with task and test coverage statuses.
- **[Architecture Specification](docs/ARCHITECTURE.md)** — Subsystem layers, structural rules, and one-way dependency boundaries.
- **[Security Model](docs/SECURITY_MODEL.md)** — Workspace isolation, path traversal defense, and non-naive shell risk classification.
- **[Agent Execution Lifecycle](docs/AGENT_LIFECYCLE.md)** — 6-stage runtime execution pipeline.
- **[Phase 0 Audit Report](docs/PHASE_0_AUDIT.md)** — Verified engineering baseline and readiness analysis.
- **Architectural Decision Records**:
  - [ADR 0001: SQLite Storage with WAL Mode](docs/adr/0001-sqlite-storage.md)
  - [ADR 0002: Bubble Tea Framework for Terminal UI](docs/adr/0002-bubbletea-tui.md)
  - [ADR 0003: Transactional Mutation Engine](docs/adr/0003-mutation-engine-safety.md)

---

## 🛠️ Technology Stack

| Layer | Technology |
|---|---|
| **Language** | Go (`1.24+`) |
| **Terminal UI** | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| **Local Database** | Embedded SQLite (`modernc.org/sqlite`) with WAL mode & foreign key enforcement |
| **Configuration** | TOML (`pelletier/go-toml/v2`) + Environment variables + CLI flags |
| **CI / Security** | GitHub Actions (`ci.yml`, `security.yml`) |

---

## 🧪 Development & Quality Gates

Run the verification suite before committing any change:

```bash
# 1. Format code
gofmt -w .

# 2. Static analysis
go vet ./...

# 3. Unit and integration tests
go test -v ./...

# 4. Data race detector
go test -race -v ./...

# 5. Build application
go build ./...
```

---

## 📋 Feature Roadmap & Phase Status

| Phase | Description | Status |
|---|---|---|
| **Phase 0** | Project Skeleton & Architecture Foundation | `PHASE_1_READY_WITH_WARNINGS` |
| **Phase 1** | Application Lifecycle & CLI Setup | `NOT_STARTED` |
| **Phase 2** | Bubble Tea Terminal UI | `NOT_STARTED` |
| **Phase 3** | Workspace Discovery & Engine | `NOT_STARTED` |
| **Phase 4** | Mutation Engine & Optimistic OCC | `NOT_STARTED` |
| **Phase 5** | Tool Runtime & Shell Sandbox | `NOT_STARTED` |
| **Phase 6** | Security Policy Engine & AST Classifier | `NOT_STARTED` |
| **Phase 7** | Single Agent Execution Loop | `NOT_STARTED` |
| **Phase 8** | Context Budgeting & Compaction | `NOT_STARTED` |
| **Phase 9** | Scheduler & Bounded Worker Pool | `NOT_STARTED` |
| **Phase 10** | Hierarchical Multi-Agent System | `NOT_STARTED` |
| **Phase 11** | Multi-Provider Engine (OpenAI/Anthropic/Gemini) | `NOT_STARTED` |
| **Phase 12** | Production Hardening | `NOT_STARTED` |

---

## 📜 License

Distributed under the MIT License.
