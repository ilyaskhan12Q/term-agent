package unit

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/config"
)

func TestRedactSecrets_DefaultPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "OpenAI Key Redaction",
			input:    "Connecting to OpenAI with key sk-1234567890abcdef1234567890abcdef",
			expected: "Connecting to OpenAI with key [REDACTED]",
		},
		{
			name:     "GitHub Token Redaction",
			input:    "Exporting token ghp_1234567890abcdef1234567890abcdef to environment",
			expected: "Exporting token [REDACTED] to environment",
		},
		{
			name:     "Bearer Token Redaction",
			input:    "Authorization: Bearer 1234567890abcdef1234567890abcdef",
			expected: "Authorization: [REDACTED]",
		},
		{
			name:     "Normal Text Untouched",
			input:    "This is a normal log message with workspace dir /tmp/project",
			expected: "This is a normal log message with workspace dir /tmp/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.RedactSecrets(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRedactSecrets_DynamicSecretRegistration(t *testing.T) {
	customSecret := "super-secret-user-defined-token-998877"
	config.RegisterSecret(customSecret)

	input := "User authenticated with " + customSecret
	expected := "User authenticated with [REDACTED]"

	result := config.RedactSecrets(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStructuredLogger_RedactsOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := config.SetupLogger("info", &buf)

	logger.Info("User logged in with key sk-1234567890abcdef1234567890abcdef", "token", "ghp_1234567890abcdef1234567890abcdef")

	output := buf.String()
	if strings.Contains(output, "sk-1234567890abcdef") {
		t.Errorf("log output leaked OpenAI key: %s", output)
	}
	if strings.Contains(output, "ghp_1234567890abcdef") {
		t.Errorf("log output leaked GitHub key: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got: %s", output)
	}
}
