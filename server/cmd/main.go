package main

import (
	"agar-server/internal/server"
	"agar-server/internal/server/clients"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 8075, "Port to listen on")
)

func main() {
	flag.Parse()

	hub := server.NewHub();
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})

	go hub.Run()
	addr := fmt.Sprint(":", *port)
	
	log.Printf("Starting server on %s port", addr)
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
}