package main

import (
	"fmt"
	"os"

	"github.com/davrodpin/mole/cli"
	"github.com/davrodpin/mole/tunnel"
	log "github.com/sirupsen/logrus"
)

var version = "unversioned"

func main() {
	app := cli.New(os.Args)
	err := app.Parse()
	if err != nil {
		app.PrintUsage()
		os.Exit(1)
	}

	switch app.Command {
	case "help":
		app.PrintUsage()
		os.Exit(0)
	case "version":
		fmt.Printf("mole %s\n", version)
		os.Exit(0)
	case "new":
		t, err := newTunnel(*app)
		if err != nil {
			log.WithFields(log.Fields{
				"tunnel": t.String(),
			}).Errorf("%v", err)

			os.Exit(1)
		}
	case "new-alias":
		err := newAlias(*app)
		if err != nil {
			log.WithFields(log.Fields{
				"alias": app.Alias,
			}).Errorf("alias could not be created: %v", err)

			os.Exit(1)
		}
	}
}

func newTunnel(app cli.App) (*tunnel.Tunnel, error) {
	log.SetOutput(os.Stdout)

	if app.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.WithFields(log.Fields{
		"options": app.String(),
	}).Debug("cli options")

	s, err := tunnel.NewServer(app.Server.User, app.Server.Address(), app.Key)
	if err != nil {
		log.Fatalf("error processing server options: %v\n", err)
	}

	log.Debugf("server: %s", s)

	t := tunnel.New(app.Local.String(), s, app.Remote.String())

	return t, t.Start()
}

func newAlias(all cli.App) error {
	return nil
}
