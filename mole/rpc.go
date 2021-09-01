package mole

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/davrodpin/mole/fsutils"
	"github.com/davrodpin/mole/rpc"
)

func init() {
	rpc.Register("show-instance", ShowRpc)
}

// ShowRpc is a rpc callback that returns runtime information about the mole client.
func ShowRpc(params interface{}) (json.RawMessage, error) {
	if cli == nil {
		return nil, fmt.Errorf("client configuration could not be found.")
	}

	runtime := cli.Runtime()

	cj, err := json.Marshal(runtime)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(cj), nil
}

// Rpc calls a remote procedure on another mole instance given its id or alias.
func Rpc(id, method string, params interface{}) (string, error) {
	d, err := fsutils.InstanceDir(id)
	if err != nil {
		return "", err
	}

	rf := filepath.Join(d.Dir, "rpc")

	addr, err := ioutil.ReadFile(rf)
	if err != nil {
		return "", err
	}

	resp, err := rpc.Call(context.Background(), string(addr), method, params)
	if err != nil {
		return "", err
	}

	r, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", err
	}

	return string(r), nil
}
