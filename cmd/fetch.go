package cmd
import (
	"fmt"
	"github.com/reujab/wallpaper"
	"github.com/spf13/cobra"
	
)
var noSet bool
var fetchCmd = &cobra.Command{
	Use: "fetch",
	Short: "Fetch wallpapers from a source",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Fetching wallpapers...")
		img, err := appSource.Fetch(cmd.Context(), imagesDir)
		if err != nil{
			return fmt.Errorf("failed to fetch wallpapers: %w", err)
		}
		if err = appStore.Save(img); err != nil {
			return fmt.Errorf("failed to save wallpapers: %w", err)
		}
		if !noSet {
			fmt.Printf("Setting %s as wallpaper...\n", img.Title)
			if err := wallpaper.SetFromFile(img.LocalPath); err != nil {
				return fmt.Errorf("failed to set wallpaper: %w", err)
			}
			fmt.Printf("Wallpaper applied successfully!\n")
		}else{
			fmt.Println("Wallpaper downloaded to cache")
		}
		return nil
	},
}

func init(){
	fetchCmd.Flags().BoolVarP(&noSet, "no-set", "n", false, "Download and save to history without setting the wallpaper")
	rootCmd.AddCommand(fetchCmd)
}