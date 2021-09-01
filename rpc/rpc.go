package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/sourcegraph/jsonrpc2"
)

var registeredMethods = sync.Map{}

const (
	// DefaultAddress is the network address used by the rpc server if none is given.
	DefaultAddress = "127.0.0.1:0"
)

// Start initializes the jsonrpc 2.0 server which will be waiting for
// connections on a random port.
func Start(address string) (net.Addr, error) {
	var err error

	if address == "" {
		address = DefaultAddress
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	h := &Handler{}

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				log.WithError(err).Warnf("error establishing connection with rpc client.")
			}
			stream := jsonrpc2.NewBufferedStream(conn, jsonrpc2.VarintObjectCodec{})
			jsonrpc2.NewConn(ctx, stream, h)
		}
	}()

	return lis.Addr(), nil
}

// Handler handles JSON-RPC requests and notifications.
type Handler struct{}

// Handle manages JSON-RPC requests and notifications, executing the requested
// method and responding back to the client when needed.
func (h *Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	log.WithFields(log.Fields{
		"notification": req.Notif,
		"method":       req.Method,
		"id":           req.ID,
	}).Info("rpc request received")

	if _, ok := registeredMethods.Load(req.Method); !ok {
		log.Errorf("rpc request method %s not supported", req.Method)

		if !req.Notif {
			resp := &jsonrpc2.Response{}
			resp.SetResult(jsonrpc2.Error{
				Code:    jsonrpc2.CodeMethodNotFound,
				Message: fmt.Sprintf("method %s not found", req.Method),
			})

			sendResponse(ctx, conn, req, resp)
		}

		return
	}

	params, err := req.Params.MarshalJSON()
	if err != nil {
		log.WithFields(log.Fields{
			"notification": req.Notif,
			"method":       req.Method,
			"id":           req.ID,
		}).WithError(err).Warn("error executing rpc method.")

		if !req.Notif {
			resp := &jsonrpc2.Response{}
			resp.SetResult(jsonrpc2.Error{
				Code:    jsonrpc2.CodeInternalError,
				Message: fmt.Sprintf("error executing rpc method %s", req.Method),
			})

			sendResponse(ctx, conn, req, resp)
		}

		return
	}

	m, _ := registeredMethods.Load(req.Method)
	rm, err := m.(Method)(params)
	if err != nil {
		log.WithFields(log.Fields{
			"notification": req.Notif,
			"method":       req.Method,
			"id":           req.ID,
		}).WithError(err).Warn("error executing rpc method.")

		if !req.Notif {
			resp := &jsonrpc2.Response{}
			resp.SetResult(jsonrpc2.Error{
				Code:    jsonrpc2.CodeInternalError,
				Message: fmt.Sprintf("error executing rpc method %s", req.Method),
			})

			sendResponse(ctx, conn, req, resp)
		}

		return
	}

	if !req.Notif {
		resp := &jsonrpc2.Response{ID: req.ID, Result: &rm}

		sendResponse(ctx, conn, req, resp)
	}
}

// Register adds a new method that can be called remotely.
func Register(name string, method Method) {
	registeredMethods.Store(name, method)
}

// Method represents a procedure that can be called remotely.
type Method func(params interface{}) (json.RawMessage, error)

func sendResponse(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request, resp *jsonrpc2.Response) error {
	if err := conn.SendResponse(ctx, resp); err != nil {
		log.WithFields(log.Fields{
			"notification": req.Notif,
			"method":       req.Method,
			"id":           req.ID,
		}).WithError(err).Warn("error sending rpc response.")

		return err
	}

	log.WithFields(log.Fields{
		"notification": req.Notif,
		"method":       req.Method,
		"id":           req.ID,
	}).Info("rpc response sent.")

	return nil

}
