package cmd

import (
	"errors"
	"fmt"

	"github.com/Ali-Hady/wallctl/internal/core"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:     "set <id|alias>",
	Aliases: []string{"use", "apply"},
	Short:   "Set the desktop wallpaper by ID or alias",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		img, err := appStore.Get(target)
		if err != nil {
			if errors.Is(err, core.ErrNotFound) {
				return fmt.Errorf("no wallpaper found matching %q (check 'wallctl history')", target)
			}
			return fmt.Errorf("resolving wallpaper %q: %w", target, err)
		}

		fmt.Printf("Setting %q as wallpaper...\n", img.Title)
		if err := core.SetDesktopWallpaper(img.LocalPath); err != nil {
			return fmt.Errorf("failed to apply wallpaper: %w", err)
		}

		// Update CurID so 'next' and 'back' continue navigation from here
		if err := appStore.Save(img); err != nil {
			return fmt.Errorf("updating active wallpaper state: %w", err)
		}

		fmt.Printf("Applied: %s (%s)\n", img.Title, img.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
