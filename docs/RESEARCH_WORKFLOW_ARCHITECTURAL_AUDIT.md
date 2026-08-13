# term-agent — Architectural Audit & Research Workflow Blueprint

> **System Evolution Document**  
> **Status:** PROPOSED & APPROVED FOR IMPLEMENTATION  
> **Date:** August 13, 2026  
> **Target Release:** Phase 2 Extensibility Update

---

## 1. Executive Summary

This document presents the **Architectural Audit** of `term-agent` and details the strategy to evolve the system from a single-purpose coding runtime into an **Extensible Multi-Agent Platform**. 

The core `term-agent` engine will remain the foundation—providing task orchestration, model provider abstraction, scheduling, tool execution, safety boundaries, event tracking, persistence, and terminal UI (TUI). On top of this core runtime, we introduce a **Specialized Workflow Architecture (`internal/workflow`)**, with **Research Agent Mode** established as the first specialized workflow.

---

## 2. Repository Component Audit

A comprehensive code audit was performed across all packages in `internal/` and `tests/`. Each component has been audited and classified according to its implementation completeness.

| Subsystem | Package / File | Classification | Current State & Assessment | Strategy for Research Workflow Evolution |
|---|---|---|---|---|
| **App Lifecycle** | `internal/app/app.go`, `lifecycle.go` | `IMPLEMENTED` | Manages app startup, signal catching (`SIGINT`/`SIGTERM`), cold-start orphan session recovery (`INTERRUPTED`). | Update CLI flags to accept `--mode=coding\|research` and load the appropriate workflow. |
| **Config & Secrets** | `internal/config/config.go`, `logging.go` | `IMPLEMENTED` | TOML config, CLI flag overrides, secrets redactor in `slog.Handler`. | Add `[research]` config section for default templates, search API keys, and concurrency bounds. |
| **Persistence** | `internal/persistence/database.go`, `writer.go`, `repository/` | `IMPLEMENTED` | SQLite WAL driver, initial 11-table DDL schema, non-blocking `ChannelAsyncWriter`, session repo. | Add `000002_research_workflow.sql` migration for research projects, findings, claims, evidence, and papers. |
| **Event Bus** | `internal/events/memory_bus.go`, `event.go` | `IMPLEMENTED` | Thread-safe synchronized pub/sub event bus with typed payload support. | Emit research pipeline events (`EventResearchQuestionDecomposed`, `EventFindingCollected`, `EventPaperDrafted`). |
| **Security Jail** | `internal/security/path.go` | `IMPLEMENTED` | Lexical path boundary check (`filepath.Rel`) enforcing workspace isolation. | Reuse security boundary to ensure research outputs/notes stay strictly inside workspace directory. |
| **Security Policy & Classifier** | `internal/security/permissions.go`, `classifier.go` | `INTERFACE_ONLY` | Risk classification and policy interfaces defined. | Define read-only tool permissions for research agents (search, paper reading, web scrape). |
| **Agent Interface** | `internal/agent/agent.go`, `orchestrator.go`, `planner.go` | `INTERFACE_ONLY` | Defines `Agent`, `Orchestrator`, `Planner`, `Worker`, `StepResult`, `TaskSpec`, `Plan`. | Retain `Agent` and `Orchestrator` contracts. Specialize research agents as implementations of `Agent`. |
| **Scheduler & DAG** | `internal/scheduler/dependency.go` | `IMPLEMENTED` | Cycle detection via Kahn’s topological sort and `ReadyTasks()` determination. | Direct reuse! Research task graphs (RQ1, RQ2 $\rightarrow$ Synthesis $\rightarrow$ Review) will run on `DependencyGraph`. |
| **Scheduler Pool** | `internal/scheduler/scheduler.go`, `worker_pool.go` | `INTERFACE_ONLY` | Interfaces for bounded parallel execution worker pool. | Implement bounded worker pool to execute independent research agents in parallel. |
| **Model Abstraction** | `internal/model/provider.go`, `registry.go` | `PARTIALLY_IMPLEMENTED` | `ModelProvider` interface & `ProviderRegistry` implemented. Provider drivers (OpenAI, Anthropic, Gemini) pending. | Implement LLM drivers so research agents can make structured completion requests. |
| **Tool Engine** | `internal/tools/tool.go`, `registry.go` | `PARTIALLY_IMPLEMENTED` | `Tool` contract and `Registry` implemented. Standard file/shell tool skeletons present. | Add research-specific tools (e.g. `academic_search`, `pdf_reader`, `web_fetch`, `citation_extractor`). |
| **Mutation Engine** | `internal/mutation/transaction.go` | `INTERFACE_ONLY` | Two-phase transaction engine interfaces for file edits. | Research paper generation writes final markdown/LaTeX files using formal mutation transactions. |
| **TUI Layer** | `internal/tui/model.go` | `PARTIALLY_IMPLEMENTED` | Initial Bubble Tea model skeleton. | Extend TUI to render Research Project Progress, DAG execution graphs, and Evidence Verification drawers. |

---

## 3. Extensible Workflow Architecture

To support specialized workflows without contaminating the core agent engine, we introduce the **Workflow Subsystem (`internal/workflow`)**.

```
                               ┌───────────────────────────┐
                               │     TUI / CLI Layer       │
                               └─────────────┬─────────────┘
                                             │
                                             ▼
                               ┌───────────────────────────┐
                               │      Workflow Engine      │
                               │(internal/workflow/engine) │
                               └───────┬───────────┬───────┘
                                       │           │
                     ┌─────────────────┘           └─────────────────┐
                     ▼                                               ▼
     ┌───────────────────────────────┐               ┌───────────────────────────────┐
     │   Coding Workflow (Default)   │               │   Research Workflow (Mode 1)  │
     │ (internal/workflows/coding)   │               │ (internal/workflows/research) │
     └───────────────┬───────────────┘               └───────────────┬───────────────┘
                     │                                               │
                     └───────────────────────┬───────────────────────┘
                                             │ (Reuses Core Infrastructure)
                                             ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                              Core Multi-Agent Runtime                                  │
├───────────────────────┬─────────────────────────┬───────────────────────┬──────────────┤
│  Agent Orchestration  │ Scheduler & DAG Engine  │ Model Provider Layer  │ Tools System │
│   (internal/agent)    │  (internal/scheduler)   │   (internal/model)    │(internal/tool)│
├───────────────────────┼─────────────────────────┼───────────────────────┼──────────────┤
│  Security & Jail      │ Mutation & Workspace    │ Events Bus            │ Persistence  │
│  (internal/security)  │ (internal/mutation)     │  (internal/events)    │(internal/per)│
└───────────────────────┴─────────────────────────┴───────────────────────┴──────────────┘
```

### 3.1 `Workflow` Core Contract (`internal/workflow/workflow.go`)

```go
package workflow

import (
	"context"
	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/events"
)

type WorkflowType string

const (
	WorkflowTypeCoding   WorkflowType = "CODING"
	WorkflowTypeResearch WorkflowType = "RESEARCH"
)

// Workflow defines the execution contract for any specialized agent mode.
type Workflow interface {
	Name() string
	Type() WorkflowType
	Initialize(ctx context.Context, input string) error
	BuildPlan(ctx context.Context) (*agent.Plan, error)
	Execute(ctx context.Context, bus events.EventBus) (*WorkflowResult, error)
	Status() WorkflowStatus
}

type WorkflowStatus string

type WorkflowResult struct {
	ID        string
	Output    string
	Data      interface{} // Specialized structured payload
	Artifacts []string
	Error     error
}
```

---

## 4. Research Workflow Architecture

### 4.1 End-to-End Pipeline

```
USER INPUT (Topic / Question / Paper / Objective)
    │
    ▼
RESEARCH ORCHESTRATOR
    │
    ▼
UNDERSTAND RESEARCH OBJECTIVE
    │
    ▼
DECOMPOSE INTO RESEARCH QUESTIONS (RQ1, RQ2, RQ3...)
    │
    ▼
BUILD RESEARCH TASK DAG (scheduler.DependencyGraph)
    │
    ▼
CREATE SPECIALIZED RESEARCH AGENTS (Literature, Evidence, Methodology, Comparative)
    │
    ▼
RUN IN PARALLEL WHERE POSSIBLE (scheduler.WorkerPool)
    │
    ▼
COLLECT STRUCTURED FINDINGS (ResearchFinding struct)
    │
    ▼
CRITICAL REVIEWER (Contradiction Analysis, Evidence Gaps, Unsupported Claims)
    │
    ▼
SYNTHESIZER (Theme Aggregation & Provenance Graph)
    │
    ▼
GLOBAL RESEARCH TEMPLATE (Academic, Literature Review, Technical Report, etc.)
    │
    ▼
PAPER WRITER (Drafts Document Sections based strictly on Verified Evidence)
    │
    ▼
CITATION / FACT / STRUCTURE REVIEWER (Verification Gate)
    │
    ▼
FINAL RESEARCH PAPER (.md / .tex / .pdf)
```

---

### 4.2 Mandatory Structural Findings Separation Rule

To guarantee academic rigor and eliminate hallucinated prose, **research agents MUST NOT directly output unformatted prose**. Instead, every research agent returns a strongly typed `ResearchFinding` structure:

```go
package domain

import "time"

// ResearchFinding represents the mandatory structured output of any research agent step.
type ResearchFinding struct {
	ID                 string      `json:"id"`
	ResearchProjectID  string      `json:"research_project_id"`
	ResearchQuestionID string      `json:"research_question_id"`
	TaskID             string      `json:"task_id"`
	AgentID            string      `json:"agent_id"`
	AgentType          string      `json:"agent_type"`
	Scope              string      `json:"scope"`
	Findings           []string    `json:"findings"`
	Evidence           []Evidence  `json:"evidence"`
	Sources            []Source    `json:"sources"`
	Claims             []Claim     `json:"claims"`
	Limitations        []string    `json:"limitations"`
	Confidence         float64     `json:"confidence"` // 0.0 to 1.0
	CreatedAt          time.Time   `json:"created_at"`
}
```

---

### 4.3 Complete Source Provenance Model

Every statement in the generated paper MUST be traceable back to verified evidence:

$$\text{Claim} \longrightarrow \text{Evidence} \longrightarrow \text{Source} \longrightarrow \text{Research Agent} \longrightarrow \text{Research Question} \longrightarrow \text{Paper Section}$$

#### Domain Entities (`internal/workflows/research/domain/`)

```go
package domain

import "time"

type SourceType string

const (
	SourceTypeAcademicPaper SourceType = "ACADEMIC_PAPER"
	SourceTypeDocumentation SourceType = "DOCUMENTATION"
	SourceTypeWebPage       SourceType = "WEB_PAGE"
	SourceTypeDataset       SourceType = "DATASET"
	SourceTypeBook          SourceType = "BOOK"
)

type Source struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	URI         string     `json:"uri"`
	Authors     []string   `json:"authors"`
	Year        int        `json:"year"`
	Publisher   string     `json:"publisher"`
	SourceType  SourceType `json:"source_type"`
	TrustScore  float64    `json:"trust_score"` // 0.0 to 1.0
	FetchedAt   time.Time  `json:"fetched_at"`
}

type EvidenceVerification string

const (
	EvidenceStatusVerified   EvidenceVerification = "VERIFIED"
	EvidenceStatusUnverified EvidenceVerification = "UNVERIFIED"
	EvidenceStatusMismatch   EvidenceVerification = "MISMATCH"
)

type Evidence struct {
	ID                 string               `json:"id"`
	SourceID           string               `json:"source_id"`
	Snippet            string               `json:"snippet"`
	Location           string               `json:"location"` // e.g. Page 4, Section 3.2, Line 120
	VerificationStatus EvidenceVerification `json:"verification_status"`
	ExtractorAgentID   string               `json:"extractor_agent_id"`
}

type ClaimStrength string

const (
	ClaimDirect       ClaimStrength = "DIRECT"
	ClaimInferential  ClaimStrength = "INFERENTIAL"
	ClaimSpeculative  ClaimStrength = "SPECULATIVE"
)

type Claim struct {
	ID          string        `json:"id"`
	Statement   string        `json:"statement"`
	EvidenceIDs []string      `json:"evidence_ids"`
	Strength    ClaimStrength `json:"strength"`
}
```

---

### 4.4 Specialized Research Agent Suite

`term-agent` will support 10 specialized agent roles operating under the orchestrator:

1. **Research Planner Agent:** Understands the research objective, generates sub-questions ($RQ_1 \dots RQ_n$), and builds the task dependency graph.
2. **Literature Research Agent:** Investigates existing papers, documentation, repositories, and published literature.
3. **Evidence Agent:** Extracts exact quotes, snippets, data points, and builds evidence mappings against claims.
4. **Methodology Agent:** Analyzes methodologies, experimental protocols, algorithm steps, or study designs.
5. **Comparative Agent:** Compares competing frameworks, algorithm performance, or conflicting paper claims.
6. **Data/Statistics Agent:** Parses quantitative datasets, tables, benchmark metrics, and statistical measures.
7. **Critical Reviewer Agent:** Performs adversarial checks for unsupported claims, weak evidence, missing limitations, or contradictions.
8. **Synthesizer Agent:** Merges findings across all agents, deduplicates claims, and constructs global thematic clusters.
9. **Paper Writer Agent:** Generates paper section drafts strictly adhering to synthesized findings and the selected Global Research Template.
10. **Citation/Quality Reviewer Agent:** Verifies citation links, section ordering, formatting constraints, and ensures no hallucinated references exist.

---

### 4.5 Global Research Template System

The research paper structure is decoupled from agent logic via JSON/YAML template definitions (`internal/workflows/research/templates/`).

#### Supported Template Types:
- **Academic Research Paper** (Abstract, Intro, Problem, Literature Review, Methodology, Results, Discussion, Limitations, Conclusion, References)
- **Literature Review** (Abstract, Background, Scope, Categorized Themes, Comparative Analysis, Open Gaps, Conclusion, References)
- **Systematic Review** (Abstract, Protocol, Search Strategy, Included Studies, Quality Assessment, Data Synthesis, Implications, References)
- **Technical Research Report** (Executive Summary, Technical Context, Architecture, Evaluation, Risk Matrix, Recommendations)
- **Comparative Research** (Overview, Evaluation Criteria, Option Analysis, Benchmarks, Trade-off Matrix, Conclusion)
- **Thesis-Style Document**

#### Example Template Schema (`academic_research.json`):

```json
{
  "template_id": "academic_research",
  "name": "Academic Research Paper",
  "version": "1.0",
  "sections": [
    {
      "id": "title",
      "name": "Title & Author Block",
      "required": true,
      "max_words": 30,
      "requires_citations": false
    },
    {
      "id": "abstract",
      "name": "Abstract",
      "required": true,
      "max_words": 300,
      "requires_citations": false
    },
    {
      "id": "introduction",
      "name": "Introduction",
      "required": true,
      "requires_citations": true,
      "min_citations": 2
    },
    {
      "id": "literature_review",
      "name": "Literature Review",
      "required": true,
      "requires_citations": true,
      "min_sources": 3
    },
    {
      "id": "methodology",
      "name": "Methodology",
      "required": true,
      "requires_citations": false
    },
    {
      "id": "results",
      "name": "Results & Analysis",
      "required": true,
      "requires_evidence": true
    },
    {
      "id": "discussion",
      "name": "Discussion & Contradiction Analysis",
      "required": true,
      "requires_citations": true
    },
    {
      "id": "limitations",
      "name": "Limitations & Gaps",
      "required": true,
      "requires_citations": false
    },
    {
      "id": "conclusion",
      "name": "Conclusion",
      "required": true,
      "requires_citations": false
    },
    {
      "id": "references",
      "name": "References",
      "required": true,
      "validation": "STRICT_PROVENANCE_CHECK"
    }
  ]
}
```

---

## 5. Database Schema Extensions

To persist research workflows across terminal restarts, migration `000002_research_workflow.sql` will add 8 specialized tables:

```sql
-- Migration 000002: Research Workflow Persistence Schema

-- 1. Research Projects Table
CREATE TABLE IF NOT EXISTS research_projects (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    title TEXT NOT NULL,
    objective TEXT NOT NULL,
    template_id TEXT NOT NULL DEFAULT 'academic_research',
    status TEXT NOT NULL DEFAULT 'CREATED', -- CREATED, PLANNING, EXECUTING, SYNTHESIZING, DRAFTING, REVIEWING, COMPLETED, FAILED
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- 2. Research Questions Table
CREATE TABLE IF NOT EXISTS research_questions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    question TEXT NOT NULL,
    scope TEXT NOT NULL,
    priority INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'PENDING',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE
);

-- 3. Research Sources Table
CREATE TABLE IF NOT EXISTS research_sources (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    uri TEXT NOT NULL,
    authors TEXT NOT NULL, -- JSON array of strings
    year INT NOT NULL DEFAULT 0,
    publisher TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT 'ACADEMIC_PAPER',
    trust_score REAL NOT NULL DEFAULT 1.0,
    fetched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE
);

-- 4. Research Evidence Table
CREATE TABLE IF NOT EXISTS research_evidence (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    source_id TEXT NOT NULL,
    snippet TEXT NOT NULL,
    location TEXT NOT NULL,
    verification_status TEXT NOT NULL DEFAULT 'UNVERIFIED',
    extractor_agent_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE,
    FOREIGN KEY (source_id) REFERENCES research_sources(id) ON DELETE CASCADE
);

-- 5. Research Claims Table
CREATE TABLE IF NOT EXISTS research_claims (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    statement TEXT NOT NULL,
    strength TEXT NOT NULL DEFAULT 'DIRECT',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE
);

-- 6. Claim Evidence Junction Table
CREATE TABLE IF NOT EXISTS claim_evidence_map (
    claim_id TEXT NOT NULL,
    evidence_id TEXT NOT NULL,
    PRIMARY KEY (claim_id, evidence_id),
    FOREIGN KEY (claim_id) REFERENCES research_claims(id) ON DELETE CASCADE,
    FOREIGN KEY (evidence_id) REFERENCES research_evidence(id) ON DELETE CASCADE
);

-- 7. Structured Findings Table
CREATE TABLE IF NOT EXISTS research_findings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    question_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    agent_type TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    confidence REAL NOT NULL DEFAULT 0.0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES research_questions(id) ON DELETE CASCADE
);

-- 8. Final Research Papers Table
CREATE TABLE IF NOT EXISTS research_papers (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL UNIQUE,
    template_id TEXT NOT NULL,
    title TEXT NOT NULL,
    paper_json TEXT NOT NULL, -- Structured sections + references
    markdown_output TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT', -- DRAFT, REVIEWED, PASSED, REJECTED
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE
);
```

---

## 6. Phase-by-Phase Implementation Roadmap

The evolution will be implemented across **5 distinct phases**, following strict TDD, clean commits (`COMMIT_GUIDE.md`), and atomic verification.

```
Phase 2.1: Workflow Core & Research Domain
    │
    ▼
Phase 2.2: Research Tools, Templates & Provenance Engine
    │
    ▼
Phase 2.3: Research Agents & Task Decomposition
    │
    ▼
Phase 2.4: Synthesis, Reviewer & Paper Generator
    │
    ▼
Phase 2.5: Terminal UI Research Dashboard & E2E Integration
```

### Phase 2.1: Workflow Core & Research Domain
- Define `internal/workflow/workflow.go` interface and `WorkflowRegistry`.
- Create domain models in `internal/workflows/research/domain/`.
- Add migration `000002_research_workflow.sql` and SQLite persistence repos.
- Write unit tests for domain validation and database persistence.
- **Commits:**
  - `feat(workflow): add core workflow interface and registry`
  - `feat(research): add research project domain models`
  - `feat(persistence): add research workflow database migration and repository`
  - `test(research): add unit tests for research domain persistence`

### Phase 2.2: Research Tools, Templates & Provenance Engine
- Implement research tools in `internal/tools/research/` (`paper_search`, `pdf_extractor`, `web_fetcher`, `citation_verifier`).
- Implement the `TemplateEngine` in `internal/workflows/research/templates/`.
- Implement `ProvenanceTracker` verifying `Claim` $\rightarrow$ `Evidence` $\rightarrow$ `Source`.
- Write unit tests for template loading, section validation, and evidence verification.
- **Commits:**
  - `feat(tools): add academic search and text extraction tools`
  - `feat(research): add template engine for research papers`
  - `feat(research): add source provenance tracker`
  - `test(research): add template validation and provenance tests`

### Phase 2.3: Research Agents & Task Decomposition
- Implement `ResearchPlanner` converting user input into `ResearchQuestion` items & DAG.
- Implement `LiteratureAgent`, `EvidenceAgent`, `MethodologyAgent`, and `ComparativeAgent`.
- Connect Research DAG to `scheduler.DependencyGraph` and `scheduler.WorkerPool`.
- Write unit tests for research task decomposition and parallel execution.
- **Commits:**
  - `feat(research): add research planner and question decomposer`
  - `feat(research): implement specialized research agents`
  - `test(research): add planner DAG decomposition tests`

### Phase 2.4: Synthesis, Reviewer & Paper Generator
- Implement `CriticalReviewer` for identifying weak claims and contradictions.
- Implement `Synthesizer` for consolidating findings and generating theme maps.
- Implement `PaperWriter` generating paper sections based strictly on verified findings.
- Implement `CitationReviewer` enforcing zero-hallucination verification.
- Write unit tests for synthesis, reviewer gates, and paper generation.
- **Commits:**
  - `feat(research): add critical reviewer for contradiction detection`
  - `feat(research): add synthesizer and evidence consolidator`
  - `feat(research): add paper writer and citation reviewer`
  - `test(research): add paper generation and citation verification tests`

### Phase 2.5: Terminal UI Research Dashboard & E2E Integration
- Extend Bubble Tea TUI to include a **Research Mode Dashboard** (DAG graph view, Evidence viewer, Paper preview).
- Add `--mode=research` CLI flag and research objective commands.
- Run end-to-end integration tests generating a complete research paper from a topic.
- Update documentation (`docs/RESEARCH_WORKFLOW.md`).
- **Commits:**
  - `feat(tui): add research workflow interactive dashboard`
  - `feat(cli): enable research workflow mode via flags`
  - `test(integration): add end-to-end research workflow test`
  - `docs(research): document research agent mode architecture`

---

## 7. Verification Matrix & Quality Gates

Every phase must satisfy the following quality gates before merging:

1. **Compilation & Formatting:** `gofmt -w . && go vet ./...`
2. **Unit Test Pass Rate:** 100% pass on all unit tests (`go test -v ./...`)
3. **Race Conditions:** Zero data races (`go test -race ./...`)
4. **Provenance Integrity Test:** Reject paper generation if any claim lacks a valid `EvidenceID` or `SourceID`.
5. **No Prose Violations:** Verify that all research worker agents return `ResearchFinding` structs instead of raw unstructured text.
