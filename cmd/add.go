package cmd

import (
	"fmt"
	"os"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/rainoffallingstar/codexswitch/internal/types"
	"github.com/rainoffallingstar/codexswitch/internal/ui"
	"github.com/spf13/cobra"
)

var addOpts types.AddProviderInput

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new provider",
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addOpts.Slug, "slug", "", "Provider slug")
	addCmd.Flags().StringVar(&addOpts.DisplayName, "name", "", "Provider display name")
	addCmd.Flags().StringVar(&addOpts.APIKey, "api-key", "", "Provider API key")
	addCmd.Flags().StringVar(&addOpts.Model, "model", "", "Default model")
	addCmd.Flags().StringVar(&addOpts.BaseURL, "base-url", "", "Provider base URL")
	addCmd.Flags().StringVar(&addOpts.WireAPI, "wire-api", "", "Wire API (default: responses)")

	rootCmd.AddCommand(addCmd)
}

func runAdd(_ *cobra.Command, _ []string) error {
	interactive := isTerminal(os.Stdin)

	input, err := ui.PromptAddProvider(addOpts, interactive)
	if err != nil {
		return err
	}

	wireAPI := input.WireAPI
	if wireAPI == "" {
		wireAPI = "responses"
	}

	p := types.Provider{
		Slug:        input.Slug,
		DisplayName: input.DisplayName,
		APIKey:      input.APIKey,
		Model:       input.Model,
		BaseURL:     input.BaseURL,
		WireAPI:     wireAPI,
	}

	if err := store.SaveProvider(p); err != nil {
		return fmt.Errorf("save provider: %w", err)
	}

	fmt.Printf("Provider '%s' (%s) added successfully.\n", p.DisplayName, p.Slug)
	return nil
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
