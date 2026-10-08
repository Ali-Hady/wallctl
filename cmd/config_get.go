package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var configGetCmd = &cobra.Command{
	Use:       "get <key>",
	Short:     "Get a configuration option",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"set-on-fetch"},
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])

		switch key {
		case "set-on-fetch":
			val := appStore.GetSetOnNew()
			fmt.Printf("%s: %t\n", key, val)
			return nil

		default:
			return fmt.Errorf("unknown configuration key: %q", key)
		}
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
}
