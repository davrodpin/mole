package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes an alias for a ssh tunneling configuration",
	Args:  cobra.MinimumNArgs(1),
	Run:   func(cmd *cobra.Command, arg []string) {},
}
