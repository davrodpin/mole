package tunnel

import (
	"errors"
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
	HostMissing       = "server host has to be provided as part of the server address"
	RandomPortAddress = "127.0.0.1:0"
	NoRemoteGiven     = "cannot create a tunnel without at least one remote address"
)

// Server holds the SSH Server attributes used for the client to connect to it.
type Server struct {
	Name    string
	Address string
	User    string
	Key     *PemKey
	// Insecure is a flag to indicate if the host keys should be validated.
	Insecure bool
	Timeout  time.Duration
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
		// Ignore file doesnt exists
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error accessing %s: %v", host, err)
		}
	}

	// // If ssh config file doesnt exists, create an empty ssh config struct to avoid nil pointer deference
	if errors.Is(err, os.ErrNotExist) {
		c = NewEmptySSHConfigStruct()
	}

	h := c.Get(host)
	hostname = reconcile(h.Hostname, host)
	port = reconcile(port, h.Port)
	user = reconcile(user, h.User)
	key = reconcile(key, h.Key)

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
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not obtain user home directory: %v", err)
		}

		key = filepath.Join(home, ".ssh", "id_rsa")
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

type SSHChannel struct {
	Local    string
	Remote   string
	listener net.Listener
	conn     net.Conn
}

func (ch *SSHChannel) Close() {
	if ch.listener != nil {
		ch.listener.Close()
	}
}

func (ch SSHChannel) String() string {
	return fmt.Sprintf("[local=%s, remote=%s]", ch.Local, ch.Remote)
}

// Tunnel represents the ssh tunnel and the channels connecting local and
// remote endpoints.
type Tunnel struct {
	// Ready tells when the Tunnel is ready to accept connections
	Ready chan bool
	// KeepAliveInterval is the time period used to send keep alive packets to
	// the remote ssh server
	KeepAliveInterval time.Duration

	server   *Server
	channels []*SSHChannel
	done     chan error
	client   *ssh.Client
}

// New creates a new instance of Tunnel.
func New(server *Server, channels []*SSHChannel) (*Tunnel, error) {

	for _, channel := range channels {
		if channel.Local == "" || channel.Remote == "" {
			return nil, fmt.Errorf("invalid ssh channel: local=%s, remote=%s", channel.Local, channel.Remote)
		}
	}

	return &Tunnel{
		Ready:    make(chan bool, 1),
		channels: channels,
		server:   server,
		done:     make(chan error),
	}, nil
}

// Start creates the ssh tunnel and initialized all channels allowing data
// exchange between local and remote enpoints.
func (t *Tunnel) Start() error {
	err := t.Listen()
	if err != nil {
		return err
	}
	defer func() {
		for _, ch := range t.channels {
			ch.Close()
		}
	}()

	log.Debugf("tunnel: %s", t)

	//connecting to ssh server
	err = t.dial()
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(t.channels))

	// wait for all ssh channels to be ready to accept connections then sends a
	// single message signalling all tunnel are ready
	go func(tunnel *Tunnel, waitgroup *sync.WaitGroup) {
		waitgroup.Wait()
		tunnel.Ready <- true
	}(t, wg)

	for _, ch := range t.channels {
		go func(channel *SSHChannel, waitgroup *sync.WaitGroup) {
			var err error
			var once sync.Once

			for {

				once.Do(func() {
					log.WithFields(log.Fields{
						"local":  channel.Local,
						"remote": channel.Remote,
					}).Info("tunnel is ready")

					waitgroup.Done()
				})

				channel.conn, err = channel.listener.Accept()
				if err != nil {
					t.done <- fmt.Errorf("error while establishing new connection: %v", err)
					return
				}

				log.WithFields(log.Fields{
					"address": channel.conn.RemoteAddr(),
				}).Debug("new connection")

				err = t.startChannel(channel)
				if err != nil {
					t.done <- err
					return
				}
			}
		}(ch, wg)
	}

	select {
	case err = <-t.done:
		if t.client != nil {
			t.client.Conn.Close()
		}
		return err
	}
}

// Listen creates tcp listeners for each channel defined.
func (t *Tunnel) Listen() error {
	for _, ch := range t.channels {
		if ch.listener == nil {
			l, err := net.Listen("tcp", ch.Local)
			if err != nil {
				return err
			}

			ch.listener = l
			ch.Local = l.Addr().String()
		}
	}

	return nil
}

func (t *Tunnel) startChannel(channel *SSHChannel) error {
	if t.client == nil {
		return fmt.Errorf("new channel can't be established: missing connection to the ssh server")
	}

	remoteConn, err := t.client.Dial("tcp", channel.Remote)
	if err != nil {
		return fmt.Errorf("remote dial error: %s", err)
	}

	go copyConn(channel.conn, remoteConn)
	go copyConn(remoteConn, channel.conn)

	log.WithFields(log.Fields{
		"remote": channel.Remote,
		"server": t.server,
	}).Debug("new connection established to remote")

	return nil
}

// Stop cancels the tunnel, closing all connections.
func (t Tunnel) Stop() {
	t.done <- nil
}

// String returns a string representation of a Tunnel.
func (t Tunnel) String() string {
	return fmt.Sprintf("[channels:%s, server:%s]", t.channels, t.server.Address)
}

func (t *Tunnel) dial() error {
	if t.client != nil {
		return nil
	}

	c, err := sshClientConfig(*t.server)
	if err != nil {
		return fmt.Errorf("error generating ssh client config: %s", err)
	}

	t.client, err = ssh.Dial("tcp", t.server.Address, c)
	if err != nil {
		return fmt.Errorf("server dial error: %s", err)
	}

	go t.keepAlive()

	log.WithFields(log.Fields{
		"server": t.server,
	}).Debug("new connection established to server")

	return nil
}

func (t *Tunnel) keepAlive() {
	ticker := time.NewTicker(t.KeepAliveInterval)

	for {
		select {
		case <-ticker.C:
			_, _, err := t.client.SendRequest("keepalive@mole", true, nil)
			if err != nil {
				log.Warnf("error sending keep-alive request to ssh server: %v", err)
			}
		}
	}
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
		Timeout:         server.Timeout,
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
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not obtain user home directory :%v", err)
		}

		knownHostFile := filepath.Join(home, ".ssh", "known_hosts")
		log.Debugf("known_hosts file used: %s", knownHostFile)

		clb, err = knownhosts.New(knownHostFile)
		if err != nil {
			return nil, fmt.Errorf("error while parsing 'known_hosts' file: %s: %v", knownHostFile, err)
		}
	}

	return clb, nil
}

func reconcile(precident, subsequent string) string {
	if precident != "" {
		return precident
	}

	return subsequent
}

func expandAddress(address string) string {
	if strings.HasPrefix(address, ":") {
		return fmt.Sprintf("127.0.0.1%s", address)
	}

	return address
}

// BuildSSHChannels normalizes the given set of local and remote addresses,
// combining them to build a set of ssh channel objects.
func BuildSSHChannels(serverName string, local, remote []string) ([]*SSHChannel, error) {
	// if not local and remote were given, try to find the addresses from the SSH
	// configuration file.
	if len(local) == 0 && len(remote) == 0 {
		lf, err := getLocalForward(serverName)
		if err != nil {
			return nil, err
		}

		local = []string{lf.Local}
		remote = []string{lf.Remote}
	} else {

		lSize := len(local)
		rSize := len(remote)

		if lSize > rSize {
			// if there are more local than remote addresses given, the additional
			// addresses must be removed.
			if rSize == 0 {
				return nil, fmt.Errorf(NoRemoteGiven)
			}

			local = local[0:rSize]
		} else if lSize < rSize {
			// if there are more remote than local addresses given, the missing local
			// addresses should be configured as localhost with random ports.
			nl := make([]string, rSize)

			for i, _ := range remote {
				if i < lSize {
					if local[i] != "" {
						nl[i] = local[i]
					} else {
						nl[i] = RandomPortAddress
					}
				} else {
					nl[i] = RandomPortAddress
				}
			}

			local = nl
		}
	}

	for i, addr := range local {
		local[i] = expandAddress(addr)
	}

	for i, addr := range remote {
		remote[i] = expandAddress(addr)
	}

	channels := make([]*SSHChannel, len(remote))
	for i, r := range remote {
		channels[i] = &SSHChannel{Local: local[i], Remote: r}
	}

	return channels, nil
}

func getLocalForward(serverName string) (*LocalForward, error) {
	cfg, err := NewSSHConfigFile()
	if err != nil {
		return nil, fmt.Errorf("error reading ssh configuration file: %v", err)
	}

	sh := cfg.Get(serverName)

	if sh.LocalForward == nil {
		return nil, fmt.Errorf("LocalForward could not be found or has invalid syntax for host %s", serverName)
	}

	return sh.LocalForward, nil
}
