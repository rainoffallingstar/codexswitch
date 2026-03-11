package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	syncpkg "github.com/rainoffallingstar/codexswitch/internal/sync"
	"github.com/rainoffallingstar/codexswitch/internal/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync provider configs via WebDAV",
}

var syncPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload ~/.codexswitch to WebDAV (PUT)",
	RunE:  runSyncPush,
}

var syncPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download ~/.codexswitch from WebDAV (GET)",
	RunE:  runSyncPull,
}

type webdavFlags struct {
	url      string
	username string
	password string
	insecure bool
	timeout  time.Duration
	backup   bool
}

var webdav webdavFlags

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncPushCmd)
	syncCmd.AddCommand(syncPullCmd)

	for _, c := range []*cobra.Command{syncPushCmd, syncPullCmd} {
		c.Flags().StringVar(&webdav.url, "webdav-url", "", "WebDAV file URL (or CODEXSWITCH_WEBDAV_URL)")
		c.Flags().StringVar(&webdav.username, "webdav-user", "", "WebDAV username (or CODEXSWITCH_WEBDAV_USER)")
		c.Flags().StringVar(&webdav.password, "webdav-pass", "", "WebDAV password (or CODEXSWITCH_WEBDAV_PASS)")
		c.Flags().BoolVar(&webdav.insecure, "webdav-insecure", false, "Skip TLS verification (dangerous)")
		c.Flags().DurationVar(&webdav.timeout, "webdav-timeout", 30*time.Second, "HTTP timeout (e.g. 10s, 1m)")
	}
	syncPullCmd.Flags().BoolVar(&webdav.backup, "backup", true, "Backup existing ~/.codexswitch before replacing")
}

func webdavOptionsFromFlags() syncpkg.WebDAVOptions {
	u := webdav.url
	if u == "" {
		u = os.Getenv("CODEXSWITCH_WEBDAV_URL")
	}
	user := webdav.username
	if user == "" {
		user = os.Getenv("CODEXSWITCH_WEBDAV_USER")
	}
	pass := webdav.password
	if pass == "" {
		pass = os.Getenv("CODEXSWITCH_WEBDAV_PASS")
	}

	return syncpkg.WebDAVOptions{
		URL:      strings.TrimSpace(u),
		Username: strings.TrimSpace(user),
		Password: pass,
		Insecure: webdav.insecure,
		Timeout:  webdav.timeout,
	}
}

func resolveWebDAVOptions(interactive bool) (syncpkg.WebDAVOptions, error) {
	opt := webdavOptionsFromFlags()

	// If required fields are missing, fall back to ~/.codexswitch/.webdav.json.
	if opt.URL == "" {
		cfg, ok, err := store.LoadWebDAVConfig()
		if err != nil {
			return syncpkg.WebDAVOptions{}, err
		}
		if ok {
			if opt.URL == "" && strings.TrimSpace(cfg.URL) != "" {
				opt.URL = strings.TrimSpace(cfg.URL)
			}
			if opt.Username == "" && strings.TrimSpace(cfg.Username) != "" {
				opt.Username = strings.TrimSpace(cfg.Username)
			}
			if opt.Password == "" && cfg.Password != "" {
				opt.Password = cfg.Password
			}
			if !webdav.insecure {
				opt.Insecure = cfg.Insecure
			}
			if webdav.timeout == 30*time.Second && strings.TrimSpace(cfg.Timeout) != "" {
				if parsed, err := time.ParseDuration(cfg.Timeout); err == nil {
					opt.Timeout = parsed
				}
			}
		}
	}

	if opt.URL != "" {
		return opt, nil
	}

	// Still missing: prompt in interactive mode, then save ~/.codexswitch/.webdav.json.
	cfg, err := ui.PromptWebDAVConfig(store.WebDAVConfig{}, interactive)
	if err != nil {
		return syncpkg.WebDAVOptions{}, err
	}
	if err := store.SaveWebDAVConfig(cfg); err != nil {
		return syncpkg.WebDAVOptions{}, fmt.Errorf("save webdav config: %w", err)
	}
	return ui.ToWebDAVOptions(cfg)
}

func runSyncPush(_ *cobra.Command, _ []string) error {
	dir, err := store.SwitchDir()
	if err != nil {
		return err
	}
	interactive := isTerminal(os.Stdin)
	opt, err := resolveWebDAVOptions(interactive)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := syncpkg.PushWebDAV(ctx, dir, opt); err != nil {
		return fmt.Errorf("sync push: %w", err)
	}
	fmt.Println("Uploaded ~/.codexswitch to WebDAV.")
	return nil
}

func runSyncPull(_ *cobra.Command, _ []string) error {
	dir, err := store.SwitchDir()
	if err != nil {
		return err
	}
	interactive := isTerminal(os.Stdin)
	opt, err := resolveWebDAVOptions(interactive)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := syncpkg.PullWebDAV(ctx, dir, syncpkg.PullOptions{
		WebDAV:         opt,
		BackupExisting: webdav.backup,
	}); err != nil {
		return fmt.Errorf("sync pull: %w", err)
	}
	fmt.Println("Downloaded ~/.codexswitch from WebDAV.")
	return nil
}
