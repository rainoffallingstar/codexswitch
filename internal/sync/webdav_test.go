package sync

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestArchiveAndExtract_RoundTrip(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "openai"), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, ".current"), []byte("openai\n"), 0o600); err != nil {
		t.Fatalf("write current: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "openai", "auth.json"), []byte(`{"openai_api_key":"sk-test"}`), 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "openai", "config.toml"), []byte("model = \"gpt-5\"\n"), 0o600); err != nil {
		t.Fatalf("write toml: %v", err)
	}

	blob, err := archiveTarGz(src)
	if err != nil {
		t.Fatalf("archive: %v", err)
	}

	dst := t.TempDir()
	if err := extractTarGz(blob, dst); err != nil {
		t.Fatalf("extract: %v", err)
	}

	gotCurrent, err := os.ReadFile(filepath.Join(dst, ".current"))
	if err != nil {
		t.Fatalf("read current: %v", err)
	}
	if string(gotCurrent) != "openai\n" {
		t.Fatalf("current mismatch: %q", string(gotCurrent))
	}
	gotAuth, err := os.ReadFile(filepath.Join(dst, "openai", "auth.json"))
	if err != nil {
		t.Fatalf("read auth: %v", err)
	}
	if string(gotAuth) != `{"openai_api_key":"sk-test"}` {
		t.Fatalf("auth mismatch: %q", string(gotAuth))
	}
}

func TestExtractTarGz_RejectsTraversal(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	_ = tw.WriteHeader(&tar.Header{Name: "../evil", Mode: 0o600, Size: 4})
	_, _ = tw.Write([]byte("oops"))
	_ = tw.Close()
	_ = gw.Close()

	err := extractTarGz(buf.Bytes(), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "invalid archive path") {
		t.Fatalf("expected traversal error, got: %v", err)
	}
}

func TestWebDAV_PushThenPull(t *testing.T) {
	t.Parallel()

	var stored []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			stored = body
			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			if stored == nil {
				http.Error(w, "missing", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(stored)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)

	local := filepath.Join(t.TempDir(), "store")
	if err := os.MkdirAll(filepath.Join(local, "p1"), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(local, "p1", "auth.json"), []byte(`{"openai_api_key":"sk-1"}`), 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	ctx := context.Background()
	opt := WebDAVOptions{URL: srv.URL + "/codexswitch.tar.gz", Timeout: 5 * time.Second}

	if err := PushWebDAV(ctx, local, opt); err != nil {
		t.Fatalf("push: %v", err)
	}

	// Create differing content to ensure pull replaces.
	if err := os.MkdirAll(local, 0o700); err != nil {
		t.Fatalf("mkdir local: %v", err)
	}
	if err := os.WriteFile(filepath.Join(local, "junk.txt"), []byte("junk"), 0o600); err != nil {
		t.Fatalf("write junk: %v", err)
	}

	if err := PullWebDAV(ctx, local, PullOptions{WebDAV: opt, BackupExisting: true}); err != nil {
		t.Fatalf("pull: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(local, "p1", "auth.json"))
	if err != nil {
		t.Fatalf("read restored auth: %v", err)
	}
	if string(got) != `{"openai_api_key":"sk-1"}` {
		t.Fatalf("restored mismatch: %q", string(got))
	}

	// Backup should exist alongside local.
	parent := filepath.Dir(local)
	entries, _ := os.ReadDir(parent)
	foundBackup := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "store.bak-") {
			foundBackup = true
		}
	}
	if !foundBackup {
		t.Fatalf("expected backup dir, none found in %s", parent)
	}
}
