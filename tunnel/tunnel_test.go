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
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

const NoSshRetries = -1

var sshDir string
var keyPath string
var encryptedKeyPath string
var publicKeyPath string
var knownHostsPath string

func TestServerOptions(t *testing.T) {
	k1, _ := NewPemKey("testdata/.ssh/id_rsa", "")
	k2, _ := NewPemKey("testdata/.ssh/other_key", "")

	tests := []struct {
		user          string
		address       string
		key           string
		expected      *Server
		expectedError error
	}{
		{
			"mole_user",
			"172.17.0.10:2222",
			"testdata/.ssh/id_rsa",
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
			nil,
			errors.New(HostMissing),
		},
	}

	for _, test := range tests {
		s, err := NewServer(test.user, test.address, test.key, "")
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

func TestTunnel(t *testing.T) {
	tun, _, _ := prepareTunnel(1, false, NoSshRetries)

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
	tun, _, _ := prepareTunnel(1, true, NoSshRetries)

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

func TestTunnelMultipleRemotes(t *testing.T) {
	tun, _, _ := prepareTunnel(2, false, NoSshRetries)

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
	tun, ssh, _ := prepareTunnel(1, false, 3)

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

	_, err = createSSHServer(ssh.Addr().String(), keyPath)
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
	prepareTestEnv()

	code := m.Run()

	os.RemoveAll(sshDir)

	os.Exit(code)
}

func TestBuildSSHChannels(t *testing.T) {
	tests := []struct {
		serverName    string
		local         []string
		remote        []string
		expected      int
		expectedError error
	}{
		{
			serverName:    "test",
			local:         []string{":3360"},
			remote:        []string{":3360"},
			expected:      1,
			expectedError: nil,
		},
		{
			serverName:    "test",
			local:         []string{":3360", ":8080"},
			remote:        []string{":3360", ":8080"},
			expected:      2,
			expectedError: nil,
		},
		{
			serverName:    "test",
			local:         []string{},
			remote:        []string{":3360"},
			expected:      1,
			expectedError: nil,
		},
		{
			serverName:    "test",
			local:         []string{":3360"},
			remote:        []string{":3360", ":8080"},
			expected:      2,
			expectedError: nil,
		},
		{
			serverName:    "hostWithLocalForward",
			local:         []string{},
			remote:        []string{},
			expected:      1,
			expectedError: nil,
		},
		{
			serverName:    "test",
			local:         []string{":3360", ":8080"},
			remote:        []string{":3360"},
			expected:      1,
			expectedError: nil,
		},
		{
			serverName:    "test",
			local:         []string{":3360"},
			remote:        []string{},
			expected:      0,
			expectedError: fmt.Errorf(NoRemoteGiven),
		},
	}

	for testId, test := range tests {
		sshChannels, err := BuildSSHChannels(test.serverName, test.local, test.remote)
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

		localSize := len(test.local)
		remoteSize := len(test.remote)

		// check if the local addresses match only if any address is given
		if localSize > 0 && remoteSize > 0 {
			for i, sshChannel := range sshChannels {
				local := ""
				if i < localSize {
					local = test.local[i]
				} else {
					local = RandomPortAddress
				}

				local = expandAddress(local)

				if sshChannel.Local != local {
					t.Errorf("local address don't match for test %d: expected: %s, value: %s", testId, sshChannel.Local, local)
				}

			}
		}
	}
}

// prepareTunnel creates a Tunnel object making sure all infrastructure
// dependencies (ssh and http servers) are ready.
//
// The 'remotes' argument tells how many remote endpoints will be available
// through the tunnel.
func prepareTunnel(remotes int, insecure bool, sshConnectionRetries int) (tun *Tunnel, ssh net.Listener, hss []*http.Server) {
	hss = make([]*http.Server, remotes)

	ssh, err := createSSHServer("", keyPath)
	if err != nil {
		// FIXME: return the error
		fmt.Printf("error while creating ssh server: %s", err)
		return
	}

	srv, _ := NewServer("mole", ssh.Addr().String(), "", "")

	srv.Insecure = insecure

	if !insecure {
		generateKnownHosts(ssh.Addr().String(), publicKeyPath, knownHostsPath)
	}

	sshChannels := []*SSHChannel{}
	for i := 1; i <= remotes; i++ {
		l, hs := createHttpServer()
		sshChannels = append(sshChannels, &SSHChannel{Local: "127.0.0.1:0", Remote: l.Addr().String()})
		hss = append(hss, hs)
	}

	tun, _ = New("local", srv, sshChannels)
	tun.ConnectionRetries = sshConnectionRetries
	tun.WaitAndRetry = 3 * time.Second
	tun.KeepAliveInterval = 10 * time.Second

	go func(tun *Tunnel) {
		var err error
		err = tun.Start()
		// FIXME: this message should be shown through *testing.t but using it here
		// would cause the message to be printed after the test ends (goroutine),
		// making the test to fail
		if err != nil {
			fmt.Printf("error returned from tunnel start: %v", err)
		}
	}(tun)

	return tun, ssh, hss
}

func prepareTestEnv() {
	home := "testdata"
	fixtureDir := filepath.Join(home, "dotssh")
	testDir := filepath.Join(home, ".ssh")

	keyPath = filepath.Join(testDir, "id_rsa")
	encryptedKeyPath = filepath.Join(testDir, "id_rsa_encrypted")
	publicKeyPath = filepath.Join(testDir, "id_rsa.pub")
	knownHostsPath = filepath.Join(testDir, "known_hosts")
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

	os.Mkdir(testDir, os.ModeDir|os.ModePerm)

	for _, f := range fixtures {
		os.Link(f["from"], f["to"])
	}

	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home)
}

// get performs a http request using the given client appending the given
// resource to a hard-coded URL.
//
// The request performed by this function is designed to reach the other side
// through a pipe (net.Pipe()) and this is the reason the URL is hard-coded.
func get(client http.Client, resource string) (string, error) {
	resp, err := client.Get(fmt.Sprintf("%s%s", "http://any-url-is.fine", resource))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body), nil
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
func createSSHServer(address string, keyPath string) (net.Listener, error) {
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

			go ssh.DiscardRequests(reqs)

			go func(chans <-chan ssh.NewChannel) {
				for newChan := range chans {
					go func(newChan ssh.NewChannel) {

						if t := newChan.ChannelType(); t != "direct-tcpip" {
							newChan.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
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
func generateKnownHosts(sshAddr, pubKeyPath, knownHostsPath string) {
	i := strings.Split(sshAddr, ":")[0]
	p := strings.Split(sshAddr, ":")[1]

	kc, _ := ioutil.ReadFile(pubKeyPath)
	t := strings.Split(string(kc), " ")[0]
	k := strings.Split(string(kc), " ")[1]

	c := fmt.Sprintf("[%s]:%s %s %s", i, p, t, k)
	ioutil.WriteFile(knownHostsPath, []byte(c), 0600)
}

type MockConn struct {
	isConnectionOpen bool
}

func (c MockConn) User() string          { return "" }
func (c MockConn) SessionID() []byte     { return []byte{} }
func (c MockConn) ClientVersion() []byte { return []byte{} }
func (c MockConn) ServerVersion() []byte { return []byte{} }
func (c MockConn) RemoteAddr() net.Addr  { return nil }
func (c MockConn) LocalAddr() net.Addr   { return nil }
func (c MockConn) SendRequest(name string, wantReply bool, payload []byte) (bool, []byte, error) {
	return false, []byte{}, nil
}
func (c MockConn) OpenChannel(name string, data []byte) (ssh.Channel, <-chan *ssh.Request, error) {
	return nil, nil, nil
}
func (c MockConn) Close() error { return nil }
func (c MockConn) Wait() error  { return nil }
