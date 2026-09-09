package cmd
import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/spf13/cobra"
	"github.com/Ali-Hady/wallctl/internal/core"
)

var(
	appStore *core.JSONStore
	appSource core.Source
	imagesDir string
)
var rootCmd = &cobra.Command{
	Use:	"wallctl",
	Short:	"Wallpaper manager for Linux",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error{
		cacheRoot, err := os.UserCacheDir()
		if err != nil{
			return fmt.Errorf("failed to get cache dir: %w", err)
		}
		imagesDir = filepath.Join(cacheRoot, "wallctl", "images")
		if err:=os.MkdirAll(imagesDir,0755); err!=nil{
			return fmt.Errorf("failed to create cache dir: %w", err)
		}
		appStore = &core.JSONStore{
			FilePath: filepath.Join(cacheRoot,"wallctl","index.json"),
			Data: &core.Data{},
		}
		appSource = core.NewBingSource()
		return nil
	},
	
}
func Execute(){
	if err:= rootCmd.Execute(); err!=nil{
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

