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

	"golang.org/x/crypto/ssh/agent"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	HostMissing        = "server host has to be provided as part of the server address"
	RandomPortAddress  = "127.0.0.1:0"
	NoDestinationGiven = "cannot create a tunnel without at least one remote address"
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
	// SSHAgent is the path to the unix socket where an ssh agent is listening
	SSHAgent string
}

// NewServer creates a new instance of Server using $HOME/.ssh/config to
// resolve the missing connection attributes (e.g. user, hostname, port, key
// and ssh agent) required to connect to the remote server, if any.
func NewServer(user, address, key, sshAgent string) (*Server, error) {
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
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error accessing %s: %v", host, err)
		}
	}

	// If ssh config file doesnt exists, create an empty ssh config struct to avoid nil pointer deference
	if errors.Is(err, os.ErrNotExist) {
		c = NewEmptySSHConfigStruct()
	}

	h := c.Get(host)
	hostname = reconcile(h.Hostname, host)
	port = reconcile(port, h.Port)
	user = reconcile(user, h.User)
	key = reconcile(key, h.Key)
	sshAgent = reconcile(sshAgent, h.IdentityAgent)

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

	if strings.HasPrefix(sshAgent, "$") {
		sshAgent = os.Getenv(sshAgent[1:])
	}

	return &Server{
		Name:     host,
		Address:  fmt.Sprintf("%s:%s", hostname, port),
		User:     user,
		Key:      pk,
		SSHAgent: sshAgent,
	}, nil
}

// String provided a string representation of a Server.
func (s Server) String() string {
	return fmt.Sprintf("[name=%s, address=%s, user=%s]", s.Name, s.Address, s.User)
}

type SSHChannel struct {
	ChannelType string
	Source      string
	Destination string
	listener    net.Listener
	conn        net.Conn
}

// Listen creates tcp listeners for each channel defined.
func (ch *SSHChannel) Listen(serverClient *ssh.Client) error {
	var l net.Listener
	var err error

	if ch.listener == nil {
		if ch.ChannelType == "local" {
			l, err = net.Listen("tcp", ch.Source)
		} else if ch.ChannelType == "remote" {
			l, err = serverClient.Listen("tcp", ch.Source)
		} else {
			return fmt.Errorf("channel can't listen on endpoint: unknown channel type %s", ch.ChannelType)
		}

		if err != nil {
			return err
		}

		ch.listener = l

		// update the endpoint value with assigned port for the cases where the user
		// haven't explicitily specified one
		ch.Source = l.Addr().String()
	}

	return nil
}

// Accept waits for and return the next connection to the SSHChannel.
func (ch *SSHChannel) Accept() error {
	var err error

	if ch.conn, err = ch.listener.Accept(); err != nil {
		return fmt.Errorf("error while establishing connection: %v", err)
	}

	return nil
}

// String returns a string representation of a SSHChannel
func (ch SSHChannel) String() string {
	return fmt.Sprintf("[source=%s, destination=%s]", ch.Source, ch.Destination)
}

// Tunnel represents the ssh tunnel and the channels connecting local and
// remote endpoints.
type Tunnel struct {
	// Type tells what kind of port forwarding this tunnel will handle: local or remote
	Type string

	// Ready tells when the Tunnel is ready to accept connections
	Ready chan bool

	// KeepAliveInterval is the time period used to send keep alive packets to
	// the remote ssh server
	KeepAliveInterval time.Duration

	// ConnectionRetries is the number os attempts to reconnect to the ssh server
	// when the current connection fails
	ConnectionRetries int

	// WaitAndRetry is the time waited before trying to reconnect to the ssh
	// server
	WaitAndRetry time.Duration

	server        *Server
	channels      []*SSHChannel
	done          chan error
	client        *ssh.Client
	stopKeepAlive chan bool
	reconnect     chan error
}

// New creates a new instance of Tunnel.
func New(tunnelType string, server *Server, source, destination []string) (*Tunnel, error) {
	var channels []*SSHChannel
	var err error

	channels, err = buildSSHChannels(server.Name, tunnelType, source, destination)
	if err != nil {
		return nil, err
	}

	for _, channel := range channels {
		if channel.Source == "" || channel.Destination == "" {
			return nil, fmt.Errorf("invalid ssh channel: source=%s, destination=%s", channel.Source, channel.Destination)
		}
	}

	return &Tunnel{
		Type:          tunnelType,
		Ready:         make(chan bool, 1),
		channels:      channels,
		server:        server,
		reconnect:     make(chan error, 1),
		done:          make(chan error, 1),
		stopKeepAlive: make(chan bool, 1),
	}, nil
}

// Start creates the ssh tunnel and initialized all channels allowing data
// exchange between local and remote enpoints.
func (t *Tunnel) Start() error {
	log.Debugf("tunnel: %s", t)

	t.connect()

	for {
		select {
		case err := <-t.reconnect:
			if err != nil {
				log.WithError(err).Warnf("reconnecting to ssh server")

				t.stopKeepAlive <- true
				t.client.Close()

				log.Debugf("restablishing the tunnel after disconnection: %s", t)

				t.connect()
			}
		case err := <-t.done:
			if t.client != nil {
				t.stopKeepAlive <- true
				t.client.Close()
			}

			return err
		}
	}
}

// Listen creates tcp listeners for each channel defined.
func (t *Tunnel) Listen() error {
	for _, ch := range t.channels {
		if err := ch.Listen(t.client); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tunnel) startChannel(channel *SSHChannel) error {
	var err error

	err = channel.Accept()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"channel": channel,
	}).Debug("connection established")

	if t.client == nil {
		return fmt.Errorf("tunnel channel can't be established: missing connection to the ssh server")
	}

	var destinationConn net.Conn

	if t.Type == "local" {
		destinationConn, err = t.client.Dial("tcp", channel.Destination)
	} else if t.Type == "remote" {
		destinationConn, err = net.Dial("tcp", channel.Destination)
	} else {
		return fmt.Errorf("unknown tunnel type %s", t.Type)
	}

	if err != nil {
		return fmt.Errorf("dial error: %s", err)
	}

	go copyConn(channel.conn, destinationConn)
	go copyConn(destinationConn, channel.conn)

	log.WithFields(log.Fields{
		"channel": channel,
		"server":  t.server,
	}).Debug("tunnel channel has been established")

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
		t.client.Close()
	}

	c, err := sshClientConfig(*t.server)
	if err != nil {
		return fmt.Errorf("error generating ssh client config: %s", err)
	}

	retries := 0
	for {
		if t.ConnectionRetries > 0 && retries == t.ConnectionRetries {
			log.WithFields(log.Fields{
				"server":  t.server,
				"retries": retries,
			}).Error("maximum number of connection retries to the ssh server reached")

			return fmt.Errorf("error while connecting to ssh server")
		}

		t.client, err = ssh.Dial("tcp", t.server.Address, c)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"server":  t.server,
				"retries": retries,
			}).Error("error while connecting to ssh server")

			if t.ConnectionRetries < 0 {
				break
			}

			retries = retries + 1

			time.Sleep(t.WaitAndRetry)
			continue
		}

		break
	}

	go t.keepAlive()

	if t.ConnectionRetries > 0 {
		go t.waitAndReconnect()
	}

	log.WithFields(log.Fields{
		"server": t.server,
	}).Debug("connection to the ssh server is established")

	return nil
}

func (t *Tunnel) waitAndReconnect() {
	t.reconnect <- t.client.Wait()
}

func (t *Tunnel) connect() {
	var err error

	err = t.dial()
	if err != nil {
		t.done <- err
		return
	}

	err = t.Listen()
	if err != nil {
		t.done <- err
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(t.channels))

	// wait for all ssh channels to be ready to accept connections then sends a
	// single message signalling all tunnels are ready
	go func(tunnel *Tunnel, waitgroup *sync.WaitGroup) {
		waitgroup.Wait()
		t.Ready <- true
	}(t, wg)

	for _, ch := range t.channels {
		go func(channel *SSHChannel, waitgroup *sync.WaitGroup) {
			var err error
			var once sync.Once

			for {
				once.Do(func() {
					log.WithFields(log.Fields{
						"source":      channel.Source,
						"destination": channel.Destination,
					}).Info("tunnel channel is waiting for connection")

					waitgroup.Done()
				})

				err = t.startChannel(channel)
				if err != nil {
					t.done <- err
					return
				}
			}
		}(ch, wg)
	}

}

func (t *Tunnel) keepAlive() {
	ticker := time.NewTicker(t.KeepAliveInterval)

	log.Debug("start sending keep alive packets")

	for {
		select {
		case <-ticker.C:
			_, _, err := t.client.SendRequest("keepalive@mole", true, nil)
			if err != nil {
				log.Warnf("error sending keep-alive request to ssh server: %v", err)
			}
		case <-t.stopKeepAlive:
			log.Debug("stop sending keep alive packets")
			return
		}
	}
}

func sshClientConfig(server Server) (*ssh.ClientConfig, error) {
	var signers []ssh.Signer

	signer, err := server.Key.Parse()
	if err != nil {
		return nil, err
	}
	signers = append(signers, signer)

	if server.SSHAgent != "" {
		if _, err := os.Stat(server.SSHAgent); err == nil {
			agentSigners, err := getAgentSigners(server.SSHAgent)
			if err != nil {
				return nil, err
			}
			signers = append(signers, agentSigners...)
		} else {
			log.WithError(err).Warnf("%s cannot be read. Will not try to talk to ssh agent", server.SSHAgent)
		}
	}

	clb, err := knownHostsCallback(server.Insecure)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
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

func getAgentSigners(addr string) ([]ssh.Signer, error) {
	log.Debugf("ssh agent address: %s", addr)
	conn, err := net.Dial("unix", addr)
	if err != nil {
		return nil, err
	}
	client := agent.NewClient(conn)
	return client.Signers()
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

func buildSSHChannels(serverName, channelType string, source, destination []string) ([]*SSHChannel, error) {
	// if source and destination were not given, try to find the addresses from the
	// SSH configuration file.
	if len(source) == 0 && len(destination) == 0 {
		f, err := getForward(channelType, serverName)
		if err != nil {
			return nil, err
		}

		source = []string{f.Source}
		destination = []string{f.Destination}
	} else {

		lSize := len(source)
		rSize := len(destination)

		if lSize > rSize {
			// if there are more source than destination addresses given, the additional
			// addresses must be removed.
			if rSize == 0 {
				return nil, fmt.Errorf(NoDestinationGiven)
			}

			source = source[0:rSize]
		} else if lSize < rSize {
			// if there are more destination than source addresses given, the missing
			// source addresses should be configured as localhost with random ports.
			nl := make([]string, rSize)

			for i := range destination {
				if i < lSize {
					if source[i] != "" {
						nl[i] = source[i]
					} else {
						nl[i] = RandomPortAddress
					}
				} else {
					nl[i] = RandomPortAddress
				}
			}

			source = nl
		}
	}

	for i, addr := range source {
		source[i] = expandAddress(addr)
	}

	for i, addr := range destination {
		destination[i] = expandAddress(addr)
	}

	channels := make([]*SSHChannel, len(destination))
	for i, d := range destination {
		channels[i] = &SSHChannel{ChannelType: channelType, Source: source[i], Destination: d}
	}

	return channels, nil
}

func getForward(channelType, serverName string) (*ForwardConfig, error) {
	var f *ForwardConfig

	cfg, err := NewSSHConfigFile()
	if err != nil {
		return nil, fmt.Errorf("error reading ssh configuration file: %v", err)
	}

	sh := cfg.Get(serverName)

	if channelType == "local" {
		f = sh.LocalForward
	} else if channelType == "remote" {
		f = sh.RemoteForward
	} else {
		return nil, fmt.Errorf("could not retrieve forwarding information from ssh configuration file: unsupported channel type %s", channelType)
	}

	if f == nil {
		return nil, fmt.Errorf("forward config could not be found or has invalid syntax for host %s", serverName)
	}

	return f, nil
}
