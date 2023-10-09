package main

import (
	"log"
	"net/http"
)

func main() {
	room := newRoom()
	handler := newHandler(&room)
	go room.run()
	server := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
