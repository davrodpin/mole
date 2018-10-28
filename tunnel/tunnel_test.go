package tunnel

import (
	"encoding/binary"
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

var sshDir string
var keyPath string
var encryptedKeyPath string
var publicKeyPath string
var knownHostsPath string

func TestServerOptions(t *testing.T) {
	tests := []struct {
		user     string
		address  string
		key      string
		expected *Server
	}{
		{
			"mole_user",
			"172.17.0.10:2222",
			"testdata/.ssh/id_rsa",
			&Server{
				Name:    "172.17.0.10",
				Address: "172.17.0.10:2222",
				User:    "mole_user",
				Key:     "testdata/.ssh/id_rsa",
			},
		},
		{
			"",
			"test",
			"",
			&Server{
				Name:    "test",
				Address: "127.0.0.1:2222",
				User:    "mole_test",
				Key:     "testdata/.ssh/id_rsa",
			},
		},
		{
			"",
			"test.something",
			"",
			&Server{
				Name:    "test.something",
				Address: "172.17.0.1:2223",
				User:    "mole_test2",
				Key:     "testdata/.ssh/other_key",
			},
		},
		{
			"mole_user",
			"test:3333",
			"testdata/.ssh/other_key",
			&Server{
				Name:    "test",
				Address: "127.0.0.1:3333",
				User:    "mole_user",
				Key:     "testdata/.ssh/other_key",
			},
		},
	}

	for _, test := range tests {
		s, err := NewServer(test.user, test.address, test.key)
		if err != nil {
			t.Errorf("%v\n", err)
		}

		if !reflect.DeepEqual(test.expected, s) {
			t.Errorf("unexpected result : expected: %s, result: %s", test.expected, s)
		}
	}

}

func TestTunnelOptions(t *testing.T) {
	server := &Server{Name: "s"}
	tests := []struct {
		local    string
		server   *Server
		remote   string
		expected *Tunnel
	}{
		{
			"172.17.0.10:2222",
			server,
			"172.17.0.10:2222",
			&Tunnel{
				local:  "172.17.0.10:2222",
				server: server,
				remote: "172.17.0.10:2222",
			},
		},
		{
			"",
			server,
			"172.17.0.10:2222",
			&Tunnel{
				local:  "127.0.0.1:0",
				server: server,
				remote: "172.17.0.10:2222",
			},
		},
		{
			":8443",
			server,
			":443",
			&Tunnel{
				local:  "127.0.0.1:8443",
				server: server,
				remote: "127.0.0.1:443",
			},
		},
	}

	for _, test := range tests {
		tun := New(test.local, test.server, test.remote)

		if test.expected.local != tun.local {
			t.Errorf("unexpected local result : expected: %s, result: %s", test.expected, tun)
		}

		if test.expected.remote != tun.remote {
			t.Errorf("unexpected remote result : expected: %s, result: %s", test.expected, tun)
		}

		if !reflect.DeepEqual(test.expected.server, tun.server) {
			t.Errorf("unexpected server result : expected: %s, result: %s", test.expected, tun)
		}

	}

}

func TestTunnel(t *testing.T) {
	expected := "ABC"
	c := prepareTunnel(t)
	client := newHTTPClient(c)

	response, err := get(client, fmt.Sprintf("/%s", expected))
	if err != nil {
		t.Errorf("tunnel failed: %v", err)
	}

	if expected != response {
		t.Errorf("expected: %s, value: %s", expected, response)
	}
}

func TestRandomLocalPort(t *testing.T) {
	expected := "127.0.0.1:0"
	local := ""
	remote := "172.17.0.1:80"
	server, _ := NewServer("", "test", "")

	tun := New(local, server, remote)

	if tun.local != expected {
		t.Errorf("unexpected local endpoint: expected: %s, value: %s", expected, tun.local)
	}
}

func TestMain(m *testing.M) {
	prepareTestEnv()

	code := m.Run()

	os.RemoveAll(sshDir)

	os.Exit(code)
}

func TestLoadPrivateKey(t *testing.T) {
	b, err := ioutil.ReadFile(encryptedKeyPath)
	if err != nil {
		t.Errorf("error reading encrypted key file: %v", err)
	}

	ReadPassword = func(fd int) ([]byte, error) {
		passByte := []byte("password")
		return passByte, nil
	}
	p, err := loadPrivateKey(b)
	if p == nil {
		t.Errorf("Signer not received from private key: %v", err)
	}
	if err != nil {
		t.Errorf("Decoding encrypted key failed: %v", err)
	}
}

// prepareTunnel creates a Tunnel object making sure all infrastructure
// dependencies (ssh and http servers) are ready returning a connection that
// can be use to reach the remote http server through the tunnel.
func prepareTunnel(t *testing.T) net.Conn {
	sshAddr := createSSHServer(keyPath)
	generateKnownHosts(sshAddr.String(), publicKeyPath, knownHostsPath)
	s, _ := NewServer("mole", sshAddr.String(), "")

	httpAddr := createWebServer()
	httpIP := strings.Split(httpAddr.String(), ":")[0]
	httpPort := strings.Split(httpAddr.String(), ":")[1]
	r := fmt.Sprintf("%s:%s", httpIP, httpPort)

	tun := &Tunnel{server: s, remote: r, done: make(chan error)}

	c1, c2 := net.Pipe()

	go func(t *testing.T) {
		err := tun.forward(c1)
		if err != nil {
			t.Errorf("tunnel could not be started: %v", err)
		}
	}(t)

	return c2
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
}

// newHTTPClient create an http client that will always use the given
// connection to perform http requests.
func newHTTPClient(conn net.Conn) http.Client {
	tr := &http.Transport{
		Dial: func(network, address string) (net.Conn, error) {
			return conn, nil
		},
	}

	client := http.Client{
		Timeout:   1 * time.Second,
		Transport: tr,
	}

	return client
}

// get performs a http request using the given client appending the given
// resource to to a hard-coded URL.
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

// createWebServer spawns a new http server, listening on a random user port
// and providing a response identical to the resource provided by the request.
//
// Example: If the request URI is /this-is-a-test, the response will be
// this-is-a-test
func createWebServer() net.Addr {

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, r.URL.Path[1:])
	}
	http.HandleFunc("/", handler)
	l, _ := net.Listen("tcp", "127.0.0.1:0")

	go func(l net.Listener) {
		http.Serve(l, nil)
	}(l)

	return l.Addr()
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
func createSSHServer(keyPath string) net.Addr {
	conf := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{}, nil
		},
	}

	b, _ := ioutil.ReadFile(keyPath)
	p, _ := ssh.ParsePrivateKey(b)
	conf.AddHostKey(p)

	l, _ := net.Listen("tcp", "127.0.0.1:0")

	go func(l net.Listener) {
		for {
			conn, _ := l.Accept()
			_, chans, reqs, _ := ssh.NewServerConn(conn, conf)

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

	return l.Addr()
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
