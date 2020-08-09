package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Starts a ssh local port forwarding tunnel",
	Long: `Local Forwarding allows anyone to access outside services like they were
running locally on the source machine.

This could be particular useful for accesing web sites, databases or any kind of
service the source machine does not have direct access to.

Source endpoints are addresses on the same machine where mole is getting executed where clients can connect to access services on the corresponding destination endpoints.
Destination endpoints are adrresess that can be reached from the jump server.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		tunnelFlags.TunnelType = "local"
		return nil
	},
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
