package app_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/davrodpin/mole/app"
)

var (
	// home is a temporary directory that acts as the user home directory.
	// It is used on most tests to validates files created by Mole and their
	// contents.
	home string
)

func TestDetachedInstanceFileLocations(t *testing.T) {
	id := "TestDetachedInstanceFileLocations"

	di, err := app.NewDetachedInstance(id)
	if err != nil {
		t.Errorf("error creating a new detached instance: %v", err)
	}

	if _, err := os.Stat(di.LogFile); os.IsNotExist(err) {
		t.Errorf("log file does not exist: %v", err)
	}

	if _, err := os.Stat(di.PidFile); os.IsNotExist(err) {
		t.Errorf("pid file does not exist: %v", err)
	}

	lfl, err := app.GetLogFileLocation(id)
	if err != nil {
		t.Errorf("error retrieving log file location: %v", err)
	}

	if _, err := os.Stat(lfl); os.IsNotExist(err) {
		t.Errorf("log file does not exist: %v", err)
	}

}

func TestDetachedInstanceGeneratedId(t *testing.T) {

	di, err := app.NewDetachedInstance("")
	if err != nil {
		t.Errorf("error creating a new detached instance: %v", err)
	}

	if di.Id == "" {
		t.Errorf("detached instance id is empty")
	}
}

func TestDetachedInstanceAlreadyRunning(t *testing.T) {
	id := "TestDetachedInstanceAlreadyRunning"

	os.MkdirAll(filepath.Join(home, ".mole", id), 0755)
	pidFileLocation := filepath.Join(home, ".mole", id, app.InstancePidFile)
	ioutil.WriteFile(pidFileLocation, []byte("1234"), 0644)

	_, err := app.NewDetachedInstance(id)

	if err == nil {
		t.Errorf("error expected but got nil")
	}

}

func TestShowLogs(t *testing.T) {
	id := "TestDetachedInstanceAlreadyRunning"

	os.MkdirAll(filepath.Join(home, ".mole", id), 0755)
	logFileLocation := filepath.Join(home, ".mole", id, app.InstanceLogFile)
	ioutil.WriteFile(logFileLocation, []byte("first log message\nsecond log message\nthird log message\n"), 0644)

	err := app.ShowLogs(id, false)

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
