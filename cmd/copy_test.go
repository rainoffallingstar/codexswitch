package cmd

import (
	"strings"
	"testing"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/rainoffallingstar/codexswitch/internal/types"
)

func TestRunCopy_WithSlug_Succeeds(t *testing.T) {
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

	copySlug = "openai"
	defer func() { copySlug = "" }()

	if err := runCopy(nil, nil); err != nil {
		t.Fatalf("runCopy() error = %v", err)
	}

	cloned, err := store.FindProviderBySlug("openai-copy1")
	if err != nil {
		t.Fatalf("FindProviderBySlug(openai-copy1) error = %v", err)
	}
	if cloned.DisplayName != "OpenAI copy1" {
		t.Fatalf("cloned display name = %q, want %q", cloned.DisplayName, "OpenAI copy1")
	}
}

func TestRunCopy_WithUnknownSlug_ReturnsError(t *testing.T) {
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

	copySlug = "missing"
	defer func() { copySlug = "" }()

	err := runCopy(nil, nil)
	if err == nil {
		t.Fatalf("runCopy() expected error for missing slug")
	}
	if !strings.Contains(err.Error(), "provider not found: missing") {
		t.Fatalf("runCopy() error = %v, want provider not found", err)
	}
}

func TestResolveProviderTarget_ForCopy_NonInteractiveMissingSlug_ReturnsError(t *testing.T) {
	list := types.ProviderList{
		Providers: []types.Provider{
			{Slug: "openai", DisplayName: "OpenAI"},
		},
	}

	_, err := resolveProviderTarget(list, "", false, "Select provider to copy")
	if err == nil {
		t.Fatalf("resolveProviderTarget() expected missing --slug error")
	}
	if !strings.Contains(err.Error(), "missing required option: --slug") {
		t.Fatalf("resolveProviderTarget() error = %v, want missing --slug", err)
	}
}
