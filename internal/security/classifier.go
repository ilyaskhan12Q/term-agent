package security

// CommandRiskLevel categorizes command execution risk.
type CommandRiskLevel string

const (
	CommandRiskAllowed      CommandRiskLevel = "ALLOWED"       // Safe read-only commands (ls, git status)
	CommandRiskRequiresUser CommandRiskLevel = "REQUIRES_USER" // Modifying/building commands (go test, git commit)
	CommandRiskBlocked      CommandRiskLevel = "BLOCKED"       // Dangerous/prohibited commands (rm -rf /, sudo, curl | sh)
)

// Classification contains the security classification of a proposed shell command.
type Classification struct {
	Command        string
	RiskLevel      CommandRiskLevel
	Category       string
	RequiresPrompt bool
	Reason         string
}

// Classifier defines the contract for command classification.
type Classifier interface {
	Classify(command string) (*Classification, error)
}
