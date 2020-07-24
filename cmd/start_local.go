package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Starts a ssh local port forwarding tunnel",
	Long:  "Starts a ssh local port forwarding tunnel",
	Run: func(cmd *cobra.Command, arg []string) {
		start("", tunnelFlags)
	},
}

func init() {
	err := bindFlags(tunnelFlags, localCmd)
	if err != nil {
		log.WithError(err).Error("error parsing command line arguments")
		os.Exit(1)
	}

	startCmd.AddCommand(localCmd)
}
