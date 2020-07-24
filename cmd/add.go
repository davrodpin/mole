package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds an alias for a ssh tunneling configuration",
	Args:  cobra.MinimumNArgs(1),
	Run:   func(cmd *cobra.Command, arg []string) {},
}
