package main

import (
	"log"
	"net/http"
)

var clients []string

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

	listen := "0.0.0.0:8081"
	log.Printf("Serving on http://%s\n", listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
