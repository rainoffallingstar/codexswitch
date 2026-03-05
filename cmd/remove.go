package cmd

import (
	"fmt"
	"os"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/rainoffallingstar/codexswitch/internal/ui"
	"github.com/spf13/cobra"
)

var removeSlug string
var removeYes bool

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a provider",
	RunE:  runRemove,
}

func init() {
	removeCmd.Flags().StringVar(&removeSlug, "slug", "", "Provider slug to remove")
	removeCmd.Flags().BoolVar(&removeYes, "yes", false, "Skip confirmation prompt")

	rootCmd.AddCommand(removeCmd)
}

func runRemove(_ *cobra.Command, _ []string) error {
	list, err := store.LoadProviders()
	if err != nil {
		return fmt.Errorf("load providers: %w", err)
	}
	if len(list.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	interactive := isTerminal(os.Stdin)
	target, err := resolveProviderTarget(list, removeSlug, interactive, "Select provider to remove")
	if err != nil {
		return err
	}

	if target.Slug == list.CurrentSlug {
		return fmt.Errorf("cannot remove active provider '%s'; switch to another provider first", target.Slug)
	}

	confirmed, err := ui.ConfirmRemoval(target.Slug, target.DisplayName, interactive, removeYes)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Removal cancelled.")
		return nil
	}

	if err := store.RemoveProvider(target.Slug); err != nil {
		return err
	}

	fmt.Printf("Provider '%s' (%s) removed successfully.\n", target.DisplayName, target.Slug)
	return nil
}
