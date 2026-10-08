package cmd

import (
	"fmt"

	"github.com/Ali-Hady/wallctl/internal/core"
	"github.com/spf13/cobra"
)

var (
	noSet bool
	count int
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch and apply wallpapers",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Resolve whether to apply wallpaper
		// Defaults to stored config unless user explicitly passes --no-set
		applyWallpaper := appStore.GetSetOnNew()
		if cmd.Flags().Changed("no-set") {
			applyWallpaper = !noSet
		}

		history, _ := appStore.List(1, false)
		isFirstRun := len(history) == 0

		// 2. Batch fetch logic
		if batcher, ok := appSource.(core.BatchSource); ok && (isFirstRun || count > 1) {
			batchCount := count
			if isFirstRun && batchCount <= 1 {
				batchCount = 7
				fmt.Printf("First run detected! Seeding history with the last %d Bing wallpapers...\n", batchCount)
			} else {
				fmt.Printf("Fetching batch of %d wallpapers...\n", batchCount)
			}

			images, err := batcher.FetchBatch(cmd.Context(), imagesDir, batchCount)
			if err != nil {
				return fmt.Errorf("batch fetch failed: %w", err)
			}
			if len(images) == 0 {
				fmt.Println("No wallpapers returned by source.")
				return nil
			}

			// Bing returns [Today, Yesterday, ..., 7 days ago].
			// Save older images in reverse so newest (index 0) is saved last.
			for i := len(images) - 1; i >= 1; i-- {
				if err := appStore.SaveNoActive(images[i]); err != nil {
					return fmt.Errorf("failed to save image %s: %w", images[i].ID, err)
				}
			}

			newestImg := images[0]

			if !applyWallpaper {
				if err := appStore.SaveNoActive(newestImg); err != nil {
					return fmt.Errorf("failed to save image %s: %w", newestImg.ID, err)
				}
				fmt.Printf("Downloaded %d wallpapers to cache (desktop not modified).\n", len(images))
				return nil
			}

			fmt.Printf("Setting latest (%q) as wallpaper...\n", newestImg.Title)
			if err := core.SetDesktopWallpaper(newestImg.LocalPath); err != nil {
				appStore.SaveNoActive(newestImg)
				return fmt.Errorf("failed to apply wallpaper: %w", err)
			}

			if err := appStore.Save(newestImg); err != nil {
				return fmt.Errorf("failed to save image %s: %w", newestImg.ID, err)
			}
			fmt.Println("Wallpaper applied successfully!")
			return nil
		}

		// 3. Single fetch fallback
		fmt.Println("Fetching latest wallpaper...")
		img, err := appSource.Fetch(cmd.Context(), imagesDir)
		if err != nil {
			return fmt.Errorf("failed to fetch wallpaper: %w", err)
		}

		if !applyWallpaper {
			if err := appStore.SaveNoActive(img); err != nil {
				return fmt.Errorf("failed to save wallpaper metadata: %w", err)
			}
			fmt.Printf("Saved %q to cache (%s)\n", img.Title, img.LocalPath)
			return nil
		}

		fmt.Printf("Setting %q as wallpaper...\n", img.Title)
		if err := core.SetDesktopWallpaper(img.LocalPath); err != nil {
			appStore.SaveNoActive(img)
			return fmt.Errorf("failed to set wallpaper: %w", err)
		}

		if err := appStore.Save(img); err != nil {
			return fmt.Errorf("failed to save wallpaper metadata: %w", err)
		}

		fmt.Println("Wallpaper applied successfully!")
		return nil
	},
}

func init() {
	// Set default to false and 1 so it doesn't break single-fetch mode or panic on uninitialized appStore
	fetchCmd.Flags().BoolVarP(&noSet, "no-set", "n", false, "Download without setting as desktop background")
	fetchCmd.Flags().IntVarP(&count, "count", "c", 1, "Number of past wallpapers to fetch (if supported by source)")
	rootCmd.AddCommand(fetchCmd)
}
