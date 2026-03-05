package ui

import (
	"strings"
	"testing"

	"github.com/rainoffallingstar/codexswitch/internal/types"
)

func TestPromptAddProvider_DefaultReasoningEffortInNonInteractive(t *testing.T) {
	input, err := PromptAddProvider(types.AddProviderInput{
		Slug:        "openai",
		DisplayName: "OpenAI",
		APIKey:      "sk-test",
		Model:       "gpt-5",
		BaseURL:     "https://api.openai.com/v1",
	}, false)
	if err != nil {
		t.Fatalf("PromptAddProvider() error = %v", err)
	}
	if input.ReasoningEffort != types.DefaultReasoningEffort {
		t.Fatalf("ReasoningEffort = %q, want %q", input.ReasoningEffort, types.DefaultReasoningEffort)
	}
}

func TestPromptAddProvider_RejectsInvalidReasoningEffort(t *testing.T) {
	_, err := PromptAddProvider(types.AddProviderInput{
		Slug:            "openai",
		DisplayName:     "OpenAI",
		APIKey:          "sk-test",
		Model:           "gpt-5",
		BaseURL:         "https://api.openai.com/v1",
		ReasoningEffort: "ultra",
	}, false)
	if err == nil {
		t.Fatalf("PromptAddProvider() expected error for invalid reasoning effort")
	}
	if !strings.Contains(err.Error(), "invalid reasoning effort") {
		t.Fatalf("PromptAddProvider() error = %v, want invalid reasoning effort message", err)
	}
}

func TestPromptEditProvider_AppliesValidReasoningEffort(t *testing.T) {
	seed := types.Provider{
		Slug:            "openai",
		DisplayName:     "OpenAI",
		APIKey:          "sk-test",
		Model:           "gpt-5",
		BaseURL:         "https://api.openai.com/v1",
		WireAPI:         "responses",
		ReasoningEffort: "medium",
	}

	out, err := PromptEditProvider(seed, types.AddProviderInput{ReasoningEffort: "xhigh"}, false)
	if err != nil {
		t.Fatalf("PromptEditProvider() error = %v", err)
	}
	if out.ReasoningEffort != "xhigh" {
		t.Fatalf("ReasoningEffort = %q, want %q", out.ReasoningEffort, "xhigh")
	}
}

func TestPromptEditProvider_RejectsInvalidReasoningEffort(t *testing.T) {
	seed := types.Provider{
		Slug:        "openai",
		DisplayName: "OpenAI",
		APIKey:      "sk-test",
		Model:       "gpt-5",
		BaseURL:     "https://api.openai.com/v1",
		WireAPI:     "responses",
	}

	_, err := PromptEditProvider(seed, types.AddProviderInput{ReasoningEffort: "ultra"}, false)
	if err == nil {
		t.Fatalf("PromptEditProvider() expected error for invalid reasoning effort")
	}
	if !strings.Contains(err.Error(), "invalid reasoning effort") {
		t.Fatalf("PromptEditProvider() error = %v, want invalid reasoning effort message", err)
	}
}
