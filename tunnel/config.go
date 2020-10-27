package tunnel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
	log "github.com/sirupsen/logrus"
)

const homeVar = "$HOME"

// SSHConfigFile finds specific attributes of a ssh server configured on a
// ssh config file.
type SSHConfigFile struct {
	sshConfig *ssh_config.Config
}

// NewSSHConfigFile creates a new instance of SSHConfigFile based on the
// ssh config file from configPath
func NewSSHConfigFile(configPath string) (*SSHConfigFile, error) {
	if strings.Contains(configPath, homeVar) {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		configPath = strings.ReplaceAll(configPath, homeVar, home)
	}

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

	localForward, err := r.getForward("LocalForward", host)
	if err != nil {
		log.Warningf("error reading local forwarding configuration from ssh config file: %v", err)
	}

	remoteForward, err := r.getForward("RemoteForward", host)
	if err != nil {
		log.Warningf("error reading remote configuration from ssh config file: %v", err)
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
		RemoteForward: remoteForward,
	}
}

func (r SSHConfigFile) getHostname(host string) string {
	hostname, err := r.sshConfig.Get(host, "Hostname")
	if err != nil {
		return ""
	}

	return hostname
}

func (r SSHConfigFile) getForward(forwardType, host string) (*ForwardConfig, error) {
	c, err := r.sshConfig.Get(host, forwardType)
	if err != nil {
		return nil, err
	}

	if c == "" {
		return nil, nil
	}

	l := strings.Fields(c)

	if len(l) < 2 {
		return nil, fmt.Errorf("malformed forwarding configuration on ssh config file: %s", l)
	}

	source := l[0]
	destination := l[1]

	if strings.HasPrefix(source, ":") {
		source = fmt.Sprintf("127.0.0.1%s", source)
	}

	if source != "" && !strings.Contains(source, ":") {
		source = fmt.Sprintf("127.0.0.1:%s", source)
	}

	return &ForwardConfig{Source: source, Destination: destination}, nil

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
	LocalForward  *ForwardConfig
	RemoteForward *ForwardConfig
}

// String returns a string representation of a SSHHost.
func (h SSHHost) String() string {
	return fmt.Sprintf("[hostname=%s, port=%s, user=%s, key=%s, identity_agent=%s, local_forward=%s, remote_forward=%s]", h.Hostname, h.Port, h.User, h.Key, h.IdentityAgent, h.LocalForward, h.RemoteForward)
}

// ForwardConfig represents either a LocalForward or a RemoteForward configuration
// for SSHHost.
type ForwardConfig struct {
	Source      string
	Destination string
}

// String returns a string representation of ForwardConfig.
func (f ForwardConfig) String() string {
	return fmt.Sprintf("[source=%s, destination=%s]", f.Source, f.Destination)
}
