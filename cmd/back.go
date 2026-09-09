package cmd

import (
	"errors"
	"fmt"

	"github.com/Ali-Hady/wallctl/internal/core"
	"github.com/spf13/cobra"
)

var backCmd = &cobra.Command{
	Use:   "back",
	Short: "Fetch and apply the previous wallpaper in history",
	RunE: func(cmd *cobra.Command, args []string) error {
		favs, err := cmd.Flags().GetBool("favs")
		if err != nil {
			return err
		}

		img, err := appStore.Shift(-1, favs)
		if err != nil {
			if errors.Is(err, core.ErrOutOfBounds) {
				if favs {
					fmt.Println("Already at the oldest favorite wallpaper.")
					return nil
				}
				fmt.Println("Already at the oldest wallpaper in history.")
				return nil
			}
			if errors.Is(err, core.ErrEmptyStore) {
				fmt.Println("History is empty. Run 'wallctl fetch' first.")
				return nil
			}
			return fmt.Errorf("navigating history: %w", err)
		}

		fmt.Printf("Setting %q as wallpaper...\n", img.Title)
		if err := core.SetDesktopWallpaper(img.LocalPath); err != nil {
			return fmt.Errorf("failed to apply wallpaper: %w", err)
		}

		fmt.Printf("Applied: %s (%s)\n", img.Title, img.ID)
		return nil
	},
}

func init() {
	backCmd.Flags().BoolP("favs", "f", false, "Traverse only favorite wallpapers")
	rootCmd.AddCommand(backCmd)
}
