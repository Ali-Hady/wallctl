package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var alias string

var favouriteCmd = &cobra.Command{
	Use:     "favourite [id|alias]",
	Aliases: []string{"favorite", "fav"},
	Short:   "Mark a wallpaper as favorite",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetID := ""

		if len(args) == 0 || args[0] == "current" {
			currentImg, err := appStore.Current()
			if err != nil {
				return fmt.Errorf("failed to get current wallpaper: %w", err)
			}
			targetID = currentImg.ID
		} else {
			// Resolve by ID or existing alias to get the true ID
			img, err := appStore.Get(args[0])
			if err != nil {
				return fmt.Errorf("wallpaper not found: %w", err)
			}
			targetID = img.ID
		}

		if err := appStore.MarkFavorite(targetID, alias); err != nil {
			return fmt.Errorf("failed to mark wallpaper as favorite: %w", err)
		}

		if alias != "" {
			fmt.Printf("Wallpaper %s marked as favorite (alias: %q)\n", targetID, alias)
		} else {
			fmt.Printf("Wallpaper %s marked as favorite\n", targetID)
		}

		return nil
	},
}

func init() {
	favouriteCmd.Flags().StringVarP(&alias, "alias", "a", "", "Assign a memorable alias to this wallpaper")
	rootCmd.AddCommand(favouriteCmd)
}
