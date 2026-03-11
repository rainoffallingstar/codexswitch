package sync

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WebDAVOptions struct {
	URL      string
	Username string
	Password string
	Insecure bool
	Timeout  time.Duration
}

func (o WebDAVOptions) validate() error {
	if strings.TrimSpace(o.URL) == "" {
		return errors.New("webdav url is required")
	}
	parsed, err := url.Parse(o.URL)
	if err != nil {
		return fmt.Errorf("parse webdav url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("webdav url must be http(s): %s", parsed.Scheme)
	}
	if parsed.Host == "" {
		return errors.New("webdav url must include host")
	}
	return nil
}

func httpClient(insecure bool, timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	tr := &http.Transport{}
	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Explicit user opt-in.
	}
	return &http.Client{Transport: tr, Timeout: timeout}
}

func PushWebDAV(ctx context.Context, localDir string, opt WebDAVOptions) error {
	if err := opt.validate(); err != nil {
		return err
	}
	data, err := archiveTarGz(localDir)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, opt.URL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/gzip")
	if opt.Username != "" || opt.Password != "" {
		req.SetBasicAuth(opt.Username, opt.Password)
	}

	resp, err := httpClient(opt.Insecure, opt.Timeout).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return fmt.Errorf("webdav put: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

type PullOptions struct {
	WebDAV WebDAVOptions
	// BackupExisting controls whether localDir is first renamed to <localDir>.bak-<unix>.
	BackupExisting bool
}

func PullWebDAV(ctx context.Context, localDir string, opt PullOptions) error {
	if err := opt.WebDAV.validate(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opt.WebDAV.URL, nil)
	if err != nil {
		return err
	}
	if opt.WebDAV.Username != "" || opt.WebDAV.Password != "" {
		req.SetBasicAuth(opt.WebDAV.Username, opt.WebDAV.Password)
	}

	resp, err := httpClient(opt.WebDAV.Insecure, opt.WebDAV.Timeout).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return fmt.Errorf("webdav get: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	archiveBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parent := filepath.Dir(localDir)
	tmpDir, err := os.MkdirTemp(parent, ".codexswitch-sync-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := extractTarGz(archiveBytes, tmpDir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	if opt.BackupExisting {
		if _, err := os.Stat(localDir); err == nil {
			backup := fmt.Sprintf("%s.bak-%d", localDir, time.Now().Unix())
			if err := os.Rename(localDir, backup); err != nil {
				return fmt.Errorf("backup existing store: %w", err)
			}
		}
	} else {
		if err := os.RemoveAll(localDir); err != nil {
			return fmt.Errorf("remove existing store: %w", err)
		}
	}

	if err := os.MkdirAll(parent, 0o700); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	if err := os.Rename(tmpDir, localDir); err != nil {
		return fmt.Errorf("replace store dir: %w", err)
	}
	return nil
}

func archiveTarGz(root string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	walkErr := filepath.WalkDir(cleanRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == cleanRoot {
			return nil
		}

		rel, err := filepath.Rel(cleanRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." || rel == "" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if !info.Mode().IsRegular() && !d.IsDir() {
			return nil
		}

		if d.IsDir() {
			h := &tar.Header{
				Name:     rel + "/",
				Typeflag: tar.TypeDir,
				Mode:     0o700,
				ModTime:  time.Unix(0, 0),
			}
			return tw.WriteHeader(h)
		}

		h := &tar.Header{
			Name:    rel,
			Mode:    0o600,
			Size:    info.Size(),
			ModTime: time.Unix(0, 0),
		}
		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})

	closeErr := tw.Close()
	gzipErr := gw.Close()

	if walkErr != nil {
		return nil, walkErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if gzipErr != nil {
		return nil, gzipErr
	}
	return buf.Bytes(), nil
}

func extractTarGz(data []byte, dest string) error {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		name := strings.TrimPrefix(h.Name, "/")
		name = filepath.Clean(name)
		if name == "." || name == "" {
			continue
		}
		if strings.HasPrefix(name, ".."+string(filepath.Separator)) || name == ".." {
			return fmt.Errorf("invalid archive path: %q", h.Name)
		}

		target := filepath.Join(destAbs, name)
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(targetAbs, destAbs+string(filepath.Separator)) && targetAbs != destAbs {
			return fmt.Errorf("path escapes dest: %q", h.Name)
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetAbs, 0o700); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetAbs), 0o700); err != nil {
				return err
			}
			f, err := os.OpenFile(targetAbs, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		default:
			// ignore other types
		}
	}
}
