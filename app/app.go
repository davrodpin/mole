package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/davrodpin/mole/fsutils"
	"github.com/gofrs/uuid"
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
// the instance directory is created.
func NewDetachedInstance(id string) (*DetachedInstance, error) {
	instanceDir, err := fsutils.Dir()
	if err != nil {
		return nil, err
	}

	if id == "" {
		u, err := uuid.NewV4()
		if err != nil {
			return nil, fmt.Errorf("could not auto generate app instance id: %v", err)
		}
		id = u.String()[:8]
	}

	home := filepath.Join(instanceDir, id)

	if _, err := os.Stat(home); os.IsNotExist(err) {
		err := os.Mkdir(home, 0755)
		if err != nil {
			return nil, err
		}
	}

	lfp, err := GetPidFileLocation(id)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(lfp); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(lfp)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while opening pid file %s: %v", lfp, err)
		}

		pid := string(data)

		if pid != "" {
			return nil, fmt.Errorf("an instance of mole with pid %s seems to be already running", pid)
		}

	}

	lf, err := os.Create(lfp)
	if err != nil {
		return nil, fmt.Errorf("could not create log file for application instance %s: %v", id, err)
	}
	defer lf.Close()

	pfp := filepath.Join(home, InstancePidFile)
	pf, err := os.Create(pfp)
	if err != nil {
		return nil, fmt.Errorf("could not create pid file for application instance %s: %v", id, err)
	}
	defer pf.Close()
	pf.WriteString(strconv.Itoa(os.Getpid()))

	return &DetachedInstance{
		Id:      id,
		LogFile: lfp,
		PidFile: pfp,
	}, nil
}

// GetPidFileLocation return the file system location of the application
// instance in the file system.
func GetPidFileLocation(id string) (string, error) {
	d, err := fsutils.Dir()
	if err != nil {
		return "", err
	}

	lfp := filepath.Join(d, id, InstancePidFile)

	return lfp, nil
}
