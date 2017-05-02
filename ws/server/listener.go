package server

import (
	"log"
	"net"

	"zombiezen.com/go/capnproto2/rpc"

	"golang.org/x/net/websocket"
)

type wsConn struct {
	*websocket.Conn

	close chan struct{}
}

// Close overrides the default behaviour to pass a message back to
// the websocket handler
func (w *wsConn) Close() error {
	w.close <- struct{}{}
	return nil
}

type WebsocketListener struct {
	connections chan net.Conn
}

func (l *WebsocketListener) Accept() (rpc.Transport, error) {
	c := <-l.connections
	return rpc.StreamTransport(c), nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *WebsocketListener) Close() error {
	close(l.connections)
	return nil
}

// Addr returns the listener's network address.
func (l *WebsocketListener) Addr() net.Addr {
	return nil
}

func (l *WebsocketListener) Handler(ws *websocket.Conn) {
	log.Println("Accepted new websocket connection")

	// Set payload type
	ws.PayloadType = websocket.BinaryFrame

	c := &wsConn{
		Conn:  ws,
		close: make(chan struct{}),
	}
	l.connections <- c

	// Wait for the close signal
	<-c.close
}

func NewListener() *WebsocketListener {

	listener := &WebsocketListener{
		connections: make(chan net.Conn),
	}

	return listener
}
