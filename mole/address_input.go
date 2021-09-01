package mole

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	AddressFormat = "%s:%s"
)

var re = regexp.MustCompile(`(?P<user>.+@)?(?P<host>[[:alpha:][:digit:]\_\-\.]+)?(?P<port>:[0-9]+)?`)

// AddressInput holds information about a host
type AddressInput struct {
	User string `mapstructure:"user" toml:"user"`
	Host string `mapstructure:"host" toml:"host"`
	Port string `mapstructure:"port" toml:"port"`
}

// String returns a string representation of a AddressInput
func (ai AddressInput) String() string {
	var s string
	if ai.User == "" {
		s = ai.Address()
	} else {
		s = fmt.Sprintf("%s@%s", ai.User, ai.Address())
	}

	return s
}

// Set parses a string representation of AddressInput into its proper attributes.
func (ai *AddressInput) Set(value string) error {
	result := parseServerInput(value)
	ai.User = strings.Trim(result["user"], "@")
	ai.Host = result["host"]
	ai.Port = strings.Trim(result["port"], ":")

	return nil
}

// Type return a string representation of AddressInput.
func (ai *AddressInput) Type() string {
	return "[<user>@][<host>]:<port>"
}

// Address returns a string representation of AddressInput to be used to perform
// network connections.
func (ai AddressInput) Address() string {
	if ai.Port == "" {
		return ai.Host
	}

	return fmt.Sprintf(AddressFormat, ai.Host, ai.Port)
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

// AddressInputList represents a collection of AddressInput objects
type AddressInputList []AddressInput

// String return the string representation of AddressInputList
func (il AddressInputList) String() string {
	ils := []string{}

	for _, i := range il {
		ils = append(ils, i.String())
	}

	return strings.Join(ils, ",")
}

// Set adds a string representation of a AddressInput to the AddressInputList
// object
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
	return "([<host>]:<port>)..."
}

// List returns an array of the string representation of each AddressInput kept
// on the AddressInputList
func (il AddressInputList) List() []string {
	sl := []string{}

	for _, i := range il {
		sl = append(sl, i.String())
	}

	return sl

}
