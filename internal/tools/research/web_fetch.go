package research

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/security"
	"github.com/ilyaskhan/term-agent/internal/tools"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

type WebFetchArgs struct {
	URI       string `json:"uri"`
	ProjectID string `json:"project_id"`
	MaxBytes  int    `json:"max_bytes,omitempty"`
}

type WebFetchResponse struct {
	Source      domain.Source `json:"source"`
	Title       string        `json:"title"`
	StatusCode  int           `json:"status_code"`
	IsPaywalled bool          `json:"is_paywalled"`
	WrappedText string        `json:"wrapped_text"`
}

type WebFetchTool struct {
	HTTPClient *http.Client
}

func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewWebFetchToolWithClient constructs a WebFetchTool with a custom HTTP client for unit testing.
func NewWebFetchToolWithClient(client *http.Client) *WebFetchTool {
	return &WebFetchTool{HTTPClient: client}
}

func (t *WebFetchTool) Spec() tools.ToolSpec {
	params, _ := json.Marshal(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"uri": map[string]interface{}{
				"type":        "string",
				"description": "Web page URL (HTTP/HTTPS) to fetch and parse",
			},
			"project_id": map[string]interface{}{
				"type":        "string",
				"description": "Research project ID to attach fetched web source to",
			},
			"max_bytes": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum bytes to read from response (default 1,000,000)",
			},
		},
		"required": []string{"uri", "project_id"},
	})

	return tools.ToolSpec{
		Name:        "web_fetch",
		Description: "Fetches and parses web documents, stripping boilerplate and wrapping text in security envelopes to prevent prompt injections.",
		Parameters:  params,
		RiskLevel:   tools.RiskLevelRead,
	}
}

func (t *WebFetchTool) ValidateArgs(args json.RawMessage) error {
	var a WebFetchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return fmt.Errorf("invalid arguments for web_fetch: %w", err)
	}
	if strings.TrimSpace(a.URI) == "" {
		return errors.New("uri is required")
	}
	if strings.TrimSpace(a.ProjectID) == "" {
		return errors.New("project_id is required")
	}

	parsed, err := url.Parse(a.URI)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid URL scheme for web_fetch: '%s' (must be http or https)", a.URI)
	}
	return nil
}

func (t *WebFetchTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	start := time.Now()
	var a WebFetchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return &tools.ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	if a.MaxBytes <= 0 {
		a.MaxBytes = 1000000
	}

	parsedURL, _ := url.Parse(a.URI)
	domainName := parsedURL.Hostname()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URI, nil)
	if err != nil {
		return &tools.ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("failed to create HTTP request: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}
	req.Header.Set("User-Agent", "TermAgent-ResearchEngine/1.0 (Mozilla/5.0 Compatible)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9")

	client := t.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		// Network fetch error -> Return fallback source record indicating fetch failure
		srcID := fmt.Sprintf("src-%s-web", a.ProjectID)
		src, _ := domain.NewSource(srcID, a.ProjectID, fmt.Sprintf("Web Page at %s", domainName), a.URI, domain.SourceTypeWebPage, 0.3)
		src.Publisher = domainName

		res := WebFetchResponse{
			Source:      *src,
			Title:       fmt.Sprintf("Web Page at %s", domainName),
			StatusCode:  0,
			IsPaywalled: false,
			WrappedText: security.WrapUntrustedContent(fmt.Sprintf("Fetch failed: %v", err), "web-fetch-error"),
		}
		outJSON, _ := json.Marshal(res)
		return &tools.ToolResult{
			Output:   string(outJSON),
			Duration: time.Since(start),
			IsError:  false,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		srcID := fmt.Sprintf("src-%s-paywalled", a.ProjectID)
		src, _ := domain.NewSource(srcID, a.ProjectID, fmt.Sprintf("Paywalled Content at %s", domainName), a.URI, domain.SourceTypeWebPage, 0.2)
		src.Publisher = domainName

		res := WebFetchResponse{
			Source:      *src,
			Title:       fmt.Sprintf("Paywalled Content at %s", domainName),
			StatusCode:  resp.StatusCode,
			IsPaywalled: true,
			WrappedText: security.WrapUntrustedContent("Access Restricted: Page requires authentication or payment.", "web-fetch-paywalled"),
		}
		outJSON, _ := json.Marshal(res)
		return &tools.ToolResult{
			Output:   string(outJSON),
			Duration: time.Since(start),
			IsError:  false,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &tools.ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("HTTP request returned status %d", resp.StatusCode),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	// Guard: reject binary / non-text content types to prevent prompt injection via binary blobs.
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		ct := strings.ToLower(contentType)
		if !strings.Contains(ct, "text/") &&
			!strings.Contains(ct, "application/xhtml") &&
			!strings.Contains(ct, "application/json") &&
			!strings.Contains(ct, "application/xml") {
			return &tools.ToolResult{
				Output:   "",
				Error:    fmt.Sprintf("web_fetch: unsupported content-type %q (only text/* content is permitted)", contentType),
				Duration: time.Since(start),
				IsError:  true,
			}, nil
		}
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, int64(a.MaxBytes)))
	if err != nil {
		return &tools.ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("failed to read HTTP body: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	rawContent := string(bodyBytes)
	title := extractHTMLTitle(rawContent, domainName)
	cleanText := stripHTMLTags(rawContent)

	srcID := fmt.Sprintf("src-%s-%d", a.ProjectID, time.Now().UnixNano()%10000)
	trustScore := 0.85
	if strings.HasSuffix(domainName, ".edu") || strings.HasSuffix(domainName, ".gov") || strings.HasSuffix(domainName, ".org") {
		trustScore = 0.95
	}

	src, _ := domain.NewSource(srcID, a.ProjectID, title, a.URI, domain.SourceTypeWebPage, trustScore)
	if src != nil {
		src.Publisher = domainName
		src.Year = time.Now().Year()
	}

	wrappedText := security.WrapUntrustedContent(cleanText, "web-fetch-content")

	res := WebFetchResponse{
		Source:      *src,
		Title:       title,
		StatusCode:  resp.StatusCode,
		IsPaywalled: false,
		WrappedText: wrappedText,
	}

	outJSON, err := json.Marshal(res)
	if err != nil {
		return &tools.ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("failed to serialize web fetch output: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	return &tools.ToolResult{
		Output:   string(outJSON),
		Duration: time.Since(start),
		IsError:  false,
	}, nil
}

func extractHTMLTitle(htmlContent, defaultTitle string) string {
	re := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	match := re.FindStringSubmatch(htmlContent)
	if len(match) > 1 {
		title := strings.TrimSpace(match[1])
		title = security.SanitizeUntrustedInput(title)
		if title != "" {
			return title
		}
	}
	return fmt.Sprintf("Web Document from %s", defaultTitle)
}

func stripHTMLTags(htmlContent string) string {
	// Strip script and style blocks
	reScript := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reStyle := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reComments := regexp.MustCompile(`(?s)<!--.*?-->`)
	reTags := regexp.MustCompile(`<[^>]+>`)
	reSpaces := regexp.MustCompile(`\s+`)

	text := reScript.ReplaceAllString(htmlContent, "")
	text = reStyle.ReplaceAllString(text, "")
	text = reComments.ReplaceAllString(text, "")
	text = reTags.ReplaceAllString(text, " ")
	text = reSpaces.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}
