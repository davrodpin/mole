package main

import (
	"testing"
)

func TestHandleArgs(t *testing.T) {

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
			"new",
		},
	}

	var c cmd
	for _, test := range tests {
		c = cmd{}
		c.Parse(test.args)
		if test.expected != c.command {
			t.Errorf("test failed. Expected: %s, value: %s", test.expected, c.command)
		}
	}
}
