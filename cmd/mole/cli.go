package main

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
)

type cmd struct {
	command string
	local   hostFlag
	remote  hostFlag
	server  hostFlag
	key     string
	verbose bool
	help    bool
	version bool
	flag    *flag.FlagSet
}

func (c *cmd) Parse(args []string) error {
	f := flag.NewFlagSet(usage(), flag.ExitOnError)

	f.Var(&c.local, "local", "(optional) Set local endpoint address: [<host>]:<port>")
	f.Var(&c.remote, "remote", "set remote endpoing address: [<host>]:<port>")
	f.Var(&c.server, "server", "set server address: [<user>@]<host>[:<port>]")
	f.StringVar(&c.key, "key", "", "(optional) Set server authentication key file path")
	f.BoolVar(&c.verbose, "v", false, "(optional) Increase log verbosity")
	f.BoolVar(&c.help, "help", false, "list all options available")
	f.BoolVar(&c.version, "version", false, "display the mole version")

	f.Parse(args[1:])

	c.flag = f

	if len(args[1:]) == 0 {
		return fmt.Errorf("not enough arguments provided")
	}

	if c.help {
		c.command = "help"
	} else if c.version {
		c.command = "version"
	} else {
		c.command = "new"
	}

	return nil

}

func (c cmd) PrintUsage() {
	fmt.Printf("%s\n", usage())
	c.flag.PrintDefaults()
}

func (c cmd) String() string {
	return fmt.Sprintf("[local=%s, remote=%s, server=%s, key=%s, verbose=%t, help=%t, version=%t]",
		c.local, c.remote, c.server, c.key, c.verbose, c.help, c.version)
}

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
	re := regexp.MustCompile("(?P<user>.+@)?(?P<host>[0-9a-zA-Z\\.-]+)?(?P<port>:[0-9]+)?")

	match := re.FindStringSubmatch(value)
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i == 0 {
			continue
		}

		result[name] = match[i]
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

func usage() string {
	return `usage:
  mole [-v] [-local [<host>]:<port>] -remote [<host>]:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -help
  mole -version
	`
}
