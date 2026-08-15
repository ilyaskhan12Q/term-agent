package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ansiRegex matches standard ANSI color and terminal escape sequences.
	ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b[()][A-Z0-9]`)

	// injectionTagRegex matches raw model system delimiters used in prompt injection attacks.
	injectionTagRegex = regexp.MustCompile(`(?i)<system>|</system>|<instructions>|</instructions>|\[INST\]|\[/INST\]|<<SYS>>|</<SYS>>|<\|im_start\|>|<\|im_end\|>`)
)

// SanitizeUntrustedInput strips ANSI escape sequences, non-printable control characters,
// and neutralizes raw system prompt hijacking tags.
func SanitizeUntrustedInput(input string) string {
	if input == "" {
		return ""
	}

	// 1. Strip ANSI escape codes
	cleaned := ansiRegex.ReplaceAllString(input, "")

	// 2. Strip non-printable control characters (preserving \n, \t, \r)
	var sb strings.Builder
	sb.Grow(len(cleaned))
	for _, r := range cleaned {
		if r == '\n' || r == '\t' || r == '\r' || (r >= 32 && r != 127) {
			sb.WriteRune(r)
		}
	}
	cleaned = sb.String()

	// 3. Neutralize system prompt injection delimiters by replacing brackets/chevrons with safe unicode homoglyphs/escapes
	cleaned = injectionTagRegex.ReplaceAllStringFunc(cleaned, func(match string) string {
		matchLower := strings.ToLower(match)
		switch {
		case strings.Contains(matchLower, "system"):
			return "[system_tag_neutralized]"
		case strings.Contains(matchLower, "instructions"):
			return "[instructions_tag_neutralized]"
		case strings.Contains(matchLower, "inst"):
			return "[inst_tag_neutralized]"
		case strings.Contains(matchLower, "sys"):
			return "[sys_tag_neutralized]"
		case strings.Contains(matchLower, "im_start"):
			return "[im_start_tag_neutralized]"
		case strings.Contains(matchLower, "im_end"):
			return "[im_end_tag_neutralized]"
		default:
			return "[injection_tag_neutralized]"
		}
	})

	return cleaned
}

// WrapUntrustedContent sanitizes content from external/untrusted sources (shell outputs, file content, web results)
// and encapsulates it within an immutable <untrusted_content> envelope with a SHA-256 hash.
func WrapUntrustedContent(source string, content string) string {
	sanitized := SanitizeUntrustedInput(content)

	hasher := sha256.New()
	hasher.Write([]byte(sanitized))
	hashStr := hex.EncodeToString(hasher.Sum(nil))[:16]

	cleanSource := strings.ReplaceAll(source, `"`, `'`)
	return fmt.Sprintf("<untrusted_content source=\"%s\" hash=\"%s\">\n%s\n</untrusted_content>", cleanSource, hashStr, sanitized)
}

// ValidateContextBoundary inspects a assembled model prompt string to ensure no un-escaped prompt injection attempt
// attempts to hijack system-level instructions. Returns an error if an active injection attack pattern is detected.
func ValidateContextBoundary(prompt string) error {
	// Detect raw attempt to inject system prompt boundaries inside context
	matches := injectionTagRegex.FindAllString(prompt, -1)
	if len(matches) > 0 {
		return fmt.Errorf("context boundary validation failure: prompt contains %d un-neutralized prompt injection delimiters (e.g., '%s')", len(matches), matches[0])
	}
	return nil
}
