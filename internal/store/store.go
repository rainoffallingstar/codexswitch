package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rainoffallingstar/codexswitch/internal/config"
	"github.com/rainoffallingstar/codexswitch/internal/types"
)

const currentFile = ".current"

func switchDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codexswitch"), nil
}

func codexDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

func providerDir(base, slug string) string {
	return filepath.Join(base, slug)
}

// SaveProvider writes auth.json and config.toml for the provider.
func SaveProvider(p types.Provider) error {
	slug, err := types.NormalizeSlug(p.Slug)
	if err != nil {
		return err
	}
	p.Slug = slug

	base, err := switchDir()
	if err != nil {
		return err
	}
	dir := providerDir(base, p.Slug)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create provider dir: %w", err)
	}

	authBytes, err := config.GenerateAuthJSON(p)
	if err != nil {
		return fmt.Errorf("generate auth.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), authBytes, 0o600); err != nil {
		return fmt.Errorf("write auth.json: %w", err)
	}

	tomlBytes, err := config.GenerateTOML(p)
	if err != nil {
		return fmt.Errorf("generate config.toml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), tomlBytes, 0o600); err != nil {
		return fmt.Errorf("write config.toml: %w", err)
	}

	return nil
}

// LoadProviders scans ~/.codexswitch/ and returns all configured providers.
func LoadProviders() (types.ProviderList, error) {
	base, err := switchDir()
	if err != nil {
		return types.ProviderList{}, err
	}
	if err := os.MkdirAll(base, 0o700); err != nil {
		return types.ProviderList{}, err
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return types.ProviderList{}, err
	}

	current, _ := readCurrent(base)
	if current != "" {
		normalizedCurrent, err := types.NormalizeSlug(current)
		if err == nil {
			current = normalizedCurrent
		} else {
			current = ""
		}
	}

	var providers []types.Provider
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		slug, slugErr := types.NormalizeSlug(e.Name())
		if slugErr != nil {
			continue
		}
		p, err := loadProvider(base, slug)
		if err != nil {
			continue // skip malformed directories
		}
		providers = append(providers, p)
	}

	return types.ProviderList{
		Providers:   providers,
		CurrentSlug: current,
	}, nil
}

// loadProvider reads auth.json and config.toml from the provider directory.
func loadProvider(base, slug string) (types.Provider, error) {
	normalizedSlug, err := types.NormalizeSlug(slug)
	if err != nil {
		return types.Provider{}, err
	}
	slug = normalizedSlug

	dir := providerDir(base, slug)

	authData, err := os.ReadFile(filepath.Join(dir, "auth.json"))
	if err != nil {
		return types.Provider{}, err
	}
	var auth types.AuthConfig
	if err := json.Unmarshal(authData, &auth); err != nil {
		return types.Provider{}, err
	}

	p := types.Provider{
		Slug:            slug,
		DisplayName:     slug,
		APIKey:          auth.OpenAIAPIKey,
		WireAPI:         "responses",
		ReasoningEffort: types.DefaultReasoningEffort,
	}

	tomlData, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err == nil {
		tomlProvider, parseErr := config.ParseTOML(tomlData, slug)
		if parseErr == nil {
			p.DisplayName = tomlProvider.DisplayName
			p.Model = tomlProvider.Model
			p.BaseURL = tomlProvider.BaseURL
			p.WireAPI = tomlProvider.WireAPI
			p.ReasoningEffort = tomlProvider.ReasoningEffort
		}
	}

	return p, nil
}

// Activate copies provider files to ~/.codex/ and records the current slug.
func Activate(slug string) error {
	normalizedSlug, err := types.NormalizeSlug(slug)
	if err != nil {
		return err
	}
	slug = normalizedSlug

	base, err := switchDir()
	if err != nil {
		return err
	}
	src := providerDir(base, slug)

	codex, err := codexDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(codex, 0o700); err != nil {
		return fmt.Errorf("create codex dir: %w", err)
	}

	for _, name := range []string{"auth.json", "config.toml"} {
		if err := copyFile(filepath.Join(src, name), filepath.Join(codex, name)); err != nil {
			return fmt.Errorf("copy %s: %w", name, err)
		}
	}

	return writeCurrent(base, slug)
}

// GetCurrentSlug returns the currently active provider slug.
func GetCurrentSlug() (string, error) {
	base, err := switchDir()
	if err != nil {
		return "", err
	}
	return readCurrent(base)
}

// FindProviderBySlug returns a configured provider by slug.
func FindProviderBySlug(slug string) (types.Provider, error) {
	normalizedSlug, err := types.NormalizeSlug(slug)
	if err != nil {
		return types.Provider{}, err
	}

	list, err := LoadProviders()
	if err != nil {
		return types.Provider{}, err
	}
	for _, p := range list.Providers {
		if p.Slug == normalizedSlug {
			return p, nil
		}
	}
	return types.Provider{}, fmt.Errorf("provider not found: %s", normalizedSlug)
}

// CopyProvider clones an existing provider into <slug>-copyN.
func CopyProvider(sourceSlug string) (types.Provider, error) {
	normalizedSource, err := types.NormalizeSlug(sourceSlug)
	if err != nil {
		return types.Provider{}, err
	}

	list, err := LoadProviders()
	if err != nil {
		return types.Provider{}, err
	}

	var source types.Provider
	found := false
	maxCopy := 0
	prefix := normalizedSource + "-copy"

	for _, p := range list.Providers {
		if p.Slug == normalizedSource {
			source = p
			found = true
		}

		if !strings.HasPrefix(p.Slug, prefix) {
			continue
		}
		suffix := strings.TrimPrefix(p.Slug, prefix)
		n, parseErr := strconv.Atoi(suffix)
		if parseErr != nil || n < 1 {
			continue
		}
		if n > maxCopy {
			maxCopy = n
		}
	}

	if !found {
		return types.Provider{}, fmt.Errorf("provider not found: %s", normalizedSource)
	}

	copyN := maxCopy + 1
	cloned := source
	cloned.Slug = fmt.Sprintf("%s-copy%d", normalizedSource, copyN)
	cloned.DisplayName = fmt.Sprintf("%s copy%d", source.DisplayName, copyN)

	if err := SaveProvider(cloned); err != nil {
		return types.Provider{}, err
	}
	return cloned, nil
}

// RemoveProvider deletes a provider directory from ~/.codexswitch.
func RemoveProvider(slug string) error {
	normalizedSlug, err := types.NormalizeSlug(slug)
	if err != nil {
		return err
	}
	slug = normalizedSlug

	base, err := switchDir()
	if err != nil {
		return err
	}
	dir := providerDir(base, slug)

	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("provider not found: %s", slug)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("provider path is not a directory: %s", slug)
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove provider dir: %w", err)
	}
	return nil
}

func readCurrent(base string) (string, error) {
	data, err := os.ReadFile(filepath.Join(base, currentFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func writeCurrent(base, slug string) error {
	return os.WriteFile(filepath.Join(base, currentFile), []byte(slug+"\n"), 0o600)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
