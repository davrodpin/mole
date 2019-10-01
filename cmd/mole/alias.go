package main

import (
	"fmt"
	"strings"

	"github.com/davrodpin/mole/cli"
	"github.com/davrodpin/mole/storage"
)

func lsAliases() error {
	aliases, err := storage.FindAll()
	if err != nil {
		return err
	}

	as := []string{}
	for a := range aliases {
		as = append(as, a)
	}

	fmt.Printf("alias list: %s\n", strings.Join(as, ", "))

	return nil
}

func app2alias(app cli.App) *storage.Alias {
	return &storage.Alias{
		Local:             app.Local.List(),
		Remote:            app.Remote.List(),
		Server:            app.Server.String(),
		Key:               app.Key,
		Verbose:           app.Verbose,
		Help:              app.Help,
		Version:           app.Version,
		Detach:            app.Detach,
		Insecure:          app.Insecure,
		KeepAliveInterval: app.KeepAliveInterval,
		Timeout:           app.Timeout,
	}
}

func alias2app(t *storage.Alias) (*cli.App, error) {
	sla, err := t.ReadLocal()
	if err != nil {
		return nil, err
	}

	lal := cli.AddressInputList{}
	for _, la := range sla {
		lal.Set(la)
	}

	sra, err := t.ReadRemote()
	if err != nil {
		return nil, err
	}

	ral := cli.AddressInputList{}
	for _, ra := range sra {
		ral.Set(ra)
	}

	server := cli.AddressInput{}
	server.Set(t.Server)

	return &cli.App{
		Command:           "start",
		Local:             lal,
		Remote:            ral,
		Server:            server,
		Key:               t.Key,
		Verbose:           t.Verbose,
		Help:              t.Help,
		Version:           t.Version,
		Detach:            t.Detach,
		Insecure:          t.Insecure,
		KeepAliveInterval: t.KeepAliveInterval,
		Timeout:           t.Timeout,
	}, nil
}
