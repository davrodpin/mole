package mole

import (
	"fmt"
	"os"

	"github.com/davrodpin/mole/fsutils"

	"github.com/hpcloud/tail"
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

	dirInfo, err := fsutils.CreateInstanceDir(id)
	if err != nil {
		return nil, err
	}

	lfl, err := fsutils.GetLogFileLocation(id)
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
		PidFile: dirInfo.PidFile,
	}, nil
}

// ShowLogs displays all logs messages from a detached applications instance.
func ShowLogs(id string, follow bool) error {
	lfl, err := fsutils.GetLogFileLocation(id)
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
