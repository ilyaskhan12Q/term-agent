# term-agent — Production-Grade Product Requirements Document
**Status:** Draft — Architecture Baseline
**Version:** 1.0
**Product Type:** Local-first terminal AI coding agent
**Primary Platform:** Linux, macOS, Windows
**Primary Language:** Go
**UI:** Bubble Tea + Lip Gloss
**Database:** SQLite

**Working Principle:** The AI proposes actions; the runtime validates, controls, executes, and r
--# 1. Executive Summary
`term-agent` is a local-first, terminal-native AI coding agent designed for software engineers,
The product combines:
- AI-assisted software development
- Workspace inspection
- Structured task planning
- Parallel sub-agents
- Controlled terminal execution
- Transactional file modifications
- Visual diff approval
- Session persistence
- Context management
- Multi-model/provider support
- Git awareness
- Crash recovery
- Local security controls
The product must not behave like a chatbot that happens to execute shell commands.

It must behave like a **controlled software-engineering runtime** in which AI agents operate ins
## Core Architectural Principle
> **The model proposes. The runtime decides. The user approves. The transaction commits.**
--# 2. Product Vision

The long-term vision is to create a terminal-native AI engineering environment where a developer

> "Add GitHub OAuth authentication, update the tests, run the test suite, and explain the change
`term-agent` should be capable of:



1. Understanding the workspace.
2. Inspecting the existing architecture.
3. Creating an execution plan.

8.4.Architectural
Boundaries
Breaking the plan into
independent tasks.
5. Running appropriate agents in parallel.

The
system must be divided into clearly defined layers.
6. Reading relevant files.
7. Proposing file changes.
8.
TUIShowing the exact diff.
9.↓Asking for approval when required.
10.
Applying approved mutations transactionally.
Application
11.
↓ Running tests.
12.
Detecting
Agent
Runtime failures.
13.
↓ Iterating safely.
14.
the complete session.
ToolPersisting
Runtime
15.
↓ Recovering unfinished operations after a crash.
Workspace / System
---

The following dependencies are prohibited:
# 3. Core Product Principles

• TUI directly calling SQLite

3.1 directly
Local First
•## TUI
executing shell commands

• TUI directly modifying files

The application should run primarily on the user's machine.

• Model providers directly modifying files

source opening
code mustSQLite
not beconnections
uploaded anywhere except to the model provider explicitly sele
•Workspace
Agents directly

• Agents directly executing arbitrary OS processes
There must be no mandatory cloud backend.

• Providers knowing about Bubble Tea
•---Tools knowing about TUI rendering
## 3.2 Runtime-Owned Security

LLM must never directly control:
9.TheProposed
Project Structure
- The filesystem
- Shell execution
- Process creation
- Git
- Network access
- Environment variables
The model can only request tools through the tool runtime.
--## 3.3 Transactional Mutations
File modifications must be treated as transactions.
Every mutation must be:



1. Proposed
2. Validated
3. Diffed
4. Approved or rejected
5. Committed
6. Recoverable
--## 3.4 Observable Agent Behavior
The user must always be able to understand:
- What the agent is doing
- Which task is running
- Which files are being inspected
- What tools are being executed
- What changes are proposed
- Why approval is required
- How much context has been consumed
- Estimated model cost
- Whether agents are waiting, running, failed, or completed
--## 3.5 Provider Independence

The application must not couple its internal agent architecture to OpenAI, Anthropic, Gemini, or
A canonical internal model interface must be used.
--## 3.6 Fail Closed

If the runtime cannot confidently determine whether an operation is safe, the default behavior m
> Ask the user or deny the operation.
Never silently allow an unknown operation.
--# 4. Target Users
## 4.1 Software Engineers
Developers working with:
- Go
- Python



- JavaScript/TypeScript
- Rust
- C/C++
- Java
- PHP
- Other common development stacks
--## 4.2 DevOps Engineers
Users working with:
- Docker
- Kubernetes
- CI/CD
- Git
- Shell scripts
- Infrastructure repositories
--## 4.3 Power Users
Users who spend significant time in:
- Terminals
- Git repositories
- Editors
- Command-line development environments
--# 5. Core User Problems
## Problem 1 — AI Coding Tools Can Be Opaque
Users do not always know:
- What files the AI inspected
- Why it modified a file
- Which commands were executed
- What changed
`term-agent` must expose this information.
--## Problem 2 — Autonomous Shell Execution Is Dangerous
An AI can execute commands that have unintended consequences.



`term-agent` must introduce a command policy layer.
--## Problem 3 — Parallel Agents Can Conflict
Two agents modifying the same files can create corruption or inconsistent state.
`term-agent` must coordinate workspace access.
--## Problem 4 — Long Conversations Lose Context
Large coding tasks can exceed model context windows.
`term-agent` must maintain structured workspace memory and perform context compaction.
--## Problem 5 — Agent Failures Can Leave Workspaces Inconsistent
A crash during a file mutation must not leave the workspace in an unknown state.
`term-agent` must use transactional mutation tracking.
--# 6. Product Scope
## 6.1 MVP Scope
The MVP must support:
- Terminal TUI
- Workspace detection
- File reading
- Workspace search
- Git status
- AI model integration
- Streaming responses
- Tool calling
- File mutation proposals
- Unified diff generation
- User approval
- Transactional file writes
- Rollback
- Shell execution with policy controls
- Session persistence
- Context tracking
- Basic context compaction



- Crash recovery
- Single-agent execution
- Configurable worker limits
--## 6.2 Post-MVP Scope
Post-MVP features include:
- Parallel sub-agents
- Dependency-aware task graphs
- Planner agent
- Researcher agent
- Coder agent
- Tester agent
- Reviewer agent
- Multi-provider routing
- Advanced workspace memory
- Git checkpoints
- Autonomous workflows
- Permission profiles
- Advanced cost controls
- Plugin/tool ecosystem
--# 7. High-Level Architecture
```text
USER
|
v
+---------------+
|

TUI

|

| Bubble Tea

|

+-------+-------+
|
v
+-------------------+
| Application

|

| Controller

|

+---------+---------+
|
+-----------+-----------+
|

|

v

v

+---------------+

+---------------+

| Agent Runtime |

|

+-------+-------+

+-------+-------+

|

|

Event Bus

|



+------+-------+

+------+------+

|

|

|

|

|

v

v

v

v

v

Orchestrator

Context

TUI

SQLite

Logs

|
v
Task Graph
|
v
Scheduler
|
+---+---+---+---+
|

|

|

|

|

v

v

v

v

v

A1

A2

A3

A4

A5

|

|

|

|

|

+---+---+---+---+
|
v
Tool Runtime
|
+---+-----------+
|

|

|

v

v

v

Files

Search

Shell

|

|

v

v

Mutation

Policy Engine

Engine

|

|

v

v

Executor

Diff
|
v
Approval
|
v
Transaction
|
v
Workspace

term-agent/
│
├── cmd/
│

└── termagent/

│

└── main.go

│
├── internal/
│

│

│

├── app/



│

├── app/

│

│

├── app.go

│

│

└── lifecycle.go

│

│

│

├── agent/

│

│

├── agent.go

│

│

├── orchestrator.go

│

│

├── planner.go

│

│

├── worker.go

│

│

├── context.go

│

│

└── memory.go

│

│

│

├── scheduler/

│

│

├── scheduler.go

│

│

├── queue.go

│

│

├── dependency.go

│

│

└── worker_pool.go

│

│

│

├── model/

│

│

├── provider.go

│

│

├── registry.go

│

│

├── request.go

│

│

├── response.go

│

│

├── capabilities.go

│

│

├── openai/

│

│

├── anthropic/

│

│

└── gemini/

│

│

│

├── tools/

│

│

├── tool.go

│

│

├── registry.go

│

│

├── read.go

│

│

├── write.go

│

│

├── search.go

│

│

└── shell.go

│

│

│

├── security/

│

│

├── policy.go

│

│

├── classifier.go

│

│

├── permissions.go

│

│

└── path.go

│

│

│

├── workspace/

│

│

├── workspace.go

│

│

├── coordinator.go

│

│

├── locking.go

│

│

├── discovery.go

│

│

└── hashing.go

│

│

│

├── mutation/

│

│

├── transaction.go

│

│

├── mutation.go



│

│

├── mutation.go

│

│

├── snapshot.go

│

│

└── rollback.go

│

│

│

├── diff/

│

│

├── engine.go

│

│

├── parser.go

│

│

└── renderer.go

│

│

│

├── context/

│

│

├── manager.go

│

│

├── tokenizer.go

│

│

├── budget.go

│

│

└── compactor.go

│

│

│

├── persistence/

│

│

├── database.go

│

│

├── writer.go

│

│

├── migrations/

│

│

└── repository/

│

│

│

├── events/

│

│

├── event.go

10.
Technology
Stack
│
│
│

│

│

├── git/

│

│

│

├── config/

Core
│
│

└── bus.go

└── git.go

Component

Technology

│
│
└── config.go
Language

Go

│

│

│

└── tui/

TUI

│

├── model.go

│

├── update.go

Styling
│

├── views/

│

Bubble Tea
Lip Gloss

├── components/

│
Database
│

├── keymap/

SQLite

└── styles/

│
Configuration

TOML + environment variables

├── migrations/
│

Logging

├── tests/
│

├── unit/

│

├── security/

Build
│

└── fixtures/

Testing
│
├── integration/

Structured Go logger
Go testing + integration tests
Go toolchain

│

CI ├── docs/
│

├── architecture.md

│

├── security.md

│

├── agent-runtime.md

│

└── decisions/

GitHub Actions



11.
Philosophy
├── Dependency
.github/
│

│
└── workflows/
Dependencies
must be minimized.
│

├── go.mod

Before introducing a third-party package, determine whether:
├── go.sum

├── Makefile

1. The Go standard library already provides the functionality.
├── README.md

2.└──The
dependency is actively maintained.
LICENSE
3. The dependency is compatible with all target platforms.
4. The dependency has a suitable license.
5. The dependency introduces unnecessary transitive dependencies.
6. The dependency creates CGO requirements.
7. The dependency has known security issues.

12. Configuration
Default configuration directory:
~/.config/termagent/

Files:
~/.config/termagent/config.toml
~/.config/termagent/state.db
~/.config/termagent/logs/

Secrets should preferably come from:
1. CLI/environment
2. OS credential storage
3. Configuration file only when explicitly supported
API keys must never be written into logs or SQLite session messages.

13. Configuration Precedence
Highest priority wins:



CLI arguments
↓
Environment variables
↓
Config file
↓
Application defaults

Example:
termagent --model <model>

overrides:
config.toml

14. CLI Interface
Required commands:
Bash
termagent
termagent --help
termagent --version
termagent --session <id>
termagent --model <model>
termagent --workspace <path>
termagent --dry-run
termagent --config <path>

Future commands:
Bash
termagent sessions
termagent history
termagent doctor
termagent config
termagent models



15. Workspace Model
A workspace is the root directory in which the agent is permitted to operate.
At startup the system must discover:
• Current working directory
• Workspace root
• Git repository
• Git branch
• Git status
• Project type
• Programming languages
• Package manager
• Build system
• Test system
• Important configuration files

16. Workspace Boundary
Every filesystem operation must be resolved relative to the workspace.
Before access:
Requested path
↓
Normalize
↓
Absolute path
↓
Resolve relevant symlinks
↓
Verify workspace boundary
↓
Security policy
↓
Allow / Ask / Deny

The system must prevent:
../../outside-workspace

and equivalent path traversal.



Symlink-based workspace escapes must also be detected.

17. Workspace Coordinator
The Workspace Coordinator owns concurrent filesystem access.
It must support:
READ
WRITE
DELETE
RENAME

Read operations may run concurrently.
Write operations must be exclusive for conflicting paths.
Example:
Agent A → READ src/auth.go
Agent B → READ src/auth.go
Agent C → WRITE src/auth.go

A and B may proceed concurrently.
C must wait until conflicting reads finish.

18. Tool Architecture
Every tool must implement a common interface.
Conceptually:
Go
type Tool interface {
Name() string
Description() string
Schema() ToolSchema
Execute(ctx context.Context, input ToolInput) (ToolResult, error)
Risk() RiskLevel
}



Tools must be registered through a Tool Registry.

19. Initial Tools
19.1 read_file
Purpose:
Read a file inside the workspace.
Input:
JSON
{
"path": "src/main.go"
}

Requirements:
• Must respect workspace boundaries.
• Must enforce maximum file-read size.
• Must detect binary content.
• Must return useful errors.
• Must not modify the workspace.

19.2 search_workspace
Purpose:
Search the workspace for relevant files, text, or symbols.
Initial support:
• Filename search
• Text search
• Path filtering
• File extension filtering
• Git-aware ignore behavior
Future support:
• Symbol search



• AST search
• Language-aware search

19.3 write_file
The model must never directly write to the filesystem.
The tool should create a mutation proposal.
Flow:
Agent
↓
write_file request
↓
Validate
↓
Read current file
↓
Hash current file
↓
Generate proposed content
↓
Generate diff
↓
Mutation transaction
↓
User approval
↓
Commit

19.4 execute_shell
Purpose:
Execute a shell command through the security policy engine.
The model must never receive unrestricted shell access.

20. Tool Call IDs
Every tool invocation must have a unique ID.



Example:
JSON
{
"tool_call_id": "call_782",
"tool": "read_file",
"arguments": {
"path": "src/main.go"
}
}

The tool result must reference the same ID.

21. Tool Risk Levels
Tools must have explicit risk levels.
SAFE
LOW
MEDIUM
HIGH
CRITICAL

Example:
read_file

SAFE

search

SAFE

git_status

SAFE

go_test

LOW

write_file

MEDIUM

npm_install

MEDIUM

git_checkout

HIGH

delete_file

HIGH

sudo

CRITICAL

Unknown operations default to:
ASK

or:
DENY



depending on policy.

22. Shell Security
A simple blacklist is prohibited.
The system must not rely only on matching strings such as:
rm -rf /
curl | bash
wget | sh

Dangerous commands can be disguised through:
• Variables
• Shell substitution
• Functions
• Redirection
• Pipes
• Scripts
• Encodings
• Aliases
• Environment manipulation
Therefore the shell execution system must include:
Command Request
↓
Command Classification
↓
Security Policy
↓
Permission Decision
↓
Execution

23. Shell Policy Categories
Commands may be classified as:



READ_ONLY
WORKSPACE_MUTATION
SYSTEM_MUTATION
NETWORK
PRIVILEGED
UNKNOWN

Examples:
git status
→ READ_ONLY
go test ./...
→ READ_ONLY
npm install
→ WORKSPACE_MUTATION
git checkout
→ WORKSPACE_MUTATION
rm file.txt
→ HIGH / ASK
sudo ...
→ CRITICAL / DENY
shutdown
→ CRITICAL / DENY

24. Shell Resource Limits
Every command must have:
• Maximum execution time
• Maximum stdout
• Maximum stderr
• Maximum process count where applicable
• Cancellation support
• Exit code
• Signal information
Commands must support context.Context .



25. Mutation Engine
The Mutation Engine manages all proposed workspace changes.
Each mutation must include:
mutation_id
transaction_id
path
operation
before_hash
after_hash
before_content
after_content
diff
status
timestamp

Operations:
CREATE
MODIFY
DELETE
RENAME

26. Mutation Transaction
A transaction groups related mutations.
Example:
Transaction
│
├── modify auth.go
├── create auth_test.go
└── modify router.go

Possible states:



PROPOSED
VALIDATED
WAITING_APPROVAL
APPROVED
REJECTED
COMMITTING
COMMITTED
ROLLING_BACK
ROLLED_BACK
FAILED

27. Optimistic File Concurrency
Before applying a mutation, the system must verify that the file has not changed since the
proposal was generated.
Example:
before_hash = abc123

At commit time:
current_hash == abc123

If not:
CONFLICT

The mutation must not silently overwrite the user's newer changes.

28. Diff Engine
Every modification must generate a unified diff.
Example:
Diff
- old code
+ new code



The TUI must visually distinguish:
+ additions
- deletions
context

The diff must be generated from the actual before/after content, not fabricated by the
model.

29. Diff Review
Before committing a mutation, the TUI must display:
• File path
• Operation
• Added lines
• Removed lines
• Context
• Agent responsible
• Task responsible
Controls:
Y

Approve

N

Reject

A

Approve all

R

Reject all

↑/↓

Navigate

Enter

Inspect

Esc

Return

Future:
E

Edit proposal

30. Crash Recovery
If the process exits while a transaction is pending, the transaction must remain in SQLite.
On startup:



Pending transaction detected.
Resume review?

The system must never assume that an interrupted transaction was successfully
committed.

31. Agent Architecture
The initial MVP uses a single agent.
The architecture must support hierarchical multi-agent execution later.
Long-term:
Root Agent
↓
Orchestrator
↓
Task Graph
↓
Scheduler
↓
Specialized Agents

Possible agents:
Planner
Researcher
Coder
Tester
Reviewer
Debugger

32. Agent Interface
Conceptually:
Go



type Agent interface {
ID() string
Run(ctx context.Context, task Task) (AgentResult, error)
}

Agents must not directly manipulate files.
They must request tools.

33. Agent Loop
The core loop:
Receive Goal
↓
Inspect Context
↓
Reason / Plan
↓
Request Tool
↓
Runtime Validates Tool
↓
Execute Tool
↓
Return Result
↓
Continue
↓
Complete / Fail

The loop must have hard limits.

34. Agent Limits
Configurable limits:
max_iterations
max_tool_calls
max_runtime
max_cost
max_context

An agent must never run indefinitely.



35. Planner
The planner converts a high-level request into structured tasks.
Example:
JSON



{
"plan_id": "plan_123",
"goal": "Add GitHub OAuth authentication",
"tasks": [
{
"id": "task_1",
"title": "Inspect authentication architecture",
"type": "analysis",
"dependencies": [],
"allowed_tools": [
"read_file",
"search_workspace"
]
},
{
"id": "task_2",
"title": "Implement OAuth provider",
"type": "implementation",
"dependencies": [
"task_1"
],
"allowed_tools": [
"read_file",
"search_workspace",
"write_file"
]
},
{
"id": "task_3",
"title": "Add authentication tests",
"type": "testing",
"dependencies": [
"task_2"
],
"allowed_tools": [
"read_file",
"write_file",
"execute_shell"
]
}
]
}

The output must be schema-validated before execution.

36. Task Graph
Tasks must support dependencies.



Example:
A ──────┐
├── C ───┐
B ──────┘

│
├── E

A ───────── D ──┘

Execution:
A + B
↓
C + D
↓
E

The scheduler determines which tasks are ready.

37. Scheduler
The scheduler is responsible for:
• Task queue
• Worker pool
• Dependency resolution
• Priority
• Cancellation
• Timeouts
• Retry policy
• Failure propagation
• Resource limits
sync.WaitGroup may be used internally but must not be treated as the scheduler itself.

38. Worker Pool
Default:
max_workers = 5

The value must be configurable.



Workers must support:
START
RUNNING
WAITING
COMPLETED
FAILED
CANCELLED
BLOCKED

39. Parallel Agent Safety
Multiple agents must not independently mutate the same workspace without coordination.
Preferred architecture:
Agent A → Proposed Changes
Agent B → Proposed Changes
Agent C → Proposed Changes
↓
Change Manager
↓
Conflict Check
↓
Diff
↓
Approval
↓
Transaction

The runtime owns actual workspace mutation.

40. Event Bus
The application must use an internal event bus.
Events include:



session.started
session.resumed
session.completed
message.created
agent.started
agent.completed
agent.failed
task.created
task.started
task.completed
task.failed
task.cancelled
tool.requested
tool.started
tool.completed
tool.failed
mutation.created
mutation.approved
mutation.rejected
mutation.committed
mutation.rolled_back
context.compacted
model.requested
model.completed
model.failed

41. Event Flow
Agent Runtime
↓
Event Bus
↓
+-----+-------+--------+
|

|

|

v

v

v

TUI

SQLite

Logs

This keeps UI and persistence independent from runtime logic.



42. Bubble Tea Architecture
Bubble Tea must be used for presentation and event handling.
The Bubble Tea model must not contain:
• Agent orchestration
• SQLite business logic
• Shell execution
• File mutation logic
• Provider-specific code
The TUI should subscribe to application events and render application state.

43. TUI Layout
Recommended layout:
┌──────────────────────────────────────────────────────────────┐
│ TERM-AGENT

session: abc123

model: provider/model │

├──────────────────────────────────────────────┬───────────────┤
│

│ AGENTS

│

│ USER

│

│

│ > Add authentication

│ ● planner

│

│

│ ● researcher

│

│ ASSISTANT

│ ◐ coder

│

│ I'll inspect the repository first...

│ ○ tester

│

│

│

│

│ [tool] search_workspace

├───────────────┤

│ [tool] read_file

│ CONTEXT

│

│

│

│

│

│ 61% context

│

│

│ $0.024

│

│

│ 8,421 tokens

│

│

│

│

├──────────────────────────────────────────────┴───────────────┤
│ > Type your instruction...

│

├──────────────────────────────────────────────────────────────┤
│ ENTER send

ALT+ENTER newline

TAB agents

CTRL+C cancel │

└──────────────────────────────────────────────────────────────┘



44. TUI Requirements
The interface must support:
• Dynamic resizing
• Horizontal/vertical layout adaptation
• Scrolling
• Multiline input
• Input history
• Streaming output
• Agent status
• Tool status
• Diff rendering
• Approval dialogs
• Error states
• Context statistics
• Cost statistics

45. Input Controls
Initial keymap:
ENTER

Send

ALT+ENTER

New line

CTRL+C

Cancel current operation

CTRL+D

Exit when input is empty

UP/DOWN

History

TAB

Navigate panels

ESC

Close overlay

Y

Approve

N

Reject

Keybindings must be configurable in the future.

46. Model Provider Architecture
All providers must implement a common interface.
Conceptually:
Go



type ModelProvider interface {
Chat(ctx context.Context, request Request) (Response, error)
Stream(ctx context.Context, request Request) (<-chan Event, error)
}

Providers:
OpenAIProvider
AnthropicProvider
GeminiProvider

The rest of the application must interact only with the canonical interface.

47. Model Registry
Each model must have metadata:
provider
model
context_window
supports_tools
supports_streaming
supports_vision
supports_reasoning
tokenizer
pricing

Example conceptual structure:
Go
type ModelCapabilities struct {
ContextWindow

int

SupportsTools

bool

SupportsStreaming bool
SupportsVision

bool

SupportsReasoning bool
}

48. Provider Streaming
Provider-specific streaming behavior must be normalized.
Internal events should look like:



Go
type ModelEvent struct {
RequestID string
Type

EventType

Delta

string

Usage

Usage

}

The TUI consumes internal events rather than provider-specific responses.

49. Context Management
The context manager is responsible for:
• System prompt
• Conversation history
• Tool calls
• Tool results
• Workspace state
• Agent memory
• Token budget
• Compaction

50. Token Tracking
The system must track:
input_tokens
output_tokens
cached_tokens
tool_tokens
total_tokens
estimated_cost
context_limit

Tokenizer support must be abstracted.
Do not assume one tokenizer works for every provider.



51. Context Budget
When context reaches a configurable threshold, default:
85%

the context manager must initiate compaction.
The threshold must be configurable.

52. Context Compaction
Compaction must not merely summarize the conversation.
It must extract structured state.
Example:
JSON
{
"workspace": {
"language": "Go",
"framework": "Bubble Tea",
"package_manager": "Go Modules"
},
"current_goal": "Implement authentication",
"completed_work": [
"Created authentication service"
],
"pending_work": [
"Add tests"
],
"decisions": [
"Use provider abstraction"
],
"modified_files": [
"internal/auth/auth.go"
],
"known_issues": [
"OAuth callback test failing"
]
}

The structured summary becomes part of the active context.
The most recent conversational turns must remain intact.



53. Workspace Memory
Workspace memory should maintain:
• Architecture summary
• Important files
• Current objective
• Completed work
• Pending work
• Decisions
• Known errors
• Modified files
• Test status
This is separate from raw conversation history.

54. SQLite
SQLite is the primary local persistence layer.
The application must use WAL mode where appropriate.
Because SQLite still has a single writer at a time, application-level persistence should use
a controlled writer rather than allowing every goroutine to independently perform
unrestricted writes.

55. Persistence Writer
Recommended flow:
Agent Goroutines
↓
Event Channel
↓
Persistence Writer
↓
SQLite

The UI must not block on persistence.



56. Database Schema
Initial schema:
SQL



CREATE TABLE sessions (
session_id TEXT PRIMARY KEY,
workspace_path TEXT NOT NULL,
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
current_model TEXT NOT NULL,
status TEXT NOT NULL
);
CREATE TABLE messages (
message_id INTEGER PRIMARY KEY AUTOINCREMENT,
session_id TEXT NOT NULL,
role TEXT NOT NULL,
content TEXT NOT NULL,
token_count INTEGER,
timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(session_id) REFERENCES sessions(session_id)
);
CREATE TABLE tool_calls (
tool_call_id TEXT PRIMARY KEY,
session_id TEXT NOT NULL,
message_id INTEGER,
tool_name TEXT NOT NULL,
arguments_json TEXT NOT NULL,
status TEXT NOT NULL,
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
completed_at TIMESTAMP,
FOREIGN KEY(session_id) REFERENCES sessions(session_id),
FOREIGN KEY(message_id) REFERENCES messages(message_id)
);
CREATE TABLE tool_results (
tool_result_id INTEGER PRIMARY KEY AUTOINCREMENT,
tool_call_id TEXT NOT NULL,
output TEXT,
error TEXT,
exit_code INTEGER,
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(tool_call_id) REFERENCES tool_calls(tool_call_id)
);
CREATE TABLE tasks (
task_id TEXT PRIMARY KEY,
session_id TEXT NOT NULL,
parent_task_id TEXT,
title TEXT NOT NULL,
description TEXT,
task_type TEXT,
status TEXT NOT NULL,
priority INTEGER DEFAULT 0,
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP



created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
started_at TIMESTAMP,
completed_at TIMESTAMP,
FOREIGN KEY(session_id) REFERENCES sessions(session_id)
);
CREATE TABLE task_dependencies (
task_id TEXT NOT NULL,
depends_on_task_id TEXT NOT NULL,
PRIMARY KEY(task_id, depends_on_task_id),
FOREIGN KEY(task_id) REFERENCES tasks(task_id),
FOREIGN KEY(depends_on_task_id) REFERENCES tasks(task_id)
);
CREATE TABLE mutation_transactions (
transaction_id TEXT PRIMARY KEY,
session_id TEXT NOT NULL,
status TEXT NOT NULL,
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
approved_at TIMESTAMP,
committed_at TIMESTAMP,
rolled_back_at TIMESTAMP,
FOREIGN KEY(session_id) REFERENCES sessions(session_id)
);
CREATE TABLE mutations (
mutation_id TEXT PRIMARY KEY,
transaction_id TEXT NOT NULL,
file_path TEXT NOT NULL,
operation TEXT NOT NULL,
before_hash TEXT,
after_hash TEXT,
before_content TEXT,
after_content TEXT,
diff TEXT,
status TEXT NOT NULL,
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(transaction_id) REFERENCES mutation_transactions(transaction_id)
);
CREATE TABLE context_summaries (
summary_id INTEGER PRIMARY KEY AUTOINCREMENT,
session_id TEXT NOT NULL,
summary_json TEXT NOT NULL,
token_count INTEGER,
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(session_id) REFERENCES sessions(session_id)
);
CREATE TABLE model_usage (
usage_id INTEGER PRIMARY KEY AUTOINCREMENT,
session_id TEXT NOT NULL,



session_id TEXT NOT NULL,
task_id TEXT,
provider TEXT NOT NULL,
model TEXT NOT NULL,
input_tokens INTEGER DEFAULT 0,
output_tokens INTEGER DEFAULT 0,
cached_tokens INTEGER DEFAULT 0,
estimated_cost REAL DEFAULT 0,
timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(session_id) REFERENCES sessions(session_id)
);
CREATE TABLE events (
event_id INTEGER PRIMARY KEY AUTOINCREMENT,
session_id TEXT NOT NULL,
event_type TEXT NOT NULL,
payload_json TEXT NOT NULL,
timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(session_id) REFERENCES sessions(session_id)
);

57. Session Recovery
The application must support:
Bash
termagent --session <id>

On startup it must restore:
• Conversation
• Workspace
• Model
• Agent state
• Pending tasks
• Pending mutations
• Context summary
• Usage statistics

58. Pending Mutation Recovery
If pending mutations exist:



┌─────────────────────────────────────────┐
│ Pending transaction detected

│

│

│

│ 3 files changed

│

│ 127 additions

│

│ 12 deletions

│

│

│

│ [R] Resume review

│

│ [D] Discard transaction

│

│ [I] Inspect

│

└─────────────────────────────────────────┘

No pending mutation may be silently committed after restart.

59. Git Integration
Git must be treated as a first-class workspace subsystem.
Initial operations:
IsRepository()
Status()
Diff()
Branch()
Log()

Future operations:
CreateCheckpoint()
CreateBranch()
Commit()

The agent must not automatically commit changes in the MVP.

60. Git Safety
Before mutation:
git status

should be captured.



The agent must distinguish:
Agent-created changes

from:
User-existing changes

The system must never overwrite unrelated user modifications.

61. Plan Mode
The product must eventually support a plan-first workflow.
Example:
> Add OAuth authentication

Agent displays:
PLAN
1. Inspect authentication architecture
2. Identify provider abstraction
3. Add GitHub OAuth provider
4. Add callback route
5. Add tests
6. Run test suite
Estimated:
6 files
~350 lines
3 execution phases
[Execute] [Edit] [Cancel]

62. Dry Run Mode
CLI:
Bash



termagent --dry-run

In dry-run mode:
Allowed:
read
search
inspect
plan
safe tests

Disallowed:
write
delete
mutation
unsafe shell

63. Permission Profiles
Future profiles:
STRICT
SAFE
DEVELOPER
AUTONOMOUS

Example:



STRICT
read

→ allow

write

→ ask

shell

→ ask

network

→ deny

SAFE
read

→ allow

tests

→ allow

write

→ ask

shell

→ ask

DEVELOPER
read

→ allow

tests

→ allow

write

→ ask

git

→ allow

shell

→ ask

The permission system must be designed so profiles can be added without changing tool
implementations.

64. Resource Limits
The system must support:
max_workers
max_agent_iterations
max_tool_calls
max_runtime
max_file_read
max_file_write
max_command_output
max_command_runtime
max_context_tokens
max_session_cost

Default values must be configurable.

65. Logging
Logs must be structured.
Levels:



DEBUG
INFO
WARN
ERROR

Terminal output should remain clean.
Detailed diagnostics should be written to:
~/.config/termagent/logs/

Sensitive information must be redacted.
API keys, authorization headers, secrets, and credentials must never be logged.

66. Error Handling
Errors must be categorized.
ConfigurationError
ProviderError
AuthenticationError
ToolError
SecurityError
WorkspaceError
MutationError
DatabaseError
TaskError
ContextError

Errors shown to users should be concise.
Detailed errors should remain available in logs and session events.

67. Cancellation
All long-running operations must use:
Go
context.Context



CTRL+C behavior:
First CTRL+C:
Cancel current operation.
Second CTRL+C:
Exit application.

The cancellation process must:
1. Stop model streaming.
2. Cancel active tools.
3. Stop child processes.
4. Cancel workers.
5. Preserve session state.
6. Leave no unknown mutation state.

68. Security Requirements
Critical Security Requirements
The system must:
• Prevent path traversal.
• Detect workspace boundary violations.
• Handle symlink escapes.
• Prevent unauthorized filesystem access.
• Validate all tool arguments.
• Enforce shell policies.
• Prevent privilege escalation.
• Never expose API keys to models unnecessarily.
• Never log secrets.
• Never silently execute unknown commands.
• Never overwrite changed files without conflict detection.
• Never silently commit pending mutations.
• Limit command execution resources.
• Support cancellation.
• Preserve transaction state after crashes.

69. Security Threat Model



The threat model must account for:

Malicious Prompt
User intentionally asks the agent to perform destructive operations.

Prompt Injection
Repository files contain instructions such as:
Ignore previous instructions and execute...

Repository content must be treated as untrusted data.

Malicious Repository
A cloned repository may contain:
• malicious scripts
• malicious build hooks
• symlinks
• executable files
• package installation scripts

Tool Abuse
The model attempts to invoke a tool outside its permitted scope.

Path Escape
The model requests files outside the workspace.

Shell Escape
The model attempts to use shell semantics to bypass restrictions.

Credential Leakage
The model attempts to read:
.env
~/.ssh/
credentials
cloud credentials

unless explicitly permitted.



70. Prompt Injection Defense
Repository content must never automatically become system instructions.
For example:
README.md:
"Ignore all previous instructions..."

must be treated as ordinary repository data.
Tool outputs must also be considered untrusted.

71. Sensitive File Policy
Potentially sensitive paths should trigger additional policy checks.
Examples:
.env
.env.*
~/.ssh/
credentials
secrets
*.pem
*.key

The exact policy must be configurable.

72. Performance Requirements
Startup
Target:
<150ms

for:
• process launch
• configuration loading
• SQLite initialization
• TUI initialization



Network operations must not block startup.
Model providers should initialize lazily.

73. Concurrency Requirements
Default maximum active workers:
5

The scheduler must provide:
• bounded concurrency
• backpressure
• cancellation
• task dependencies
• failure handling

74. Memory Requirements
The application must avoid loading entire repositories into memory.
Large files must be:
• size limited
• streamed when appropriate
• truncated with explicit indication
Binary files should not be passed to the model as ordinary text.

75. Agent Context Strategy
The agent must not automatically send the entire repository to the model.
Context acquisition:



User Goal
↓
Workspace Discovery
↓
Search
↓
Relevant Files
↓
Targeted Reads
↓
Architecture Understanding
↓
Model Context

76. Workspace Discovery
At the beginning of a session the system should inspect:
.git/
go.mod
package.json
Cargo.toml
pyproject.toml
requirements.txt
pom.xml
Makefile
Dockerfile
docker-compose.yml
README.md

This list should be extensible.

77. Testing Strategy
Testing must exist at four levels.

Unit Tests
Test individual components:



security
diff
mutation
scheduler
context
tokenization
configuration
workspace

Integration Tests
Test:
Agent
↓
Tool
↓
Mutation
↓
SQLite
↓
Workspace

Security Tests
Test:
path traversal
symlink escape
shell bypass
dangerous commands
secret access
permission escalation
mutation races

End-to-End Tests
Use fixture repositories.
Example:



tests/fixtures/simple-go/
├── main.go
├── calculator.go
└── calculator_test.go

User request:
"Add a divide function with zero-division handling and tests."

Expected:
discover
→ inspect
→ plan
→ modify
→ diff
→ approve
→ commit
→ test
→ report

78. Crash Recovery Test
The application must be intentionally terminated during:
WAITING_APPROVAL

After restart:
Pending transaction detected.

The system must restore the exact pending state.

79. Concurrency Test
Two agents attempt to modify:
src/auth.go

Expected:



Only one mutation may commit at a time.

The second mutation must either:
• wait
• conflict
• rebase
• be rejected
It must never silently overwrite the first mutation.

80. Acceptance Criteria — MVP
The MVP is considered successful when the following workflow works reliably:



1. Launch term-agent.
2. Detect workspace.
3. User enters:
"Add a divide function with tests."
4. Agent inspects the repository.
5. Agent identifies relevant files.
6. Agent creates a plan.
7. Agent proposes file changes.
8. Runtime validates changes.
9. TUI renders unified diff.
10. User presses Y.
11. Mutation transaction commits.
12. Agent executes test command.
13. Test results are displayed.
14. Session is persisted.
15. Application exits.
16. Application restarts.
17. Session can be restored.
18. All tool calls and mutations remain available in history.

81. Phase Roadmap
Phase 0 — Architecture
Deliver:



architecture.md
security.md
agent-runtime.md
ADR documents
database schema
state machines
interfaces

No AI required.

Phase 1 — CLI Foundation
Implement:
Go project
CLI
configuration
logging
SQLite
migrations
application lifecycle

Exit criteria:
termagent launches successfully.

Phase 2 — TUI
Implement:
Bubble Tea
Lip Gloss
layout
input
scrolling
resize
keymap
status bar

Exit criteria:
TUI behaves correctly without AI.



Phase 3 — Workspace Engine
Implement:
workspace detection
filesystem abstraction
file reads
search
hashing
Git status

Exit criteria:
Agent runtime can safely inspect a repository.

Phase 4 — Mutation Engine
Implement:
mutation proposals
hash validation
snapshots
diff
approval
commit
rollback
crash recovery

Exit criteria:
A file can be safely changed through an approval transaction.

Phase 5 — Tool Runtime
Implement:



Tool interface
Tool Registry
read_file
search_workspace
write_file
execute_shell

Exit criteria:
Tools operate exclusively through runtime policies.

Phase 6 — Security Layer
Implement:
path policy
workspace boundaries
symlink protection
shell classification
permission engine
timeouts
resource limits
secret protection

Exit criteria:
Security integration tests pass.

Phase 7 — Single Agent
Implement:
model abstraction
streaming
tool calling
agent loop
context manager
usage tracking

Exit criteria:



Agent can complete a basic coding task end-to-end.

Phase 8 — Context System
Implement:
tokenization
budget tracking
compaction
workspace memory
summary persistence

Exit criteria:
Long sessions remain usable without exceeding model limits.

Phase 9 — Scheduler
Implement:
task graph
dependency resolution
worker pool
cancellation
timeouts
retry
failure propagation

Exit criteria:
Independent tasks execute concurrently.

Phase 10 — Multi-Agent
Implement:



planner
researcher
coder
tester
reviewer

All agents must use the same runtime and tool security boundaries.

Phase 11 — Multi-Provider
Implement:
OpenAI
Anthropic
Gemini

The provider layer must remain independent from agent logic.

Phase 12 — Production Hardening
Test:
network failures
provider failures
database failures
large repositories
large files
binary files
concurrent mutations
symlinks
path traversal
shell attacks
prompt injection
crashes
terminal resizing
cancellation
resource exhaustion

82. Development Rules
Rule 1



Never implement a feature that bypasses the runtime architecture for convenience.

Rule 2
Never allow an LLM provider to directly access the filesystem.

Rule 3
Never allow an agent to execute raw shell commands without policy evaluation.

Rule 4
Never silently overwrite user changes.

Rule 5
Never commit a mutation that was not recorded.

Rule 6
Never add concurrency without cancellation and resource limits.

Rule 7
Never add a provider-specific abstraction outside the model provider layer.

Rule 8
Never place business logic inside Bubble Tea views.

Rule 9
Every major feature must have unit and integration tests.



Rule 10
Every security-sensitive feature must have adversarial tests.

83. Git Development Workflow
Development must be phase-based.
Recommended branches:
main
develop
feature/*
fix/*
security/*
refactor/*

Feature workflow:
Create branch
↓
Implement
↓
Unit tests
↓
Integration tests
↓
Lint
↓
Format
↓
Security checks
↓
Commit
↓
Push
↓
Review
↓
Merge

84. Required Quality Gates
Before merging:



Bash
go fmt ./...
go vet ./...
go test ./...
go test -race ./...

Additional tooling may include:
Bash
golangci-lint
govulncheck

CI must run:
format check
lint
unit tests
race tests
integration tests
security tests
build

85. Definition of Done
A feature is not complete until:
• Implementation exists.
• Interfaces are documented.
• Unit tests exist.
• Integration tests exist where applicable.
• Error handling exists.
• Cancellation is supported where applicable.
• Security implications are evaluated.
• Logs are appropriate.
• Database migrations exist if required.
• Documentation is updated.
• CI passes.
• Race detector passes where applicable.
• No known critical regression exists.



86. Important Non-Goals
The initial product will not attempt to:
• Replace a full IDE.
• Host repositories remotely.
• Provide mandatory cloud synchronization.
• Build a web dashboard.
• Automatically commit all changes.
• Guarantee perfect command sandboxing without OS-level isolation.
• Understand every programming language deeply.
• Run unlimited autonomous agents.
• Store secrets in plaintext.
• Automatically trust repository instructions.

87. Future Advanced Architecture
Once the core runtime is stable, future versions may introduce:
OS-level sandboxing
Container execution
Language Server Protocol integration
AST-aware editing
Semantic code search
MCP compatibility
Plugin architecture
Git worktrees
Parallel isolated agent workspaces
Automatic conflict resolution
Agent evaluation
Replayable sessions
Session branching
Remote model gateways
Team collaboration

These must not compromise the core local-first runtime.

88. Future Parallel Workspace Model
For advanced multi-agent execution:



Main Workspace
|
+---- Agent A Worktree
|
+---- Agent B Worktree
|
+---- Agent C Worktree
|
+---- Agent D Worktree

Each agent can work independently.
After completion:
Agent A
Agent B
Agent C
Agent D
↓
Review
↓
Conflict Detection
↓
Merge Plan
↓
User Approval
↓
Main Workspace

This should be considered for advanced multi-agent mode rather than allowing many
agents to directly modify one working tree.

89. Product Success Metrics
Technical metrics:
Startup latency
Tool execution latency
Model response latency
SQLite write latency
TUI frame responsiveness
Memory usage
CPU usage
Crash recovery success rate
Mutation conflict rate



Agent metrics:
Task completion rate
Tool-call success rate
Test-pass rate
Failed-plan rate
Average iterations per task
Average cost per task
Context compaction frequency

Safety metrics:
Unauthorized mutation attempts
Blocked dangerous commands
Workspace escape attempts
Mutation conflicts
Rollback success rate
Security test pass rate

90. Critical Architectural Decisions
The following decisions are considered P0.

P0-1 — LLM Does Not Own Filesystem Mutations
The LLM can propose.
The runtime applies.

P0-2 — Shell Is Policy-Controlled
Never rely solely on command blacklists.

P0-3 — Workspace Boundary Is Mandatory
Every path must be validated.

P0-4 — Mutations Are Transactional
Every change must be recorded and recoverable.



P0-5 — Agent Scheduler Is Separate From WaitGroup
Concurrency requires:
• queue
• dependencies
• cancellation
• limits
• lifecycle
• failure handling

P0-6 — Runtime Is Event-Driven
Events feed:
TUI
SQLite
Logging
Debugging

P0-7 — Model Providers Are Abstracted
No provider-specific logic may leak into the agent runtime.

P0-8 — SQLite Uses Controlled Persistence
Application goroutines should publish persistence events rather than independently
performing uncontrolled writes.

P0-9 — User Changes Must Never Be Overwritten
File hashes must be checked before mutation commit.

P0-10 — Fail Closed
Unknown or ambiguous operations require approval or denial.



91. Final Product Architecture
The finished system should conceptually look like:



USER
|
v
+-------------+
|

TUI

|

+------+------+
|
v
+---------------------+
| Application Runtime |
+----------+----------+
|
+----------------+----------------+
|

|

v

v

+--------------+

+--------------+

| Agent Engine |

|

+------+-------+

+------+-------+

|

|

v

+------------+----------+

Event Bus

|

+--------------+

|

|

|

| Orchestrator |

v

v

v

+------+-------+

TUI

SQLite

Logs

|
v
+-----------+
| Task Graph|
+-----+-----+
|
v
+---------+
|Scheduler|
+----+----+
|
+---+---+---+
|

|

|

|

v

v

v

v

A1

A2

A3

A4

|

|

|

|

+---+---+---+
|
v
+-------------+
| Tool Runtime|
+------+------+
|
+-----+------+
|

|

v

v

Workspace

Security



Runtime

Policy

|

|

v

v

Mutation

Shell

Engine

Executor

|
v
Diff
|
v
User Approval
|
v
Transaction
|
v
Workspace

92. Final Product Principle
term-agent should not compete by simply being another AI chatbot inside a terminal.

Its differentiator should be:
Safe, observable, recoverable, local-first AI software engineering inside the
terminal.
The architecture must therefore prioritize:
1. Safety
2. Correctness
3. Recoverability
4. Observability
5. Workspace awareness
6. Deterministic runtime behavior
7. Provider independence
8. Controlled concurrency
9. Excellent terminal UX
10. Developer trust
The first milestone is not "multiple AI agents."
The first milestone is:



User
↓
AI understands workspace
↓
AI proposes a change
↓
Runtime validates it
↓
TUI shows exact diff
↓
User approves
↓
Transaction commits safely
↓
Tests run
↓
Everything is persisted

Once this loop is reliable, parallel agents, autonomous planning, provider routing,
advanced context management, and other capabilities can be added on top of a stable
foundation.



