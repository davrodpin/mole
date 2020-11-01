package alias

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/davrodpin/mole/fsutils"
)

// Alias holds all attributes required to start a ssh port forwarding tunnel.
type Alias struct {
	Name              string   `toml:"name"`
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
	SshConfig         string   `toml:"config"`
}

func (a Alias) String() string {
	return fmt.Sprintf("[verbose: %t, insecure: %t, detach: %t, source: %s, destination: %s, server: %s, key: %s, keep-alive-interval: %s, connection-retries: %d, wait-and-retry: %s, ssh-agent: %s, timeout: %s, config: %s]",
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
		a.SshConfig,
	)
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
	mp, err := fsutils.Dir()

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
	a, err := Get(aliasName)
	if err != nil {
		return "", fmt.Errorf("could not show alias %s configuration: %v", aliasName, err)
	}

	var aliases bytes.Buffer
	e := toml.NewEncoder(&aliases)

	if err = e.Encode(a); err != nil {
		return "", err
	}

	return aliases.String(), nil
}

// ShowAll displays the configuration parameters for all persisted aliases.
func ShowAll() (string, error) {
	mp, err := fsutils.Dir()
	if err != nil {
		return "", err
	}

	aliases := aliases{}
	aliases.Aliases = make(map[string]*Alias)

	err = filepath.Walk(mp, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".toml" {
				var err error
				an := strings.TrimSuffix(filepath.Base(path), ".toml")
				al, err := Get(an)
				if err != nil {
					return err
				}

				aliases.Aliases[al.Name] = al
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer

	e := toml.NewEncoder(&buff)

	if err = e.Encode(aliases); err != nil {
		return "", err
	}

	return buff.String(), nil
}

// Get returns an alias previously created
func Get(aliasName string) (*Alias, error) {
	mp, err := fsutils.Dir()
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

func createDir() (string, error) {
	mp, err := fsutils.Dir()
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

//FIXME terrible struct name. Change it.
type aliases struct {
	Aliases map[string]*Alias `toml:"aliases"`
}
