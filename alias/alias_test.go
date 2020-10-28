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

	"github.com/davrodpin/mole/alias"
)

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
		SshConfig:         "/home/user/.ssh/config",
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
