package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Store contains the map of aliases, where key contains the alias name.
type Store struct {
	Aliases map[string]*Alias `toml:"tunnels"`
}

// Alias represents settings of the ssh tunnel.
type Alias struct {
	// Local holds all local addresses configured on an alias.
	//
	// The type is specified as `interface{}` for backward-compatibility reasons
	// since only a single value was supported before.
	Local interface{} `toml:"local"`
	// Remote holds all remote addresses configured on an alias.
	//
	// The type is specified as `interface{}` for backward-compatibility reasons
	// since only a single value was supported before.
	Remote interface{} `toml:"remote"`

	Server            string        `toml:"server"`
	Key               string        `toml:"key"`
	Verbose           bool          `toml:"verbose"`
	Help              bool          `toml:"help"`
	Version           bool          `toml:"version"`
	Detach            bool          `toml:"detach"`
	Insecure          bool          `toml:"insecure"`
	KeepAliveInterval time.Duration `toml:"keep-alive-interval"`
}

func (t Alias) ReadLocal() ([]string, error) {
	return readAddress(t.Local)
}

func (t Alias) ReadRemote() ([]string, error) {
	return readAddress(t.Remote)
}

func readAddress(address interface{}) ([]string, error) {
	switch v := address.(type) {
	case string:
		return []string{v}, nil
	case []interface{}:
		sv := []string{}
		for _, e := range v {
			sv = append(sv, e.(string))
		}
		return sv, nil
	default:
		return nil, fmt.Errorf("couldn't load addresses: %v", address)
	}

}

func (t Alias) String() string {
	return fmt.Sprintf("[local=%s, remote=%s, server=%s, key=%s, verbose=%t, help=%t, version=%t, detach=%t, insecure=%t, ka-interval=%v]",
		t.Local, t.Remote, t.Server, t.Key, t.Verbose, t.Help, t.Version, t.Detach, t.Insecure, t.KeepAliveInterval)
}

// Save stores Alias to the Store.
func Save(name string, alias *Alias) (*Alias, error) {
	store, err := loadStore()
	if err != nil {
		return nil, fmt.Errorf("error while loading mole configuration: %v", err)
	}

	store.Aliases[name] = alias

	_, err = createStore(store)
	if err != nil {
		return nil, fmt.Errorf("error while saving mole configuration: %v", err)
	}

	return alias, nil
}

// FindByName finds the Alias in Store by name.
func FindByName(name string) (*Alias, error) {
	store, err := loadStore()
	if err != nil {
		return nil, fmt.Errorf("error while loading mole configuration: %v", err)
	}

	tun := store.Aliases[name]

	if tun == nil {
		return nil, fmt.Errorf("alias could not be found: %s", name)
	}

	return tun, nil
}

// FindAll finds all the Aliass in Store.
func FindAll() (map[string]*Alias, error) {
	store, err := loadStore()
	if err != nil {
		return nil, fmt.Errorf("error while loading mole configuration: %v", err)
	}

	return store.Aliases, nil
}

// Remove deletes Alias from the Store by name.
func Remove(name string) (*Alias, error) {
	store, err := loadStore()
	if err != nil {
		return nil, fmt.Errorf("error while loading mole configuration: %v", err)
	}

	tun := store.Aliases[name]

	if tun != nil {
		delete(store.Aliases, name)
		_, err := createStore(store)
		if err != nil {
			return nil, err
		}
	}

	return tun, nil
}

func loadStore() (*Store, error) {
	var store *Store

	sp, err := storePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(sp); err != nil {
		store = &Store{Aliases: make(map[string]*Alias)}
		store, err = createStore(store)
		if err != nil {
			return nil, err
		}

		return store, nil
	}

	if _, err := toml.DecodeFile(sp, &store); err != nil {
		return nil, err
	}

	return store, nil
}

func createStore(store *Store) (*Store, error) {
	sp, err := storePath()
	if err != nil {
		return nil, err
	}

	f, err := os.Create(sp)
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

func storePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".mole.conf"), nil
}
