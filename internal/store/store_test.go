package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rainoffallingstar/codexswitch/internal/types"
)

func TestSaveProvider_RejectsInvalidSlugTraversal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := SaveProvider(types.Provider{
		Slug:            "../evil",
		DisplayName:     "Bad",
		APIKey:          "sk-test",
		Model:           "gpt-5",
		BaseURL:         "https://api.openai.com/v1",
		WireAPI:         "responses",
		ReasoningEffort: "medium",
	})
	if err == nil {
		t.Fatalf("SaveProvider() expected error for invalid slug")
	}

	outsidePath := filepath.Join(home, "evil")
	if _, statErr := os.Stat(outsidePath); !os.IsNotExist(statErr) {
		t.Fatalf("unexpected file/dir created outside provider root: %s", outsidePath)
	}
}

func TestLoadProviders_SkipsInvalidDirectories(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	base := filepath.Join(home, ".codexswitch")
	if err := os.MkdirAll(base, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(base, "..bad"), 0o700); err != nil {
		t.Fatalf("MkdirAll(invalid) error = %v", err)
	}

	if err := SaveProvider(types.Provider{
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

	list, err := LoadProviders()
	if err != nil {
		t.Fatalf("LoadProviders() error = %v", err)
	}
	if len(list.Providers) != 1 {
		t.Fatalf("LoadProviders() providers len = %d, want 1", len(list.Providers))
	}
	if list.Providers[0].Slug != "openai" {
		t.Fatalf("LoadProviders() slug = %q, want %q", list.Providers[0].Slug, "openai")
	}
}

func TestActivateAndRemoveProvider_RejectInvalidSlug(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := Activate("../evil"); err == nil {
		t.Fatalf("Activate() expected error for invalid slug")
	}
	if err := RemoveProvider("../evil"); err == nil {
		t.Fatalf("RemoveProvider() expected error for invalid slug")
	}
}
