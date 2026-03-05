package cmd

import (
	"strings"
	"testing"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/rainoffallingstar/codexswitch/internal/types"
)

func TestRunSwitch_WithSlug_Succeeds(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := store.SaveProvider(types.Provider{
		Slug:            "openai",
		DisplayName:     "OpenAI",
		APIKey:          "sk-test",
		Model:           "gpt-5",
		BaseURL:         "https://api.openai.com/v1",
		WireAPI:         "responses",
		ReasoningEffort: "medium",
	}); err != nil {
		t.Fatalf("SaveProvider() error = %v", err)
	}

	switchSlug = "openai"
	defer func() { switchSlug = "" }()

	if err := runSwitch(nil, nil); err != nil {
		t.Fatalf("runSwitch() error = %v", err)
	}

	current, err := store.GetCurrentSlug()
	if err != nil {
		t.Fatalf("GetCurrentSlug() error = %v", err)
	}
	if current != "openai" {
		t.Fatalf("current slug = %q, want %q", current, "openai")
	}
}

func TestRunSwitch_WithUnknownSlug_ReturnsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := store.SaveProvider(types.Provider{
		Slug:            "openai",
		DisplayName:     "OpenAI",
		APIKey:          "sk-test",
		Model:           "gpt-5",
		BaseURL:         "https://api.openai.com/v1",
		WireAPI:         "responses",
		ReasoningEffort: "medium",
	}); err != nil {
		t.Fatalf("SaveProvider() error = %v", err)
	}

	switchSlug = "missing"
	defer func() { switchSlug = "" }()

	err := runSwitch(nil, nil)
	if err == nil {
		t.Fatalf("runSwitch() expected error for missing slug")
	}
	if !strings.Contains(err.Error(), "provider not found: missing") {
		t.Fatalf("runSwitch() error = %v, want provider not found", err)
	}
}

func TestResolveProviderTarget_NonInteractiveMissingSlug_ReturnsError(t *testing.T) {
	list := types.ProviderList{
		Providers: []types.Provider{
			{Slug: "openai", DisplayName: "OpenAI"},
		},
	}

	_, err := resolveProviderTarget(list, "", false, "Select provider to switch")
	if err == nil {
		t.Fatalf("resolveProviderTarget() expected missing --slug error")
	}
	if !strings.Contains(err.Error(), "missing required option: --slug") {
		t.Fatalf("resolveProviderTarget() error = %v, want missing --slug", err)
	}
}

func TestResolveProviderTarget_InvalidSlug_ReturnsError(t *testing.T) {
	list := types.ProviderList{
		Providers: []types.Provider{
			{Slug: "openai", DisplayName: "OpenAI"},
		},
	}

	_, err := resolveProviderTarget(list, "../bad", false, "Select provider to switch")
	if err == nil {
		t.Fatalf("resolveProviderTarget() expected invalid slug error")
	}
	if !strings.Contains(err.Error(), "invalid provider slug") {
		t.Fatalf("resolveProviderTarget() error = %v, want invalid provider slug", err)
	}
}
