package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var peers = make(map[string]time.Time)

func main() {

	// Set up HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/index.html")
	})
	http.HandleFunc("/client.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/client.js")
	})
	http.HandleFunc("/client.js.map", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/client.js.map")
	})

	// Get list of peers
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		peerID := r.FormValue("id")
		if peerID == "" {
			http.Error(w, "No ID specified", http.StatusBadRequest)
			return
		}

		json, err := json.Marshal(peers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(json)

		peers[peerID] = time.Now()
	})

	listen := "0.0.0.0:8081"
	log.Printf("Serving on http://%s\n", listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
