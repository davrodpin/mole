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

var version = "unversioned"

var (
	localFlag   hostFlag
	remoteFlag  hostFlag
	serverFlag  hostFlag
	keyFlag     string
	vFlag       bool
	helpFlag    bool
	versionFlag bool
)

func init() {
	flag.Var(&localFlag, "local", "set local endpoint address: <host>:<port>")
	flag.Var(&remoteFlag, "remote", "set remote endpoing address: <host>:<port>")
	flag.Var(&serverFlag, "server", "set server address: [<user>@]<host>[:<port>]")
	flag.StringVar(&keyFlag, "key", "", "set server authentication key file path")
	flag.BoolVar(&vFlag, "v", false, "increase log verbosity")
	flag.BoolVar(&helpFlag, "help", false, "list all options available")
	flag.BoolVar(&versionFlag, "version", false, "display the mole version")
}

func main() {

	code := handleCLIOptions()
	if code > -1 {
		os.Exit(code)
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

// handleCLIOptions parses the options given through CLI returning the program
// exit code.
// The function returns -1 if there is no need to exit the program.
func handleCLIOptions() int {
	if len(os.Args[1:]) == 0 {
		fmt.Printf("%s\n", usage())
		flag.PrintDefaults()
		return 1
	}

	flag.Parse()

	if versionFlag {
		fmt.Printf("mole %s\n", version)
		return 0
	}

	if helpFlag {
		fmt.Printf("%s\n", usage())
		flag.PrintDefaults()
		return 0
	}

	return -1
}

func usage() string {
	return `usage:
  mole [-v] [-local <host>:<port>] -remote <host>:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -help
  mole -version
	`
}
