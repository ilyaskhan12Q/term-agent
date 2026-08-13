# ADR 0002: Bubble Tea Framework for Terminal UI

## Context
Term-Agent needs a responsive, interactive terminal user interface supporting live streaming, diff reviews, vim keybindings, and dynamic window resizing.

## Decision
We select Charm's `Bubble Tea` (Elm architecture) along with `Lip Gloss` for design tokens and styling.

## Consequences
- Clean separation of UI state, message updates, and view rendering.
- UI layer communicates with backend subsystems strictly through asynchronous event messages and command channels.
- TUI layer has zero direct dependencies on SQLite repositories, file mutation routines, or shell execution.
