# ADR 0003: Transactional Mutation Engine & Optimistic Concurrency

## Context
AI agents modifying files on disk risk corrupting codebases if writes are partial, unapproved, or collide with concurrent edits.

## Decision
All filesystem modifications MUST pass through a dedicated `MutationEngine`.
- Operations are grouped into atomic two-phase transactions (`Transaction`).
- Optimistic concurrency control checks `before_hash` against live file hash prior to commit.
- Backup snapshots (`FileSnapshot`) enable full rollback if a transaction fails or is rejected.

## Consequences
- Agents cannot write directly to disk.
- User can inspect complete diffs before any change commits.
- File edit conflicts are detected before disk mutations occur.
