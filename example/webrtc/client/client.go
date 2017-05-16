package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"bitbucket.org/mikehouston/webconsole"
	"github.com/PalmStoneGames/gopherjs-net-http"
	"zombiezen.com/go/capnproto2/rpc"

	"github.com/kothar/capngopher/example/service"
	"github.com/kothar/capngopher/webrtc"
)

func init() {
	webconsole.Enable()
}

func requestPeers(id string) (map[string]time.Time, error) {
	req, err := http.NewRequest("GET", "/register?id="+id, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("did not get acceptable status code: %s body: %q", resp.Status, string(body))
		return nil, nil
	}

	var peers map[string]time.Time
	err = json.NewDecoder(resp.Body).Decode(&peers)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func serve(s *service.PingerServer, l *webrtc.PeerListener) {
	for {
		t, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Received connection, starting server")

		// Create a new locally implemented Pinger.
		main := service.Pinger_ServerToClient(s)

		// Listen for calls, using the Pinger as the bootstrap interface.
		rpc.NewConn(t, rpc.MainInterface(main.Client))
	}
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

	// Start local ping server
	log.Printf("Starting local Pinger server")
	s := &service.PingerServer{}
	l, err := peer.Listen()
	if err != nil {
		log.Fatal(err)
	}
	go serve(s, l)

	id, err := peer.ID()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to broker: id = %s", id)

	// Signal presence
	peers, err := requestPeers(id)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received list of peers from server: %v", peers)

	// Connect to other peers
	for remote, lastSeen := range peers {
		if !lastSeen.After(time.Now().Add(-time.Second * 30)) {
			continue
		}

		t, err := peer.Connect(remote)
		if err != nil {
			log.Fatal(err)
		}

		conn := rpc.NewConn(t)
		defer conn.Close()

		ctx := context.Background()
		pinger := service.Pinger{Client: conn.Bootstrap(ctx)}

		log.Println("Sending ping")
		response, err := pinger.Ping(ctx, func(p service.Pinger_ping_Params) error {
			p.SetMsg("Hello World from " + id + " over WebRTC")
			return nil
		}).Struct()
		if err != nil {
			log.Println("Failed to send ping: ", err)
		} else {
			log.Println("Sent ping")
		}

		msg, err := response.Msg()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Received ping response: %s", msg)
	}

	log.Printf("Waiting 20m...")
	<-time.After(time.Minute * 20)

	log.Printf("Exiting")
}
