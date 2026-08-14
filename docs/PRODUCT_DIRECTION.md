# Permanent Product & Architectural Direction

> **CRITICAL ARCHITECTURAL DIRECTIVE FOR ALL AGENTS & DEVELOPERS**  
> `term-agent` is **NOT** just a coding agent. It is a **generic, extensible multi-agent orchestration engine**, capable of running specialized vertical domain workflows. **Research Agent Mode** is its first specialized production workflow.

---

## 1. Core Mission & Vision

`term-agent` is built as an enterprise-grade, local-first multi-agent runtime. Its core mission is to provide foundational infrastructure for agent orchestration, planning, task decomposition, parallel execution, context management, and interactive terminal visualization for **any specialized domain workflow**.

---

## 2. Platform Architecture Layers

```
┌────────────────────────────────────────────────────────────────────────┐
│                        Terminal User Interface                         │
│       Bubble Tea / Lip Gloss (Agent, Plan, Diff, Log, Research Views)  │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼────────────────────────────────────┐
│                    Specialized Workflow Layer                          │
│   ┌──────────────────────────────────┐  ┌───────────────────────────┐  │
│   │    Research Agent Mode           │  │   Future Workflows        │  │
│   │   (Planner, Search, Extractor,   │  │   (Coding, Legal,         │  │
│   │    Verifier, Synthesis Agents)   │  │    Finance, Security)     │  │
│   └──────────────────────────────────┘  └───────────────────────────┘  │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼────────────────────────────────────┐
│                    Core Engine Runtime (`internal/`)                   │
│  - Task Scheduler & DAG Engine       - Event Bus & Diagnostics        │
│  - Agent Execution & Lifecycle       - Context Window Budgeting       │
│  - Model Provider Abstractions       - State Persistence (SQLite)     │
│  - Tool Registry & Sandbox           - Diff & Safety Review Engine    │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Core Engine Responsibilities vs Specialized Workflows

### Core Engine Responsibilities (`internal/`)
* **Agent Orchestration & Scheduling**: DAG dependency graph resolution, parallel execution, worker pool management.
* **Model & Provider Abstraction**: Uniform interface over OpenAI, Anthropic, Gemini, local LLMs.
* **Context & Memory Management**: Session state tracking, context compaction, persistent memory.
* **Persistence & Event System**: SQLite schema migration, structured event bus, state hashing.
* **Safety & Diff Review**: Workspace boundary enforcement, dangerous command filtering, interactive transaction authorization.
* **Terminal User Interface (TUI)**: Rich 5-view Bubble Tea interface visualizing logs, task DAGs, agent streams, diffs, and domain outputs.

### Specialized Workflow 1: Research Agent Mode (`internal/workflows/research/`)
* **Goal**: Conduct automated deep research, literature review, provenance tracking, evidence verification, and academic paper synthesis.
* **Worker Squad**: `ResearchPlannerAgent`, `LiteratureSearchAgent`, `EvidenceExtractorAgent`, `CitationVerifierAgent`, `SynthesisAgent`.
* **Tools**: `academic_search`, `pdf_extractor`, `citation_verifier`.
* **Outputs**: Markdown/LaTeX Research Papers with strict Claim-Evidence-Source provenance reports.

---

## 4. Invariant Rules for Future Development

1. **Engine Decoupling**: Core engine abstractions (`internal/agent`, `internal/scheduler`, `internal/events`, `internal/tui`) MUST remain domain-agnostic. Never hardcode domain logic into core engine components.
2. **Workflow Pluggability**: All specialized workflows (Research Mode, future Coding Mode, Legal Mode, etc.) MUST implement the `workflow.Workflow` interface and register cleanly with `workflow.Registry`.
3. **No Coding-Agent Assumption**: Never assume `term-agent` only performs software development tasks. All prompt templates, UI labels, and data structures must support arbitrary multi-agent workflows.
4. **Verification Gate**: Before executing any new phase, PR, or feature, verify compliance against this document and the core PRD.
