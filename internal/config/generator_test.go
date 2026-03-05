package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rainoffallingstar/codexswitch/internal/types"
)

func TestGenerateAuthJSON_UsesAPIKeyMode(t *testing.T) {
	out, err := GenerateAuthJSON(types.Provider{APIKey: "sk-test"})
	if err != nil {
		t.Fatalf("GenerateAuthJSON() error = %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got["auth_mode"] != "apikey" {
		t.Fatalf("auth_mode = %q, want %q", got["auth_mode"], "apikey")
	}
}

func TestGenerateTOML_IncludesReasoningEffortBeforeModelProviders(t *testing.T) {
	out, err := GenerateTOML(types.Provider{
		Slug:        "openai",
		DisplayName: "OpenAI",
		Model:       "gpt-5",
		BaseURL:     "https://api.openai.com/v1",
		WireAPI:     "responses",
	})
	if err != nil {
		t.Fatalf("GenerateTOML() error = %v", err)
	}

	content := string(out)
	if !strings.Contains(content, "model_reasoning_effort = 'medium'") {
		t.Fatalf("missing model_reasoning_effort in TOML:\n%s", content)
	}

	reasoningPos := strings.Index(content, "model_reasoning_effort")
	modelProvidersPos := strings.Index(content, "[model_providers]")
	if reasoningPos == -1 || modelProvidersPos == -1 {
		t.Fatalf("required sections missing in TOML:\n%s", content)
	}
	if reasoningPos > modelProvidersPos {
		t.Fatalf("model_reasoning_effort should appear before [model_providers]:\n%s", content)
	}
}

func TestParseTOML_ReasoningEffortCompatibility(t *testing.T) {
	t.Run("with field", func(t *testing.T) {
		data := []byte(`model = "gpt-5"
model_provider = "openai"
model_reasoning_effort = "high"

[model_providers.openai]
name = "OpenAI"
base_url = "https://api.openai.com/v1"
wire_api = "responses"
requires_openai_auth = true
`)
		p, err := ParseTOML(data, "openai")
		if err != nil {
			t.Fatalf("ParseTOML() error = %v", err)
		}
		if p.ReasoningEffort != "high" {
			t.Fatalf("ReasoningEffort = %q, want %q", p.ReasoningEffort, "high")
		}
	})

	t.Run("missing field falls back to default", func(t *testing.T) {
		data := []byte(`model = "gpt-5"
model_provider = "openai"

[model_providers.openai]
name = "OpenAI"
base_url = "https://api.openai.com/v1"
wire_api = "responses"
requires_openai_auth = true
`)
		p, err := ParseTOML(data, "openai")
		if err != nil {
			t.Fatalf("ParseTOML() error = %v", err)
		}
		if p.ReasoningEffort != types.DefaultReasoningEffort {
			t.Fatalf("ReasoningEffort = %q, want %q", p.ReasoningEffort, types.DefaultReasoningEffort)
		}
	})
}
