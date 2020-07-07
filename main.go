package main

import (
	"os"

	"github.com/davrodpin/mole/cmd"

	"github.com/awnumar/memguard"
	log "github.com/sirupsen/logrus"
)

func main() {
	// memguard is used to securely keep sensitive information in memory.
	// This call makes sure all data will be destroy when the program exits.
	defer memguard.Purge()

	log.SetOutput(os.Stdout)

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
