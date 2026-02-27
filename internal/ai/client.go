package ai

import (
	"context"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	openai "github.com/sashabaranov/go-openai"
)

type Config struct {
	AnthropicKey string
	OpenAIKey    string
	Model        string
}

type Message struct {
	Role    string
	Content string
}

type Client struct{}

func New() *Client {
	return &Client{}
}

func IsOpenAIModel(model string) bool {
	return strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3")
}

func (c *Client) Send(ctx context.Context, cfg Config, systemPrompt string, messages []Message) (string, error) {
	if IsOpenAIModel(cfg.Model) {
		return c.sendOpenAI(ctx, cfg, systemPrompt, messages)
	}
	return c.sendAnthropic(ctx, cfg, systemPrompt, messages)
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
			apiMessages = append(apiMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
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
		msgs = append(msgs, openai.ChatCompletionMessage{Role: role, Content: m.Content})
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
