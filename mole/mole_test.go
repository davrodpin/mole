package mole_test

import (
	"testing"

	"github.com/davrodpin/mole/alias"
	"github.com/davrodpin/mole/mole"
)

func TestAliasMerge(t *testing.T) {
	tests := []struct {
		alias      *alias.Alias
		givenFlags []string
		conf       mole.Configuration
		expected   mole.Configuration
	}{
		{
			&alias.Alias{
				Verbose:           false,
				Insecure:          false,
				Detach:            false,
				Source:            []string{"127.0.0.1:80"},
				Destination:       []string{"172.17.0.100:8080"},
				Server:            "user@example.com:22",
				Key:               "path/to/key/1",
				KeepAliveInterval: "3s",
				ConnectionRetries: 3,
				WaitAndRetry:      "10s",
				SshAgent:          "path/to/sshagent",
				Timeout:           "3s",
			},
			[]string{},
			mole.Configuration{
				Verbose:  true,
				Insecure: true,
				Detach:   true,
			},
			mole.Configuration{
				Verbose:  false,
				Insecure: false,
				Detach:   false,
			},
		},
		{
			&alias.Alias{
				Verbose:           false,
				Insecure:          false,
				Detach:            false,
				Source:            []string{"127.0.0.1:80"},
				Destination:       []string{"172.17.0.100:8080"},
				Server:            "user@example.com:22",
				Key:               "path/to/key/1",
				KeepAliveInterval: "3s",
				ConnectionRetries: 3,
				WaitAndRetry:      "10s",
				SshAgent:          "path/to/sshagent",
				Timeout:           "3s",
			},
			[]string{"verbose", "insecure", "detach"},
			mole.Configuration{
				Verbose:  true,
				Insecure: true,
				Detach:   true,
			},
			mole.Configuration{
				Verbose:  true,
				Insecure: true,
				Detach:   true,
			},
		},
		{
			&alias.Alias{
				Verbose:           false,
				Insecure:          false,
				Detach:            false,
				Source:            []string{"127.0.0.1:80"},
				Destination:       []string{"172.17.0.100:8080"},
				Server:            "user@example.com:22",
				Key:               "path/to/key/1",
				KeepAliveInterval: "3s",
				ConnectionRetries: 3,
				WaitAndRetry:      "10s",
				SshAgent:          "path/to/sshagent",
				Timeout:           "3s",
			},
			[]string{},
			mole.Configuration{
				SshConfig: "path/to/config",
			},
			mole.Configuration{
				SshConfig: "path/to/config",
			},
		},
	}

	for id, test := range tests {
		conf := test.conf
		conf.Merge(test.alias, test.givenFlags)

		if test.expected.Verbose != conf.Verbose {
			t.Errorf("verbose doesn't match on test %d: expected: %t, value: %t", id, test.expected.Verbose, conf.Verbose)
		}

		if test.expected.Insecure != conf.Insecure {
			t.Errorf("insecure doesn't match on test %d: expected: %t, value: %t", id, test.expected.Insecure, conf.Insecure)
		}

		if test.expected.Detach != conf.Detach {
			t.Errorf("detach doesn't match on test %d: expected: %t, value: %t", id, test.expected.Detach, conf.Detach)
		}
		if test.expected.SshConfig != conf.SshConfig {
			t.Errorf("sshConfig doesn't match on test %d: expected: %s, value: %s", id, test.expected.SshConfig, conf.SshConfig)
		}
	}

}
