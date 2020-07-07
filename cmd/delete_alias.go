package cmd

import (
	"errors"
	"os"

	"github.com/davrodpin/mole/alias"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteAliasCmd = &cobra.Command{
	Use:   "alias [name]",
	Short: "Deletes an alias for a ssh tunneling configuration",
	Long:  "Deletes an alias for a ssh tunneling configuration",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("alias name not provided")
		}

		aliasName = args[0]

		return nil
	},
	Run: func(cmd *cobra.Command, arg []string) {

		err := alias.Delete(aliasName)
		if err != nil {
			log.Errorf("failed to delete tunnel alias: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteAliasCmd)
}
