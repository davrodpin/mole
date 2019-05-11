package storage_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/davrodpin/mole/storage"
)

func TestSaveTunnel(t *testing.T) {
	alias := "example-save-443"
	expected := &storage.Tunnel{
		Local:   "",
		Remote:  ":443",
		Server:  "example",
		Verbose: true,
	}

	_, err := storage.Save(alias, expected)
	if err != nil {
		t.Errorf("Test failed while saving tunnel configuration: %v", err)
	}

	value, err := storage.FindByName(alias)
	if err != nil {
		t.Errorf("Test failed while retrieving tunnel configuration: %v", err)
	}

	if !reflect.DeepEqual(expected, value) {
		t.Errorf("Test failed.\n\texpected: %s\n\tvalue   : %s", expected, value)
	}
}

func TestRemoveTunnel(t *testing.T) {
	alias := "example-rm-443"
	expected := &storage.Tunnel{
		Local:   "",
		Remote:  ":443",
		Server:  "example",
		Verbose: true,
	}

	storage.Save(alias, expected)
	value, err := storage.Remove(alias)
	if err != nil {
		t.Errorf("Test failed while removing tunnel configuration: %v", err)
	}

	if !reflect.DeepEqual(expected, value) {
		t.Errorf("Test failed.\n\texpected: %s\n\tvalue   : %s", expected, value)
	}

	value, _ = storage.FindByName(alias)

	if value != nil {
		t.Errorf("Test failed. Alias %s is not suppose to exist after deletion.", alias)
	}

}

func TestFindAll(t *testing.T) {
	alias1 := "example-save-443"
	expected1 := &storage.Tunnel{
		Local:   "",
		Remote:  ":443",
		Server:  "example",
		Verbose: true,
	}

	alias2 := "example-save-80"
	expected2 := &storage.Tunnel{
		Local:   "",
		Remote:  ":80",
		Server:  "example",
		Verbose: true,
	}

	storage.Save(alias1, expected1)
	storage.Save(alias2, expected2)

	expectedTunnelList := make(map[string]*storage.Tunnel)
	expectedTunnelList[alias1] = expected1
	expectedTunnelList[alias2] = expected2

	tunnels, err := storage.FindAll()
	if err != nil {
		t.Errorf("Test failed while retrieving all tunnels: %v", err)
	}

	if !reflect.DeepEqual(expectedTunnelList, tunnels) {
		t.Errorf("Test failed.\n\texpected: %v\n\tvalue   : %v", expectedTunnelList, tunnels)
	}
}

func TestMain(m *testing.M) {
	dir, err := ioutil.TempDir("", "mole-testing")
	if err != nil {
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	os.Setenv("USERPROFILE", dir)

	code := m.Run()

	os.Exit(code)
}
