package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Starts a ssh remote port forwarding tunnel",
	Long: `Remote Forwarding allows anyone to expose a service running locally to a remote machine.

This could be particular useful for giving someone on the outside access to an internal web application, for example.

Source endpoints are addresses on the jump server where clients can connect to access services running on the corresponding destination endpoints.
Destination endpoints are addresses of services running on the same machine where mole is getting executed.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		tunnelFlags.TunnelType = "remote"
		return nil
	},
	Run: func(cmd *cobra.Command, arg []string) {
		start("", tunnelFlags)
	},
}

func init() {
	err := bindFlags(tunnelFlags, remoteCmd)
	if err != nil {
		log.WithError(err).Error("error parsing command line arguments")
		os.Exit(1)
	}

	startCmd.AddCommand(remoteCmd)
}
