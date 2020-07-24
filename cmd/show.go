package cmd

import (
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows configuration details about ssh tunnel aliases",
	Args:  cobra.MinimumNArgs(1),
	Run:   func(cmd *cobra.Command, arg []string) {},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
