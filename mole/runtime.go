package mole

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/davrodpin/mole/fsutils"
	ps "github.com/mitchellh/go-ps"
)

type Formatter interface {
	Format(format string) (string, error)
}

// Runtime holds runtime data about an application instance.
type Runtime Configuration

// Format parses a Runtime object into a string representation based on the given
// format (i.e. toml).
func (rt Runtime) Format(format string) (string, error) {
	if format == "toml" {
		return rt.ToToml()
	} else {
		return "", fmt.Errorf("unknown %s format", format)
	}
}

func (rt Runtime) ToToml() (string, error) {
	var buf bytes.Buffer
	e := toml.NewEncoder(&buf)

	if err := e.Encode(rt); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type InstancesRuntime []Runtime

func (ir InstancesRuntime) Format(format string) (string, error) {
	if format == "toml" {
		return ir.ToToml()
	} else {
		return "", fmt.Errorf("unknown %s format", format)
	}
}

func (ir InstancesRuntime) ToToml() (string, error) {
	rt := make(map[string]map[string]Runtime)
	rt["instances"] = make(map[string]Runtime)

	for _, instance := range ir {
		rt["instances"][instance.Id] = instance
	}

	var buf bytes.Buffer
	e := toml.NewEncoder(&buf)

	if err := e.Encode(rt); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *Client) Runtime() (*Runtime, error) {
	runtime := Runtime(*c.Conf)

	if c.Tunnel != nil {
		source := &AddressInputList{}
		destination := &AddressInputList{}

		for _, channel := range c.Tunnel.Channels() {
			var err error

			err = source.Set(channel.Source)
			if err != nil {
				return nil, err
			}

			err = destination.Set(channel.Destination)
			if err != nil {
				return nil, err
			}

		}

		runtime.Source = *source
		runtime.Destination = *destination
	}

	return &runtime, nil
}

// Running checks if an instance of mole is running on the system.
func (c *Client) Running() (bool, error) {
	d, err := fsutils.InstanceDir(c.Conf.Id)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(d.PidFile); os.IsNotExist(err) {
		return false, nil
	}

	pd, err := ioutil.ReadFile(d.PidFile)
	if err != nil {
		return false, err
	}

	pid, err := strconv.Atoi(string(pd))
	if err != nil {
		return false, err
	}

	ps, err := ps.FindProcess(pid)
	if err != nil {
		return false, err
	}

	if ps == nil {
		return false, nil
	}

	return true, nil
}
