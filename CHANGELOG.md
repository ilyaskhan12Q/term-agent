# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- OpenCode interaction model featuring a unified linear `ConversationView` (`internal/tui/views/conversation_view.go`) with threaded messaging, interactive command blocks, and sub-agent research activity trees.
- `OpenRouter` provider support via `openai.NewProviderWithURL` and custom provider name overrides in `internal/model/bootstrap` and `internal/model/openai`.
- Specialist research worker pool (`WorkerPool`) executing `LITERATURE_AGENT`, `WEB_RESEARCH_AGENT`, `EXTRACTION_AGENT`, and `VERIFICATION_AGENT` in parallel.
- `EvidenceVerifier` claim/entailment verification pipeline for research provenance tracking.
- `SynthesisAgent` enhanced with LLM-backed paper synthesis, inline citation mapping (`[N]`), and status badges.
- `PaperWriter` multi-format paper compiler supporting Markdown, LaTeX, HTML, and JSON outputs.
- `ReviewerAgent` adversarial auditing for paper hallucination risk analysis and claim verification.
- Phase 1 (Research Provider/Model Setup): Real OpenAI, Anthropic, and Gemini provider implementations with `net/http` clients.
- `model.ProviderConfig` validation that produces actionable user-facing errors for missing provider, model, or API key.
- `model.SupportedProviders` registry listing canonical provider identifiers.
- `model.DefaultFactory` injectable factory pattern to decouple provider construction from the model package.
- `internal/model/bootstrap` package that wires concrete provider implementations into the factory at startup.
- `NewProviderWithURL` test constructors on all three providers enabling httptest-based unit testing without real API keys.
- Phase 2 (Research Slash Commands): Extensible slash-command registry (`internal/commands`) with `Command` interface, `Registry.Dispatch`, `Registry.HelpText`, and `Registry.Lookup`.
- Input parser (`commands.Parse`) detecting `/cmd args` prefixes from raw prompt input.
- 12 research slash commands: `/research`, `/topic`, `/plan`, `/sources`, `/status`, `/pause`, `/resume`, `/cancel`, `/export`, `/model`, `/help`, `/clear` with aliases (`/r`, `/s`, `/m`, `/h`, `/?`, `/cls`).
- `ResearchState` shared mutable struct propagated across all commands; provider/model switchable at runtime.
- `ResearchView.AddLog` / `ResearchView.Clear` methods enabling command output rendering in the research panel.
- `ResearchState.Sources` field storing `[]domain.Source` instances collected during research sessions.
- `/sources [query]` command extended to trigger real `AcademicSearchTool` searches and append results to session state.
- `NewAcademicSearchToolWithURL`, `NewWebFetchToolWithClient`, and `NewPDFExtractorToolWithClient` constructors for dependency-injected testing.
- Content-Type guards in `WebFetchTool` and `PDFExtractorTool` to reject non-text and non-PDF responses.
- `ExecuteStepWithProject` method on `ResearchWorkerAgent` providing dynamic project ID context and automated extraction of `domain.Source` and `domain.Evidence` from tool output.
- Thread-safe repository persistence and provenance registration during parallel worker pool execution.
- `UpdateEvidenceStatus` thread-safe method on `ProvenanceTracker` for updating evidence verification states.
- Evidence verification summary integration into `ResearchWorkflow.Execute` output payload.
- Embedded 4 standard paper templates (`academic_research`, `technical_survey`, `executive_briefing`, `system_architecture`) into `TemplateEngine`.
- Comprehensive unit and integration test suites covering research E2E execution, worker pools, synthesis, paper compilation, and claim verification.

### Changed
- Refactored `internal/tui/model.go` and `internal/tui/update.go` to route event streams, command results, user input, and viewport scrolling through the new `ConversationView`.
- Refactored `ResearchState` handling to dynamically update status bar telemetry (provider and model name) upon `/model` slash command invocation.
- Updated `/export` slash command to support `html` and `json` formats alongside `markdown`, `latex`, and `pdf`.

### Fixed
- Resolved `domain.Evidence` field name inconsistencies across research slash commands (`Snippet`, `Location`, `VerificationStatus`).
- Resolved command alias collision between `/questions` (renamed alias to `ques`) and `/quit` (alias `q`).
- Fixed `openrouter` provider identification issue where `Name()` previously returned `openai` instead of `openrouter`, causing unit test failures in `TestBootstrap_BuildProvider_AllSupportedProviders`.
- Fixed foreign key constraint failures during SQLite persistence by seeding parent session records in `ResearchWorkflow`.
- Fixed `containsNegation` in `CitationVerifierTool` to use word boundary regex (`\b(not|no|...)\b`), eliminating false positives on words like "nodes".

### Security
- Prompt injection defense using immutable `<untrusted_content>` envelopes with SHA-256 integrity tags in `internal/security/isolation.go`.

### Deprecated

### Removed

## [0.8.0] - 2026-08-15

### Added
- Research Planner Agent (`ResearchPlannerAgent`) for research objective decomposition.
- Directed Acyclic Graph (DAG) plan validation (`ValidateDAG`) to enforce non-cyclic dependency resolution.
- Dynamic research replanning (`Replan`) to inject fallback web discovery tasks upon worker failure.

### Fixed
- Fixed `gofmt` code formatting in `internal/security/isolation.go`.
- Fixed temporary PDF file isolation handling in unit test suite.

## [0.7.0] - 2026-08-15

### Added
- Academic Search Tool (`AcademicSearchTool`) with arXiv Atom XML feed retrieval.
- Web Content Fetcher (`WebFetchTool`) with HTML parsing and boilerplate removal.
- Document Extractor (`PDFExtractorTool`) with stream-based local and remote PDF text parsing.
- Citation Verification (`CitationVerifierTool`) featuring semantic entailment scoring, token overlap calculation, and negation conflict detection.
- URI-based source deduplication and conflict detection in `ProvenanceTracker`.
- Standardized bibliography generation supporting APA and IEEE citation formats.

### Security
- Prompt injection defense using immutable `<untrusted_content>` envelopes with SHA-256 integrity tags in `internal/security/isolation.go`.
- Encapsulated external web content, PDF stream contents, and academic search responses within security envelopes.

## [0.6.0] - 2026-08-15

### Added
- Security isolation subsystem (`internal/security/isolation.go`).
- Untrusted content input sanitization to strip control characters and prompt injection signatures.

## [0.5.0] - 2026-08-15

### Added
- POSIX AST command risk classifier (`internal/security/classifier.go`).
- Shell tool execution engine (`internal/tools/shell.go`) with execution timeouts and safety policy gating.
- Sensitive path validation blocking unsafe access to system configuration files and secrets.

## [0.4.0] - 2026-08-15

### Added
- Transactional mutation engine supporting optimistic concurrency control (OCC).
- Disk state hashing (`before_hash`) verification before workspace commits.
- Unified diff generation and rendering engine (`internal/diff`).

## [0.3.0] - 2026-08-15

### Added
- Workspace file discovery and metadata extraction (`internal/workspace`).
- Read-only file inspection, pattern search, and memory-bounded directory listing tools.

## [0.1.0] - 2026-08-15

### Added
- Foundation architecture including SQLite WAL database initialization, event bus, configuration management, and baseline security path traversal prevention.
