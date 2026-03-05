package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rainoffallingstar/codexswitch/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured providers",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(_ *cobra.Command, _ []string) error {
	list, err := store.LoadProviders()
	if err != nil {
		return fmt.Errorf("load providers: %w", err)
	}

	if len(list.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SLUG\tDISPLAY NAME\tSTATUS")
	fmt.Fprintln(w, "----\t------------\t------")
	for _, p := range list.Providers {
		status := ""
		if p.Slug == list.CurrentSlug {
			status = "* active"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Slug, p.DisplayName, status)
	}
	w.Flush()
	return nil
}
