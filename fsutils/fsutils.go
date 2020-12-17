package fsutils

import (
	"os"
	"path/filepath"
)

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
func CreateInstanceDir(appId string) (string, error) {
	home, err := Dir()
	if err != nil {
		return "", err
	}

	d := filepath.Join(home, appId)

	if _, err := os.Stat(d); os.IsNotExist(err) {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			return "", err
		}
	}

	return d, nil
}

// InstanceDir returns the location where all files related to a specific mole
// instance are persisted.
func InstanceDir(id string) (string, error) {
	home, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, id), nil
}
