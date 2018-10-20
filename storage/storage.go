package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Store contains tunnels map by name
type Store struct {
	Tunnels map[string]*Tunnel `toml:"tunnels"`
}

// Tunnel struct
type Tunnel struct {
	Local   string `toml:"local"`
	Remote  string `toml:"remote"`
	Server  string `toml:"server"`
	Key     string `toml:"key"`
	Verbose bool   `toml:"verbose"`
	Help    bool   `toml:"help"`
	Version bool   `toml:"version"`
}

func (t Tunnel) String() string {
	return fmt.Sprintf("[local=%s, remote=%s, server=%s, key=%s, verbose=%t, help=%t, version=%t]",
		t.Local, t.Remote, t.Server, t.Key, t.Verbose, t.Help, t.Version)
}

// Save tunnel by name
func Save(name string, tunnel *Tunnel) (*Tunnel, error) {
	store, err := loadStore()
	if err != nil {
		return nil, fmt.Errorf("error while loading mole configuration: %v", err)
	}

	store.Tunnels[name] = tunnel

	_, err = createStore(store)
	if err != nil {
		return nil, fmt.Errorf("error while saving mole configuration: %v", err)
	}

	return tunnel, nil
}

// FindByName finds a tunnel by name
func FindByName(name string) (*Tunnel, error) {
	store, err := loadStore()
	if err != nil {
		return nil, fmt.Errorf("error while loading mole configuration: %v", err)
	}

	tun := store.Tunnels[name]

	if tun == nil {
		return nil, fmt.Errorf("tunnel could not be found: %s", name)
	}

	return tun, nil
}

// Remove tunnel by name
func Remove(name string) (*Tunnel, error) {
	store, err := loadStore()
	if err != nil {
		return nil, fmt.Errorf("error while loading mole configuration: %v", err)
	}

	tun := store.Tunnels[name]

	if tun != nil {
		delete(store.Tunnels, name)
		_, err := createStore(store)
		if err != nil {
			return nil, err
		}
	}

	return tun, nil
}

func loadStore() (*Store, error) {
	var store *Store

	if _, err := os.Stat(storePath()); err != nil {
		store = &Store{Tunnels: make(map[string]*Tunnel)}
		store, err = createStore(store)
		if err != nil {
			return nil, err
		}

		return store, nil
	}

	if _, err := toml.DecodeFile(storePath(), &store); err != nil {
		return nil, err
	}

	return store, nil
}

func createStore(store *Store) (*Store, error) {
	f, err := os.Create(storePath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b := bufio.NewWriter(f)
	e := toml.NewEncoder(b)

	if err := e.Encode(&store); err != nil {
		return nil, err
	}

	return store, nil
}

func storePath() string {
	return filepath.Join(os.Getenv("HOME"), ".mole.conf")
}
