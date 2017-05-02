package client

import (
	"github.com/goxjs/websocket"
	"zombiezen.com/go/capnproto2/rpc"
)

func Dial(addr string) (rpc.Transport, error) {
	c, err := websocket.Dial(addr, "") // Blocks until connection is established
	if err != nil {
		return nil, err
	}

	return rpc.StreamTransport(c), nil
}
