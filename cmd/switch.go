package cmd

import (
	"fmt"
	"os"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/spf13/cobra"
)

var switchSlug string

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch active provider",
	RunE:  runSwitch,
}

func init() {
	switchCmd.Flags().StringVar(&switchSlug, "slug", "", "Provider slug to switch to")
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(_ *cobra.Command, _ []string) error {
	list, err := store.LoadProviders()
	if err != nil {
		return fmt.Errorf("load providers: %w", err)
	}
	if len(list.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	interactive := isTerminal(os.Stdin)
	chosen, err := resolveProviderTarget(list, switchSlug, interactive, "Select provider to switch")
	if err != nil {
		return err
	}

	if err := store.Activate(chosen.Slug); err != nil {
		return fmt.Errorf("activate: %w", err)
	}

	fmt.Printf("Switched to provider: %s (%s)\n", chosen.DisplayName, chosen.Slug)
	return nil
}
