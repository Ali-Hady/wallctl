package cmd

import (
	"fmt"

	"github.com/reujab/wallpaper"
	"github.com/spf13/cobra"
)

var noSet bool

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch and apply the latest wallpaper",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Fetching latest wallpaper...")

		img, err := appSource.Fetch(cmd.Context(), imagesDir)
		if err != nil {
			return fmt.Errorf("failed to fetch wallpaper: %w", err)
		}

		if err := appStore.Save(img); err != nil {
			return fmt.Errorf("failed to save wallpaper metadata: %w", err)
		}

		if noSet {
			fmt.Printf("Saved %q to cache (%s)\n", img.Title, img.LocalPath)
			return nil
		}

		fmt.Printf("Setting %q as wallpaper...\n", img.Title)
		if err := wallpaper.SetFromFile(img.LocalPath); err != nil {
			return fmt.Errorf("failed to set wallpaper: %w", err)
		}

		fmt.Println("Wallpaper applied successfully!")
		return nil
	},
}

func init() {
	fetchCmd.Flags().BoolVarP(&noSet, "no-set", "n", false, "Download and save to history without setting as desktop background")
	rootCmd.AddCommand(fetchCmd)
}
