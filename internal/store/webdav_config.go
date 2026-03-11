package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const webdavConfigFile = ".webdav.json"

type WebDAVConfig struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
	Timeout  string `json:"timeout,omitempty"` // Go duration string, e.g. "30s"
}

func LoadWebDAVConfig() (WebDAVConfig, bool, error) {
	base, err := switchDir()
	if err != nil {
		return WebDAVConfig{}, false, err
	}
	path := filepath.Join(base, webdavConfigFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return WebDAVConfig{}, false, nil
		}
		return WebDAVConfig{}, false, err
	}

	var cfg WebDAVConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return WebDAVConfig{}, false, fmt.Errorf("parse webdav config: %w", err)
	}
	return cfg, true, nil
}

func SaveWebDAVConfig(cfg WebDAVConfig) error {
	base, err := switchDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(base, 0o700); err != nil {
		return err
	}
	path := filepath.Join(base, webdavConfigFile)

	// Normalize/validate timeout string if present.
	if cfg.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Timeout); err != nil {
			return fmt.Errorf("invalid webdav timeout: %w", err)
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}
