package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/davrodpin/mole/alias"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	addAliasLocalDoc = `Adds an alias for a ssh tunneling configuration by saving a set of start
command flags so it can be reused later.

The alias configuration file is saved under the ".mole" directory, inside the
user home directory.`
)

var addAliasLocalCmd = &cobra.Command{
	Use:   "local [name]",
	Short: "Adds an alias for a ssh tunneling configuration",
	Long:  fmt.Sprintf("%s\n%s", addAliasLocalDoc, LocalForwardDoc),
	Args: func(cmd *cobra.Command, args []string) error {
		fmt.Printf(">>> %v\n", args)
		if len(args) < 1 {
			return errors.New("alias name not provided")
		}

		conf.TunnelType = "local"
		aliasName = args[0]

		return nil
	},
	Run: func(cmd *cobra.Command, arg []string) {
		if err := alias.Add(conf.ParseAlias(aliasName)); err != nil {
			log.WithError(err).Error("failed to add tunnel alias")
			os.Exit(1)
		}
	},
}

func init() {
	err := bindFlags(conf, addAliasLocalCmd)
	if err != nil {
		log.WithError(err).Error("error parsing command line arguments")
		os.Exit(1)
	}

	addAliasCmd.AddCommand(addAliasLocalCmd)
}
