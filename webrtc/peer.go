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
		go func() {
			close(peer.onOpen)
			peer.onOpen = nil
		}()
	})

	o.Call("on", "error", func(err *PeerError) {
		go func() {
			log.Println("Peer:", err)
		}()
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

type PeerListener struct {
	peer      *Peer
	onConnect chan *PeerConnection
}

func (p *Peer) Listen() (*PeerListener, error) {

	l := &PeerListener{
		peer:      p,
		onConnect: make(chan *PeerConnection),
	}

	p.o.Call("on", "connection", func(conn *js.Object) {
		c := newPeerConnection(conn)
		log.Println("Received connection from remote peer ", c.Peer)
		l.onConnect <- c
	})

	return l, nil
}

func (l *PeerListener) Accept() (rpc.Transport, error) {
	c := <-l.onConnect
	log.Println("Accepted connection from remote peer ", c.Peer)
	t := rpc.StreamTransport(c)
	return t, nil
}

type PeerConnection struct {
	o *js.Object

	Peer string `js:"peer"`

	status  string
	onReady chan struct{}

	err   error
	onErr chan error

	buffer []byte
	onData chan []byte
}

func (p *Peer) Connect(remoteID string) (rpc.Transport, error) {
	conn := p.o.Call("connect", remoteID)
	log.Println("Connecting to remote peer ", remoteID)

	c := newPeerConnection(conn)
	t := rpc.StreamTransport(c)

	return t, nil
}

func newPeerConnection(conn *js.Object) *PeerConnection {
	c := &PeerConnection{
		o: conn,
	}

	c.onData = make(chan []byte)
	c.onErr = make(chan error)
	c.onReady = make(chan struct{})

	go func() {
		// TODO hook the peer errors and look for peer unavailable messages
		<-time.After(time.Second * 5)
		if c.onReady != nil && c.onErr != nil {
			c.err = errors.New("Connection to " + c.Peer + " timed out")

			onErr := c.onErr
			c.onErr = nil
			log.Println("Marked connection error")
			onErr <- c.err
			close(onErr)
		}
	}()

	conn.Call("on", "open", func() {
		go func() {
			log.Println("Connection to " + c.Peer + " open")
			c.status = "open"

			onReady := c.onReady
			c.onReady = nil
			log.Println("Marked connection as ready")
			onReady <- struct{}{}
			close(onReady)
		}()
	})
	conn.Call("on", "close", func() {
		go func() {
			log.Println("Connection to " + c.Peer + " closed")
			c.status = "closed"

			c.err = errors.New("Closed")
			onErr := c.onErr
			c.onErr = nil
			log.Println("Marked connection as closed")
			onErr <- c.err
			close(onErr)
		}()
	})
	conn.Call("on", "data", func(data *js.Object) {
		go func() {
			bytes := js.Global.Get("Uint8Array").New(data).Interface().([]byte)
			c.onData <- bytes
		}()
	})
	conn.Call("on", "error", func(err *PeerError) {
		go func() {
			log.Println("Conn:", err.Type)
			c.err = err

			if c.onErr != nil {
				onErr := c.onErr
				c.onErr = nil
				log.Println("Marked connection error")
				onErr <- c.err
				close(onErr)
			}
		}()
	})

	return c
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
		n = len(p)
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
		log.Println("Waiting for channel " + c.Peer + " to connect")
		select {
		case <-c.onReady:
			log.Println("Connected to " + c.Peer + ": writing")
		case err = <-c.onErr:
			return 0, err
		}
	}

	c.o.Call("send", js.NewArrayBuffer(p))
	return len(p), c.err
}

func (c *PeerConnection) Close() error {
	if c.err != nil {
		return c.err
	}

	c.o.Call("close")
	return c.err
}
