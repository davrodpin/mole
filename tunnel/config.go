package tunnel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
	log "github.com/sirupsen/logrus"
)

// Resolver finds specific attributes of a ssh server configured on a ssh config
// file.
type Resolver struct {
	sshConfig *ssh_config.Config
}

// NewResolver creates a new instance of Resolver based on the given ssh config
// file path.
func NewResolver(configPath string) (*Resolver, error) {
	f, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		return nil, err
	}

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}

	log.Debugf("using ssh config file from: %s", configPath)

	return &Resolver{sshConfig: cfg}, nil
}

// Resolve consults a ssh config file to extract some ssh server attributes
// from it, returning a ResolvedHost. Any attribute which its value is an empty
// string is an attribute that could not be found in the ssh config file.
func (r Resolver) Resolve(host string) *ResolvedHost {
	hostname := r.resolveHostname(host)

	port, err := r.sshConfig.Get(host, "Port")
	if err != nil {
		port = ""
	}

	user, err := r.sshConfig.Get(host, "User")
	if err != nil {
		user = ""
	}

	key := r.resolveKey(host)

	return &ResolvedHost{
		Hostname: hostname,
		Port:     port,
		User:     user,
		Key:      key,
	}
}

func (r Resolver) resolveHostname(host string) string {
	hostname, err := r.sshConfig.Get(host, "Hostname")

	if err != nil {
		return host
	}

	if hostname == "" {
		hostname = host
	}

	return hostname
}

func (r Resolver) resolveKey(host string) string {
	id, err := r.sshConfig.Get(host, "IdentityFile")

	if err != nil {
		return ""
	}

	if id != "" {
		if strings.HasPrefix(id, "~") {
			return filepath.Join(os.Getenv("HOME"), id[1:])
		} else {
			return id
		}
	}

	return ""
}

// ResolvedHost holds information extracted from a ssh config file.
type ResolvedHost struct {
	Hostname string
	Port     string
	User     string
	Key      string
}

// String returns a string representation of a ResolvedHost.
func (rh ResolvedHost) String() string {
	return fmt.Sprintf("[hostname=%s, port=%s, user=%s, key=%s]", rh.Hostname, rh.Port, rh.User, rh.Key)
}
