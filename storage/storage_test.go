package storage_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/davrodpin/mole/storage"
)

func TestSaveTunnel(t *testing.T) {
	alias := "hpe-halon-443"
	expected := &storage.Tunnel{
		Local:   "",
		Remote:  ":443",
		Server:  "hpe-halon",
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

func TestMain(m *testing.M) {
	dir, err := ioutil.TempDir("", "mole-testing")
	if err != nil {
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)

	code := m.Run()

	os.Exit(code)
}
