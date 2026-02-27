package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"lay/internal/ai"
)

type App struct {
	ctx               context.Context
	aiClient          *ai.Client
	currentTranscript string
	liveCancel        context.CancelFunc
	liveChunkSeq      int
	liveSegments      []string
	liveMu            sync.Mutex
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

func New() *App {
	return &App{aiClient: ai.New()}
}

func (a *App) Startup(ctx context.Context) {
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

func (a *App) SendMessage(conversationJSON string) (string, error) {
	cfg := a.GetConfig()

	var messages []Message
	if err := json.Unmarshal([]byte(conversationJSON), &messages); err != nil {
		return "", fmt.Errorf("invalid conversation format: %w", err)
	}

	aiCfg := ai.Config{
		AnthropicKey: cfg.AnthropicKey,
		OpenAIKey:    cfg.OpenAIKey,
		Model:        cfg.Model,
	}

	aiMessages := make([]ai.Message, 0, len(messages))
	for _, m := range messages {
		aiMessages = append(aiMessages, ai.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	return a.aiClient.Send(context.Background(), aiCfg, a.systemPrompt(), aiMessages)
}

func (a *App) systemPrompt() string {
	base := "You are a helpful meeting assistant. Be concise and practical. Format responses in markdown when it aids clarity."

	if a.currentTranscript != "" {
		return base + "\n\nThe user has a meeting transcript from this session. Use it to answer questions about the meeting.\n\n<transcript>\n" + a.currentTranscript + "\n</transcript>"
	}

	a.liveMu.Lock()
	live := strings.Join(a.liveSegments, "\n")
	a.liveMu.Unlock()

	if live == "" {
		return base
	}
	return base + "\n\nThe meeting is currently being recorded. Below is the live transcript so far — it may be incomplete.\n\n<transcript>\n" + live + "\n</transcript>"
}
