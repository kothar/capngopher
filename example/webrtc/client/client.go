package main

import (
	"context"
	"log"

	"github.com/gopherjs/gopherjs/js"
	"zombiezen.com/go/capnproto2/rpc"

	"time"

	"bitbucket.org/mikehouston/capngopher/example/service"
	"bitbucket.org/mikehouston/capngopher/ws/client"
	"bitbucket.org/mikehouston/webconsole"
)

func init() {
	webconsole.Enable()
}

func main() {
	// Get the current host
	location := js.Global.Get("window").Get("location")
	host := location.Get("host")
	protocol := location.Get("protocol")

	var path string
	if protocol.String() == "https:" {
		path = "wss://" + host.String() + "/ws"
	} else {
		path = "ws://" + host.String() + "/ws"
	}

	log.Println("Connecting to websocket")
	t, err := client.Dial(path)
	if err != nil {
		log.Fatal(err)
	}

	conn := rpc.NewConn(t)
	defer conn.Close()

	ctx := context.Background()
	pinger := service.Pinger{Client: conn.Bootstrap(ctx)}

	response, err := pinger.Ping(ctx, func(p service.Pinger_ping_Params) error {
		p.SetMsg("Hello World over websockets")
		return nil
	}).Struct()
	if err != nil {
		log.Fatal(err)
	}

	msg, err := response.Msg()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received ping response: %s", msg)

	log.Printf("Waiting 2s...")
	<-time.After(time.Second * 2)

	log.Printf("Exiting")
}
