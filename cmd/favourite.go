package cmd
import (
	"fmt"
	"github.com/spf13/cobra"
)
var alias string
var favouriteCmd = &cobra.Command{
	Use: "favourite [id|current]",
	Short: "Mark a wallpaper as favorite",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		if target == "current"{
			currentImg, err := appStore.Current()
			if err!=nil{
				return fmt.Errorf("failed to get current wallpaper: %w", err)
			}
			target = currentImg.ID
		}
		err := appStore.MarkFavorite(target,alias)
		if err!=nil{
			return fmt.Errorf("failed to mark wallpaper as favorite: %w", err)
		}
		if alias != ""{
			fmt.Printf("Wallpaper %s marked as favorite (Alias: %s)\n", target, alias)
		}else{
			fmt.Printf("Wallpaper %s marked as favorite\n", target)
		}
		return nil		
	},
}
func init(){
	favouriteCmd.Flags().StringVarP(&alias, "alias", "a", "", "Alias for the wallpaper")
	rootCmd.AddCommand(favouriteCmd)
}
	
