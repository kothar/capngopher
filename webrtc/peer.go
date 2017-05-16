package webrtc

import (
	"errors"
	"log"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"zombiezen.com/go/capnproto2/rpc"
)

type PeerError struct {
	o *js.Object

	Type string `js:"type"`
}

func (err *PeerError) Error() string {
	return "[" + err.Type + "] " + err.o.String()
}

type PeerListener struct {
	peer *Peer
}

type PeerConfig struct {
	o     *js.Object
	ID    string `js:"id"`
	Key   string `js:"key"`
	Debug int    `js:"debug"`
}

type ConfigOption func(config *PeerConfig)

func WithID(id string) ConfigOption {
	return func(config *PeerConfig) {
		config.ID = id
	}
}

func WithKey(key string) ConfigOption {
	return func(config *PeerConfig) {
		config.Key = key
	}
}

func WithDebug(debug int) ConfigOption {
	return func(config *PeerConfig) {
		config.Debug = debug
	}
}

func NewPeerConfig(options ...ConfigOption) *PeerConfig {
	config := &PeerConfig{o: js.Global.Get("Object").New()}
	for _, option := range options {
		option(config)
	}
	return config
}

type Peer struct {
	o  *js.Object
	id string `js:"id"`

	onOpen chan struct{}
}

func NewPeer(config *PeerConfig) *Peer {
	o := js.Global.Get("Peer").New(config)
	peer := &Peer{o: o}

	peer.onOpen = make(chan struct{})
	o.Call("on", "open", func(id string) {
		close(peer.onOpen)
		peer.onOpen = nil
	})

	o.Call("on", "error", func(err *PeerError) {
		log.Println("Peer:", err)
	})

	return peer
}

func (p *Peer) ID() (string, error) {
	if p.onOpen != nil {
		for range p.onOpen {
		}
	}
	return p.id, nil
}

type PeerConnection struct {
	o *js.Object

	status  string
	onReady chan struct{}

	err   error
	onErr chan error

	buffer []byte
	onData chan []byte
}

func (p *Peer) Connect(remoteID string) (rpc.Transport, error) {
	conn := p.o.Call("connect", remoteID)
	c := &PeerConnection{
		o: conn,
	}

	c.onData = make(chan []byte)
	c.onReady = make(chan struct{})
	c.onErr = make(chan error)

	go func() {
		// TODO hook the peer errors and look for peer unavailable messages
		<-time.After(time.Second * 5)
		if c.onReady != nil && c.onErr != nil {
			c.err = errors.New("Connection to " + remoteID + " timed out")

			c.onErr <- c.err
			close(c.onErr)
			c.onErr = nil

			close(c.onReady)
			c.onReady = nil
		}
	}()

	conn.Call("on", "open", func() {
		log.Println("open")
		c.status = "open"
		c.onReady <- struct{}{}
		close(c.onReady)
		c.onReady = nil
	})
	conn.Call("on", "close", func() {
		log.Println("close")
		c.status = "closed"
		close(c.onReady)
		c.onReady = nil
	})
	conn.Call("on", "data", func(data []byte) { c.onData <- data })
	conn.Call("on", "error", func(err *PeerError) {
		log.Println("Conn:", err.Type)
		c.err = err

		if c.onErr != nil {
			c.onErr <- c.err
			close(c.onErr)
			c.onErr = nil

			close(c.onReady)
			c.onReady = nil
		}
	})

	t := rpc.StreamTransport(c)

	return t, nil
}

func (c *PeerConnection) Read(p []byte) (n int, err error) {
	if c.err != nil {
		return 0, c.err
	}

	remaining := len(c.buffer)
	if remaining == 0 {
		select {
		case err = <-c.onErr:
			return 0, err
		case c.buffer = <-c.onData:
			remaining = len(c.buffer)
		}
	}

	if remaining > 0 {
		n := len(p)
		if remaining < n {
			n = remaining
		}

		copy(p[:n], c.buffer[:n])
		c.buffer = c.buffer[n:]
	}

	return n, c.err
}

func (c *PeerConnection) Write(p []byte) (n int, err error) {
	if c.err != nil {
		return 0, c.err
	}

	if c.onReady != nil {
		log.Println("Waiting for channel to connect")
		select {
		case <-c.onReady:
			log.Println("Connected")
		case err = <-c.onErr:
			return 0, err
		}
	}

	c.o.Call("send", p)
	return len(p), c.err
}

func (c *PeerConnection) Close() error {
	if c.err != nil {
		return c.err
	}

	c.o.Call("close")
	return c.err
}
