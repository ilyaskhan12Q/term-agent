package unit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/model"
	"github.com/ilyaskhan/term-agent/internal/model/anthropic"
	"github.com/ilyaskhan/term-agent/internal/model/bootstrap"
	"github.com/ilyaskhan/term-agent/internal/model/gemini"
	"github.com/ilyaskhan/term-agent/internal/model/openai"
)

// ---------------------------------------------------------------------------
// ProviderConfig.Validate tests
// ---------------------------------------------------------------------------

func TestProviderConfig_Validate_MissingProvider(t *testing.T) {
	cfg := model.ProviderConfig{Model: "gpt-4o", APIKey: "sk-test"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing provider, got nil")
	}
}

func TestProviderConfig_Validate_UnknownProvider(t *testing.T) {
	cfg := model.ProviderConfig{ProviderName: "fakeai", Model: "model-x", APIKey: "key"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

func TestProviderConfig_Validate_MissingModel(t *testing.T) {
	cfg := model.ProviderConfig{ProviderName: "openai", APIKey: "sk-test"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing model, got nil")
	}
}

func TestProviderConfig_Validate_MissingAPIKey(t *testing.T) {
	cfg := model.ProviderConfig{ProviderName: "anthropic", Model: "claude-3-5-sonnet-20241022"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing API key, got nil")
	}
	// Error message should tell user which environment variable to set.
	if !containsString(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("expected error to mention ANTHROPIC_API_KEY, got: %s", err.Error())
	}
}

func TestProviderConfig_Validate_ValidConfig(t *testing.T) {
	for _, provider := range model.SupportedProviders {
		cfg := model.ProviderConfig{
			ProviderName: provider,
			Model:        "test-model",
			APIKey:       "test-key",
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("provider %q: unexpected validation error: %v", provider, err)
		}
	}
}

// ---------------------------------------------------------------------------
// bootstrap.Init and BuildProvider tests
// ---------------------------------------------------------------------------

func TestBootstrap_Init_RegistersFactory(t *testing.T) {
	bootstrap.Init()
	if model.DefaultFactory == nil {
		t.Fatal("DefaultFactory should be non-nil after bootstrap.Init()")
	}
}

func TestBootstrap_BuildProvider_MissingKey(t *testing.T) {
	bootstrap.Init()
	_, err := bootstrap.BuildProvider("openai", "gpt-4o", "")
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestBootstrap_BuildProvider_UnknownProvider(t *testing.T) {
	bootstrap.Init()
	_, err := bootstrap.BuildProvider("unknown", "model", "key")
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

func TestBootstrap_BuildProvider_AllSupportedProviders(t *testing.T) {
	bootstrap.Init()
	for _, p := range model.SupportedProviders {
		provider, err := bootstrap.BuildProvider(p, "test-model", "test-key")
		if err != nil {
			t.Errorf("provider %q: unexpected error: %v", p, err)
			continue
		}
		if provider.Name() != p {
			t.Errorf("provider %q: Name() returned %q", p, provider.Name())
		}
	}
}

// ---------------------------------------------------------------------------
// OpenAI provider HTTP unit test (test server)
// ---------------------------------------------------------------------------

func TestOpenAIProvider_GenerateCompletion_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"role": "assistant", "content": "Test response from OpenAI."}},
			},
			"model": "gpt-4o",
			"usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := openai.NewProviderWithURL("test-key", srv.URL)
	if err != nil {
		t.Fatalf("NewProviderWithURL: %v", err)
	}

	req := &model.CompletionRequest{
		Model:    "gpt-4o",
		Messages: []model.Message{{Role: model.RoleUser, Content: "Hello"}},
	}
	resp, err := p.GenerateCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateCompletion: %v", err)
	}
	if resp.Content != "Test response from OpenAI." {
		t.Errorf("unexpected content: %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("unexpected total tokens: %d", resp.Usage.TotalTokens)
	}
}

func TestOpenAIProvider_GenerateCompletion_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]interface{}{
			"error": map[string]string{"type": "invalid_api_key", "message": "Invalid API key."},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, _ := openai.NewProviderWithURL("bad-key", srv.URL)
	req := &model.CompletionRequest{
		Model:    "gpt-4o",
		Messages: []model.Message{{Role: model.RoleUser, Content: "Hello"}},
	}
	_, err := p.GenerateCompletion(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for API error response, got nil")
	}
}

// ---------------------------------------------------------------------------
// Anthropic provider HTTP unit test (test server)
// ---------------------------------------------------------------------------

func TestAnthropicProvider_GenerateCompletion_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			t.Error("missing x-api-key header")
		}
		resp := map[string]interface{}{
			"content": []map[string]string{{"type": "text", "text": "Test response from Anthropic."}},
			"model":   "claude-3-5-sonnet-20241022",
			"usage":   map[string]int{"input_tokens": 10, "output_tokens": 7},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := anthropic.NewProviderWithURL("test-key", srv.URL)
	if err != nil {
		t.Fatalf("NewProviderWithURL: %v", err)
	}

	req := &model.CompletionRequest{
		Model:    "claude-3-5-sonnet-20241022",
		Messages: []model.Message{{Role: model.RoleUser, Content: "Hello"}},
	}
	resp, err := p.GenerateCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateCompletion: %v", err)
	}
	if resp.Content != "Test response from Anthropic." {
		t.Errorf("unexpected content: %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 17 {
		t.Errorf("unexpected total tokens: %d", resp.Usage.TotalTokens)
	}
}

// ---------------------------------------------------------------------------
// Gemini provider HTTP unit test (test server)
// ---------------------------------------------------------------------------

func TestGeminiProvider_GenerateCompletion_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{"content": map[string]interface{}{
					"role":  "model",
					"parts": []map[string]string{{"text": "Test response from Gemini."}},
				}},
			},
			"usageMetadata": map[string]int{
				"promptTokenCount":     12,
				"candidatesTokenCount": 6,
				"totalTokenCount":      18,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := gemini.NewProviderWithURL("test-key", srv.URL)
	if err != nil {
		t.Fatalf("NewProviderWithURL: %v", err)
	}

	req := &model.CompletionRequest{
		Model:    "gemini-1.5-pro",
		Messages: []model.Message{{Role: model.RoleUser, Content: "Hello"}},
	}
	resp, err := p.GenerateCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateCompletion: %v", err)
	}
	if resp.Content != "Test response from Gemini." {
		t.Errorf("unexpected content: %q", resp.Content)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
