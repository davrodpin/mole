package cmd

import (
	"fmt"
	"os"

	"github.com/davrodpin/mole/mole"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	showInstancesCmd = &cobra.Command{
		Use:   "instances [name]",
		Short: "Shows runtime information about application instances",
		Long: `Shows runtime information about application instances.

Only instances with rpc enabled will be shown by this command.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				id = args[0]
			}

			return nil
		},
		Run: func(cmd *cobra.Command, arg []string) {
			var err error
			var formatter mole.Formatter

			if id == "" {
				formatter, err = mole.ShowInstances()
			} else {
				formatter, err = mole.ShowInstance(id)
			}

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"id": id,
				}).Error("could not retrieve information about application instance(s)")
				os.Exit(1)
			}

			out, err := formatter.Format("toml")
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"id": id,
				}).Error("error converting output")
				os.Exit(1)
			}

			fmt.Printf("%s\n", out)

		},
	}
)

func init() {
	showCmd.AddCommand(showInstancesCmd)
}
