package webrtc

import (
	"errors"

	"golang.org/x/net/context"

	"github.com/gopherjs/gopherjs/js"
	"zombiezen.com/go/capnproto2/rpc"
	rpccapnp "zombiezen.com/go/capnproto2/std/capnp/rpc"
)

type Peer struct {
	o  *js.Object
	id string `js:"id"`

	idReady chan struct{}
}

type PeerListener struct {
	peer *Peer
}

type PeerConfig struct {
	o   *js.Object
	ID  string `js:"id"`
	Key string `js:"key"`
}

type ConfigOption func(config *PeerConfig)

func WithKey(key string) ConfigOption {
	return func(config *PeerConfig) {
		config.Key = key
	}
}

func NewPeerConfig(options ...ConfigOption) *PeerConfig {
	config := &PeerConfig{o: js.Global.Get("Object").New()}
	for _, option := range options {
		option(config)
	}
	return config
}

func NewPeer(config *PeerConfig) *Peer {
	o := js.Global.Get("Peer").New(config)
	peer := &Peer{o: o}
	peer.idReady = make(chan struct{})
	peer.o.Call("on", "open", func(id string) {
		close(peer.idReady)
		peer.idReady = nil
	})
	return peer
}

func (p *Peer) ID() (string, error) {
	if p.idReady != nil {
		for range p.idReady {
		}
	}
	return p.id, nil
}

func (p *Peer) Connect(remoteID string) (rpc.Transport, error) {
	conn := p.o.Call("connect", remoteID)
	t := &PeerTransport{
		conn: conn,
	}
	return t, nil
}

type PeerTransport struct {
	conn *js.Object
}

// SendMessage sends msg.
func (t *PeerTransport) SendMessage(ctx context.Context, msg rpccapnp.Message) error {
	return errors.New("not implemented")
}

// RecvMessage waits to receive a message and returns it.
// Implementations may re-use buffers between calls, so the message is
// only valid until the next call to RecvMessage.
func (t *PeerTransport) RecvMessage(ctx context.Context) (rpccapnp.Message, error) {
	m := rpccapnp.Message{}
	return m, errors.New("not implemented")
}

// Close releases any resources associated with the transport.
func (t *PeerTransport) Close() error {
	return errors.New("not implemented")
}
