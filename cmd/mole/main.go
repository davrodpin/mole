package main

import (
	"fmt"
	"os"

	"github.com/davrodpin/mole/cli"
	"github.com/davrodpin/mole/storage"
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

	log.SetOutput(os.Stdout)

	switch app.Command {
	case "help":
		app.PrintUsage()
	case "version":
		fmt.Printf("mole %s\n", version)
	case "start":
		err := start(*app)
		if err != nil {
			os.Exit(1)
		}
	case "start-from-alias":
		err := startFromAlias(*app)
		if err != nil {
			os.Exit(1)
		}
	case "new-alias":
		err := newAlias(*app)
		if err != nil {
			os.Exit(1)
		}
	case "rm-alias":
		err := rmAlias(*app)
		if err != nil {
			os.Exit(1)
		}

	}
}

func startFromAlias(app cli.App) error {
	conf, err := storage.FindByName(app.Alias)
	if err != nil {
		log.WithFields(log.Fields{
			"alias": app.Alias,
		}).Errorf("error starting mole: %v", err)

		return err
	}

	return start(alias2app(conf))
}

func start(app cli.App) error {
	if app.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.WithFields(log.Fields{
		"options": app.String(),
	}).Debug("cli options")

	s, err := tunnel.NewServer(app.Server.User, app.Server.Address(), app.Key)
	if err != nil {
		log.Errorf("error processing server options: %v\n", err)

		return err
	}

	s.SetInsecureMode(app.InsecureMode)

	log.Debugf("server: %s", s)

	t := tunnel.New(app.Local.String(), s, app.Remote.String())

	if err = t.Start(); err != nil {
		log.WithFields(log.Fields{
			"tunnel": t.String(),
		}).Errorf("%v", err)

		return err
	}

	return nil
}

func newAlias(app cli.App) error {
	_, err := storage.Save(app.Alias, app2alias(app))

	if err != nil {
		log.WithFields(log.Fields{
			"alias": app.Alias,
		}).Errorf("alias could not be created: %v", err)

		return err
	}

	return nil
}

func rmAlias(app cli.App) error {
	_, err := storage.Remove(app.Alias)
	if err != nil {
		return err
	}

	return nil
}

func app2alias(app cli.App) *storage.Tunnel {
	return &storage.Tunnel{
		Local:   app.Local.String(),
		Remote:  app.Remote.String(),
		Server:  app.Server.String(),
		Key:     app.Key,
		Verbose: app.Verbose,
		Help:    app.Help,
		Version: app.Version,
	}
}

func alias2app(t *storage.Tunnel) cli.App {
	local := cli.HostInput{}
	local.Set(t.Local)

	remote := cli.HostInput{}
	remote.Set(t.Remote)

	server := cli.HostInput{}
	server.Set(t.Server)

	return cli.App{
		Command: "start",
		Local:   local,
		Remote:  remote,
		Server:  server,
		Key:     t.Key,
		Verbose: t.Verbose,
		Help:    t.Help,
		Version: t.Version,
	}
}
