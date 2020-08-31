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
