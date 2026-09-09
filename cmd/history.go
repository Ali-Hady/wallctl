package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)
var (
	limit int
	favsOnly bool
)
var historyCommand = &cobra.Command{
	Use:   "history",
	Short: "List wallpaper history",
	RunE: func(cmd *cobra.Command, args []string) error {
		images, err := appStore.List(limit, favsOnly)
		if err != nil{
			return fmt.Errorf("failed to list images: %w", err)
		}
		if len(images)==0{
			fmt.Println("No wallpapers in history")
			return nil
		}

		writer:= tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ' , 0)
		fmt.Fprintln(writer, "ID\tALIAS\tDATE\tTITLE\tFAV")
		fmt.Fprintln(writer, "--\t----\t---\t----\t---")
		for _, img:=range images{
			alias:=img.Alias
			if alias==""{
				alias="-"
			}
			favMark:=""
			if img.Favorite{
				favMark="*"
			}
			dateStr := img.FetchedAt.Format("2006-01-02")
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", img.ID, alias, dateStr, img.Title, favMark)
		}
		writer.Flush()
		return nil
	},
}
func init(){
	rootCmd.AddCommand(historyCommand)
	historyCommand.Flags().IntVarP(&limit, "limit", "l", 10, "Number of wallpapers to show")
	historyCommand.Flags().BoolVarP(&favsOnly, "favs", "f", false, "Show only favorite wallpapers")
}

