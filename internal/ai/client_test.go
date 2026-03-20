package ai

import (
	"context"
	"strings"
	"testing"
)

func TestIsOpenAIModel(t *testing.T) {
	cases := []struct {
		model string
		want  bool
	}{
		{model: "gpt-4.1", want: true},
		{model: "o1-mini", want: true},
		{model: "o3", want: true},
		{model: "claude-sonnet-4-6", want: false},
	}

	for _, tc := range cases {
		if got := IsOpenAIModel(tc.model); got != tc.want {
			t.Fatalf("IsOpenAIModel(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestIsAnthropicModel(t *testing.T) {
	cases := []struct {
		model string
		want  bool
	}{
		{model: "claude-sonnet-4-6", want: true},
		{model: "claude-opus-4-6", want: true},
		{model: "us.anthropic.claude-haiku-4-5-20251001-v1:0", want: true},
		{model: "gpt-5.1", want: false},
		{model: "gemini-2.5-flash-lite", want: false},
		{model: "grok-4-fast-reasoning", want: false},
	}

	for _, tc := range cases {
		if got := IsAnthropicModel(tc.model); got != tc.want {
			t.Fatalf("IsAnthropicModel(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestSendValidatesProviderKeys(t *testing.T) {
	c := New()
	msgs := []Message{{Role: "user", Content: "hello"}}

	_, err := c.Send(context.Background(), Config{
		Model: "gpt-4.1",
	}, "prompt", msgs)
	if err == nil || !strings.Contains(err.Error(), "OpenAI API key not set") {
		t.Fatalf("expected missing OpenAI key error, got: %v", err)
	}

	_, err = c.Send(context.Background(), Config{
		Model: "claude-sonnet-4-6",
	}, "prompt", msgs)
	if err == nil || !strings.Contains(err.Error(), "Anthropic API key not set") {
		t.Fatalf("expected missing Anthropic key error, got: %v", err)
	}
}

func TestSendGatewayOnlyModelWithoutGateway(t *testing.T) {
	c := New()
	msgs := []Message{{Role: "user", Content: "hello"}}

	for _, model := range []string{"gemini-2.5-flash-lite", "grok-4-fast-reasoning"} {
		_, err := c.Send(context.Background(), Config{
			Model: model,
		}, "prompt", msgs)
		if err == nil || !strings.Contains(err.Error(), "is not a recognized provider") {
			t.Fatalf("model %q: expected gateway-required error, got: %v", model, err)
		}
	}
}
