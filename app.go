package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	openai "github.com/sashabaranov/go-openai"
)

type App struct {
	ctx               context.Context
	currentTranscript string
}

type Config struct {
	AnthropicKey string `json:"anthropicKey"`
	OpenAIKey    string `json:"openaiKey"`
	Model        string `json:"model"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = os.MkdirAll(layDir(), 0o755)
}

func layDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".lay")
}

func (a *App) GetNotes() string {
	data, err := os.ReadFile(filepath.Join(layDir(), "notes.md"))
	if err != nil {
		return ""
	}
	return string(data)
}

func (a *App) SaveNotes(content string) error {
	return os.WriteFile(filepath.Join(layDir(), "notes.md"), []byte(content), 0o644)
}

func (a *App) GetConfig() Config {
	data, err := os.ReadFile(filepath.Join(layDir(), "config.json"))
	if err != nil {
		return Config{Model: "claude-sonnet-4-6"}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{Model: "claude-sonnet-4-6"}
	}
	if cfg.AnthropicKey == "" {
		var raw map[string]string
		if json.Unmarshal(data, &raw) == nil {
			if v := raw["apiKey"]; v != "" {
				cfg.AnthropicKey = v
			}
		}
	}
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-6"
	}
	return cfg
}

func (a *App) SaveConfig(anthropicKey string, openAIKey string, model string) error {
	cfg := Config{AnthropicKey: anthropicKey, OpenAIKey: openAIKey, Model: model}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(layDir(), "config.json"), data, 0o600)
}

func isOpenAIModel(model string) bool {
	return strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3")
}

func (a *App) SendMessage(conversationJSON string) (string, error) {
	cfg := a.GetConfig()

	var messages []Message
	if err := json.Unmarshal([]byte(conversationJSON), &messages); err != nil {
		return "", fmt.Errorf("invalid conversation format: %w", err)
	}

	if isOpenAIModel(cfg.Model) {
		return a.sendOpenAI(cfg, messages)
	}
	return a.sendAnthropic(cfg, messages)
}

func (a *App) systemPrompt() string {
	base := "You are a helpful meeting assistant. Be concise and practical. Format responses in markdown when it aids clarity."
	if a.currentTranscript == "" {
		return base
	}
	return base + "\n\nThe user has a meeting transcript from this session. Use it to answer questions about the meeting.\n\n<transcript>\n" + a.currentTranscript + "\n</transcript>"
}

func (a *App) sendAnthropic(cfg Config, messages []Message) (string, error) {
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

	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.Model(cfg.Model),
		MaxTokens: 2048,
		Messages:  apiMessages,
		System: []anthropic.TextBlockParam{
			{Text: a.systemPrompt()},
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

func (a *App) sendOpenAI(cfg Config, messages []Message) (string, error) {
	if cfg.OpenAIKey == "" {
		return "", fmt.Errorf("OpenAI API key not set — open Settings to add your key")
	}

	client := openai.NewClient(cfg.OpenAIKey)

	var msgs []openai.ChatCompletionMessage
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: a.systemPrompt(),
	})
	for _, m := range messages {
		role := openai.ChatMessageRoleUser
		if m.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		msgs = append(msgs, openai.ChatCompletionMessage{Role: role, Content: m.Content})
	}

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
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
