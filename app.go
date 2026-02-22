package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// App struct holds runtime state.
type App struct {
	ctx context.Context
}

// Config holds user settings persisted to ~/.lay/config.json.
type Config struct {
	APIKey string `json:"apiKey"`
	Model  string `json:"model"`
}

// Message is a single chat turn for the Claude API.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup saves the Wails context for later use.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Ensure ~/.lay/ storage directory exists.
	_ = os.MkdirAll(layDir(), 0o755)
}

// layDir returns the path to the ~/.lay storage directory.
func layDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".lay")
}

// ---------- Notes ----------

// GetNotes reads notes from ~/.lay/notes.md. Returns empty string if not found.
func (a *App) GetNotes() string {
	data, err := os.ReadFile(filepath.Join(layDir(), "notes.md"))
	if err != nil {
		return ""
	}
	return string(data)
}

// SaveNotes writes notes to ~/.lay/notes.md.
func (a *App) SaveNotes(content string) error {
	return os.WriteFile(filepath.Join(layDir(), "notes.md"), []byte(content), 0o644)
}

// ---------- Config ----------

// GetConfig reads ~/.lay/config.json. Returns defaults if not found.
func (a *App) GetConfig() Config {
	data, err := os.ReadFile(filepath.Join(layDir(), "config.json"))
	if err != nil {
		return Config{Model: "claude-sonnet-4-6"}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{Model: "claude-sonnet-4-6"}
	}
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-6"
	}
	return cfg
}

// SaveConfig persists API key and model selection to ~/.lay/config.json.
func (a *App) SaveConfig(apiKey string, model string) error {
	cfg := Config{APIKey: apiKey, Model: model}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(layDir(), "config.json"), data, 0o600)
}

// ---------- AI Chat ----------

// SendMessage sends a conversation to Claude and returns the assistant reply.
// conversationJSON is a JSON array of {role, content} objects.
func (a *App) SendMessage(conversationJSON string) (string, error) {
	cfg := a.GetConfig()
	if cfg.APIKey == "" {
		return "", fmt.Errorf("API key not set — open Settings to add your Anthropic API key")
	}

	var messages []Message
	if err := json.Unmarshal([]byte(conversationJSON), &messages); err != nil {
		return "", fmt.Errorf("invalid conversation format: %w", err)
	}

	client := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))

	// Build Anthropic message params.
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
			{Text: "You are a helpful meeting assistant. Be concise and practical."},
		},
	})
	if err != nil {
		return "", fmt.Errorf("Claude API error: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	// Extract text from the first content block.
	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in response")
}
