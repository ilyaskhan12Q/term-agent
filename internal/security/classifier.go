package security

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// CommandRiskLevel categorizes command execution risk.
type CommandRiskLevel string

const (
	CommandRiskAllowed      CommandRiskLevel = "ALLOWED"       // Safe read-only commands (ls, git status)
	CommandRiskRequiresUser CommandRiskLevel = "REQUIRES_USER" // Modifying/building commands (go test, git commit)
	CommandRiskBlocked      CommandRiskLevel = "BLOCKED"       // Dangerous/prohibited commands (rm -rf /, sudo, curl | sh)
)

// Classification contains the security classification of a proposed shell command.
type Classification struct {
	Command        string           `json:"command"`
	RiskLevel      CommandRiskLevel `json:"risk_level"`
	Category       string           `json:"category"`
	RequiresPrompt bool             `json:"requires_prompt"`
	Reason         string           `json:"reason"`
}

// Classifier defines the contract for command classification.
type Classifier interface {
	Classify(command string) (*Classification, error)
}

// POSIXClassifier evaluates shell commands using a POSIX AST parser.
type POSIXClassifier struct{}

// NewPOSIXClassifier constructs a new POSIX AST command classifier.
func NewPOSIXClassifier() *POSIXClassifier {
	return &POSIXClassifier{}
}

// Blocked binaries that are never allowed under any circumstances.
var blockedBinaries = map[string]bool{
	"sudo":      true,
	"su":        true,
	"doas":      true,
	"dd":        true,
	"mkfs":      true,
	"mkfs.ext2": true,
	"mkfs.ext3": true,
	"mkfs.ext4": true,
	"mkfs.vfat": true,
	"shutdown":  true,
	"reboot":    true,
	"init":      true,
	"halt":      true,
	"poweroff":  true,
	"useradd":   true,
	"usermod":   true,
	"userdel":   true,
	"groupadd":  true,
	"iptables":  true,
	"ufw":       true,
}

// Allowed read-only tools that default to ALLOWED unless dangerous flags/subshells are used.
var allowedReadOnlyCommands = map[string]bool{
	"ls":       true,
	"pwd":      true,
	"whoami":   true,
	"date":     true,
	"echo":     true,
	"cat":      true,
	"head":     true,
	"tail":     true,
	"grep":     true,
	"which":    true,
	"whereis":  true,
	"wc":       true,
	"diff":     true,
	"env":      true,
	"printenv": true,
	"basename": true,
	"dirname":  true,
	"stat":     true,
	"file":     true,
	"git":      true, // checked specifically for subcommands (status, log, diff vs commit, push)
	"go":       true, // checked specifically (version, doc vs test, build)
	"node":     true, // checked (-v, --version vs execution)
	"npm":      true, // checked (list, outdated vs install, run)
	"python":   true, // checked (-v, --version vs execution)
	"python3":  true,
}

// Sensitive file patterns that trigger BLOCKED status if accessed or modified.
var sensitivePaths = []string{
	".env",
	".env.local",
	".env.production",
	".env.staging",
	".env.development",
	"id_rsa",
	"id_ed25519",
	"id_dsa",
	"id_ecdsa",
	"/etc/shadow",
	"/etc/sudoers",
	"/etc/passwd",
	"~/.ssh",
	"~/.aws",
	"~/.gnupg",
}

// Classify parses the command using a POSIX AST parser and assigns a risk level.
func (c *POSIXClassifier) Classify(command string) (*Classification, error) {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return &Classification{
			Command:        command,
			RiskLevel:      CommandRiskAllowed,
			Category:       "EMPTY",
			RequiresPrompt: false,
			Reason:         "Empty command",
		}, nil
	}

	parser := syntax.NewParser(syntax.Variant(syntax.LangPOSIX))
	file, err := parser.Parse(strings.NewReader(trimmed), "")
	if err != nil {
		return &Classification{
			Command:        command,
			RiskLevel:      CommandRiskBlocked,
			Category:       "SYNTAX_ERROR",
			RequiresPrompt: true,
			Reason:         fmt.Sprintf("Invalid shell syntax: %v", err),
		}, nil
	}

	var hasSubshell bool
	var hasCmdSubst bool
	var hasPipeline bool
	var hasRedirection bool
	var blockedReason string
	var requiresUserReason string
	isBlocked := false
	requiresUser := false

	// Check raw text for quick sensitive path check
	for _, sensitive := range sensitivePaths {
		if containsPathToken(trimmed, sensitive) {
			isBlocked = true
			blockedReason = fmt.Sprintf("Access to sensitive path detected: %s", sensitive)
			break
		}
	}

	// Track commands in pipeline to catch curl ... | sh
	var pipelineCommands []string

	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil || isBlocked {
			return false
		}

		switch n := node.(type) {
		case *syntax.Subshell:
			hasSubshell = true
			requiresUser = true
			requiresUserReason = "Subshell execution detected"

		case *syntax.CmdSubst:
			hasCmdSubst = true
			requiresUser = true
			requiresUserReason = "Command substitution $(...) or `...` detected"

		case *syntax.BinaryCmd:
			if n.Op == syntax.Pipe || n.Op == syntax.PipeAll {
				hasPipeline = true
				requiresUser = true
				if requiresUserReason == "" {
					requiresUserReason = "Pipelined command execution"
				}
			} else if n.Op == syntax.AndStmt || n.Op == syntax.OrStmt {
				requiresUser = true
				if requiresUserReason == "" {
					requiresUserReason = "Chained command execution (&& / ||)"
				}
			}

		case *syntax.Redirect:
			hasRedirection = true
			requiresUser = true
			if requiresUserReason == "" {
				requiresUserReason = "File I/O output redirection (> or >>)"
			}

		case *syntax.CallExpr:
			if len(n.Args) == 0 {
				return true
			}

			cmdWord := getWordLiteral(n.Args[0])
			pipelineCommands = append(pipelineCommands, cmdWord)

			// 1. Check blocked binary
			if blockedBinaries[cmdWord] {
				isBlocked = true
				blockedReason = fmt.Sprintf("Use of dangerous binary '%s' is prohibited", cmdWord)
				return false
			}

			// 2. Check dangerous recursive rm
			if cmdWord == "rm" {
				var hasRecursive bool
				var targetsRoot bool
				for _, arg := range n.Args[1:] {
					argVal := getWordLiteral(arg)
					if strings.Contains(argVal, "r") || strings.Contains(argVal, "f") {
						if strings.HasPrefix(argVal, "-") {
							if strings.Contains(argVal, "r") || strings.Contains(argVal, "R") {
								hasRecursive = true
							}
						}
					}
					if argVal == "/" || argVal == "/*" || argVal == "~" || argVal == "~/" || argVal == "*" || argVal == ".." {
						targetsRoot = true
					}
				}
				if hasRecursive && targetsRoot {
					isBlocked = true
					blockedReason = "Dangerous recursive rm targeting root/home/wildcard directory"
					return false
				}
				requiresUser = true
				requiresUserReason = "File removal command ('rm')"
			}

			// 3. Check chmod / chown
			if cmdWord == "chmod" || cmdWord == "chown" {
				for _, arg := range n.Args[1:] {
					argVal := getWordLiteral(arg)
					if argVal == "777" || argVal == "a+rwx" || argVal == "-R" {
						requiresUser = true
						requiresUserReason = fmt.Sprintf("Permission modification via '%s'", cmdWord)
					}
				}
			}

			// 4. Check git subcommands
			if cmdWord == "git" && len(n.Args) > 1 {
				subCmd := getWordLiteral(n.Args[1])
				if subCmd != "status" && subCmd != "log" && subCmd != "diff" && subCmd != "branch" && subCmd != "show" {
					requiresUser = true
					requiresUserReason = fmt.Sprintf("Modifying git command ('git %s')", subCmd)
				}
			}

			// 5. Check go subcommands
			if cmdWord == "go" && len(n.Args) > 1 {
				subCmd := getWordLiteral(n.Args[1])
				if subCmd != "version" && subCmd != "env" && subCmd != "doc" {
					requiresUser = true
					requiresUserReason = fmt.Sprintf("Go workspace modification/execution ('go %s')", subCmd)
				}
			}

			// 6. Check npm subcommands
			if cmdWord == "npm" && len(n.Args) > 1 {
				subCmd := getWordLiteral(n.Args[1])
				if subCmd != "list" && subCmd != "outdated" && subCmd != "view" && subCmd != "version" {
					requiresUser = true
					requiresUserReason = fmt.Sprintf("NPM package modification/execution ('npm %s')", subCmd)
				}
			}

			// 7. Check general non-read-only commands
			if !allowedReadOnlyCommands[cmdWord] {
				requiresUser = true
				if requiresUserReason == "" {
					requiresUserReason = fmt.Sprintf("System command '%s' requires user confirmation", cmdWord)
				}
			}
		}

		return true
	})

	// Check pipeline for curl ... | sh pattern
	if hasPipeline && len(pipelineCommands) >= 2 {
		first := pipelineCommands[0]
		last := pipelineCommands[len(pipelineCommands)-1]
		if (first == "curl" || first == "wget" || first == "fetch") &&
			(last == "sh" || last == "bash" || last == "zsh" || last == "python" || last == "perl") {
			isBlocked = true
			blockedReason = fmt.Sprintf("Piping remote fetch ('%s') directly to shell interpreter ('%s') is blocked", first, last)
		}
	}

	if isBlocked {
		return &Classification{
			Command:        command,
			RiskLevel:      CommandRiskBlocked,
			Category:       "BLOCKED_SECURITY_VIOLATION",
			RequiresPrompt: true,
			Reason:         blockedReason,
		}, nil
	}

	if requiresUser || hasSubshell || hasCmdSubst || hasPipeline || hasRedirection {
		return &Classification{
			Command:        command,
			RiskLevel:      CommandRiskRequiresUser,
			Category:       "REQUIRES_USER_APPROVAL",
			RequiresPrompt: true,
			Reason:         requiresUserReason,
		}, nil
	}

	return &Classification{
		Command:        command,
		RiskLevel:      CommandRiskAllowed,
		Category:       "READ_ONLY_SAFE",
		RequiresPrompt: false,
		Reason:         "Safe read-only execution",
	}, nil
}

// getWordLiteral extracts string literal value from AST word.
func getWordLiteral(word *syntax.Word) string {
	if word == nil {
		return ""
	}
	var buf bytes.Buffer
	printer := syntax.NewPrinter()
	if err := printer.Print(&buf, word); err == nil {
		s := buf.String()
		s = strings.Trim(s, "'\"")
		return s
	}
	return ""
}

// containsPathToken checks if text contains a reference to a sensitive path.
func containsPathToken(text, target string) bool {
	lower := strings.ToLower(text)
	targetLower := strings.ToLower(target)
	if strings.Contains(lower, targetLower) {
		return true
	}
	base := filepath.Base(targetLower)
	words := strings.Fields(lower)
	for _, w := range words {
		if w == base || strings.HasSuffix(w, "/"+base) || strings.HasSuffix(w, "\\"+base) {
			return true
		}
	}
	return false
}
