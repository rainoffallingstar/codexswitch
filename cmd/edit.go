package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/rainoffallingstar/codexswitch/internal/types"
	"github.com/rainoffallingstar/codexswitch/internal/ui"
	"github.com/spf13/cobra"
)

var editSlug string
var editOpts types.AddProviderInput

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing provider",
	RunE:  runEdit,
}

func init() {
	editCmd.Flags().StringVar(&editSlug, "slug", "", "Provider slug to edit")
	editCmd.Flags().StringVar(&editOpts.DisplayName, "name", "", "Provider display name")
	editCmd.Flags().StringVar(&editOpts.APIKey, "api-key", "", "Provider API key")
	editCmd.Flags().StringVar(&editOpts.Model, "model", "", "Default model")
	editCmd.Flags().StringVar(&editOpts.BaseURL, "base-url", "", "Provider base URL")
	editCmd.Flags().StringVar(&editOpts.WireAPI, "wire-api", "", "Wire API")

	rootCmd.AddCommand(editCmd)
}

func runEdit(_ *cobra.Command, _ []string) error {
	list, err := store.LoadProviders()
	if err != nil {
		return fmt.Errorf("load providers: %w", err)
	}
	if len(list.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	interactive := isTerminal(os.Stdin)

	target, err := resolveProviderTarget(list, editSlug, interactive, "Select provider to edit")
	if err != nil {
		return err
	}

	updated, err := ui.PromptEditProvider(target, editOpts, interactive)
	if err != nil {
		return err
	}

	if err := store.SaveProvider(updated); err != nil {
		return fmt.Errorf("save provider: %w", err)
	}

	fmt.Printf("Provider '%s' (%s) updated successfully.\n", updated.DisplayName, updated.Slug)
	return nil
}

func resolveProviderTarget(list types.ProviderList, slug string, interactive bool, promptLabel string) (types.Provider, error) {
	slug = strings.TrimSpace(slug)
	if slug != "" {
		for _, p := range list.Providers {
			if p.Slug == slug {
				return p, nil
			}
		}
		return types.Provider{}, fmt.Errorf("provider not found: %s", slug)
	}

	if !interactive {
		return types.Provider{}, fmt.Errorf("missing required option: --slug (non-interactive mode)")
	}
	return ui.PickProviderWithLabel(list, promptLabel)
}
