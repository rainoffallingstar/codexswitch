package types

// Provider holds all configuration for a single Codex provider.
type Provider struct {
	Slug        string // folder name + config.toml provider key
	DisplayName string
	APIKey      string
	Model       string
	BaseURL     string
	WireAPI     string // default "responses"
}

// ProviderList is the full registry of configured providers.
type ProviderList struct {
	Providers   []Provider
	CurrentSlug string
}

// AuthConfig maps to ~/.codex/auth.json
type AuthConfig struct {
	AuthMode     string `json:"auth_mode"`
	OpenAIAPIKey string `json:"OPENAI_API_KEY"`
}

// ModelProviderConfig maps to [model_providers.<slug>] in config.toml
type ModelProviderConfig struct {
	Name               string `toml:"name"`
	BaseURL            string `toml:"base_url"`
	WireAPI            string `toml:"wire_api"`
	RequiresOpenAIAuth bool   `toml:"requires_openai_auth"`
}

// TOMLConfig maps to config.toml
type TOMLConfig struct {
	Model          string                         `toml:"model"`
	ModelProvider  string                         `toml:"model_provider"`
	ModelProviders map[string]ModelProviderConfig `toml:"model_providers"`
}

// AddProviderInput carries the interactive prompts result for the add command.
type AddProviderInput struct {
	Slug        string
	DisplayName string
	APIKey      string
	Model       string
	BaseURL     string
	WireAPI     string
}
