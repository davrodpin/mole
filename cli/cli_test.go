package cli_test

import (
	"testing"

	"github.com/davrodpin/mole/cli"
)

func TestHandleArgs(t *testing.T) {

	tests := []struct {
		args     []string
		expected string
	}{
		{
			[]string{"./mole", "--version"},
			"version",
		},
		{
			[]string{"./mole", "-v"},
			"version",
		},
		{
			[]string{"./mole", "--help"},
			"help",
		},
		{
			[]string{"./mole", "-h"},
			"help",
		},
		{
			[]string{"./mole", "--remote", ":443", "--server", "example1"},
			"start",
		},
		{
			[]string{"./mole", "-r", ":443", "-s", "example1"},
			"start",
		},
		{
			[]string{"./mole", "--alias", "xyz", "--remote", ":443", "--server", "example1"},
			"new-alias",
		},
		{
			[]string{"./mole", "-a", "xyz", "-r", ":443", "-s", "example1"},
			"new-alias",
		},
		{
			[]string{"./mole", "--alias", "xyz", "--delete"},
			"rm-alias",
		},
		{
			[]string{"./mole", "-a", "xyz", "-d"},
			"rm-alias",
		},
		{
			[]string{"./mole", "--start", "example1-alias"},
			"start-from-alias",
		},
		{
			[]string{"./mole", "-S", "example1-alias"},
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
			[]string{"./mole", "--alias", "xyz", "--remote", ":443", "--server", "example1"},
			true,
		},
		{
			[]string{"./mole", "-a", "xyz", "-r", ":443", "-s", "example1"},
			true,
		},
		{
			[]string{"./mole", "--alias", "xyz", "--server", "example1"},
			false,
		},
		{
			[]string{"./mole", "-a", "xyz", "-s", "example1"},
			false,
		},
		{
			[]string{"./mole", "--alias", "xyz", "--remote", ":443"},
			false,
		},
		{
			[]string{"./mole", "-a", "xyz", "-r", ":443"},
			false,
		},
		{
			[]string{"./mole", "--alias", "xyz"},
			false,
		},
		{
			[]string{"./mole", "-a", "xyz"},
			false,
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
