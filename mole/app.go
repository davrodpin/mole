package mole

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/davrodpin/mole/fsutils"
	"github.com/davrodpin/mole/rpc"

	"github.com/hpcloud/tail"
)

const (
	InstancePidFile = "pid"
	InstanceLogFile = "mole.log"
)

// DetachedInstance holds the location to directories and files associated
// with an application instance running on background.
type DetachedInstance struct {
	// Id is the unique identifier of a detached application instance. The value
	// can be either the alias or a unique alphanumeric value.
	Id string
	// LogFile points to a file path in the file system where the application
	// log file is stored.
	LogFile string
	// PidFile points to a file path in the file system where the application
	// procces identifier is stored.
	PidFile string
}

// NewDetachedInstance returns a new instance of DetachedInstance, making sure
// the application instance directory is created.
func NewDetachedInstance(id string) (*DetachedInstance, error) {
	if id == "" {
		return nil, fmt.Errorf("application instance id can't be empty")
	}

	_, err := fsutils.CreateInstanceDir(id)
	if err != nil {
		return nil, err
	}

	pfl, err := GetPidFileLocation(id)
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(pfl); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(pfl)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while reading from pid file %s: %v", pfl, err)
		}

		pid := string(data)

		if pid != "" {
			return nil, fmt.Errorf("an instance of mole with pid %s seems to be already running", pid)
		}

	}

	pf, err := os.Create(pfl)
	if err != nil {
		return nil, fmt.Errorf("could not create pid file for application instance %s: %v", id, err)
	}
	defer pf.Close()
	pf.WriteString(strconv.Itoa(os.Getpid()))

	lfl, err := GetLogFileLocation(id)
	if err != nil {
		return nil, err
	}

	lf, err := os.Create(lfl)
	if err != nil {
		return nil, fmt.Errorf("could not create log file for application instance %s: %v", id, err)
	}
	defer lf.Close()

	return &DetachedInstance{
		Id:      id,
		LogFile: lfl,
		PidFile: pfl,
	}, nil
}

// GetPidFileLocation returns the file system location of the application
// instance in the file system.
func GetPidFileLocation(id string) (string, error) {
	d, err := fsutils.Dir()
	if err != nil {
		return "", err
	}

	pfp := filepath.Join(d, id, InstancePidFile)

	return pfp, nil
}

// GetLogFileLocation returns the file system location of the file where all
// log messages are saved for an specific detached application instance.
func GetLogFileLocation(id string) (string, error) {
	d, err := fsutils.Dir()
	if err != nil {
		return "", err
	}

	lfp := filepath.Join(d, id, InstanceLogFile)

	return lfp, nil
}

// ShowLogs displays all logs messages from a detached applications instance.
func ShowLogs(id string, follow bool) error {
	lfl, err := GetLogFileLocation(id)
	if err != nil {
		return err
	}

	t, err := tail.TailFile(lfl, tail.Config{Follow: follow})
	if err != nil {
		return err
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}

	return nil
}

// Rpc calls a remote procedure on another mole instance given its id or alias.
func Rpc(id, method string, params interface{}) (string, error) {
	d, err := fsutils.InstanceDir(id)
	if err != nil {
		return "", err
	}

	rf := filepath.Join(d, "rpc")

	addr, err := ioutil.ReadFile(rf)
	if err != nil {
		return "", err
	}

	resp, err := rpc.Call(context.Background(), string(addr), method, params)
	if err != nil {
		return "", err
	}

	r, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", err
	}

	return string(r), nil
}
