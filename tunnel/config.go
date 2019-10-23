package tunnel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
	log "github.com/sirupsen/logrus"
)

// SSHConfigFile finds specific attributes of a ssh server configured on a
// ssh config file.
type SSHConfigFile struct {
	sshConfig *ssh_config.Config
}

// NewSSHConfigFile creates a new instance of SSHConfigFile based on the
// ssh config file from $HOME/.ssh/config.
func NewSSHConfigFile() (*SSHConfigFile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".ssh", "config")

	f, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		return nil, err
	}

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}

	log.Debugf("using ssh config file from: %s", configPath)

	return &SSHConfigFile{sshConfig: cfg}, nil
}

func NewEmptySSHConfigStruct() *SSHConfigFile {
	log.Debugf("generating an empty config struct")
	return &SSHConfigFile{sshConfig: &ssh_config.Config{}}
}

// Get consults a ssh config file to extract some ssh server attributes
// from it, returning a SSHHost. Any attribute which its value is an empty
// string is an attribute that could not be found in the ssh config file.
func (r SSHConfigFile) Get(host string) *SSHHost {
	hostname := r.getHostname(host)

	port, err := r.sshConfig.Get(host, "Port")
	if err != nil {
		port = ""
	}

	user, err := r.sshConfig.Get(host, "User")
	if err != nil {
		user = ""
	}

	localForward, err := r.getLocalForward(host)
	if err != nil {
		log.Warningf("error reading LocalForward configuration from ssh config file: %v", err)
	}

	key := r.getKey(host)

	identityAgent, err := r.sshConfig.Get(host, "IdentityAgent")
	if err != nil {
		identityAgent = ""
	}

	return &SSHHost{
		Hostname:      hostname,
		Port:          port,
		User:          user,
		Key:           key,
		IdentityAgent: identityAgent,
		LocalForward:  localForward,
	}
}

func (r SSHConfigFile) getHostname(host string) string {
	hostname, err := r.sshConfig.Get(host, "Hostname")
	if err != nil {
		return ""
	}

	return hostname
}

func (r SSHConfigFile) getLocalForward(host string) (*LocalForward, error) {
	var local, remote string

	c, err := r.sshConfig.Get(host, "LocalForward")
	if err != nil {
		return nil, err
	}

	if c == "" {
		return nil, nil
	}

	l := strings.Fields(c)

	if len(l) < 2 {
		return nil, fmt.Errorf("bad forwarding specification on ssh config file: %s", l)
	}

	local = l[0]
	remote = l[1]

	if strings.HasPrefix(local, ":") {
		local = fmt.Sprintf("127.0.0.1%s", local)
	}

	if local != "" && !strings.Contains(local, ":") {
		local = fmt.Sprintf("127.0.0.1:%s", local)
	}

	return &LocalForward{Local: local, Remote: remote}, nil

}

func (r SSHConfigFile) getKey(host string) string {
	id, err := r.sshConfig.Get(host, "IdentityFile")

	if err != nil {
		return ""
	}

	if id != "" {
		if strings.HasPrefix(id, "~") {
			return filepath.Join(os.Getenv("HOME"), id[1:])
		}

		return id
	}

	return ""
}

// SSHHost represents a host configuration extracted from a ssh config file.
type SSHHost struct {
	Hostname      string
	Port          string
	User          string
	Key           string
	IdentityAgent string
	LocalForward  *LocalForward
}

// String returns a string representation of a SSHHost.
func (h SSHHost) String() string {
	return fmt.Sprintf("[hostname=%s, port=%s, user=%s, key=%s, identity_agent=%s, local_forward=%s]", h.Hostname, h.Port, h.User, h.Key, h.IdentityAgent, h.LocalForward)
}

// LocalForward represents a LocalForward configuration for SSHHost.
type LocalForward struct {
	Local  string
	Remote string
}

// String returns a string representation of LocalForward.
func (f LocalForward) String() string {
	return fmt.Sprintf("[local=%s, remote=%s]", f.Local, f.Remote)
}
