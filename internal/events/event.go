package events

import (
	"time"
)

// EventType identifies system-wide event types.
type EventType string

const (
	EventAgentThought       EventType = "AGENT_THOUGHT"
	EventToolCallProposed   EventType = "TOOL_CALL_PROPOSED"
	EventToolCallApproved   EventType = "TOOL_CALL_APPROVED"
	EventToolCallRejected   EventType = "TOOL_CALL_REJECTED"
	EventToolCallExecuted   EventType = "TOOL_CALL_EXECUTED"
	EventMutationStaged     EventType = "MUTATION_STAGED"
	EventMutationCommitted  EventType = "MUTATION_COMMITTED"
	EventMutationRolledBack EventType = "MUTATION_ROLLEDBACK"
	EventContextCompacted   EventType = "CONTEXT_COMPACTED"
	EventTaskStatusChanged  EventType = "TASK_STATUS_CHANGED"
	EventErrorOccurred      EventType = "ERROR_OCCURRED"
)

// Event represents an immutable internal system event.
type Event struct {
	ID        string
	Type      EventType
	SessionID string
	Payload   interface{}
	Timestamp time.Time
}
