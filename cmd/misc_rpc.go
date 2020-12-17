package cmd

import (
	"fmt"
	"os"

	"github.com/davrodpin/mole/mole"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	method, params string

	miscRpcCmd = &cobra.Command{
		Use:   "rpc [alias or id] [method] [params]",
		Short: "Executes a remote procedure call on a given mole instance",
		Long:  "Executes a remote procedure call on a given mole instance",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments.")
			}

			id = args[0]
			method = args[1]

			if len(args) == 3 {
				params = args[2]
			}

			return nil
		},
		Run: func(cmd *cobra.Command, arg []string) {
			resp, err := mole.Rpc(id, method, params)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"id": id,
				}).Error("error executing remote procedure.")
				os.Exit(1)
			}

			fmt.Printf(resp)
		},
	}
)

func init() {
	miscCmd.AddCommand(miscRpcCmd)
}
