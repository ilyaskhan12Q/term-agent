# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Phase 9 parallel specialist researcher orchestration architecture.
- Parallel worker pool scheduling and concurrency error group management.

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
