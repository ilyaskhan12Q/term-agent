// Package openai implements a term-agent ModelProvider for the OpenAI Chat Completions API.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ilyaskhan/term-agent/internal/model"
)

const (
	defaultBaseURL = "https://api.openai.com/v1/chat/completions"
	defaultTimeout = 120 * time.Second
	providerName   = "openai"
)

// Provider implements model.ModelProvider for OpenAI (and OpenAI-compatible endpoints).
type Provider struct {
	apiKey       string
	baseURL      string
	client       *http.Client
	nameOverride string
}

// NewProvider constructs a Provider. apiKey must be non-empty.
func NewProvider(apiKey string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai: OPENAI_API_KEY is required")
	}
	return &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

// NewProviderWithURL constructs a Provider with a custom base URL, intended for unit testing or custom endpoints.
func NewProviderWithURL(apiKey, baseURL string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai: OPENAI_API_KEY is required")
	}
	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

// SetName overrides the default provider name (e.g. for openrouter).
func (p *Provider) SetName(name string) {
	p.nameOverride = name
}

func (p *Provider) Name() string {
	if p.nameOverride != "" {
		return p.nameOverride
	}
	return providerName
}

func (p *Provider) Capabilities() model.ModelCapabilities {
	return model.ModelCapabilities{
		SupportsNativeToolCalling: true,
		SupportsStreaming:         false,
		SupportsSystemPrompt:      true,
		MaxContextWindow:          128000,
	}
}

// chatMessage is the OpenAI API message format.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the OpenAI Chat Completions request body.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// chatChoice is a single completion choice.
type chatChoice struct {
	Message chatMessage `json:"message"`
}

// usageInfo contains token usage from the API.
type usageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// chatResponse is the OpenAI Chat Completions response body.
type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Model   string       `json:"model"`
	Usage   usageInfo    `json:"usage"`
	Error   *apiError    `json:"error,omitempty"`
}

// apiError is the OpenAI error response body.
type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// GenerateCompletion sends a chat completion request to the OpenAI API.
func (p *Provider) GenerateCompletion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	messages := make([]chatMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, chatMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	payload := chatRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("openai: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai: failed to build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai: failed to read response body: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("openai: failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("openai API error [%s]: %s", chatResp.Error.Type, chatResp.Error.Message)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai: unexpected HTTP status %d", resp.StatusCode)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no completion choices returned")
	}

	return &model.CompletionResponse{
		Content: chatResp.Choices[0].Message.Content,
		Model:   chatResp.Model,
		Usage: model.TokenUsage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
	}, nil
}
