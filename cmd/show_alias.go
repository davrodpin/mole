package cmd

import (
	"fmt"

	"github.com/davrodpin/mole/alias"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var showAliasCmd = &cobra.Command{
	Use:   "alias [name]",
	Short: "Shows configuration details about ssh tunnel aliases",
	Long:  "Shows configuration details about ssh tunnel aliases",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			aliasName = args[0]
		}

		return nil
	},
	Run: func(cmd *cobra.Command, arg []string) {
		var aliases string
		var err error

		if aliasName == "" {
			aliases, err = alias.ShowAll()
		} else {
			aliases, err = alias.Show(aliasName)
		}

		if err != nil {
			log.Errorf("could not show alias: %v", err)
		}

		fmt.Printf("%s\n", aliases)
	},
}

func init() {
	showCmd.AddCommand(showAliasCmd)
}
