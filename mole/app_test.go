package mole_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/davrodpin/mole/fsutils"
	"github.com/davrodpin/mole/mole"
)

var (
	// home is a temporary directory that acts as the user home directory.
	// It is used on most tests to validates files created by Mole and their
	// contents.
	home string
)

func TestDetachedInstanceFileLocations(t *testing.T) {
	id := "TestDetachedInstanceFileLocations"

	di, err := mole.NewDetachedInstance(id)
	if err != nil {
		t.Errorf("error creating a new detached instance: %v", err)
	}

	if _, err := os.Stat(di.LogFile); os.IsNotExist(err) {
		t.Errorf("log file does not exist: %v", err)
	}

	if _, err := os.Stat(di.PidFile); os.IsNotExist(err) {
		t.Errorf("pid file does not exist: %v", err)
	}

	lfl, err := fsutils.GetLogFileLocation(id)
	if err != nil {
		t.Errorf("error retrieving log file location: %v", err)
	}

	if _, err := os.Stat(lfl); os.IsNotExist(err) {
		t.Errorf("log file does not exist: %v", err)
	}

}

func TestShowLogs(t *testing.T) {
	id := "TestDetachedInstanceAlreadyRunning"

	os.MkdirAll(filepath.Join(home, ".mole", id), 0755)
	logFileLocation := filepath.Join(home, ".mole", id, fsutils.InstanceLogFile)
	ioutil.WriteFile(logFileLocation, []byte("first log message\nsecond log message\nthird log message\n"), 0644)

	err := mole.ShowLogs(id, false)

	if err != nil {
		t.Errorf("error showing logs: %v", err)
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
