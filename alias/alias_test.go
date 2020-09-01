package alias_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/davrodpin/mole/alias"
)

func TestParseTunnelFlags(t *testing.T) {

	tests := []struct {
		tunnelType        string
		verbose           bool
		insecure          bool
		detach            bool
		source            []string
		destination       []string
		server            string
		key               string
		keepAliveInterval string
		connectionRetries int
		waitAndRetry      string
		sshAgent          string
		timeout           string
	}{
		{
			"local",
			true,
			true,
			true,
			[]string{":1234", ":2345"},
			[]string{"192.168.1.1:80", "192.168.1.1:8080"},
			"server.com",
			"path/to/key",
			"10s",
			3,
			"5s",
			"path/to/ssh/agent",
			"1m0s",
		},
		{
			"local",
			true,
			false,
			true,
			[]string{"172.17.2.1:1234", "172.172.1:2345"},
			[]string{"192.168.1.1:80", "192.168.1.1:8080"},
			"user@192.168.10.1:22",
			"",
			"10s",
			3,
			"5s",
			"path/to/ssh/agent",
			"1m0s",
		},
	}

	for id, test := range tests {
		ai := &alias.Alias{
			TunnelType:        test.tunnelType,
			Verbose:           test.verbose,
			Insecure:          test.insecure,
			Detach:            test.detach,
			Source:            test.source,
			Destination:       test.destination,
			Server:            test.server,
			Key:               test.key,
			KeepAliveInterval: test.keepAliveInterval,
			ConnectionRetries: test.connectionRetries,
			WaitAndRetry:      test.waitAndRetry,
			SshAgent:          test.sshAgent,
			Timeout:           test.timeout,
		}

		tf, err := ai.ParseTunnelFlags()
		if err != nil {
			t.Errorf("%v\n", err)
		}

		if test.tunnelType != tf.TunnelType {
			t.Errorf("tunnelType doesn't match on test %d: expected: %s, value: %s", id, test.tunnelType, tf.TunnelType)
		}

		if test.verbose != tf.Verbose {
			t.Errorf("verbose doesn't match on test %d: expected: %t, value: %t", id, test.verbose, tf.Verbose)
		}

		if test.insecure != tf.Insecure {
			t.Errorf("insecure doesn't match on test %d: expected: %t, value: %t", id, test.insecure, tf.Insecure)
		}

		if test.detach != tf.Detach {
			t.Errorf("detach doesn't match on test %d: expected: %t, value: %t", id, test.detach, tf.Detach)
		}

		for i, tsrc := range test.source {
			src := tf.Source[i].String()
			if tsrc != src {
				t.Errorf("source %d doesn't match on test %d: expected: %s, value: %s", id, i, tsrc, src)
			}
		}

		for i, tdst := range test.destination {
			dst := tf.Destination[i].String()
			if tdst != dst {
				t.Errorf("destination %d doesn't match on test %d: expected: %s, value: %s", id, i, tdst, dst)
			}
		}

		if test.server != tf.Server.String() {
			t.Errorf("server doesn't match on test %d: expected: %s, value: %s", id, test.server, tf.Server.String())
		}

		if test.key != tf.Key {
			t.Errorf("key doesn't match on test %d: expected: %s, value: %s", id, test.key, tf.Key)
		}

		if test.keepAliveInterval != tf.KeepAliveInterval.String() {
			t.Errorf("keepAliveInterval doesn't match on test %d: expected: %s, value: %s", id, test.keepAliveInterval, tf.KeepAliveInterval.String())
		}

		if test.connectionRetries != tf.ConnectionRetries {
			t.Errorf("connectionRetries doesn't match on test %d: expected: %d, value: %d", id, test.connectionRetries, tf.ConnectionRetries)
		}

		if test.waitAndRetry != tf.WaitAndRetry.String() {
			t.Errorf("waitAndRetry doesn't match on test %d: expected: %s, value: %s", id, test.waitAndRetry, tf.WaitAndRetry.String())
		}

		if test.sshAgent != tf.SshAgent {
			t.Errorf("sshAgent doesn't match on test %d: expected: %s, value: %s", id, test.sshAgent, tf.SshAgent)
		}

		if test.timeout != tf.Timeout.String() {
			t.Errorf("timeout doesn't match on test %d: expected: %s, value: %s", id, test.timeout, tf.Timeout.String())
		}

	}

}

func TestAddThenGetThenDeleteAlias(t *testing.T) {
	dir, err := setAliasDirectory()
	if err != nil {
		t.Errorf("error during test setup: %v", err)
	}
	defer os.RemoveAll(dir)

	expectedAlias, err := addAlias()
	if err != nil {
		t.Errorf("error creating alias file %v", err)
	}

	expectedAliasFilePath := filepath.Join(dir, ".mole", fmt.Sprintf("%s.toml", expectedAlias.Name))

	if _, err := os.Stat(expectedAliasFilePath); os.IsNotExist(err) {
		t.Errorf("alias file could not be found after the attempt to create it")
	}

	al, err := alias.Get(expectedAlias.Name)
	if err != nil {
		t.Errorf("%v", err)
	}

	if !reflect.DeepEqual(expectedAlias, al) {
		t.Errorf("expected: %s, actual: %s", expectedAlias, al)
	}

	err = alias.Delete(expectedAlias.Name)
	if err != nil {
		t.Errorf("error while deleting %s alias file: %v", expectedAlias.Name, err)
	}

	if _, err := os.Stat(expectedAliasFilePath); !os.IsNotExist(err) {
		t.Errorf("alias file found after the attempt to delete it")
	}

}

func TestShow(t *testing.T) {
	dir, err := setAliasDirectory()
	if err != nil {
		t.Errorf("error during test setup: %v", err)
	}
	defer os.RemoveAll(dir)

	a, err := addAlias()
	if err != nil {
		t.Errorf("error creating alias file %v", err)
	}
	defer alias.Delete(a.Name)

	showOutput, err := alias.Show(a.Name)
	if err != nil {
		t.Errorf("error while showing all aliases")
	}

	expectedShowOutput, err := generateAliasShowOutput(a)
	if err != nil {
		t.Errorf("%v", err)
	}

	if expectedShowOutput != showOutput {
		t.Errorf("ShowAll output format has changed")
	}
}

func TestShowAll(t *testing.T) {
	dir, err := setAliasDirectory()
	if err != nil {
		t.Errorf("error during test setup: %v", err)
	}
	defer os.RemoveAll(dir)

	a, err := addAlias()
	if err != nil {
		t.Errorf("error creating alias file %v", err)
	}
	defer alias.Delete(a.Name)

	showOutput, err := alias.ShowAll()
	if err != nil {
		t.Errorf("error while showing all aliases")
	}

	expectedShowOutput, err := generateAliasShowOutput(a)
	if err != nil {
		t.Errorf("%v", err)
	}

	if expectedShowOutput != showOutput {
		t.Errorf("ShowAll output format has changed")
	}
}

func TestAliasMerge(t *testing.T) {
	keepAliveInterval, _ := time.ParseDuration("5s")

	tests := []struct {
		alias       alias.Alias
		tunnelFlags *alias.TunnelFlags
		expected    alias.Alias
	}{
		{
			alias.Alias{
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
			&alias.TunnelFlags{
				Verbose:           true,
				Insecure:          true,
				Detach:            true,
				Source:            alias.AddressInputList([]alias.AddressInput{alias.AddressInput{Host: "127.0.0.1", Port: "80"}}),
				Destination:       alias.AddressInputList([]alias.AddressInput{alias.AddressInput{Host: "172.17.0.100", Port: "8080"}}),
				Server:            alias.AddressInput{Host: "acme.com", Port: "22"},
				Key:               "path/to/key/2",
				KeepAliveInterval: keepAliveInterval,
				ConnectionRetries: 10,
				GivenFlags:        []string{"verbose", "insecure", "detach"},
			},
			alias.Alias{
				Verbose:  true,
				Insecure: true,
				Detach:   true,
			},
		},
		{
			alias.Alias{
				Verbose:           true,
				Insecure:          true,
				Detach:            true,
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
			&alias.TunnelFlags{
				Verbose:           false,
				Insecure:          false,
				Detach:            false,
				Source:            alias.AddressInputList([]alias.AddressInput{alias.AddressInput{Host: "127.0.0.1", Port: "80"}}),
				Destination:       alias.AddressInputList([]alias.AddressInput{alias.AddressInput{Host: "172.17.0.100", Port: "8080"}}),
				Server:            alias.AddressInput{Host: "acme.com", Port: "22"},
				Key:               "path/to/key/2",
				KeepAliveInterval: keepAliveInterval,
				ConnectionRetries: 10,
				GivenFlags:        []string{},
			},
			alias.Alias{
				Verbose:  true,
				Insecure: true,
				Detach:   true,
			},
		},
	}

	for id, test := range tests {
		test.alias.Merge(test.tunnelFlags)

		if test.expected.Verbose != test.alias.Verbose {
			t.Errorf("alias verbose doesn't match on test %d: expected: %t, value: %t", id, test.expected.Verbose, test.alias.Verbose)
		}

		if test.expected.Insecure != test.alias.Insecure {
			t.Errorf("alias insecure doesn't match on test %d: expected: %t, value: %t", id, test.expected.Insecure, test.alias.Insecure)
		}

		if test.expected.Detach != test.alias.Detach {
			t.Errorf("alias detach doesn't match on test %d: expected: %t, value: %t", id, test.expected.Detach, test.alias.Detach)
		}
	}

}

func TestParseAlias(t *testing.T) {
	kai, _ := time.ParseDuration("3s")
	war, _ := time.ParseDuration("5s")
	tim, _ := time.ParseDuration("1s")

	flags := alias.TunnelFlags{
		TunnelType:        "local",
		Verbose:           false,
		Insecure:          false,
		Detach:            false,
		Source:            alias.AddressInputList{alias.AddressInput{Host: "127.0.0.1", Port: "8080"}},
		Destination:       alias.AddressInputList{alias.AddressInput{Host: "172.17.0.100", Port: "80"}},
		Server:            alias.AddressInput{},
		Key:               "path/to/key/1",
		KeepAliveInterval: kai,
		ConnectionRetries: 3,
		WaitAndRetry:      war,
		SshAgent:          "path/to/sshagent",
		Timeout:           tim,
	}

	al := flags.ParseAlias("aliasName")

	if flags.TunnelType != al.TunnelType {
		t.Errorf("tunnelType does not match: expected: %s, value: %s", flags.TunnelType, al.TunnelType)
	}

	if flags.Verbose != al.Verbose {
		t.Errorf("verbose does not match: expected: %t, value: %t", flags.Verbose, al.Verbose)
	}

	if flags.Insecure != al.Insecure {
		t.Errorf("insecure does not match: expected: %t, value: %t", flags.Insecure, al.Insecure)
	}

	if flags.Detach != al.Detach {
		t.Errorf("detach does not match: expected: %t, value: %t", flags.Detach, al.Detach)
	}

	if !reflect.DeepEqual(flags.Source.List(), al.Source) {
		t.Errorf("source does not match: expected: %s, value: %s", flags.Source.List(), al.Source)
	}

	if !reflect.DeepEqual(flags.Destination.List(), al.Destination) {
		t.Errorf("destination does not match: expected: %s, value: %s", flags.Destination.List(), al.Destination)
	}

	if flags.Server.String() != al.Server {
		t.Errorf("server does not match: expected: %s, value: %s", flags.Server.String(), al.Server)
	}

	if flags.Key != al.Key {
		t.Errorf("key does not match: expected: %s, value: %s", flags.Key, al.Key)
	}

	if flags.KeepAliveInterval.String() != al.KeepAliveInterval {
		t.Errorf("keep alive interval does not match: expected: %s, value: %s", flags.KeepAliveInterval.String(), al.KeepAliveInterval)
	}

	if flags.ConnectionRetries != al.ConnectionRetries {
		t.Errorf("connection retries does not match: expected: %d, value: %d", flags.ConnectionRetries, al.ConnectionRetries)
	}

	if flags.WaitAndRetry.String() != al.WaitAndRetry {
		t.Errorf("wait and retry does not match: expected: %s, value: %s", flags.WaitAndRetry.String(), al.WaitAndRetry)
	}

	if flags.SshAgent != al.SshAgent {
		t.Errorf("ssh agent does not match: expected: %s, value: %s", flags.SshAgent, al.SshAgent)
	}

	if flags.Timeout.String() != al.Timeout {
		t.Errorf("timeout does not match: expected: %s, value: %s", flags.Timeout.String(), al.Timeout)
	}

}

func addAlias() (*alias.Alias, error) {
	a := &alias.Alias{
		Name:              "alias",
		TunnelType:        "local",
		Verbose:           true,
		Insecure:          true,
		Detach:            true,
		Source:            []string{":1234"},
		Destination:       []string{"192.168.1.1:80"},
		Server:            "server.com",
		Key:               "path/to/key",
		KeepAliveInterval: "5s",
		ConnectionRetries: 3,
		WaitAndRetry:      "10s",
		SshAgent:          "path/to/agent",
		Timeout:           "1m",
	}

	err := alias.Add(a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func generateAliasShowOutput(a *alias.Alias) (string, error) {
	var expectedShowOutput bytes.Buffer

	tmpl := template.Must(template.New("aliases").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(alias.ShowTemplate))
	if err := tmpl.Execute(&expectedShowOutput, a); err != nil {
		return "", fmt.Errorf("error generating expected")
	}

	return expectedShowOutput.String(), nil
}

func setAliasDirectory() (string, error) {

	dir, err := ioutil.TempDir("", "mole")
	if err != nil {
		return "", err
	}

	os.Setenv("HOME", dir)
	os.Setenv("USERPROFILE", dir)

	return dir, nil
}
