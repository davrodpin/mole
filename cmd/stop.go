package cmd

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	stopCmd = &cobra.Command{
		Use:   "stop [alias name or id]",
		Short: "Stops a ssh tunnel",
		Long:  "Stops a ssh tunnel by either an auto generated id or a given alias",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("alias name or id not provided")
			}

			id = args[0]

			return nil
		},
		Run: func(cmd *cobra.Command, arg []string) {
			err := stop(id)
			if err != nil {
				log.WithError(err).Error("error stopping detached mole instance")
				os.Exit(1)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(stopCmd)
}
