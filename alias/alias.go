package alias

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	InstancePidFile = "pid"
	InstanceLogFile = "mole.log"
	ShowTemplate    = `{{.Name}}
              verbose: {{.Verbose}}
             insecure: {{.Insecure}}
               detach: {{.Detach}}
               source: {{ StringsJoin .Source ", " }}
          destination: {{ StringsJoin .Destination ", " }}
               server: {{.Server}}
                  key: {{.Key}}
  keep alive interval: {{.KeepAliveInterval}}
   connection retries: {{.ConnectionRetries}}
       wait and retry: {{.WaitAndRetry}}
            ssh agent: {{.SshAgent}}
              timeout: {{.Timeout}}
`
)

// TunnelFlags is a struct that holds all flags required to establish a ssh
// port forwarding tunnel.
type TunnelFlags struct {
	TunnelType        string
	Verbose           bool
	Insecure          bool
	Detach            bool
	Source            AddressInputList
	Destination       AddressInputList
	Server            AddressInput
	Key               string
	KeepAliveInterval time.Duration
	ConnectionRetries int
	WaitAndRetry      time.Duration
	SshAgent          string
	Timeout           time.Duration
}

// ParseAlias translates a TunnelFlags object to an Alias object
func (tf TunnelFlags) ParseAlias(name string) *Alias {
	return &Alias{
		Name:              name,
		TunnelType:        tf.TunnelType,
		Verbose:           tf.Verbose,
		Insecure:          tf.Insecure,
		Detach:            tf.Detach,
		Source:            tf.Source.List(),
		Destination:       tf.Destination.List(),
		Server:            tf.Server.String(),
		Key:               tf.Key,
		KeepAliveInterval: tf.KeepAliveInterval.String(),
		ConnectionRetries: tf.ConnectionRetries,
		WaitAndRetry:      tf.WaitAndRetry.String(),
		SshAgent:          tf.SshAgent,
		Timeout:           tf.Timeout.String(),
	}
}

func (tf TunnelFlags) String() string {
	return fmt.Sprintf("[verbose: %t, insecure: %t, detach: %t, source: %s, destination: %s, server: %s, key: %s, keep-alive-interval: %s, connection-retries: %d, wait-and-retry: %s, ssh-agent: %s, timeout: %s]",
		tf.Verbose,
		tf.Insecure,
		tf.Detach,
		tf.Source,
		tf.Destination,
		tf.Server,
		tf.Key,
		tf.KeepAliveInterval,
		tf.ConnectionRetries,
		tf.WaitAndRetry,
		tf.SshAgent,
		tf.Timeout,
	)
}

// Alias holds all attributes required to start a ssh port forwarding tunnel.
type Alias struct {
	Name              string   `toml:"-"`
	TunnelType        string   `toml:"type"`
	Verbose           bool     `toml:"verbose"`
	Insecure          bool     `toml:"insecure"`
	Detach            bool     `toml:"detach"`
	Source            []string `toml:"source"`
	Destination       []string `toml:"destination"`
	Server            string   `toml:"server"`
	Key               string   `toml:"key"`
	KeepAliveInterval string   `toml:"keep-alive-interval"`
	ConnectionRetries int      `toml:"connection-retries"`
	WaitAndRetry      string   `toml:"wait-and-retry"`
	SshAgent          string   `toml:"ssh-agent"`
	Timeout           string   `toml:"timeout"`
}

// ParseTunnelFlags parses an Alias into a TunnelFlags
func (a Alias) ParseTunnelFlags() (*TunnelFlags, error) {
	var err error

	tf := &TunnelFlags{}

	tf.TunnelType = a.TunnelType
	tf.Verbose = a.Verbose
	tf.Insecure = a.Insecure
	tf.Detach = a.Detach

	srcl := AddressInputList{}
	for _, src := range a.Source {
		err = srcl.Set(src)
		if err != nil {
			return nil, err
		}
	}
	tf.Source = srcl

	dstl := AddressInputList{}
	for _, dst := range a.Destination {
		err = dstl.Set(dst)
		if err != nil {
			return nil, err
		}
	}
	tf.Destination = dstl

	srv := AddressInput{}
	err = srv.Set(a.Server)
	if err != nil {
		return nil, err
	}
	tf.Server = srv

	tf.Key = a.Key

	kai, err := time.ParseDuration(a.KeepAliveInterval)
	if err != nil {
		return nil, err
	}
	tf.KeepAliveInterval = kai

	tf.ConnectionRetries = a.ConnectionRetries

	war, err := time.ParseDuration(a.WaitAndRetry)
	if err != nil {
		return nil, err
	}
	tf.WaitAndRetry = war

	tf.SshAgent = a.SshAgent

	tim, err := time.ParseDuration(a.Timeout)
	if err != nil {
		return nil, err
	}
	tf.Timeout = tim

	return tf, nil
}

// Merge overwrites certain Alias attributes based on the given TunnelFlags.
func (a *Alias) Merge(tunnelFlags *TunnelFlags) {
	a.Verbose = tunnelFlags.Verbose
	a.Insecure = tunnelFlags.Insecure
	a.Detach = tunnelFlags.Detach
}

func (a Alias) String() string {
	return fmt.Sprintf("[verbose: %t, insecure: %t, detach: %t, source: %s, destination: %s, server: %s, key: %s, keep-alive-interval: %s, connection-retries: %d, wait-and-retry: %s, ssh-agent: %s, timeout: %s]",
		a.Verbose,
		a.Insecure,
		a.Detach,
		a.Source,
		a.Destination,
		a.Server,
		a.Key,
		a.KeepAliveInterval,
		a.ConnectionRetries,
		a.WaitAndRetry,
		a.SshAgent,
		a.Timeout,
	)
}

type InstanceConfiguration struct {
	Home    string
	LogFile string
	PidFile string
}

func NewInstanceConfiguration(aliasName string) (*InstanceConfiguration, error) {
	aliasDir, err := Dir()
	if err != nil {
		return nil, err
	}

	home := filepath.Join(aliasDir, aliasName)

	return &InstanceConfiguration{
		Home:    home,
		LogFile: filepath.Join(home, InstanceLogFile),
		PidFile: filepath.Join(home, InstancePidFile),
	}, nil
}

// Add persists an tunnel alias to the disk
func Add(alias *Alias) error {
	mp, err := createDir()
	if err != nil {
		return err
	}

	ap := filepath.Join(mp, fmt.Sprintf("%s.toml", alias.Name))

	f, err := os.Create(ap)
	if err != nil {
		return err
	}
	defer f.Close()

	e := toml.NewEncoder(f)

	if err = e.Encode(alias); err != nil {
		return err
	}

	return nil
}

// Delete destroys a alias configuration file.
func Delete(alias string) error {
	mp, err := Dir()

	if err != nil {
		return err
	}

	afp := filepath.Join(mp, fmt.Sprintf("%s.toml", alias))

	if _, err := os.Stat(afp); os.IsNotExist(err) {
		return fmt.Errorf("alias %s does not exist", alias)
	}

	err = os.Remove(afp)
	if err != nil {
		return err
	}

	return nil
}

// Show displays the configuration parameters for the given alias name.
func Show(aliasName string) (string, error) {
	mp, err := Dir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(mp, fmt.Sprintf("%s.toml", aliasName))

	return showAlias(path)
}

// ShowAll displays the configuration parameters for all persisted aliases.
func ShowAll() (string, error) {
	mp, err := Dir()
	if err != nil {
		return "", err
	}

	var aliases bytes.Buffer

	err = filepath.Walk(mp, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".toml" {
				al, err := showAlias(path)
				if err != nil {
					return err
				}

				aliases.WriteString(al)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return aliases.String(), nil
}

// Get returns an alias previously created
func Get(aliasName string) (*Alias, error) {
	mp, err := Dir()
	if err != nil {
		return nil, err
	}

	p := filepath.Join(mp, fmt.Sprintf("%s.toml", aliasName))

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil, fmt.Errorf("alias %s does not exist", aliasName)
	}

	a := &Alias{}
	if _, err := toml.DecodeFile(p, a); err != nil {
		return nil, err
	}
	a.Name = aliasName

	return a, nil
}

// Dir returns directory path where all alias files are stored.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	mp := filepath.Join(home, ".mole")

	return mp, nil
}

func showAlias(filePath string) (string, error) {
	an := strings.TrimSuffix(filepath.Base(filePath), ".toml")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("alias %s does not exist", an)
	}

	var aliases bytes.Buffer

	a, err := Get(an)
	if err != nil {
		return "", fmt.Errorf("could not show alias configuration %s: %v", filePath, err)
	}

	t := template.Must(template.New("aliases").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(ShowTemplate))
	if err := t.Execute(&aliases, a); err != nil {
		return "", err
	}

	return aliases.String(), nil
}

func createDir() (string, error) {
	mp, err := Dir()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(mp); !os.IsNotExist(err) {
		return mp, nil
	}

	err = os.MkdirAll(mp, os.ModePerm)
	if err != nil {
		return "", err
	}

	return mp, nil
}
