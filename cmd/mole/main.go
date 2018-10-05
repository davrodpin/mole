package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/davrodpin/mole/tunnel"
	log "github.com/sirupsen/logrus"
)

type hostFlag struct {
	User string
	Host string
	Port string
}

func (f hostFlag) String() string {
	var s string
	if f.User == "" {
		s = f.Address()
	} else {
		s = fmt.Sprintf("%s@%s", f.User, f.Address())
	}

	return s
}

func (f *hostFlag) Set(value string) error {
	re := regexp.MustCompile("(?P<user>.+@)?(?P<host>[0-9a-zA-Z.-]+)(?P<port>:[0-9]+)?")

	match := re.FindStringSubmatch(value)
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i == 0 {
			continue
		}

		result[name] = match[i]
	}

	if result["host"] == "" {
		return fmt.Errorf("error parsing argument. Host must be provided.")
	}

	f.User = strings.Trim(result["user"], "@")
	f.Host = result["host"]
	f.Port = strings.Trim(result["port"], ":")

	return nil
}

func (f hostFlag) Address() string {
	if f.Port == "" {
		return fmt.Sprintf("%s", f.Host)
	}

	return fmt.Sprintf("%s:%s", f.Host, f.Port)
}

var (
	localFlag  hostFlag
	remoteFlag hostFlag
	serverFlag hostFlag
	keyFlag    string
	vFlag      bool
	helpFlag   bool
)

func init() {
	flag.Var(&localFlag, "local", "local endpoint address: <host>:<port>")
	flag.Var(&remoteFlag, "remote", "remote endpoing address: <host>:<port>")
	flag.Var(&serverFlag, "server", "server address: [<user>@]<host>[:<port>]")
	flag.StringVar(&keyFlag, "key", "", "server authentication key file path")
	flag.BoolVar(&vFlag, "v", false, "increases log verbosity")
	flag.BoolVar(&helpFlag, "help", false, "list all options available")
}

func main() {
	if len(os.Args[1:]) == 0 {
		fmt.Printf("%s\n", usage())
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()

	if helpFlag {
		fmt.Printf("%s\n", usage())
		flag.PrintDefaults()
		os.Exit(0)
	}

	log.SetOutput(os.Stdout)

	if vFlag {
		log.SetLevel(log.DebugLevel)
	}

	log.WithFields(log.Fields{
		"local":  localFlag.String(),
		"remote": remoteFlag.String(),
		"server": serverFlag.String(),
		"key":    keyFlag,
		"v":      vFlag,
	}).Debug("cli options")

	s, err := tunnel.NewServer(serverFlag.User, serverFlag.Address(), keyFlag)
	if err != nil {
		log.Fatalf("error processing server options: %v\n", err)
	}

	log.Debugf("server: %s", s)

	t := tunnel.New(localFlag.String(), s, remoteFlag.String())

	err = t.Start()
	if err != nil {
		log.WithFields(log.Fields{
			"tunnel": t.String(),
		}).Errorf("%v", err)

		os.Exit(1)
	}
}

func usage() string {
	return `usage:
  mole [-v] [-local <host>:<port>] -remote <host>:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -help
	`
}
