package tunnel_test

import (
	"log"

	"github.com/davrodpin/mole/tunnel"
)

// This example shows the basic usage of the package: define both the local and
// remote endpoints, the ssh server and then start the tunnel that will
// exchange data from the local address to the remote address through the
// established ssh channel.
func Example() {
	sshChan := &tunnel.SSHChannel{Local: "127.0.0.1:8080", Remote: "user@example.com:22"}

	// Initialize the SSH Server configuration providing all values so
	// tunnel.NewServer will not try to lookup any value using $HOME/.ssh/config
	server, err := tunnel.NewServer("user", "172.17.0.20:2222", "/home/user/.ssh/key", "")
	if err != nil {
		log.Fatalf("error processing server options: %v\n", err)
	}

	t, err := tunnel.New("local", server, []*tunnel.SSHChannel{sshChan})
	if err != nil {
		log.Fatalf("error creating tunnel: %v\n", err)
	}

	// Start the tunnel
	err = t.Start()
	if err != nil {
		log.Fatalf("error starting tunnel: %v\n", err)
	}
}
