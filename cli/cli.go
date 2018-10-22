package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

var re = regexp.MustCompile("(?P<user>.+@)?(?P<host>[0-9a-zA-Z\\.-]+)?(?P<port>:[0-9]+)?")

// App contains main settings of application.
type App struct {
	args []string
	flag *pflag.FlagSet

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
	AliasList   bool
}

// New creates a new instance of App.
func New(args []string) *App {
	return &App{args: args}
}

// Parse grabs arguments and flags from CLI.
func (c *App) Parse() error {
	pf := pflag.NewFlagSet("", pflag.ExitOnError)
	pf.Usage = c.PrintUsage
	c.flag = pf

	pf.StringVarP(&c.Alias, "alias", "a", "", "Create a tunnel alias")
	pf.BoolVarP(&c.AliasDelete, "delete", "d", false, "Create a tunnel alias")
	pf.BoolVarP(&c.AliasList, "aliases", "A", false, "list all aliases")
	pf.StringVarP(&c.Start, "start", "S", "", "Start a tunnel using a given alias")
	pf.VarP(&c.Local, "local", "l", "(optional) Set local endpoint address: [<host>]:<port>")
	pf.VarP(&c.Remote, "remote", "r", "set remote endpoint address: [<host>]:<port>")
	pf.VarP(&c.Server, "server", "s", "set remote endpoint address: [<host>]:<port>")
	pf.StringVarP(&c.Key, "key", "k", "", "(optional) Set server authentication key file path")
	pf.BoolVarP(&c.Verbose, "verbose", "V", false, "(optional) Increase log verbosity")
	pf.BoolVarP(&c.Help, "help", "h", false, "list all options available")
	pf.BoolVarP(&c.Version, "version", "v", false, "display the mole version")

	pf.Parse(c.args[1:])

	if len(c.args[1:]) == 0 {
		return fmt.Errorf("not enough arguments provided")
	}

	if c.Help {
		c.Command = "help"
	} else if c.Version {
		c.Command = "version"
	} else if c.AliasList {
		c.Command = "aliases"
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

	err := c.Validate()
	if err != nil {
		return err
	}

	return nil
}

// Validate checks parsed params.
func (c App) Validate() error {
	if len(c.args[1:]) == 0 {
		return fmt.Errorf("not enough arguments provided")
	}

	switch c.Command {
	case "start", "new-alias":
		if c.Remote.String() == "" {
			return fmt.Errorf("required flag is missing: -remote")
		} else if c.Server.String() == "" {
			return fmt.Errorf("required flag is missing: -server")
		}
	}
	return nil
}

// PrintUsage prints, to the standard output, the informational text on how to
// use the tool.
func (c *App) PrintUsage() {
	fmt.Fprintf(os.Stderr, "%s\n\n", `usage:
	mole [-v] [-local [<host>]:<port>] -remote [<host>]:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
	mole -alias <alias_name> [-v] [-local [<host>]:<port>] -remote [<host>]:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
	mole -alias <alias_name> -delete
	mole -start <alias_name>
	mole -help
	mole -version`)
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

func (h *HostInput) Type() string {
	return "string"
}

// Address returns a string representation of HostInput to be used to perform
// network connections.
func (h HostInput) Address() string {
	if h.Port == "" {
		return fmt.Sprintf("%s", h.Host)
	}

	return fmt.Sprintf("%s:%s", h.Host, h.Port)
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
