package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ali-Hady/wallctl/internal/core"
	"github.com/spf13/cobra"
)

var (
	appStore  *core.JSONStore
	appSource core.Source
	imagesDir string
)

var rootCmd = &cobra.Command{
	Use:   "wallctl",
	Short: "Wallpaper manager for Linux",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cacheRoot, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("failed to get cache dir: %w", err)
		}

		appDir := filepath.Join(cacheRoot, "wallctl")
		imagesDir = filepath.Join(appDir, "images")
		if err := os.MkdirAll(imagesDir, 0755); err != nil {
			return fmt.Errorf("failed to create images dir: %w", err)
		}

		indexPath := filepath.Join(appDir, "index.json")
		store, err := core.NewJSONStore(indexPath)
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}
		appStore = store

		appSource = core.NewBingSource()
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
