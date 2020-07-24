package cmd

import (
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a ssh tunnel",
	Long:  "Starts a ssh tunnel by either its port forwarding type or by a given alias",
	Args:  cobra.MinimumNArgs(1),
	Run:   func(cmd *cobra.Command, arg []string) {},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
