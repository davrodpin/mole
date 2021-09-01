package mole

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
)

type Formatter interface {
	Format(format string) (string, error)
}

// Runtime holds runtime data about an application instances.
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
