package tunnel

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	HostMissing = "server host has to be provided as part of the server address"
)

// Server holds the SSH Server attributes used for the client to connect to it.
type Server struct {
	Name    string
	Address string
	User    string
	Key     *PemKey
	// Insecure is a flag to indicate if the host keys should be validated.
	Insecure bool
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

	c, err := NewSSHConfigFile()
	if err != nil {
		return nil, fmt.Errorf("error accessing %s: %v", host, err)
	}

	h := c.Get(host)
	hostname = reconcileHostname(host, h.Hostname)
	port = reconcilePort(port, h.Port)
	user = reconcileUser(user, h.User)
	key = reconcileKey(key, h.Key)

	if host == "" {
		return nil, fmt.Errorf(HostMissing)
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

	pk, err := NewPemKey(key, "")
	if err != nil {
		return nil, fmt.Errorf("error while reading key %s: %v", key, err)
	}

	return &Server{
		Name:    host,
		Address: fmt.Sprintf("%s:%s", hostname, port),
		User:    user,
		Key:     pk,
	}, nil
}

// String provided a string representation of a Server.
func (s Server) String() string {
	return fmt.Sprintf("[name=%s, address=%s, user=%s]", s.Name, s.Address, s.User)
}

// Tunnel represents the ssh tunnel used to forward a local connection to a
// a remote endpoint through a ssh server.
type Tunnel struct {
	// Ready tells when the Tunnel is ready to accept connections
	Ready    chan bool
	local    string
	server   *Server
	remote   string
	done     chan error
	client   *ssh.Client
	listener net.Listener
}

// New creates a new instance of Tunnel.
func New(localAddress string, server *Server, remoteAddress string) *Tunnel {
	cfg, err := NewSSHConfigFile()
	if err != nil {
		log.Warningf("error to read ssh config: %v", err)
	}

	sh := cfg.Get(server.Name)
	localAddress = reconcileLocal(localAddress, sh.LocalForward.Local)
	remoteAddress = reconcileRemote(remoteAddress, sh.LocalForward.Remote)

	return &Tunnel{
		Ready:  make(chan bool, 1),
		local:  localAddress,
		server: server,
		remote: remoteAddress,
		done:   make(chan error),
	}
}

// Start creates a new ssh tunnel, allowing data exchange between the local and
// remote endpoints.
func (t *Tunnel) Start() error {
	var once sync.Once

	_, err := t.Listen()
	if err != nil {
		return err
	}
	defer t.listener.Close()

	log.Debugf("tunnel: %s", t)

	log.WithFields(log.Fields{
		"local_address": t.local,
	}).Info("listening on local address")

	go func(t *Tunnel) {
		for {

			once.Do(func() {
				t.Ready <- true
			})

			conn, err := t.listener.Accept()
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
	}(t)

	select {
	case err = <-t.done:
		if t.client != nil {
			t.client.Conn.Close()
		}
		return err
	}
}

// Listen binds the local address configured on Tunnel.
func (t *Tunnel) Listen() (net.Listener, error) {

	if t.listener != nil {
		return t.listener, nil
	}

	local, err := net.Listen("tcp", t.local)
	if err != nil {
		return nil, err
	}

	t.listener = local
	t.local = local.Addr().String()

	return t.listener, nil
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

	if t.client == nil {
		t.client, err = ssh.Dial("tcp", t.server.Address, c)
		if err != nil {
			return nil, fmt.Errorf("server dial error: %s", err)
		}

		log.WithFields(log.Fields{
			"server": t.server,
		}).Debug("new connection established to server")

	}

	remoteConn, err := t.client.Dial("tcp", t.remote)
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
	signer, err := server.Key.Parse()
	if err != nil {
		return nil, err
	}

	clb, err := knownHostsCallback(server.Insecure)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: clb,
		Timeout:         3 * time.Second,
	}, nil
}

func copyConn(writer, reader net.Conn) {
	_, err := io.Copy(writer, reader)
	if err != nil {
		log.Errorf("%v", err)
	}
}

func knownHostsCallback(insecure bool) (ssh.HostKeyCallback, error) {
	var clb func(hostname string, remote net.Addr, key ssh.PublicKey) error

	if insecure {
		clb = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}
	} else {
		var err error
		knownHostFile := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
		log.Debugf("known_hosts file used: %s", knownHostFile)

		clb, err = knownhosts.New(knownHostFile)
		if err != nil {
			return nil, fmt.Errorf("error while parsing 'known_hosts' file: %s: %v", knownHostFile, err)
		}
	}

	return clb, nil
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

func reconcileLocal(givenLocal, resolvedLocal string) string {

	if givenLocal == "" && resolvedLocal != "" {
		return resolvedLocal
	}
	if givenLocal == "" {
		return "127.0.0.1:0"
	}
	if strings.HasPrefix(givenLocal, ":") {
		return fmt.Sprintf("127.0.0.1%s", givenLocal)
	}

	return givenLocal
}

func reconcileRemote(givenRemote, resolvedRemote string) string {

	if givenRemote == "" && resolvedRemote != "" {
		return resolvedRemote
	}
	if strings.HasPrefix(givenRemote, ":") {
		return fmt.Sprintf("127.0.0.1%s", givenRemote)
	}

	return givenRemote
}
