package fsutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const (
	InstancePidFile = "pid"
	InstanceLogFile = "mole.log"
)

type InstanceDirInfo struct {
	Id      string
	Dir     string
	PidFile string
}

// Dir returns the location where all mole related files are persisted,
// including alias configuration and log files.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	mp := filepath.Join(home, ".mole")

	return mp, nil
}

// CreateHomeDir creates then returns the location where all mole related files
// are persisted, including alias configuration and log files.
func CreateHomeDir() (string, error) {

	home, err := Dir()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(home); os.IsNotExist(err) {
		err := os.MkdirAll(home, 0755)
		if err != nil {
			return "", err
		}
	}

	return home, err
}

// CreateInstanceDir creates and then returns the location where all files
// related to a specific mole instance are persisted.
// Along with the directory this function also created a pid file where the
// instance process id is stored.
func CreateInstanceDir(appId string) (*InstanceDirInfo, error) {
	home, err := Dir()
	if err != nil {
		return nil, err
	}

	d := filepath.Join(home, appId)

	if _, err := os.Stat(d); !os.IsNotExist(err) {
		return InstanceDir(appId)
	}

	err = os.MkdirAll(d, 0755)
	if err != nil {
		return nil, err
	}

	pfl, err := createPidFile(appId)
	if err != nil {
		return nil, err
	}

	return &InstanceDirInfo{
		Id:      appId,
		Dir:     d,
		PidFile: pfl,
	}, nil
}

// InstanceDir returns the location where all files related to a specific mole
// instance are persisted.
func InstanceDir(id string) (*InstanceDirInfo, error) {
	home, err := Dir()
	if err != nil {
		return nil, err
	}

	d := filepath.Join(home, id)

	pfl, err := GetPidFileLocation(id)
	if err != nil {
		return nil, err
	}

	return &InstanceDirInfo{
		Id:      id,
		Dir:     d,
		PidFile: pfl,
	}, nil
}

// GetPidFileLocation returns the file system location of the application
// instance in the file system.
func GetPidFileLocation(id string) (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}

	pfp := filepath.Join(d, id, InstancePidFile)

	return pfp, nil
}

// GetLogFileLocation returns the file system location of the file where all
// log messages are saved for an specific detached application instance.
func GetLogFileLocation(id string) (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}

	lfp := filepath.Join(d, id, InstanceLogFile)

	return lfp, nil
}

func createPidFile(id string) (string, error) {
	pfl, err := GetPidFileLocation(id)
	if err != nil {
		return "", err
	}

	if _, err = os.Stat(pfl); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(pfl)
		if err != nil {
			return "", fmt.Errorf("something went wrong while reading from pid file %s: %v", pfl, err)
		}

		pid := string(data)

		if pid != "" {
			return "", fmt.Errorf("an instance of mole with pid %s seems to be already running", pid)
		}

	}

	pf, err := os.Create(pfl)
	if err != nil {
		return "", fmt.Errorf("could not create pid file for application instance %s: %v", id, err)
	}
	defer pf.Close()
	pf.WriteString(strconv.Itoa(os.Getpid()))

	return pfl, nil

}
