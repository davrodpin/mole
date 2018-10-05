package tunnel

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Server holds the SSH Server attributes used for the client to connect to it.
type Server struct {
	Name    string
	Address string
	User    string
	Key     string
}

// NewServer creates a new instance of Server using $HOME/.ssh/config to
// resolve the missing connection attributes (e.g. user, hostname, port and
// key) required to connect to the remote server, if any.
func NewServer(user, address, key string) (*Server, error) {
	var host string
	var hostname string
	var port string

	host = address
	if strings.Contains(host, ":") {
		args := strings.Split(host, ":")
		host = args[0]
		port = args[1]
	}

	c := filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	if _, err := os.Stat(c); err != nil {
		hostname = host
	} else {
		r, err := NewResolver(c)
		if err != nil {
			return nil, fmt.Errorf("error accessing %s: %v", c, err)
		}

		rh := r.Resolve(host)
		hostname = reconcileHostname(host, rh.Hostname)
		port = reconcilePort(port, rh.Port)
		user = reconcileUser(user, rh.User)
		key = reconcileKey(key, rh.Key)
	}

	if host == "" {
		return nil, fmt.Errorf("server host has to be provided as part of the server address")
	}

	if hostname == "" {
		return nil, fmt.Errorf("no server hostname (ip) could be found for server %s", host)
	}

	if port == "" {
		port = "22"
	}

	if user == "" {
		return nil, fmt.Errorf("no user could be found for server %s", host)
	}

	if key == "" {
		key = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	}

	if _, err := os.Stat(key); err != nil {
		return nil, fmt.Errorf("key file does not exists: %s: %v", key, err)
	}

	return &Server{
		Name:    host,
		Address: fmt.Sprintf("%s:%s", hostname, port),
		User:    user,
		Key:     key,
	}, nil
}

// String provided a string representation os a Server.
func (s Server) String() string {
	return fmt.Sprintf("[name=%s, address=%s, user=%s, key=%s]", s.Name, s.Address, s.User, s.Key)
}

// Tunnel represents the ssh tunnel used to forward a local connection to a
// a remote endpoint through a ssh server.
type Tunnel struct {
	local  string
	server *Server
	remote string
	done   chan error
}

// New creates a new instance of Tunnel.
func New(localAddress string, server *Server, remoteAddress string) *Tunnel {

	if localAddress == "" {
		localAddress = "127.0.0.1:0"
	}

	return &Tunnel{
		local:  localAddress,
		server: server,
		remote: remoteAddress,
		done:   make(chan error),
	}
}

// Start creates a new ssh tunnel, allowing data exchange between the local and
// remote endpoints.
func (t *Tunnel) Start() error {
	local, err := net.Listen("tcp", t.local)
	if err != nil {
		return err
	}
	defer local.Close()

	t.local = local.Addr().String()

	log.Debugf("tunnel: %s", t)

	log.WithFields(log.Fields{
		"local_address": t.local,
	}).Info("listening on local address")

	go func(l net.Listener, t *Tunnel) {
		for {
			conn, err := l.Accept()
			if err != nil {
				t.done <- fmt.Errorf("error while establishing new connection: %v", err)
				return
			}

			log.WithFields(log.Fields{
				"address": conn.RemoteAddr(),
			}).Debug("new connection")

			err = t.forward(conn)
			if err != nil {
				t.done <- err
				return
			}
		}
	}(local, t)

	select {
	case err = <-t.done:
		return err
	}
}

func (t *Tunnel) forward(localConn net.Conn) error {
	remoteConn, err := t.proxy()
	if err != nil {
		return err
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)

	return nil
}

// Stop cancels the tunnel, closing all connections.
func (t Tunnel) Stop() {
	t.done <- nil
}

// String returns a string representation of a Tunnel.
func (t Tunnel) String() string {
	return fmt.Sprintf("[local:%s, server:%s, remote:%s]", t.local, t.server.Address, t.remote)
}

func (t *Tunnel) proxy() (net.Conn, error) {
	c, err := sshClientConfig(*t.server)
	if err != nil {
		return nil, fmt.Errorf("error generating ssh client config: %s", err)
	}

	serverConn, err := ssh.Dial("tcp", t.server.Address, c)
	if err != nil {
		return nil, fmt.Errorf("server dial error: %s", err)
	}

	log.WithFields(log.Fields{
		"server": t.server,
	}).Debug("new connection established to server")

	remoteConn, err := serverConn.Dial("tcp", t.remote)
	if err != nil {
		return nil, fmt.Errorf("remote dial error: %s", err)
	}

	log.WithFields(log.Fields{
		"remote": t.remote,
		"server": t.server,
	}).Debug("new connection established to remote")

	return remoteConn, nil
}

func sshClientConfig(server Server) (*ssh.ClientConfig, error) {
	key, err := ioutil.ReadFile(server.Key)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	callback, err := knownHostsCallback()
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: callback,
		Timeout:         3 * time.Second,
	}, nil
}

func copyConn(writer, reader net.Conn) {
	_, err := io.Copy(writer, reader)
	if err != nil {
		log.Errorf("%v", err)
	}
}

func knownHostsCallback() (ssh.HostKeyCallback, error) {
	knownHostFile := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")

	log.Debugf("known_hosts file used: %s", knownHostFile)

	callback, err := knownhosts.New(knownHostFile)
	if err != nil {
		return nil, fmt.Errorf("error while parsing 'known_hosts' file: %s: %v", knownHostFile, err)
	}

	return callback, nil
}

func reconcileHostname(givenHostname, resolvedHostname string) string {
	if resolvedHostname != "" {
		return resolvedHostname
	}

	if resolvedHostname == "" && givenHostname != "" {
		return givenHostname
	}

	return ""
}

func reconcilePort(givenPort, resolvedPort string) string {
	if givenPort != "" {
		return givenPort
	}

	if givenPort == "" && resolvedPort != "" {
		return resolvedPort
	}

	return ""
}

func reconcileUser(givenUser, resolvedUser string) string {
	if givenUser != "" {
		return givenUser
	}

	if givenUser == "" && resolvedUser != "" {
		return resolvedUser
	}

	return ""
}

func reconcileKey(givenKey, resolvedKey string) string {
	if givenKey != "" {
		return givenKey
	}

	if givenKey == "" && resolvedKey != "" {
		return resolvedKey
	}

	return ""
}
