package fsutils_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/davrodpin/mole/fsutils"
)

var (
	// home is a temporary directory that acts as the user home directory.
	// It is used on most tests to validates files created by Mole and their
	// contents.
	home string
)

func TestCreateInstanceDir(t *testing.T) {
	tests := []struct {
		id        string
		preCreate bool
	}{
		{id: "d34dbeef", preCreate: true},
		{id: "f00d", preCreate: false},
	}

	for idx, test := range tests {
		instanceDir := filepath.Join(home, ".mole", test.id)
		pidFileLocation := filepath.Join(home, ".mole", test.id, fsutils.InstancePidFile)

		if test.preCreate {
			os.MkdirAll(instanceDir, 0755)
			ioutil.WriteFile(pidFileLocation, []byte("1234"), 0644)
		}

		dirInfo, err := fsutils.CreateInstanceDir(test.id)

		if err != nil {
			t.Errorf("item %d: expected: nil; got %v", idx, err)
		}

		if test.id != dirInfo.Id {
			t.Errorf("item %d: expected: %s; got: %s", idx, test.id, dirInfo.Id)
		}

		if instanceDir != dirInfo.Dir {
			t.Errorf("item %d: expected: %s; got: %s", idx, instanceDir, dirInfo.Dir)
		}

		if pidFileLocation != dirInfo.PidFile {
			t.Errorf("item %d: expected: %s; got: %s", idx, pidFileLocation, dirInfo.PidFile)
		}

		if _, err := os.Stat(dirInfo.Dir); os.IsNotExist(err) {
			t.Errorf("item %d: no such instance directory %s", idx, dirInfo.Dir)
		}

		if _, err := os.Stat(dirInfo.PidFile); os.IsNotExist(err) {
			t.Errorf("item %d: no such pid file %s", idx, dirInfo.PidFile)
		}

	}

}

func TestMain(m *testing.M) {
	var err error

	home, err = ioutil.TempDir("", "mole")
	if err != nil {
		os.Exit(1)
	}

	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home)

	code := m.Run()

	os.RemoveAll(home)

	os.Exit(code)
}
