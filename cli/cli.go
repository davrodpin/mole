package cli

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
)

var re *regexp.Regexp = regexp.MustCompile("(?P<user>.+@)?(?P<host>[0-9a-zA-Z\\.-]+)?(?P<port>:[0-9]+)?")

type App struct {
	args []string
	flag *flag.FlagSet

	Command     string
	Local       HostInput
	Remote      HostInput
	Server      HostInput
	Key         string
	Verbose     bool
	Help        bool
	Version     bool
	Alias       string
	Start       string
	AliasDelete bool
}

func New(args []string) *App {
	return &App{args: args}
}

func (c *App) Parse() error {
	f := flag.NewFlagSet(usage(), flag.ExitOnError)

	f.StringVar(&c.Alias, "alias", "", "Create a tunnel alias")
	f.BoolVar(&c.AliasDelete, "delete", false, "delete a tunnel alias (must be used with -alias)")
	f.StringVar(&c.Start, "start", "", "Start a tunnel using a given alias")
	f.Var(&c.Local, "local", "(optional) Set local endpoint address: [<host>]:<port>")
	f.Var(&c.Remote, "remote", "set remote endpoint address: [<host>]:<port>")
	f.Var(&c.Server, "server", "set server address: [<user>@]<host>[:<port>]")
	f.StringVar(&c.Key, "key", "", "(optional) Set server authentication key file path")
	f.BoolVar(&c.Verbose, "v", false, "(optional) Increase log verbosity")
	f.BoolVar(&c.Help, "help", false, "list all options available")
	f.BoolVar(&c.Version, "version", false, "display the mole version")

	f.Parse(c.args[1:])

	c.flag = f

	if len(c.args[1:]) == 0 {
		return fmt.Errorf("not enough arguments provided")
	}

	if c.Help {
		c.Command = "help"
	} else if c.Version {
		c.Command = "version"
	} else if c.Alias != "" && c.AliasDelete {
		c.Command = "rm-alias"
	} else if c.Alias != "" {
		c.Command = "new-alias"
	} else if c.Start != "" {
		c.Command = "start-from-alias"
		c.Alias = c.Start
	} else {
		c.Command = "start"
	}

	return nil
}

// PrintUsage prints, to the standard output, the informational text on how to
// use the tool.
func (c App) PrintUsage() {
	fmt.Printf("%s\n", usage())
	c.flag.PrintDefaults()
}

// String returns a string representation of an App.
func (c App) String() string {
	return fmt.Sprintf("[local=%s, remote=%s, server=%s, key=%s, verbose=%t, help=%t, version=%t]",
		c.Local, c.Remote, c.Server, c.Key, c.Verbose, c.Help, c.Version)
}

// HostInput holds information about a host
type HostInput struct {
	User string
	Host string
	Port string
}

// String returns a string representation of a HostInput
func (h HostInput) String() string {
	var s string
	if h.User == "" {
		s = h.Address()
	} else {
		s = fmt.Sprintf("%s@%s", h.User, h.Address())
	}

	return s
}

// Set parses a string representation of HostInput into its proper attributes.
func (h *HostInput) Set(value string) error {
	result := parseServerInput(value)
	h.User = strings.Trim(result["user"], "@")
	h.Host = result["host"]
	h.Port = strings.Trim(result["port"], ":")

	return nil
}

// Address returns a string representation of HostInput to be used to perform
// network connections.
func (h HostInput) Address() string {
	if h.Port == "" {
		return fmt.Sprintf("%s", h.Host)
	}

	return fmt.Sprintf("%s:%s", h.Host, h.Port)
}

func usage() string {
	return `usage:
  mole [-v] [-local [<host>]:<port>] -remote [<host>]:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -alias <alias_name> [-v] [-local [<host>]:<port>] -remote [<host>]:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -alias <alias_name> -delete
  mole -start <alias_name>
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
