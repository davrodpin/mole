package cmd

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startAliasCmd = &cobra.Command{
	Use:   "alias [name]",
	Short: "Starts a ssh tunnel by alias",
	Long:  "Starts a ssh tunnel by alias",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("alias name not provided")
		}

		aliasName = args[0]

		return nil
	},
	Run: func(cmd *cobra.Command, arg []string) {
		err := startWithAlias(aliasName)
		if err != nil {
			log.Errorf("failed to start tunnel: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	startCmd.AddCommand(startAliasCmd)
}
