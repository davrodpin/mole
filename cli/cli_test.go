package cli_test

import (
	"reflect"
	"testing"

	"github.com/davrodpin/mole/cli"
)

func TestAddressInput(t *testing.T) {
	tests := []struct {
		input    string
		expected cli.AddressInput
	}{
		{
			"test",
			cli.AddressInput{User: "", Host: "test", Port: ""},
		},
		{
			"user@test",
			cli.AddressInput{User: "user", Host: "test", Port: ""},
		},
		{
			"user@test:2222",
			cli.AddressInput{User: "user", Host: "test", Port: "2222"},
		},
		{
			"test-1",
			cli.AddressInput{User: "", Host: "test-1", Port: ""},
		},
		{
			"test-1-2-xy",
			cli.AddressInput{User: "", Host: "test-1-2-xy", Port: ""},
		},
		{
			"test.com",
			cli.AddressInput{User: "", Host: "test.com", Port: ""},
		},
		{
			"test_1",
			cli.AddressInput{User: "", Host: "test_1", Port: ""},
		},
		{
			"user@test_1",
			cli.AddressInput{User: "user", Host: "test_1", Port: ""},
		},
		{
			"user@test_1:2222",
			cli.AddressInput{User: "user", Host: "test_1", Port: "2222"},
		},
	}

	var h cli.AddressInput
	for _, test := range tests {
		h = cli.AddressInput{}
		h.Set(test.input)

		if !reflect.DeepEqual(test.expected, h) {
			t.Errorf("test failed. Expected: %v, value: %v", test.expected, h)
		}
	}
}

func TestCommand(t *testing.T) {

	tests := []struct {
		args     []string
		expected string
	}{
		{
			[]string{"./mole", "-version"},
			"version",
		},
		{
			[]string{"./mole", "-help"},
			"help",
		},
		{
			[]string{"./mole", "-remote", ":443", "-server", "example1"},
			"start",
		},
		{
			[]string{"./mole", "-alias", "xyz", "-remote", ":443", "-server", "example1"},
			"new-alias",
		},
		{
			[]string{"./mole", "-alias", "xyz", "-delete"},
			"rm-alias",
		},
		{
			[]string{"./mole", "-aliases"},
			"aliases",
		},
		{
			[]string{"./mole", "-start", "example1-alias"},
			"start-from-alias",
		},
	}

	var c *cli.App

	for _, test := range tests {
		c = cli.New(test.args)
		c.Parse()
		if test.expected != c.Command {
			t.Errorf("test failed. Expected: %s, value: %s", test.expected, c.Command)
		}
	}
}

func TestValidate(t *testing.T) {

	tests := []struct {
		args     []string
		expected bool
	}{
		{
			[]string{"./mole"},
			false,
		},
		{
			[]string{"./mole", "-alias", "xyz", "-remote", ":443", "-server", "example1"},
			true,
		},
		{
			[]string{"./mole", "-alias", "xyz", "-remote", ":443"},
			false,
		},
		{
			[]string{"./mole", "-alias", "xyz", "-server", "example1"},
			true,
		},
		{
			[]string{"./mole", "-alias", "xyz", "-remote", ":443"},
			false,
		},
		{
			[]string{"./mole", "-alias", "xyz"},
			false,
		},
		{
			[]string{"./mole", "-local", ":8080", "-remote", ":80", "-server", "example1"},
			true,
		},
		{
			[]string{"./mole", "-remote", ":3366", "-remote", ":443", "-server", "example1"},
			true,
		},
		{
			[]string{"./mole", "-local", ":1234", "-remote", ":3366", "-remote", ":443", "-server", "example1"},
			true,
		},
	}

	var c *cli.App

	for index, test := range tests {
		c = cli.New(test.args)
		c.Parse()

		err := c.Validate()
		value := err == nil

		if value != test.expected {
			t.Errorf("test case %v failed. Expected: %v, value: %v", index, test.expected, value)
		}
	}
}
