package cmd

import (
	"github.com/spf13/cobra"
)

var addAliasCmd = &cobra.Command{
	Use:   "alias local [name]",
	Short: "Adds an alias for a ssh tunneling configuration",
	Long: `Adds an alias for a ssh tunneling configuration by saving a set of start
command flags so it can be reused later.

The alias configuration file is saved under the ".mole" directory, inside the
user home directory.
	`,

	Args: cobra.MinimumNArgs(1),
	Run:  func(cmd *cobra.Command, arg []string) {},
}

func init() {
	addCmd.AddCommand(addAliasCmd)
}
