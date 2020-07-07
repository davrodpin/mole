package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "unversioned"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the version for mole",
		Run: func(cmd *cobra.Command, arg []string) {
			fmt.Printf("%v\n", version)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
