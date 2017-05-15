package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
	"zombiezen.com/go/capnproto2/rpc"

	"github.com/kothar/capngopher/example/service"
	"github.com/kothar/capngopher/ws/server"
)

type pingerServer struct {
}

func (s *pingerServer) Ping(p service.Pinger_ping) error {
	msg, err := p.Params.Msg()
	if err != nil {
		return err
	}

	log.Printf("Ping: %s\n", msg)
	if err := p.Results.SetMsg("Ping: " + msg); err != nil {
		return err
	}

	return nil
}

func serve(s *pingerServer, listener *server.WebsocketListener) {
	for {
		t, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		// Create a new locally implemented Pinger.
		main := service.Pinger_ServerToClient(s)

		// Listen for calls, using the Pinger as the bootstrap interface.
		conn := rpc.NewConn(t, rpc.MainInterface(main.Client))

		// Wait for connection to abort.
		err = conn.Wait()
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	// Init websocket listener
	listener := server.NewListener()

	s := &pingerServer{}
	go serve(s, listener)

	// Set up HTTP handlers
	http.Handle("/ws", websocket.Handler(listener.Handler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/index.html")
	})
	http.HandleFunc("/client.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/client.js")
	})

	listen := "0.0.0.0:8081"
	log.Printf("Serving on http://%s\n", listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
