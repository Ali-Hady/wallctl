package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var configSetCmd = &cobra.Command{
	Use:       "set <key> <value>",
	Short:     "Set a configuration option",
	Args:      cobra.ExactArgs(2),
	ValidArgs: []string{"set-on-fetch"},
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		val := args[1]

		switch key {
		case "set-on-fetch":
			parsedBool, err := strconv.ParseBool(val)
			if err != nil {
				return fmt.Errorf("invalid value %q for %s: expected true or false", val, key)
			}
			if err := appStore.ChangeSetOnNew(parsedBool); err != nil {
				return fmt.Errorf("failed to persist setting: %w", err)
			}
			fmt.Printf("Updated %s to %t\n", key, parsedBool)
			return nil

		default:
			return fmt.Errorf("unknown configuration key: %q", key)
		}
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
