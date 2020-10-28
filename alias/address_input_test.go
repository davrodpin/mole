package alias_test

import (
	"fmt"
	"testing"

	"github.com/davrodpin/mole/alias"
)

func TestAddressInputSet(t *testing.T) {
	tests := []struct {
		user string
		host string
		port string
	}{
		{
			"mole",
			"mole-server",
			"22",
		},
	}

	var ai alias.AddressInput
	for id, test := range tests {
		ai.Set(fmt.Sprintf("%s@%s:%s", test.user, test.host, test.port))

		if test.user != ai.User {
			t.Errorf("user does not match on test %d: expected: %s, value: %s", id, test.user, ai.User)
		}

		if test.host != ai.Host {
			t.Errorf("host does not match on test %d: expected: %s, value: %s", id, test.host, ai.Host)
		}

		if test.port != ai.Port {
			t.Errorf("port does not match on test %d: expected: %s, value: %s", id, test.port, ai.Port)
		}
	}

}

func TestAddressInputListSet(t *testing.T) {

	tests := []struct {
		ai string
	}{
		{
			"mole@mole-server:22",
		},
	}

	for id, test := range tests {
		aiList := alias.AddressInputList{}
		aiList.Set(test.ai)

		if test.ai != aiList.String() {
			t.Errorf("test %d: expected: %s, value: %s", id, test.ai, aiList.String())
		}
	}
}

func TestAddressInputAddress(t *testing.T) {

	tests := []struct {
		host     string
		port     string
		expected string
	}{
		{
			"mole-server",
			"22",
			"mole-server:22",
		},
		{
			"mole-server",
			"",
			"mole-server",
		},
	}

	for id, test := range tests {
		ai := alias.AddressInput{Host: test.host, Port: test.port}
		addr := ai.Address()

		if test.expected != addr {
			t.Errorf("address does not match on test %d: expected: %s, value: %s", id, test.expected, addr)
		}
	}
}
