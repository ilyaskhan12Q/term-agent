// Package gemini implements a term-agent ModelProvider for the Google Gemini generateContent API.
package gemini

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
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"
	defaultTimeout = 120 * time.Second
	providerName   = "gemini"
)

// Provider implements model.ModelProvider for Google Gemini.
type Provider struct {
	apiKey  string
	testURL string // non-empty overrides URL template; used in unit tests only
	client  *http.Client
}

// NewProvider constructs a Provider. apiKey must be non-empty.
func NewProvider(apiKey string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini: GEMINI_API_KEY is required")
	}
	return &Provider{
		apiKey: apiKey,
		client: &http.Client{Timeout: defaultTimeout},
	}, nil
}

// NewProviderWithURL constructs a Provider with a fixed base URL, intended for unit testing.
func NewProviderWithURL(apiKey, testURL string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini: GEMINI_API_KEY is required")
	}
	return &Provider{
		apiKey:  apiKey,
		testURL: testURL,
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

func (p *Provider) Name() string { return providerName }

func (p *Provider) Capabilities() model.ModelCapabilities {
	return model.ModelCapabilities{
		SupportsNativeToolCalling: true,
		SupportsStreaming:         false,
		SupportsSystemPrompt:      true,
		MaxContextWindow:          1000000,
	}
}

// geminiPart is a part within a content block.
type geminiPart struct {
	Text string `json:"text"`
}

// geminiContent is a single content block in the Gemini API.
type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

// generateRequest is the Gemini generateContent request body.
type generateRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
}

// candidate is a single response candidate from Gemini.
type candidate struct {
	Content geminiContent `json:"content"`
}

// usageMetadata contains token usage from the Gemini API.
type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// generateResponse is the Gemini generateContent response body.
type generateResponse struct {
	Candidates    []candidate   `json:"candidates"`
	UsageMetadata usageMetadata `json:"usageMetadata"`
	Error         *apiError     `json:"error,omitempty"`
}

// apiError is the Gemini error response body.
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// GenerateCompletion sends a generateContent request to the Gemini API.
func (p *Provider) GenerateCompletion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	var systemInstruction *geminiContent
	contents := make([]geminiContent, 0, len(req.Messages))

	for _, m := range req.Messages {
		if m.Role == model.RoleSystem {
			systemInstruction = &geminiContent{
				Parts: []geminiPart{{Text: m.Content}},
			}
			continue
		}

		role := "user"
		if m.Role == model.RoleAssistant {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	payload := generateRequest{
		Contents:          contents,
		SystemInstruction: systemInstruction,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to marshal request: %w", err)
	}

	url := fmt.Sprintf(defaultBaseURL, req.Model, p.apiKey)
	if p.testURL != "" {
		url = p.testURL
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to read response body: %w", err)
	}

	var genResp generateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return nil, fmt.Errorf("gemini: failed to parse response: %w", err)
	}

	if genResp.Error != nil {
		return nil, fmt.Errorf("gemini API error [%s]: %s", genResp.Error.Status, genResp.Error.Message)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini: unexpected HTTP status %d", resp.StatusCode)
	}

	if len(genResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: no candidates returned")
	}

	var content string
	for _, part := range genResp.Candidates[0].Content.Parts {
		content += part.Text
	}

	return &model.CompletionResponse{
		Content: content,
		Model:   req.Model,
		Usage: model.TokenUsage{
			PromptTokens:     genResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: genResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      genResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}
