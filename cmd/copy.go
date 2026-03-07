package cmd

import (
	"fmt"
	"os"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/spf13/cobra"
)

var copySlug string

var copyCmd = &cobra.Command{
	Use:     "copy",
	Aliases: []string{"replicate"},
	Short:   "Copy an existing provider",
	RunE:    runCopy,
}

func init() {
	copyCmd.Flags().StringVar(&copySlug, "slug", "", "Provider slug to copy")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(_ *cobra.Command, _ []string) error {
	list, err := store.LoadProviders()
	if err != nil {
		return fmt.Errorf("load providers: %w", err)
	}
	if len(list.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	interactive := isTerminal(os.Stdin)
	source, err := resolveProviderTarget(list, copySlug, interactive, "Select provider to copy")
	if err != nil {
		return err
	}

	cloned, err := store.CopyProvider(source.Slug)
	if err != nil {
		return err
	}

	fmt.Printf("Provider '%s' copied to '%s' (%s).\n", source.Slug, cloned.DisplayName, cloned.Slug)
	return nil
}
