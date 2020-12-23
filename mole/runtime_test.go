package mole_test

import (
	"strings"
	"testing"

	"github.com/davrodpin/mole/mole"

	"github.com/andreyvit/diff"
)

const expectedInstance string = `id = "id1"
tunnel-type = ""
verbose = false
insecure = false
detach = false
key = ""
keep-alive-interval = 0
connection-retries = 0
wait-and-retry = 0
ssh-agent = ""
timeout = 0
ssh-config = ""
rpc = false
rpc-address = ""

[server]
  user = ""
  host = ""
  port = ""`

const expectedMultipleInstances string = `[instances]
  [instances.id1]
    id = "id1"
    tunnel-type = ""
    verbose = false
    insecure = false
    detach = false
    key = ""
    keep-alive-interval = 0
    connection-retries = 0
    wait-and-retry = 0
    ssh-agent = ""
    timeout = 0
    ssh-config = ""
    rpc = false
    rpc-address = ""
    [instances.id1.server]
      user = ""
      host = ""
      port = ""
  [instances.id2]
    id = "id2"
    tunnel-type = ""
    verbose = false
    insecure = false
    detach = false
    key = ""
    keep-alive-interval = 0
    connection-retries = 0
    wait-and-retry = 0
    ssh-agent = ""
    timeout = 0
    ssh-config = ""
    rpc = false
    rpc-address = ""
    [instances.id2.server]
      user = ""
      host = ""
      port = ""`

func TestFormatRuntimeToML(t *testing.T) {
	instances := []mole.Runtime{
		mole.Runtime{Id: "id1"},
		mole.Runtime{Id: "id2"},
	}

	runtimes := mole.InstancesRuntime(instances)

	tests := []struct {
		formatter mole.Formatter
		expected  string
	}{
		{formatter: mole.Runtime{Id: "id1"}, expected: expectedInstance},
		{formatter: runtimes, expected: expectedMultipleInstances},
	}

	for _, test := range tests {
		out, err := test.formatter.Format("toml")

		if err != nil {
			t.Errorf(err.Error())
		}

		if a, e := strings.TrimSpace(out), strings.TrimSpace(test.expected); a != e {
			t.Errorf("Result not as expected:\n%v", diff.LineDiff(e, a))
		}
	}
}
