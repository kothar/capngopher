package main

import (
	"context"
	"log"
	"time"

	"bitbucket.org/mikehouston/webconsole"
	"zombiezen.com/go/capnproto2/rpc"

	"github.com/kothar/capngopher/example/service"
	"github.com/kothar/capngopher/webrtc"
)

func init() {
	webconsole.Enable()
}

func main() {
	// Get the current host
	// location := js.Global.Get("window").Get("location")
	// host := location.Get("host")
	// protocol := location.Get("protocol")

	// Init webrtc peer
	log.Printf("Connecting to PeerJS broker")
	peer := webrtc.NewPeer(webrtc.NewPeerConfig(
		webrtc.WithKey("znaqnunoxaqt1emi"),
		webrtc.WithDebug(3),
	))

	id, err := peer.ID()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to broker: id = %s", id)

	// Signal presence
	peers := []string{"one"}

	// Connect to other peers
	for _, remote := range peers {
		t, err := peer.Connect(remote)
		if err != nil {
			log.Fatal(err)
		}

		conn := rpc.NewConn(t)
		defer conn.Close()

		ctx := context.Background()
		pinger := service.Pinger{Client: conn.Bootstrap(ctx)}

		response, err := pinger.Ping(ctx, func(p service.Pinger_ping_Params) error {
			p.SetMsg("Hello World from " + id + " over WebRTC")
			return nil
		}).Struct()
		if err != nil {
			log.Fatal("Failed to send ping: ", err)
		}

		msg, err := response.Msg()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Received ping response: %s", msg)
	}

	log.Printf("Waiting 2s...")
	<-time.After(time.Second * 2)

	log.Printf("Exiting")
}
