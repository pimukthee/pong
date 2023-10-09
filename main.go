package main

import (
	"log"
	"net/http"
)

func main() {
	hub := newHub()
	handler := newHandler(&hub)
	go hub.run()
	server := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
