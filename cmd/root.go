package cmd

import (
	"fmt"
	"os"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/rainoffallingstar/codexswitch/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "codexswitch",
	Short: "Switch between Codex provider configurations",
	Long:  "codexswitch switches between Codex provider configurations.",
	RunE:  runRoot,
}

// Execute is the CLI entry point.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}

func runRoot(_ *cobra.Command, _ []string) error {
	list, err := store.LoadProviders()
	if err != nil {
		return fmt.Errorf("load providers: %w", err)
	}

	if len(list.Providers) == 0 {
		fmt.Fprintln(os.Stderr, "No providers configured. Run 'codexswitch add' to add one.")
		return nil
	}

	chosen, err := ui.PickProvider(list)
	if err != nil {
		return err
	}

	if err := store.Activate(chosen.Slug); err != nil {
		return fmt.Errorf("activate: %w", err)
	}

	fmt.Printf("Switched to provider: %s (%s)\n", chosen.DisplayName, chosen.Slug)
	return nil
}
