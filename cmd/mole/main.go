package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/davrodpin/mole/cli"
	"github.com/davrodpin/mole/storage"
	"github.com/davrodpin/mole/tunnel"
	uuid "github.com/satori/go.uuid"
	daemon "github.com/sevlyar/go-daemon"
	log "github.com/sirupsen/logrus"
)

var version = "unversioned"
var instancesDir string

func main() {

	app := cli.New(os.Args)
	err := app.Parse()

	if err != nil {
		fmt.Printf("%v\n", err)
		app.PrintUsage()
		os.Exit(1)
	}
	log.SetOutput(os.Stdout)

	instancesDir = fmt.Sprintf("%s/.mole/instances", os.Getenv("HOME"))

	err = createInstancesDir()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

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
	case "stop":
		err := stopDaemon(*app)
		if err != nil {
			os.Exit(1)
		}
	case "aliases":
		err := lsAliases(*app)
		if err != nil {
			os.Exit(1)
		}
	}
}

func stopDaemon(app cli.App) error {
	daemonDir := fmt.Sprintf("%s/%s", instancesDir, app.Stop)
	pidPathName := fmt.Sprintf("%s/pid", daemonDir)
	logPathName := fmt.Sprintf("%s/mole.log", daemonDir)

	if _, err := os.Stat(pidPathName); os.IsNotExist(err) {
		return fmt.Errorf("an instance of mole, %s, is not running", app.Stop)
	}

	cntxt := &daemon.Context{
		PidFileName: pidPathName,
		PidFilePerm: 0644,
		LogFileName: logPathName,
		LogFilePerm: 0640,
		Umask:       027,
		Args:        os.Args,
	}

	d, err := cntxt.Search()
	if err != nil {
		return err
	}

	err = d.Kill()
	if err != nil {
		return err
	}

	removePath := fmt.Sprintf("%s/%s", instancesDir, app.Stop)
	err = os.RemoveAll(removePath)
	if err != nil {
		return err
	}

	return nil
}

func createInstancesDir() error {
	_, err := os.Stat(instancesDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(instancesDir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func startDaemonProcess(aliasName string) error {
	cntxt := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "mole.log",
		LogFilePerm: 0640,
		Umask:       027,
		Args:        os.Args,
	}
	d, err := cntxt.Reborn()
	if err != nil {
		return err
	}
	if d != nil {
		daemonDir := fmt.Sprintf("%s/%s", instancesDir, aliasName)
		pidPathName := fmt.Sprintf("%s/pid", daemonDir)
		logPathName := fmt.Sprintf("%s/mole.log", daemonDir)
		if _, err := os.Stat(daemonDir); os.IsNotExist(err) {
			err := os.Mkdir(daemonDir, 0755)
			if err != nil {
				return err
			}
		}
		if _, err := os.Stat(pidPathName); !os.IsNotExist(err) {
			return fmt.Errorf("an instance of mole, %s, seems to be already running", aliasName)
		}
		err := os.Rename("pid", pidPathName)
		if err != nil {
			return err
		}
		err = os.Rename("mole.log", logPathName)
		if err != nil {
			return err
		}
		log.Infof("execute \"mole -stop %s\" if you like to stop it at any time", aliasName)
		os.Exit(0)
	}
	defer cntxt.Release()
	return nil
}

func startFromAlias(app cli.App) error {
	conf, err := storage.FindByName(app.Alias)
	if err != nil {
		log.WithFields(log.Fields{
			"alias": app.Alias,
		}).Errorf("error starting mole: %v", err)

		return err
	}

	appFromAlias := alias2app(conf)
	appFromAlias.Alias = app.Alias
	// if use -detach when -start but none -detach in storage
	if app.Detach {
		appFromAlias.Detach = true
	}

	return start(appFromAlias)
}

func start(app cli.App) error {
	if app.Detach {
		var alias string
		if app.Alias != "" {
			alias = app.Alias
		} else {
			alias = uuid.Must(uuid.NewV4()).String()[:8]
		}
		err := startDaemonProcess(alias)
		if err != nil {
			log.WithFields(log.Fields{
				"daemon": app.Alias,
			}).Errorf("error starting mole: %v", err)
			return err
		}
	}
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

	s.Insecure = app.InsecureMode

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

func lsAliases(app cli.App) error {
	tunnels, err := storage.FindAll()
	if err != nil {
		return err
	}

	aliases := []string{}
	for alias := range tunnels {
		aliases = append(aliases, alias)
	}

	fmt.Printf("alias list: %s\n", strings.Join(aliases, ", "))

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
		Detach:  app.Detach,
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
		Detach:  t.Detach,
	}
}
