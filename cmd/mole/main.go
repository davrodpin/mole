package main

import (
	"fmt"
	"os"

	"github.com/davrodpin/mole/tunnel"
	log "github.com/sirupsen/logrus"
)

var version = "unversioned"

func main() {

	cmd := &cmd{}
	err := cmd.Parse(os.Args)
	if err != nil {
		cmd.PrintUsage()
		os.Exit(1)
	}

	switch cmd.command {
	case "help":
		cmd.PrintUsage()
		os.Exit(0)
	case "version":
		fmt.Printf("mole %s\n", version)
		os.Exit(1)
	case "new":
		log.SetOutput(os.Stdout)

		if cmd.verbose {
			log.SetLevel(log.DebugLevel)
		}

		log.WithFields(log.Fields{
			"options": cmd.String(),
		}).Debug("cli options")

		s, err := tunnel.NewServer(cmd.server.User, cmd.server.Address(), cmd.key)
		if err != nil {
			log.Fatalf("error processing server options: %v\n", err)
		}

		log.Debugf("server: %s", s)

		t := tunnel.New(cmd.local.String(), s, cmd.remote.String())

		err = t.Start()
		if err != nil {
			log.WithFields(log.Fields{
				"tunnel": t.String(),
			}).Errorf("%v", err)

			os.Exit(1)
		}
	}
}
