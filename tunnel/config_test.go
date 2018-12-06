package tunnel

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kevinburke/ssh_config"
)

func TestSSHConfigFile(t *testing.T) {

	var config = `
Host example1
  Hostname 172.17.0.1
	Port 3306
	User john
	IdentityFile /path/.ssh/id_rsa
Host example2
	LocalForward 8080 127.0.0.1:8080
Host example3
	LocalForward 9090 127.0.0.1:9090
`

	c, _ := ssh_config.Decode(strings.NewReader(config))
	cfg := &SSHConfigFile{sshConfig: c}

	tests := []struct {
		host     string
		expected *SSHHost
	}{
		{
			"example1",
			&SSHHost{
				Hostname:     "172.17.0.1",
				Port:         "3306",
				User:         "john",
				Key:          "/path/.ssh/id_rsa",
				LocalForward: &LocalForward{Local: "", Remote: ""},
			},
		},
		{
			"example2",
			&SSHHost{
				Hostname:     "",
				Port:         "",
				User:         "",
				Key:          "",
				LocalForward: &LocalForward{Local: "127.0.0.1:8080", Remote: "127.0.0.1:8080"},
			},
		},
		{
			"example3",
			&SSHHost{
				Hostname:     "",
				Port:         "",
				User:         "",
				Key:          "",
				LocalForward: &LocalForward{Local: "127.0.0.1:9090", Remote: "127.0.0.1:9090"},
			},
		},
	}

	var value *SSHHost
	for _, test := range tests {
		value = cfg.Get(test.host)

		if !reflect.DeepEqual(test.expected, value) {
			t.Errorf("unexpected result for %s:\n\texpected: %s\n\tvalue   : %s", test.host, test.expected, value)
		}
	}
}
