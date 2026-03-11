package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWebDAVConfig_SaveThenLoad(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := WebDAVConfig{
		URL:      "https://example.com/dav/codexswitch.tar.gz",
		Username: "alice",
		Password: "secret",
		Insecure: true,
		Timeout:  "45s",
	}
	if err := SaveWebDAVConfig(cfg); err != nil {
		t.Fatalf("SaveWebDAVConfig() error = %v", err)
	}

	loaded, ok, err := LoadWebDAVConfig()
	if err != nil {
		t.Fatalf("LoadWebDAVConfig() error = %v", err)
	}
	if !ok {
		t.Fatalf("LoadWebDAVConfig() ok = false, want true")
	}
	if loaded.URL != cfg.URL || loaded.Username != cfg.Username || loaded.Password != cfg.Password || loaded.Insecure != cfg.Insecure || loaded.Timeout != cfg.Timeout {
		t.Fatalf("loaded config mismatch: %+v vs %+v", loaded, cfg)
	}

	path := filepath.Join(home, ".codexswitch", ".webdav.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config perm = %v, want 0600", info.Mode().Perm())
	}
}

func TestWebDAVConfig_LoadMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	_, ok, err := LoadWebDAVConfig()
	if err != nil {
		t.Fatalf("LoadWebDAVConfig() error = %v", err)
	}
	if ok {
		t.Fatalf("LoadWebDAVConfig() ok = true, want false")
	}
}
