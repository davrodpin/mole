package tunnel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const NoSshRetries = -1

var sshDir string
var keyPath string
var encryptedKeyPath string
var publicKeyPath string
var knownHostsPath string
var configPath string

func TestServerOptions(t *testing.T) {
	k1, _ := NewPemKey("testdata/.ssh/id_rsa", "")
	k2, _ := NewPemKey("testdata/.ssh/other_key", "")

	tests := []struct {
		user          string
		address       string
		key           string
		config        string
		expected      *Server
		expectedError error
	}{
		{
			"mole_user",
			"172.17.0.10:2222",
			"testdata/.ssh/id_rsa",
			"testdata/.ssh/config",
			&Server{
				Name:    "172.17.0.10",
				Address: "172.17.0.10:2222",
				User:    "mole_user",
				Key:     k1,
			},
			nil,
		},
		{
			"",
			"test",
			"",
			"testdata/.ssh/config",
			&Server{
				Name:    "test",
				Address: "127.0.0.1:2222",
				User:    "mole_test",
				Key:     k1,
			},
			nil,
		},
		{
			"",
			"test.something",
			"",
			"testdata/.ssh/config",
			&Server{
				Name:    "test.something",
				Address: "172.17.0.1:2223",
				User:    "mole_test2",
				Key:     k2,
			},
			nil,
		},
		{
			"mole_user",
			"test:3333",
			"testdata/.ssh/other_key",
			"testdata/.ssh/config",
			&Server{
				Name:    "test",
				Address: "127.0.0.1:3333",
				User:    "mole_user",
				Key:     k2,
			},
			nil,
		},
		{
			"",
			"",
			"",
			"testdata/.ssh/config",
			nil,
			errors.New(HostMissing),
		},
	}

	for _, test := range tests {
		s, err := NewServer(test.user, test.address, test.key, "", test.config)
		if err != nil {
			if test.expectedError != nil {
				if test.expectedError.Error() != err.Error() {
					t.Errorf("error '%v' was expected, but got '%v'", test.expectedError, err)
				}
			} else {
				t.Errorf("%v\n", err)
			}
		}

		if !reflect.DeepEqual(test.expected, s) {
			t.Errorf("unexpected result : expected: %s, result: %s", test.expected, s)
		}
	}
}

func TestLocalTunnel(t *testing.T) {
	c := &tunnelConfig{t, "local", 1, false, NoSshRetries}
	tun, _, _ := prepareTunnel(c)

	select {
	case <-tun.Ready:
		t.Log("tunnel is ready to accept connections")
	case <-time.After(1 * time.Second):
		t.Errorf("error waiting for tunnel to be ready")
		return
	}

	err := validateTunnelConnectivity(t, "ABC", tun)
	if err != nil {
		t.Errorf("%v", err)
	}

	tun.Stop()
}

func TestRemoteTunnel(t *testing.T) {
	c := &tunnelConfig{t, "remote", 1, true, NoSshRetries}
	tun, _, _ := prepareTunnel(c)

	select {
	case <-tun.Ready:
		t.Log("tunnel is ready to accept connections")
	case <-time.After(1 * time.Second):
		t.Errorf("error waiting for tunnel to be ready")
		return
	}

	err := validateTunnelConnectivity(t, "ABC", tun)
	if err != nil {
		t.Errorf("%v", err)
	}

	tun.Stop()
}

func TestTunnelInsecure(t *testing.T) {
	c := &tunnelConfig{t, "local", 1, true, NoSshRetries}
	tun, _, _ := prepareTunnel(c)

	select {
	case <-tun.Ready:
		t.Log("tunnel is ready to accept connections")
	case <-time.After(1 * time.Second):
		t.Errorf("error waiting for tunnel to be ready")
		return
	}

	err := validateTunnelConnectivity(t, "ABC", tun)
	if err != nil {
		t.Errorf("%v", err)
	}

	tun.Stop()
}

func TestTunnelMultipleDestinations(t *testing.T) {
	c := &tunnelConfig{t, "local", 2, false, NoSshRetries}
	tun, _, _ := prepareTunnel(c)

	select {
	case <-tun.Ready:
		t.Log("tunnel is ready to accept connections")
	case <-time.After(1 * time.Second):
		t.Errorf("error waiting for tunnel to be ready")
		return
	}

	err := validateTunnelConnectivity(t, "ABC", tun)
	if err != nil {
		t.Errorf("%v", err)
	}

	tun.Stop()
}

func TestReconnectSSHServer(t *testing.T) {
	c := &tunnelConfig{t, "local", 1, false, 3}
	tun, ssh, _ := prepareTunnel(c)

	select {
	case <-tun.Ready:
		t.Log("tunnel is ready to accept connections")
	case <-time.After(1 * time.Second):
		t.Errorf("error waiting for tunnel to be ready")
		return
	}

	err := validateTunnelConnectivity(t, "ABC", tun)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	ssh.Close()

	// http request should fail since ssh server is not running
	err = validateTunnelConnectivity(t, "DEF", tun)
	if err == nil {
		t.Errorf("%v", err)
		return
	}

	_, err = createSSHServer(t, ssh.Addr().String(), keyPath)
	if err != nil {
		t.Errorf("error while recreating ssh server: %s", err)
		return
	}

	select {
	case <-tun.Ready:
		t.Log("tunnel is ready to accept connections")
	case <-time.After(10 * time.Second): // this is the maximum timeout based on the retries attempts
		t.Errorf("error waiting for tunnel to be ready")
		return
	}

	err = validateTunnelConnectivity(t, "GHJ", tun)
	if err != nil {
		t.Errorf("%v", err)
	}

	tun.Stop()
}

func validateTunnelConnectivity(t *testing.T, expected string, tun *Tunnel) error {
	for _, sshChan := range tun.channels {
		url := fmt.Sprintf("http://%s/%s", sshChan.listener.Addr(), expected)
		timeout := time.Duration(500 * time.Millisecond)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get(url)
		if err != nil {
			return fmt.Errorf("error while making http request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		response := string(body)

		if expected != response {
			return fmt.Errorf("expected: %s, value: %s", expected, response)
		}
	}

	return nil
}

func TestMain(m *testing.M) {
	err := prepareTestEnv()
	if err != nil {
		fmt.Printf("could not start test suite: %v\n", err)
		os.RemoveAll(sshDir)
		os.Exit(1)
	}

	code := m.Run()

	os.RemoveAll(sshDir)

	os.Exit(code)
}

func TestBuildSSHChannels(t *testing.T) {
	tests := []struct {
		serverName    string
		source        []string
		destination   []string
		config        string
		expected      int
		expectedError error
	}{
		{
			serverName:    "test",
			source:        []string{":3360"},
			destination:   []string{":3360"},
			config:        "testdata/.ssh/config",
			expected:      1,
			expectedError: nil,
		},
		{
			serverName:    "test",
			source:        []string{":3360", ":8080"},
			destination:   []string{":3360", ":8080"},
			config:        "testdata/.ssh/config",
			expected:      2,
			expectedError: nil,
		},
		{
			serverName:    "test",
			source:        []string{},
			destination:   []string{":3360"},
			config:        "testdata/.ssh/config",
			expected:      1,
			expectedError: nil,
		},
		{
			serverName:    "test",
			source:        []string{":3360"},
			destination:   []string{":3360", ":8080"},
			config:        "testdata/.ssh/config",
			expected:      2,
			expectedError: nil,
		},
		{
			serverName:    "hostWithLocalForward",
			source:        []string{},
			destination:   []string{},
			config:        "testdata/.ssh/config",
			expected:      1,
			expectedError: nil,
		},
		{
			serverName:    "test",
			source:        []string{":3360", ":8080"},
			destination:   []string{":3360"},
			config:        "testdata/.ssh/config",
			expected:      1,
			expectedError: nil,
		},
		{
			serverName:    "test",
			source:        []string{":3360"},
			destination:   []string{},
			config:        "testdata/.ssh/config",
			expected:      0,
			expectedError: fmt.Errorf(NoDestinationGiven),
		},
	}

	for testId, test := range tests {
		sshChannels, err := buildSSHChannels(test.serverName, "local", test.source, test.destination, test.config)
		if err != nil {
			if test.expectedError != nil {
				if test.expectedError.Error() != err.Error() {
					t.Errorf("error '%v' was expected, but got '%v'", test.expectedError, err)
				}
			} else {
				t.Errorf("unable to build ssh channels objects for test %d: %v", testId, err)
			}
		}

		if test.expected != len(sshChannels) {
			t.Errorf("wrong number of ssh channel objects created for test %d: expected: %d, value: %d", testId, test.expected, len(sshChannels))
		}

		sourceSize := len(test.source)
		destinationSize := len(test.destination)

		// check if the source addresses match only if any address is given
		if sourceSize > 0 && destinationSize > 0 {
			for i, sshChannel := range sshChannels {
				source := ""
				if i < sourceSize {
					source = test.source[i]
				} else {
					source = RandomPortAddress
				}

				source = expandAddress(source)

				if sshChannel.Source != source {
					t.Errorf("source address don't match for test %d: expected: %s, value: %s", testId, sshChannel.Source, source)
				}

			}
		}
	}
}

type tunnelConfig struct {
	T          *testing.T
	TunnelType string

	// Destinations indicates how many endpoints should be available through the
	// tunnel.
	Destinations int

	Insecure          bool
	ConnectionRetries int
}

// prepareTunnel creates a Tunnel object making sure all infrastructure
// dependencies (ssh and http servers) are ready.
//
// The 'remotes' argument tells how many remote endpoints will be available
// through the tunnel.
func prepareTunnel(config *tunnelConfig) (tun *Tunnel, ssh net.Listener, hss []*http.Server) {
	hss = make([]*http.Server, config.Destinations)

	ssh, err := createSSHServer(config.T, "", keyPath)
	if err != nil {
		config.T.Errorf("error while creating ssh server: %s", err)
		return
	}

	srv, _ := NewServer("mole", ssh.Addr().String(), "", "", "testdata/.ssh/config")

	srv.Insecure = config.Insecure

	if !config.Insecure {
		err = generateKnownHosts(ssh.Addr(), publicKeyPath, knownHostsPath)
		if err != nil {
			config.T.Errorf("error generating known hosts file for tests: %v\n", err)
			return
		}

	}

	source := make([]string, config.Destinations)
	destination := make([]string, config.Destinations)

	for i := 0; i <= (config.Destinations - 1); i++ {
		l, hs := createHttpServer()
		if config.TunnelType == "local" {
			source[i] = "127.0.0.1:0"
			destination[i] = l.Addr().String()
		} else if config.TunnelType == "remote" {
			source[i] = l.Addr().String()
			destination[i] = "127.0.0.1:0"
		} else {
			config.T.Errorf("could not configure destination endpoints for testing: %v\n", err)
			return
		}
		hss = append(hss, hs)
	}

	tun, _ = New(config.TunnelType, srv, source, destination, configPath)
	tun.ConnectionRetries = config.ConnectionRetries
	tun.WaitAndRetry = 3 * time.Second
	tun.KeepAliveInterval = 10 * time.Second

	go func(tun *Tunnel) {
		err := tun.Start()
		// FIXME: this message should be shown through *testing.t but using it here
		// would cause the message to be printed after the test ends (goroutine),
		// making the test to fail
		if err != nil {
			fmt.Printf("error returned from tunnel start: %v", err)
		}
	}(tun)

	return tun, ssh, hss
}

func prepareTestEnv() error {
	home := "testdata"
	fixtureDir := filepath.Join(home, "dotssh")
	testDir := filepath.Join(home, ".ssh")

	keyPath = filepath.Join(testDir, "id_rsa")
	encryptedKeyPath = filepath.Join(testDir, "id_rsa_encrypted")
	publicKeyPath = filepath.Join(testDir, "id_rsa.pub")
	knownHostsPath = filepath.Join(testDir, "known_hosts")
	configPath = filepath.Join(testDir, "config")
	sshDir = testDir

	fixtures := []map[string]string{
		{
			"from": filepath.Join(fixtureDir, "config"),
			"to":   filepath.Join(testDir, "config"),
		},
		{
			"from": filepath.Join(fixtureDir, "id_rsa.pub"),
			"to":   publicKeyPath,
		},
		{
			"from": filepath.Join(fixtureDir, "id_rsa"),
			"to":   keyPath,
		},
		{
			"from": filepath.Join(fixtureDir, "id_rsa"),
			"to":   filepath.Join(testDir, "other_key"),
		},
		{
			"from": filepath.Join(fixtureDir, "id_rsa_encrypted"),
			"to":   filepath.Join(testDir, "id_rsa_encrypted"),
		},
	}

	err := os.Mkdir(testDir, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	for _, f := range fixtures {
		err = os.Link(f["from"], f["to"])
		if err != nil {
			return err
		}
	}

	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home)

	return nil
}

// createHttpServer spawns a new http server, listening on a random port.
// The http server provided an endpoint, /XXX, that will respond, in plain
// text, with the very same given string.
//
// Example: If the request URI is /this-is-a-test, the response will be
// this-is-a-test
func createHttpServer() (net.Listener, *http.Server) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, r.URL.Path[1:])
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := &http.Server{
		Handler: mux,
	}

	l, _ := net.Listen("tcp", "127.0.0.1:0")

	go server.Serve(l)

	return l, server
}

// createSSHServer starts a SSH server that authenticates connections using
// the given keyPath, listens on a random user port and returns the SSH Server
// address.
//
// The SSH Server created by this function only responds to "direct-tcpip",
// which is used to establish local port forwarding.
//
// References:
// https://gist.github.com/jpillora/b480fde82bff51a06238
// https://tools.ietf.org/html/rfc4254#section-7.2
func createSSHServer(t *testing.T, address string, keyPath string) (net.Listener, error) {
	conf := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{}, nil
		},
	}

	b, _ := ioutil.ReadFile(keyPath)
	p, _ := ssh.ParsePrivateKey(b)
	conf.AddHostKey(p)

	if address == "" {
		address = "127.0.0.1:0"
	}

	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("error while creating listener: %s", err)
	}

	go func(listener net.Listener) {
		var conns []ssh.Conn
		for {
			var err error

			conn, err := listener.Accept()
			if err != nil {
				// closing all ssh connections if a new client can't connect to the server
				for _, sc := range conns {
					sc.Close()
				}
				break
			}

			serverConn, chans, reqs, _ := ssh.NewServerConn(conn, conf)
			conns = append(conns, serverConn)

			// go routine to handle ssh client requests. In the context of mole's test,
			// this is needed when a remote ssh forwarding listens to a port on the jump
			// server and the port needs to be randomized (port is given as 0).
			// The reply's needs to carry the port to be listened in its payload.
			// All requests but "tcpip-forward" are discarded.
			go func(reqs <-chan *ssh.Request) {
				var err error

				for newReq := range reqs {
					if newReq.Type != "tcpip-forward" {
						err = newReq.Reply(false, nil)
						if err != nil {
							t.Errorf("error replying to tcpip-forward request: %v", err)
						}
						return
					}

					if newReq.WantReply {
						ports, err := freeport.GetFreePorts(1)
						if err != nil {
							t.Errorf("could not get a free port: %v", err)
							return
						}
						port := make([]byte, 4)
						binary.BigEndian.PutUint32(port, uint32(ports[0]))
						err = newReq.Reply(true, port)
						if err != nil {
							t.Errorf("error replying to tcpip-forward request: %v", err)
							return
						}
					}
				}
			}(reqs)

			// go routine to handle requests to create new ssh channels. This particular
			// implementation only supports "direct-tcpip", which is the identifier used
			// for ssh port forwarding.
			go func(chans <-chan ssh.NewChannel) {
				for newChan := range chans {
					go func(newChan ssh.NewChannel) {
						var err error

						if ct := newChan.ChannelType(); ct != "direct-tcpip" {
							err = newChan.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", ct))
							if err != nil {
								t.Errorf("error rejecting unsupported channel: %v", err)
							}
							return
						}

						payload := newChan.ExtraData()
						pad := byte(4)
						l := payload[3]
						remoteIP := string(payload[pad : pad+l])
						remotePort := binary.BigEndian.Uint32(payload[pad+l : pad+l+4])

						conn, _, _ := newChan.Accept()
						remoteConn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", remoteIP, remotePort))

						go func() {
							io.Copy(conn, remoteConn)
						}()

						go func() {
							io.Copy(remoteConn, conn)
						}()
					}(newChan)
				}
			}(chans)
		}
	}(l)

	return l, nil
}

// generateKnownHosts creates a new "known_hosts" file on a given path with a
// single entry based on the given SSH server address and public key.
func generateKnownHosts(sshAddr net.Addr, pubKeyPath, knownHostsPath string) error {
	d, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return err
	}

	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(d))
	if err != nil {
		return err
	}

	l := knownhosts.Line([]string{sshAddr.String()}, pk)
	ioutil.WriteFile(knownHostsPath, []byte(l), 0600)

	return nil
}
