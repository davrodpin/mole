package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/davrodpin/mole/fsutils"
	"github.com/gofrs/uuid"
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
		err := os.MkdirAll(home, 0755)
		if err != nil {
			return nil, err
		}
	}

	pfl, err := GetPidFileLocation(id)
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(pfl); !os.IsNotExist(err) {
		_, err = os.Stat(pfl)
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

// GetPidFileLocation return the file system location of the application
// instance in the file system.
func GetPidFileLocation(id string) (string, error) {
	d, err := fsutils.Dir()
	if err != nil {
		return "", err
	}

	pfp := filepath.Join(d, id, InstancePidFile)

	return pfp, nil
}

// GetLogFileLocation return the file system location of the file where all
// log messages are saved for an specific detached application instance.
func GetLogFileLocation(id string) (string, error) {
	d, err := fsutils.Dir()
	if err != nil {
		return "", err
	}

	lfp := filepath.Join(d, id, InstanceLogFile)

	return lfp, nil
}

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
