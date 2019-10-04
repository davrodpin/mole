package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

var re = regexp.MustCompile(`(?P<user>.+@)?(?P<host>[[:alpha:][:digit:]\_\-\.]+)?(?P<port>:[0-9]+)?`)

// App contains all supported CLI arguments given by the user.
type App struct {
	args []string
	flag *flag.FlagSet

	Command           string
	Local             AddressInputList
	Remote            AddressInputList
	Server            AddressInput
	Key               string
	Verbose           bool
	Help              bool
	Version           bool
	Alias             string
	Start             string
	AliasDelete       bool
	Detach            bool
	Stop              string
	AliasList         bool
	Insecure          bool
	KeepAliveInterval time.Duration
	Timeout           time.Duration
	ConnectionRetries int
	WaitAndRetry      time.Duration
}

// New creates a new instance of App.
func New(args []string) *App {
	return &App{args: args}
}

// Parse grabs arguments and flags from CLI.
func (c *App) Parse() error {
	f := flag.NewFlagSet("", flag.ExitOnError)
	f.Usage = c.PrintUsage
	c.flag = f

	f.StringVarP(&c.Alias, "alias", "a", "", "create a tunnel alias")
	f.BoolVarP(&c.AliasDelete, "delete", "d", false, "delete a tunnel alias (must be used with -alias)")
	f.BoolVarP(&c.AliasList, "aliases", "I", false, "list all aliases")
	f.StringVarP(&c.Start, "start", "S", "", "start a tunnel using a given alias")
	f.VarP(&c.Local, "local", "l", "(optional) set local endpoint address: [<host>]:<port>. Multiple -local args can be provided")
	f.VarP(&c.Remote, "remote", "r", "(optional) set remote endpoint address: [<host>]:<port>. Multiple -remote args can be provided")
	f.VarP(&c.Server, "server", "s", "set server address: [<user>@]<host>[:<port>]")
	f.StringVarP(&c.Key, "key", "k", "", "(optional) Set server authentication key file path")
	f.BoolVarP(&c.Verbose, "verbose", "v", false, "(optional) Increase log verbosity")
	f.BoolVarP(&c.Help, "help", "h", false, "list all options available")
	f.BoolVar(&c.Version, "version", false, "display the mole version")
	f.BoolVarP(&c.Detach, "detach", "x", false, "(optional) run process in background")
	f.StringVarP(&c.Stop, "stop", "o", "", "stop background process")
	f.BoolVarP(&c.Insecure, "insecure", "i", false, "(optional) skip host key validation when connecting to ssh server")
	f.DurationVarP(&c.KeepAliveInterval, "keep-alive-interval", "K", 10*time.Second, "(optional) time interval for keep alive packets to be sent")
	f.DurationVarP(&c.Timeout, "timeout", "t", 3*time.Second, "(optional) ssh server connection timeout")
	f.IntVarP(&c.ConnectionRetries, "connection-retries", "R", 3, "(optional) maximum number of connection retries to the ssh server. Provide 0 if mole should never give up or negative number to disable retries")
	f.DurationVarP(&c.WaitAndRetry, "retry-wait", "w", 3*time.Second, "(optional) time to wait before trying to reconnect to ssh server")

	f.Parse(c.args[1:])

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
	} else if c.Stop != "" {
		c.Command = "stop"
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
	case "new-alias":
		if c.Server.String() == "" {
			return fmt.Errorf("required flag is missing: -server")
		}
	case "start":
		if c.Server.String() == "" {
			return fmt.Errorf("required flag is missing: -server")
		}
	}

	return nil
}

// PrintUsage prints, to the standard output, the informational text on how to
// use the tool.
func (c *App) PrintUsage() {
	fmt.Fprintf(os.Stderr, "%s\n\n", `usage:
	mole [--verbose|-v] [--insecure|-i] [--detach|-x] (--local|-l [<host>]:<port>)... (--remote|-r [<host>]:<port>)... --server|-s [<user>@]<host>[:<port>] [--key|-k <key_path>] [--keep-alive-interval|-K <time_interval>] [--connection-retries|-R <retries>] [--retry-wait|-w <time>]
	mole --alias <alias_name> [--verbose] (--local [<host>]:<port>)... (--remote [<host>]:<port>)... --server [<user>@]<host>[:<port>] [--key <key_path>] [--keep-alive-interval <time_interval>] [--connection-retries <retries>] [--retry-wait <time>]
	mole --alias <alias_name> --delete
	mole --start <alias_name>
	mole --help
	mole --version`)
	c.flag.PrintDefaults()
}

// String returns a string representation of an App.
func (c App) String() string {
	return fmt.Sprintf("[local=%s, remote=%s, server=%s, key=%s, verbose=%t, help=%t, version=%t, detach=%t, insecure=%t, keep-alive-interval=%v, timeout=%v, connection-retries=%d, retry-wait=%s]",
		c.Local, c.Remote, c.Server, c.Key, c.Verbose, c.Help, c.Version, c.Detach, c.Insecure, c.KeepAliveInterval, c.Timeout, c.ConnectionRetries, c.WaitAndRetry)
}

// AddressInput holds information about a host
type AddressInput struct {
	User string
	Host string
	Port string
}

// String returns a string representation of a AddressInput
func (h AddressInput) String() string {
	var s string
	if h.User == "" {
		s = h.Address()
	} else {
		s = fmt.Sprintf("%s@%s", h.User, h.Address())
	}

	return s
}

// Set parses a string representation of AddressInput into its proper attributes.
func (h *AddressInput) Set(value string) error {
	result := parseServerInput(value)
	h.User = strings.Trim(result["user"], "@")
	h.Host = result["host"]
	h.Port = strings.Trim(result["port"], ":")

	return nil
}

// Type return a string representation of AddressInput.
func (h *AddressInput) Type() string {
	return "AddressInput"
}

// Address returns a string representation of AddressInput to be used to perform
// network connections.
func (h AddressInput) Address() string {
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

type AddressInputList []AddressInput

func (il AddressInputList) String() string {
	ils := []string{}

	for _, i := range il {
		ils = append(ils, i.String())
	}

	return strings.Join(ils, ",")
}

func (il *AddressInputList) Set(value string) error {
	i := AddressInput{}

	err := i.Set(value)
	if err != nil {
		return err
	}

	*il = append(*il, i)

	return nil
}

// Type return a string representation of AddressInputList.
func (il *AddressInputList) Type() string {
	return "AddressInputList"
}

func (il AddressInputList) List() []string {
	sl := []string{}

	for _, i := range il {
		sl = append(sl, i.String())
	}

	return sl

}
