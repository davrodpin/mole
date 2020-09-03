package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/davrodpin/mole/app"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	showLogsCmd = &cobra.Command{
		Use:   "logs [name]",
		Short: "Shows log messages of a detached running application instance",
		Long:  "Shows log messages of a detached running application instance",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				id = args[0]
			}

			return nil
		},
		Run: func(cmd *cobra.Command, arg []string) {
			lfl, err := app.GetLogFileLocation(id)
			if err != nil {
				log.WithError(err).Error("error stopping detached mole instance")
				os.Exit(1)
			}

			file, err := os.Open(lfl)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"id": id,
				}).Error("error opening log file")
				os.Exit(1)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"id": id,
				}).Error("error reading log file")
				os.Exit(1)
			}
		},
	}
)

func init() {
	showCmd.AddCommand(showLogsCmd)
}
