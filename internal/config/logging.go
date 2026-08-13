package config

import (
	"context"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	// Default patterns matching common API keys and tokens.
	defaultRedactRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(sk-proj-[a-zA-Z0-9_-]{20,})`),
		regexp.MustCompile(`(?i)(sk-[a-zA-Z0-9]{20,})`),
		regexp.MustCompile(`(?i)(ghp_[a-zA-Z0-9]{20,})`),
		regexp.MustCompile(`(?i)(gho_[a-zA-Z0-9]{20,})`),
		regexp.MustCompile(`(?i)(glpat-[a-zA-Z0-9_-]{20,})`),
		regexp.MustCompile(`(?i)(xai-[a-zA-Z0-9]{20,})`),
		regexp.MustCompile(`(?i)(Bearer\s+[a-zA-Z0-9._-]{20,})`),
		regexp.MustCompile(`(?i)(AIzaSy[a-zA-Z0-9_-]{33})`),
	}

	registeredSecrets []string
	secretMutex       sync.RWMutex
)

// RegisterSecret registers a dynamic secret string (e.g. an loaded API key) to be redacted from logs.
func RegisterSecret(secret string) {
	if strings.TrimSpace(secret) == "" || len(secret) < 4 {
		return
	}
	secretMutex.Lock()
	defer secretMutex.Unlock()
	for _, existing := range registeredSecrets {
		if existing == secret {
			return
		}
	}
	registeredSecrets = append(registeredSecrets, secret)
}

// RedactSecrets replaces sensitive patterns and registered secret strings with [REDACTED].
func RedactSecrets(input string) string {
	if input == "" {
		return ""
	}

	result := input

	// 1. Match regex patterns
	for _, re := range defaultRedactRegexes {
		result = re.ReplaceAllString(result, "[REDACTED]")
	}

	// 2. Match exact registered secrets
	secretMutex.RLock()
	defer secretMutex.RUnlock()
	for _, secret := range registeredSecrets {
		if secret != "" {
			result = strings.ReplaceAll(result, secret, "[REDACTED]")
		}
	}

	return result
}

// RedactingHandler is an slog.Handler wrapper that redacts secrets in log entries.
type RedactingHandler struct {
	slog.Handler
}

// NewRedactingHandler creates a new slog.Handler that redacts secret patterns.
func NewRedactingHandler(h slog.Handler) *RedactingHandler {
	return &RedactingHandler{Handler: h}
}

func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	newRecord := slog.NewRecord(r.Time, r.Level, RedactSecrets(r.Message), r.PC)

	r.Attrs(func(a slog.Attr) bool {
		newRecord.AddAttrs(redactAttr(a))
		return true
	})

	return h.Handler.Handle(ctx, newRecord)
}

func redactAttr(a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindString {
		return slog.String(a.Key, RedactSecrets(a.Value.String()))
	}
	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		newAttrs := make([]slog.Attr, len(attrs))
		for i, gAttr := range attrs {
			newAttrs[i] = redactAttr(gAttr)
		}
		return slog.Group(a.Key, convertToAny(newAttrs)...)
	}
	return a
}

func convertToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, a := range attrs {
		result[i] = a
	}
	return result
}

// SetupLogger initializes the global slog logger with structured JSON or Text output and secret redaction.
func SetupLogger(levelStr string, w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stderr
	}

	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	baseHandler := slog.NewJSONHandler(w, opts)
	redactingHandler := NewRedactingHandler(baseHandler)
	logger := slog.New(redactingHandler)

	slog.SetDefault(logger)
	return logger
}
