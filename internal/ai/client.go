package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	openai "github.com/sashabaranov/go-openai"
)

type Config struct {
	AnthropicKey   string
	OpenAIKey      string
	Model          string
	GatewayURL string // if set, all requests are routed through a gateway
}

type Message struct {
	Role    string
	Content string
	Images  []string // base64-encoded image data (PNG)
}

type Client struct{}

func New() *Client {
	return &Client{}
}

func IsOpenAIModel(model string) bool {
	return strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4")
}

func IsAnthropicModel(model string) bool {
	return strings.HasPrefix(model, "claude-") ||
		strings.HasPrefix(model, "us.anthropic.")
}

func (c *Client) Send(ctx context.Context, cfg Config, systemPrompt string, messages []Message) (string, error) {
	if cfg.GatewayURL != "" {
		return c.sendGateway(ctx, cfg, systemPrompt, messages)
	}
	if IsOpenAIModel(cfg.Model) {
		return c.sendOpenAI(ctx, cfg, systemPrompt, messages)
	}
	if IsAnthropicModel(cfg.Model) {
		return c.sendAnthropic(ctx, cfg, systemPrompt, messages)
	}
	return "", fmt.Errorf("model %q is not a recognized provider — set up a gateway to use it", cfg.Model)
}

func (c *Client) sendAnthropic(ctx context.Context, cfg Config, systemPrompt string, messages []Message) (string, error) {
	if cfg.AnthropicKey == "" {
		return "", fmt.Errorf("Anthropic API key not set — open Settings to add your key")
	}

	client := anthropic.NewClient(option.WithAPIKey(cfg.AnthropicKey))

	var apiMessages []anthropic.MessageParam
	for _, m := range messages {
		switch m.Role {
		case "user":
			var blocks []anthropic.ContentBlockParamUnion
			for _, img := range m.Images {
				blocks = append(blocks, anthropic.NewImageBlockBase64("image/png", img))
			}
			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}
			apiMessages = append(apiMessages, anthropic.NewUserMessage(blocks...))
		case "assistant":
			apiMessages = append(apiMessages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(cfg.Model),
		MaxTokens: 2048,
		Messages:  apiMessages,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("Anthropic API error: %w", err)
	}
	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("no text content in response")
}

func (c *Client) sendOpenAI(ctx context.Context, cfg Config, systemPrompt string, messages []Message) (string, error) {
	if cfg.OpenAIKey == "" {
		return "", fmt.Errorf("OpenAI API key not set — open Settings to add your key")
	}

	client := openai.NewClient(cfg.OpenAIKey)

	var msgs []openai.ChatCompletionMessage
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	})
	for _, m := range messages {
		role := openai.ChatMessageRoleUser
		if m.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		if len(m.Images) > 0 && m.Role == "user" {
			var parts []openai.ChatMessagePart
			for _, img := range m.Images {
				parts = append(parts, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: "data:image/png;base64," + img,
					},
				})
			}
			if m.Content != "" {
				parts = append(parts, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeText,
					Text: m.Content,
				})
			}
			msgs = append(msgs, openai.ChatCompletionMessage{Role: role, MultiContent: parts})
		} else {
			msgs = append(msgs, openai.ChatCompletionMessage{Role: role, Content: m.Content})
		}
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    cfg.Model,
		Messages: msgs,
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}
	return resp.Choices[0].Message.Content, nil
}

// gatewayMessage is a message in the Chat Completions API format.
type gatewayMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []gatewayContentPart
}

// gatewayContentPart represents a content block in the gateway format.
type gatewayContentPart struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	Image    *gatewayImage    `json:"image,omitempty"`    // Anthropic models
	ImageURL *gatewayImageURL `json:"image_url,omitempty"` // OpenAI models
}

type gatewayImage struct {
	File     string `json:"file"`
	MimeType string `json:"mimeType"`
}

type gatewayImageURL struct {
	URL string `json:"url"`
}

// gatewayParams holds model parameters sent inside the request body.
type gatewayParams struct {
	MaxOutputTokens int `json:"max_output_tokens"`
}

// gatewayRequest is the body sent to the gateway endpoint.
type gatewayRequest struct {
	Model      string           `json:"model"`
	Messages   []gatewayMessage `json:"messages"`
	System     string           `json:"system,omitempty"`
	Parameters gatewayParams    `json:"parameters"`
}

// gatewayResponse handles both Anthropic and OpenAI Chat Completions response shapes.
type gatewayResponse struct {
	// Anthropic shape
	Content []gatewayContentBlock `json:"content"`
	// OpenAI Chat Completions shape
	Choices []gatewayChoice `json:"choices"`
	// Common
	Error *gatewayErrorPayload `json:"error,omitempty"`
}

type gatewayContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type gatewayChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type gatewayErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (c *Client) sendGateway(ctx context.Context, cfg Config, systemPrompt string, messages []Message) (string, error) {
	isOpenAI := IsOpenAIModel(cfg.Model)

	gwMessages := make([]gatewayMessage, 0, len(messages)+1)

	// OpenAI models expect system prompt as a message; Anthropic uses top-level field.
	if isOpenAI && systemPrompt != "" {
		gwMessages = append(gwMessages, gatewayMessage{
			Role:    "system",
			Content: []gatewayContentPart{{Type: "text", Text: systemPrompt}},
		})
	}

	for _, m := range messages {
		var parts []gatewayContentPart
		if len(m.Images) > 0 && m.Role == "user" {
			for _, img := range m.Images {
				if isOpenAI {
					parts = append(parts, gatewayContentPart{
						Type:     "image_url",
						ImageURL: &gatewayImageURL{URL: "data:image/png;base64," + img},
					})
				} else {
					parts = append(parts, gatewayContentPart{
						Type: "image",
						Image: &gatewayImage{
							File:     img,
							MimeType: "image/png",
						},
					})
				}
			}
		}
		if m.Content != "" {
			parts = append(parts, gatewayContentPart{Type: "text", Text: m.Content})
		}
		gwMessages = append(gwMessages, gatewayMessage{Role: m.Role, Content: parts})
	}

	reqBody := gatewayRequest{
		Model:    cfg.Model,
		Messages: gwMessages,
		Parameters: gatewayParams{
			MaxOutputTokens: 20000,
		},
	}
	if !isOpenAI {
		reqBody.System = systemPrompt
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode gateway request: %w", err)
	}

	url := cfg.GatewayURL
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create gateway request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gateway request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read gateway response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gateway returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var gwResp gatewayResponse
	if err := json.Unmarshal(respBytes, &gwResp); err != nil {
		return "", fmt.Errorf("failed to parse gateway response: %w", err)
	}

	if gwResp.Error != nil {
		return "", fmt.Errorf("gateway error: %s", gwResp.Error.Message)
	}

	// OpenAI Chat Completions response
	if len(gwResp.Choices) > 0 {
		return gwResp.Choices[0].Message.Content, nil
	}

	// Anthropic-style response
	for _, block := range gwResp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in gateway response: %s", string(respBytes))
}
