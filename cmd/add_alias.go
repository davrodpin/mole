package cmd

import (
	"errors"
	"os"

	"github.com/davrodpin/mole/alias"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var addAliasCmd = &cobra.Command{
	Use:   "alias local [name]",
	Short: "Adds an alias for a ssh tunneling configuration",
	Long: `Adds an alias for a ssh tunneling configuration by saving a set of start
command flags so it can be reused later.
The alias configuration file is saved to ".mole", under your home directory.
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("tunnel type not provided")
		}

		if len(args) < 2 {
			return errors.New("alias name not provided")
		}

		tunnelFlags.TunnelType = args[0]
		aliasName = args[1]

		return nil
	},
	Run: func(cmd *cobra.Command, arg []string) {
		if err := alias.Add(tunnelFlags.ParseAlias(aliasName)); err != nil {
			log.WithError(err).Error("failed to add tunnel alias")
			os.Exit(1)
		}
	},
}

func init() {
	err := bindFlags(tunnelFlags, addAliasCmd)
	if err != nil {
		log.WithError(err).Error("error parsing command line arguments")
		os.Exit(1)
	}

	addCmd.AddCommand(addAliasCmd)
}
