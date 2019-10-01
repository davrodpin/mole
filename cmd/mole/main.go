package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/awnumar/memguard"
	"github.com/davrodpin/mole/cli"
	"github.com/davrodpin/mole/storage"
	"github.com/davrodpin/mole/tunnel"
	"github.com/gofrs/uuid"
	daemon "github.com/sevlyar/go-daemon"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

var version = "unversioned"
var instancesDir string

func main() {
	// memguard is used to securely keep sensitive information in memory.
	// This call makes sure all data will be destroy when the program exits.
	defer memguard.Purge()

	app := cli.New(os.Args)
	err := app.Parse()

	if err != nil {
		fmt.Printf("%v\n", err)
		app.PrintUsage()
		os.Exit(1)
	}

	log.SetOutput(os.Stdout)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("error starting mole: %v", err)
		os.Exit(1)
	}

	instancesDir = fmt.Sprintf("%s/.mole/instances", home)

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
		err := start(app)
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
		err := lsAliases()
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

	appFromAlias, err := alias2app(conf)
	if err != nil {
		log.WithFields(log.Fields{
			"alias": app.Alias,
		}).Errorf("error starting mole: %v", err)

		return err
	}

	appFromAlias.Alias = app.Alias
	// if use -detach when -start but none -detach in storage
	if app.Detach {
		appFromAlias.Detach = true
	}

	return start(appFromAlias)
}

func start(app *cli.App) error {
	if app.Detach {
		var alias string
		if app.Alias != "" {
			alias = app.Alias
		} else {
			u, err := uuid.NewV4()
			if err != nil {
				log.Errorf("error could not generate uuid: %v", err)
				return err
			}
			alias = u.String()[:8]
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

	s.Insecure = app.Insecure

	s.Key.HandlePassphrase(func() ([]byte, error) {
		fmt.Printf("The key provided is secured by a password. Please provide it below:\n")
		fmt.Printf("Password: ")
		p, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Printf("\n")
		return p, err
	})

	log.Debugf("server: %s", s)

	local := make([]string, len(app.Local))
	for i, r := range app.Local {
		local[i] = r.String()
	}

	remote := make([]string, len(app.Remote))
	for i, r := range app.Remote {
		if r.Port == "" {
			err := fmt.Errorf("missing port in remote address: %s", r.String())
			log.Error(err)
			return err
		}

		remote[i] = r.String()
	}

	channels, err := tunnel.BuildSSHChannels(s.Name, local, remote)
	if err != nil {
		return err
	}

	t, err := tunnel.New(s, channels)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}

	t.KeepAliveInterval = app.KeepAliveInterval

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
