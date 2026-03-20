package main

import (
	"context"

	core "lay/internal/app"
)

type appService interface {
	Startup(ctx context.Context)
	GetNotes() string
	SaveNotes(content string) error
	GetConfig() core.Config
	GetGatewayConfig() *core.GatewayConfig
	SaveConfig(anthropicKey string, openAIKey string, model string, gatewayURL string, transcribeLang string) error
	SendMessage(conversationJSON string) (string, error)
	StartRecording() (string, error)
	StopRecording() error
	Transcribe(recordingDir string) (string, error)
	StartMicOnlyRecording() (string, error)
	TranscribeMicOnly(recordingDir string) (string, error)
	AppendTranscriptToNotes(recordingDir string) error
	ExportToFile(content string, path string) error
	GetHomePath() string
}

type App struct {
	service appService
}

func NewApp() *App {
	return &App{service: core.New()}
}

func (a *App) startup(ctx context.Context) {
	a.service.Startup(ctx)
}

func (a *App) GetNotes() string {
	return a.service.GetNotes()
}

func (a *App) SaveNotes(content string) error {
	return a.service.SaveNotes(content)
}

func (a *App) GetConfig() core.Config {
	return a.service.GetConfig()
}

func (a *App) GetGatewayConfig() *core.GatewayConfig {
	return a.service.GetGatewayConfig()
}

func (a *App) SaveConfig(anthropicKey string, openAIKey string, model string, gatewayURL string, transcribeLang string) error {
	return a.service.SaveConfig(anthropicKey, openAIKey, model, gatewayURL, transcribeLang)
}

func (a *App) SendMessage(conversationJSON string) (string, error) {
	return a.service.SendMessage(conversationJSON)
}

func (a *App) StartRecording() (string, error) {
	return a.service.StartRecording()
}

func (a *App) StopRecording() error {
	return a.service.StopRecording()
}

func (a *App) Transcribe(recordingDir string) (string, error) {
	return a.service.Transcribe(recordingDir)
}

func (a *App) StartMicOnlyRecording() (string, error) {
	return a.service.StartMicOnlyRecording()
}

func (a *App) TranscribeMicOnly(recordingDir string) (string, error) {
	return a.service.TranscribeMicOnly(recordingDir)
}

func (a *App) AppendTranscriptToNotes(recordingDir string) error {
	return a.service.AppendTranscriptToNotes(recordingDir)
}

func (a *App) ExportToFile(content string, path string) error {
	return a.service.ExportToFile(content, path)
}

func (a *App) GetHomePath() string {
	return a.service.GetHomePath()
}
