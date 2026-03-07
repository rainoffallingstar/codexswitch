package store

import (
	"os"
	"path/filepath"
	"strings"
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

func TestCopyProvider_CreatesIncrementedCloneAndKeepsCurrent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	seed := types.Provider{
		Slug:            "openai",
		DisplayName:     "OpenAI",
		APIKey:          "sk-test",
		Model:           "gpt-5",
		BaseURL:         "https://api.openai.com/v1",
		WireAPI:         "responses",
		ReasoningEffort: "high",
	}
	if err := SaveProvider(seed); err != nil {
		t.Fatalf("SaveProvider(seed) error = %v", err)
	}
	if err := SaveProvider(types.Provider{
		Slug:            "openai-copy1",
		DisplayName:     "OpenAI copy1",
		APIKey:          "sk-old",
		Model:           "gpt-4.1-mini",
		BaseURL:         "https://api.openai.com/v1",
		WireAPI:         "responses",
		ReasoningEffort: "medium",
	}); err != nil {
		t.Fatalf("SaveProvider(copy1) error = %v", err)
	}
	if err := Activate("openai"); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}

	cloned, err := CopyProvider("openai")
	if err != nil {
		t.Fatalf("CopyProvider() error = %v", err)
	}
	if cloned.Slug != "openai-copy2" {
		t.Fatalf("cloned slug = %q, want %q", cloned.Slug, "openai-copy2")
	}
	if cloned.DisplayName != "OpenAI copy2" {
		t.Fatalf("cloned display name = %q, want %q", cloned.DisplayName, "OpenAI copy2")
	}
	if cloned.APIKey != seed.APIKey || cloned.Model != seed.Model || cloned.BaseURL != seed.BaseURL || cloned.WireAPI != seed.WireAPI || cloned.ReasoningEffort != seed.ReasoningEffort {
		t.Fatalf("cloned provider fields differ from source")
	}

	current, err := GetCurrentSlug()
	if err != nil {
		t.Fatalf("GetCurrentSlug() error = %v", err)
	}
	if current != "openai" {
		t.Fatalf("current slug after copy = %q, want %q", current, "openai")
	}
}

func TestCopyProvider_NotFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	_, err := CopyProvider("missing")
	if err == nil {
		t.Fatalf("CopyProvider() expected provider not found error")
	}
	if !strings.Contains(err.Error(), "provider not found: missing") {
		t.Fatalf("CopyProvider() error = %v, want provider not found", err)
	}
}
