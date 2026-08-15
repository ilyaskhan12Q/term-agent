# term-agent — Multi-Agent Orchestration Engine & Research Agent Platform

`term-agent` is a production-grade, local-first multi-agent orchestration engine built in Go. It provides high-performance task decomposition, Directed Acyclic Graph (DAG) plan scheduling, parallel worker execution, model provider abstractions, context memory budgeting, persistent SQLite event tracking, and an interactive Bubble Tea terminal user interface.

On top of this core orchestration engine, `term-agent` supports specialized vertical domain workflows. Its primary flagship workflow is **Research Agent Mode** — an automated deep research system with literature discovery, evidence extraction, citation verification, and academic paper synthesis.

> **Core Architectural Principle:**
> *The model proposes. The engine orchestrates. The user approves. The transaction commits.*

---

## Project Documentation & Governance

- **[CHANGELOG](CHANGELOG.md)** — Version history and release notes following Keep a Changelog.
- **[SECURITY](SECURITY.md)** — Security policy, threat model, and prompt injection defense.
- **[CONTRIBUTING](CONTRIBUTING.md)** — Guidelines for contributing and development workflows.
- **[CODE_OF_CONDUCT](CODE_OF_CONDUCT.md)** — Standard contributor code of conduct.
- **[COMMIT_GUIDE](COMMIT_GUIDE.md)** — Conventional commit standards and rules.
- **[Product Direction](docs/PRODUCT_DIRECTION.md)** — Architectural directives and product direction.
- **[Research Mode Blueprint](docs/RESEARCH_WORKFLOW_ARCHITECTURAL_AUDIT.md)** — Architecture specification for Research Agent Mode.
- **[PRD](PRD.md)** — Product Requirements Document (Source of Truth).
- **[Requirement Traceability Matrix](docs/feature-matrix.md)** — Complete requirement mapping (FR, NFR, SEC, DB, CLI) with task and test coverage statuses.
- **[Architecture Specification](docs/ARCHITECTURE.md)** — Subsystem layers and one-way dependency boundaries.
- **[Security Model](docs/SECURITY_MODEL.md)** — Workspace isolation, path traversal defense, and POSIX AST shell risk classification.

---

## Technology Stack

| Layer | Technology |
|---|---|
| **Language** | Go (`1.24+`) |
| **Terminal UI** | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| **Local Database** | Embedded SQLite (`modernc.org/sqlite`) with WAL mode & foreign key enforcement |
| **Configuration** | TOML (`pelletier/go-toml/v2`) + Environment variables + CLI flags |
| **CI / Security** | GitHub Actions (`ci.yml`, `security.yml`) |

---

## Command Line Interface (CLI)

```bash
# Build the binary
go build -o term-agent ./cmd/term-agent

# Execute default terminal agent mode
./term-agent --workspace ./path/to/project

# Execute specialized Research Agent mode
./term-agent --workflow research --prompt "Investigate Transformer Architecture Memory Scaling"
```

---

## Development & Quality Gates

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

## Implementation Phase Status

| Phase | Component Description | Status |
|---|---|---|
| **Phase 0** | Project Skeleton & Architecture Foundation | `VERIFIED` |
| **Phase 1** | Application Lifecycle & CLI Setup | `VERIFIED` |
| **Phase 2** | Bubble Tea Terminal UI | `VERIFIED` |
| **Phase 3** | Workspace Discovery & Engine | `VERIFIED` |
| **Phase 4** | Mutation Engine & Optimistic OCC | `VERIFIED` |
| **Phase 5** | Tool Runtime & Shell Sandbox | `VERIFIED` |
| **Phase 6** | Security Policy Engine & AST Classifier | `VERIFIED` |
| **Phase 7** | Real Research & Source Tools (arXiv, Web, PDF, Citation Verifier) | `VERIFIED` |
| **Phase 8** | Provenance Tracking & Task Decomposition (Planner DAG, Deduplication) | `VERIFIED` |
| **Phase 9** | Parallel Specialist Researchers | `IN_PROGRESS` |
| **Phase 10** | Evidence & Citation Verification | `NOT_STARTED` |
| **Phase 11** | Research Synthesis | `NOT_STARTED` |
| **Phase 12** | Global Paper Templates & Writer | `NOT_STARTED` |
| **Phase 13** | Reviewer & Hallucination/Claim Detection | `NOT_STARTED` |
| **Phase 14** | Full End-to-End Research Mode | `NOT_STARTED` |

---

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.