package cmd

import (
	"github.com/spf13/cobra"
)

var miscCmd = &cobra.Command{
	Use:   "misc",
	Short: "A set of miscelaneous commands",
	Args:  cobra.MinimumNArgs(1),
	Run:   func(cmd *cobra.Command, arg []string) {},
}

func init() {
	rootCmd.AddCommand(miscCmd)
}
