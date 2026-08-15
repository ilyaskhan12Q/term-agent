// Package anthropic implements a term-agent ModelProvider for the Anthropic Messages API.
package anthropic

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
	defaultBaseURL   = "https://api.anthropic.com/v1/messages"
	defaultTimeout   = 120 * time.Second
	providerName     = "anthropic"
	anthropicVersion = "2023-06-01"
)

// Provider implements model.ModelProvider for Anthropic.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewProvider constructs a Provider. apiKey must be non-empty.
func NewProvider(apiKey string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic: ANTHROPIC_API_KEY is required")
	}
	return &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

// NewProviderWithURL constructs a Provider with a custom base URL, intended for unit testing.
func NewProviderWithURL(apiKey, baseURL string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic: ANTHROPIC_API_KEY is required")
	}
	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

func (p *Provider) Name() string { return providerName }

func (p *Provider) Capabilities() model.ModelCapabilities {
	return model.ModelCapabilities{
		SupportsNativeToolCalling: true,
		SupportsStreaming:         false,
		SupportsSystemPrompt:      true,
		MaxContextWindow:          200000,
	}
}

// anthropicMessage is the Anthropic API message format.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messagesRequest is the Anthropic Messages request body.
type messagesRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

// contentBlock is a content block in the Anthropic response.
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// usageInfo contains token usage from the Anthropic API.
type usageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// messagesResponse is the Anthropic Messages response body.
type messagesResponse struct {
	Content []contentBlock `json:"content"`
	Model   string         `json:"model"`
	Usage   usageInfo      `json:"usage"`
	Error   *apiError      `json:"error,omitempty"`
}

// apiError is the Anthropic error response body.
type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// GenerateCompletion sends a messages request to the Anthropic API.
func (p *Provider) GenerateCompletion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	var systemPrompt string
	messages := make([]anthropicMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == model.RoleSystem {
			systemPrompt = m.Content
			continue
		}
		messages = append(messages, anthropicMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	payload := messagesRequest{
		Model:     req.Model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  messages,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to read response body: %w", err)
	}

	var msgResp messagesResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return nil, fmt.Errorf("anthropic: failed to parse response: %w", err)
	}

	if msgResp.Error != nil {
		return nil, fmt.Errorf("anthropic API error [%s]: %s", msgResp.Error.Type, msgResp.Error.Message)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("anthropic: unexpected HTTP status %d", resp.StatusCode)
	}

	var content string
	for _, block := range msgResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &model.CompletionResponse{
		Content: content,
		Model:   msgResp.Model,
		Usage: model.TokenUsage{
			PromptTokens:     msgResp.Usage.InputTokens,
			CompletionTokens: msgResp.Usage.OutputTokens,
			TotalTokens:      msgResp.Usage.InputTokens + msgResp.Usage.OutputTokens,
		},
	}, nil
}
