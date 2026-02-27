package main

import (
	"context"
	"errors"
	"testing"

	core "lay/internal/app"
)

type fakeService struct {
	started bool
	ctx     context.Context
	notes   string
	cfg     core.Config
	err     error
}

func (f *fakeService) Startup(ctx context.Context) { f.started, f.ctx = true, ctx }
func (f *fakeService) GetNotes() string            { return f.notes }
func (f *fakeService) SaveNotes(_ string) error    { return f.err }
func (f *fakeService) GetConfig() core.Config      { return f.cfg }
func (f *fakeService) SaveConfig(_, _, _ string) error {
	return f.err
}
func (f *fakeService) SendMessage(_ string) (string, error)   { return "ok", f.err }
func (f *fakeService) StartRecording() (string, error)        { return "/tmp/r", f.err }
func (f *fakeService) StopRecording() error                   { return f.err }
func (f *fakeService) Transcribe(_ string) (string, error)    { return "tx", f.err }
func (f *fakeService) AppendTranscriptToNotes(_ string) error { return f.err }

func TestAppWrapperDelegates(t *testing.T) {
	f := &fakeService{
		notes: "N",
		cfg:   core.Config{Model: "claude-sonnet-4-6"},
	}
	a := &App{service: f}

	ctx := context.WithValue(context.Background(), "k", "v")
	a.startup(ctx)
	if !f.started || f.ctx != ctx {
		t.Fatalf("startup should delegate context to service")
	}

	if got := a.GetNotes(); got != "N" {
		t.Fatalf("GetNotes() = %q, want %q", got, "N")
	}
	if got := a.GetConfig(); got.Model != "claude-sonnet-4-6" {
		t.Fatalf("GetConfig().Model = %q", got.Model)
	}
}

func TestAppWrapperPropagatesErrors(t *testing.T) {
	expected := errors.New("boom")
	a := &App{service: &fakeService{err: expected}}

	if err := a.SaveNotes("x"); !errors.Is(err, expected) {
		t.Fatalf("SaveNotes() error = %v, want %v", err, expected)
	}
	if _, err := a.SendMessage("[]"); !errors.Is(err, expected) {
		t.Fatalf("SendMessage() error = %v, want %v", err, expected)
	}
}
