package storage_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/davrodpin/mole/storage"
)

func TestSaveAlias(t *testing.T) {
	alias := "example-save-443"
	expected := &storage.Alias{
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

func TestRemoveAlias(t *testing.T) {
	alias := "example-rm-443"
	expected := &storage.Alias{
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
	expected1 := &storage.Alias{
		Local:   "",
		Remote:  ":443",
		Server:  "example",
		Verbose: true,
	}

	alias2 := "example-save-80"
	expected2 := &storage.Alias{
		Local:   "",
		Remote:  ":80",
		Server:  "example",
		Verbose: true,
	}

	storage.Save(alias1, expected1)
	storage.Save(alias2, expected2)

	expectedAliasList := make(map[string]*storage.Alias)
	expectedAliasList[alias1] = expected1
	expectedAliasList[alias2] = expected2

	tunnels, err := storage.FindAll()
	if err != nil {
		t.Errorf("Test failed while retrieving all tunnels: %v", err)
	}

	if !reflect.DeepEqual(expectedAliasList, tunnels) {
		t.Errorf("Test failed.\n\texpected: %v\n\tvalue   : %v", expectedAliasList, tunnels)
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

func TestReadLocalAndRemote(t *testing.T) {

	tests := []struct {
		alias          *storage.Alias
		expectedLocal  []string
		expectedRemote []string
	}{
		{
			alias:          &storage.Alias{Local: ":3306", Remote: ":3306"},
			expectedLocal:  []string{":3306"},
			expectedRemote: []string{":3306"},
		},
		{
			alias:          &storage.Alias{Local: []interface{}{":3306", ":8080"}, Remote: []interface{}{":3306", ":8080"}},
			expectedLocal:  []string{":3306", ":8080"},
			expectedRemote: []string{":3306", ":8080"},
		},
		{
			alias:          &storage.Alias{Local: ":3306", Remote: []interface{}{":3306", ":8080"}},
			expectedLocal:  []string{":3306"},
			expectedRemote: []string{":3306", ":8080"},
		},
		{
			alias:          &storage.Alias{Local: []interface{}{":3306", ":8080"}, Remote: []interface{}{":3306"}},
			expectedLocal:  []string{":3306", ":8080"},
			expectedRemote: []string{":3306"},
		},
	}

	for testId, test := range tests {
		local, err := test.alias.ReadLocal()
		if err != nil {
			t.Errorf("unexpected error while reading local address from alias for %d: %v", testId, err)
		}

		remote, err := test.alias.ReadRemote()
		if err != nil {
			t.Errorf("unexpected error while reading remote address from alias for %d: %v", testId, err)
		}

		if !reflect.DeepEqual(test.expectedLocal, local) {
			t.Errorf("unexpected local address for %d: expected: %v, value: %v", testId, test.expectedLocal, local)
		}

		if !reflect.DeepEqual(test.expectedRemote, remote) {
			t.Errorf("unexpected remote address for %d: expected: %v, value: %v", testId, test.expectedRemote, remote)
		}

	}
}
