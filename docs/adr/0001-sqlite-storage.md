# ADR 0001: Embedded SQLite Persistence with WAL Mode

## Context
Term-Agent requires a fast, reliable, local-first database to store session history, message logs, task DAGs, mutation transactions, and event logs.

## Decision
We choose embedded SQLite (`modernc.org/sqlite`) as the local relational storage engine.
- WAL mode (`_pragma=journal_mode(WAL)`) is enabled for concurrent read performance.
- Foreign keys (`_pragma=foreign_keys(1)`) are strictly enforced.
- Synchronous mode (`_pragma=synchronous(NORMAL)`) balances durability and speed.

## Consequences
- Single binary deployment without external database dependencies.
- Asynchronous persistence writer (`AsyncWriter`) is required to serialize concurrent write operations from parallel agent workers.
