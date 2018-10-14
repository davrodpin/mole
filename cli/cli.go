package cli

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
)

var re *regexp.Regexp = regexp.MustCompile("(?P<user>.+@)?(?P<host>[0-9a-zA-Z\\.-]+)?(?P<port>:[0-9]+)?")

type App struct {
	Command string
	args    []string
	flag    *flag.FlagSet

	Local   hostFlag
	Remote  hostFlag
	Server  hostFlag
	Key     string
	Verbose bool
	help    bool
	version bool
	Alias   string
}

func New(args []string) *App {
	return &App{args: args}
}

func (c *App) Parse() error {
	f := flag.NewFlagSet(usage(), flag.ExitOnError)

	f.StringVar(&c.Alias, "alias", "", "Create a tunnel alias")
	f.Var(&c.Local, "local", "(optional) Set local endpoint address: [<host>]:<port>")
	f.Var(&c.Remote, "remote", "set remote endpoing address: [<host>]:<port>")
	f.Var(&c.Server, "server", "set server address: [<user>@]<host>[:<port>]")
	f.StringVar(&c.Key, "key", "", "(optional) Set server authentication key file path")
	f.BoolVar(&c.Verbose, "v", false, "(optional) Increase log verbosity")
	f.BoolVar(&c.help, "help", false, "list all options available")
	f.BoolVar(&c.version, "version", false, "display the mole version")

	f.Parse(c.args[1:])

	c.flag = f

	if len(c.args[1:]) == 0 {
		return fmt.Errorf("not enough arguments provided")
	}

	if c.help {
		c.Command = "help"
	} else if c.version {
		c.Command = "version"
	} else if c.Alias != "" {
		c.Command = "new-alias"
	} else {
		c.Command = "new"
	}

	return nil
}

func (c App) PrintUsage() {
	fmt.Printf("%s\n", usage())
	c.flag.PrintDefaults()
}

func (c App) String() string {
	return fmt.Sprintf("[local=%s, remote=%s, server=%s, key=%s, verbose=%t, help=%t, version=%t]",
		c.Local, c.Remote, c.Server, c.Key, c.Verbose, c.help, c.version)
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
	result := parseServerInput(value)
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
  mole -alias <alias_name> [-v] [-local [<host>]:<port>] -remote [<host>]:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -start <alias_name>
	mole -alias <alias_name> -delete
  mole -help
  mole -version
	`
}

func parseServerInput(input string) map[string]string {
	match := re.FindStringSubmatch(input)
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i == 0 {
			continue
		}

		result[name] = match[i]
	}

	return result
}
