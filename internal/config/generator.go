package config

import (
	"encoding/json"
	"fmt"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/rainoffallingstar/codexswitch/internal/types"
)

// GenerateAuthJSON returns the JSON bytes for auth.json.
func GenerateAuthJSON(p types.Provider) ([]byte, error) {
	auth := types.AuthConfig{
		AuthMode:     "openai",
		OpenAIAPIKey: p.APIKey,
	}
	return json.MarshalIndent(auth, "", "  ")
}

// GenerateTOML returns the TOML content for config.toml.
func GenerateTOML(p types.Provider) ([]byte, error) {
	wireAPI := p.WireAPI
	if wireAPI == "" {
		wireAPI = "responses"
	}

	cfg := types.TOMLConfig{
		Model:         p.Model,
		ModelProvider: p.Slug,
		ModelProviders: map[string]types.ModelProviderConfig{
			p.Slug: {
				Name:               p.DisplayName,
				BaseURL:            p.BaseURL,
				WireAPI:            wireAPI,
				RequiresOpenAIAuth: true,
			},
		},
	}

	out, err := toml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal toml: %w", err)
	}
	return out, nil
}

// ParseTOML extracts provider details from config.toml.
func ParseTOML(data []byte, slug string) (types.Provider, error) {
	var cfg types.TOMLConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return types.Provider{}, fmt.Errorf("parse toml: %w", err)
	}

	selectedSlug := slug
	if _, ok := cfg.ModelProviders[selectedSlug]; !ok {
		if _, ok := cfg.ModelProviders[cfg.ModelProvider]; ok {
			selectedSlug = cfg.ModelProvider
		}
	}

	providerCfg, ok := cfg.ModelProviders[selectedSlug]
	if !ok {
		return types.Provider{}, fmt.Errorf("provider section not found for slug: %s", slug)
	}

	displayName := providerCfg.Name
	if displayName == "" {
		displayName = selectedSlug
	}
	wireAPI := providerCfg.WireAPI
	if wireAPI == "" {
		wireAPI = "responses"
	}

	return types.Provider{
		Slug:        slug,
		DisplayName: displayName,
		Model:       cfg.Model,
		BaseURL:     providerCfg.BaseURL,
		WireAPI:     wireAPI,
	}, nil
}
