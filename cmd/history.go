package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	limit    int
	favsOnly bool
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "List wallpaper history",
	RunE: func(cmd *cobra.Command, args []string) error {
		images, err := appStore.List(limit, favsOnly)
		if err != nil {
			return fmt.Errorf("failed to list images: %w", err)
		}
		if len(images) == 0 {
			fmt.Println("No wallpapers in history")
			return nil
		}

		current, _ := appStore.Current()
		currentID := ""
		if current != nil {
			currentID = current.ID
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(writer, "CUR\tID\tALIAS\tDATE\tFAV\tTITLE")
		fmt.Fprintln(writer, "---\t--\t-----\t----\t---\t-----")

		for _, img := range images {
			curMark := " "
			if img.ID == currentID {
				curMark = ">"
			}

			alias := img.Alias
			if alias == "" {
				alias = "-"
			}

			favMark := " "
			if img.Favorite {
				favMark = "★"
			}

			dateStr := img.FetchedAt.Format("2006-01-02")

			title := img.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}

			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n", curMark, img.ID, alias, dateStr, favMark, title)
		}

		return writer.Flush()
	},
}

func init() {
	historyCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of wallpapers to show (0 for all)")
	historyCmd.Flags().BoolVarP(&favsOnly, "favs", "f", false, "Show only favorite wallpapers")
	rootCmd.AddCommand(historyCmd)
}
