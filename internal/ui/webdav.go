package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/rainoffallingstar/codexswitch/internal/store"
	syncpkg "github.com/rainoffallingstar/codexswitch/internal/sync"
)

func PromptWebDAVConfig(seed store.WebDAVConfig, interactive bool) (store.WebDAVConfig, error) {
	cfg := seed

	fields := []struct {
		label    string
		value    *string
		required bool
		dflt     string
		mask     rune
	}{
		{label: "WebDAV URL (file URL, e.g. https://.../codexswitch.tar.gz)", value: &cfg.URL, required: true},
		{label: "WebDAV username", value: &cfg.Username, required: false},
		{label: "WebDAV password", value: &cfg.Password, required: false, mask: '*'},
		{label: "HTTP timeout (duration, e.g. 30s)", value: &cfg.Timeout, required: false, dflt: "30s"},
	}

	for _, f := range fields {
		*f.value = strings.TrimSpace(*f.value)
		if *f.value != "" {
			continue
		}
		if !interactive {
			if f.required {
				return store.WebDAVConfig{}, fmt.Errorf("missing required option: %s (set via flags/env or configure ~/.codexswitch/%s)", f.label, ".webdav.json")
			}
			*f.value = f.dflt
			continue
		}

		p := promptui.Prompt{Label: f.label, Default: f.dflt}
		if f.mask != 0 {
			p.Mask = f.mask
		}
		val, err := p.Run()
		if err != nil {
			return store.WebDAVConfig{}, fmt.Errorf("read %s: %w", strings.ToLower(f.label), err)
		}
		val = strings.TrimSpace(val)
		if val == "" {
			val = f.dflt
		}
		if f.required && val == "" {
			return store.WebDAVConfig{}, fmt.Errorf("%s is required", strings.ToLower(f.label))
		}
		*f.value = val
	}

	cfg.Timeout = strings.TrimSpace(cfg.Timeout)
	if cfg.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Timeout); err != nil {
			return store.WebDAVConfig{}, fmt.Errorf("invalid timeout: %w", err)
		}
	}

	if interactive {
		sel := promptui.Select{
			Label: "Skip TLS verification? (dangerous)",
			Items: []string{"no", "yes"},
		}
		idx, _, err := sel.Run()
		if err == nil {
			cfg.Insecure = idx == 1
		}
	}

	return cfg, nil
}

func ToWebDAVOptions(cfg store.WebDAVConfig) (syncpkg.WebDAVOptions, error) {
	var timeout time.Duration
	var err error
	if strings.TrimSpace(cfg.Timeout) != "" {
		timeout, err = time.ParseDuration(cfg.Timeout)
		if err != nil {
			return syncpkg.WebDAVOptions{}, fmt.Errorf("parse timeout: %w", err)
		}
	}
	return syncpkg.WebDAVOptions{
		URL:      strings.TrimSpace(cfg.URL),
		Username: strings.TrimSpace(cfg.Username),
		Password: cfg.Password,
		Insecure: cfg.Insecure,
		Timeout:  timeout,
	}, nil
}
