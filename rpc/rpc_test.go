package rpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/davrodpin/mole/rpc"
)

var (
	addr net.Addr
)

func TestHandler(t *testing.T) {
	responseTemplate := `{"message":"%s"}`
	method := "test"
	paramValue := "param"
	expectedResponse := fmt.Sprintf(responseTemplate, paramValue)

	rpc.Register(method, func(params interface{}) (json.RawMessage, error) {
		var r []uint8
		var ok bool

		if r, ok = params.([]uint8); !ok {
			return nil, fmt.Errorf("invalid parameter")
		}

		m := fmt.Sprintf(responseTemplate, string(r[1:6]))

		return json.RawMessage(m), nil
	})

	response, err := rpc.Call(context.Background(), addr.String(), method, paramValue)
	if err != nil {
		t.Errorf("error while calling remote procedure: %v", err)
	}

	json, err := json.Marshal(response)
	if err != nil {
		t.Errorf("error while parsing response to string: response: %s, err: %v", response, err)
	}

	if expectedResponse != string(json) {
		t.Errorf("unexpected response for remote procedure call: want: %s, got: %s", expectedResponse, string(json))
	}
}

func TestMethodNotRegistered(t *testing.T) {
	method := "methodnotregistered"
	expectedResponse := fmt.Sprintf(`{"code":-32601,"data":null,"message":"method %s not found"}`, method)

	response, err := rpc.Call(context.Background(), addr.String(), method, "param")
	if err != nil {
		t.Errorf("error while calling remote procedure: %v", err)
	}

	json, err := json.Marshal(response)
	if err != nil {
		t.Errorf("error while parsing response to string: response: %s, err: %v", response, err)
	}

	if expectedResponse != string(json) {
		t.Errorf("unexpected response for remote procedure call: want: %s, got: %s", expectedResponse, string(json))
	}
}

func TestMethodWithError(t *testing.T) {
	method := "testwitherror"
	expectedResponse := fmt.Sprintf(`{"code":-32603,"data":null,"message":"error executing rpc method %s"}`, method)

	rpc.Register(method, func(params interface{}) (json.RawMessage, error) {
		return nil, fmt.Errorf("error")
	})

	response, err := rpc.Call(context.Background(), addr.String(), method, "param")
	if err != nil {
		t.Errorf("error while calling remote procedure: %v", err)
	}

	json, err := json.Marshal(response)
	if err != nil {
		t.Errorf("error while parsing response to string: response: %s, err: %v", response, err)
	}

	if expectedResponse != string(json) {
		t.Errorf("unexpected response for remote procedure call: want: %s, got: %s", expectedResponse, string(json))
	}
}

func TestMain(m *testing.M) {
	var err error

	addr, err = rpc.Start(rpc.DefaultAddress)
	if err != nil {
		fmt.Printf("error initializing rpc server: %v", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}
