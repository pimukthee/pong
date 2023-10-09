package main

import (
	"log"
	"net/http"
)

type Rooms struct {
	rooms     map[roomID]Room
	available []roomID
}

func main() {
  rooms := Rooms {
    rooms: make(map[roomID]Room),
  }
	handler := newHandler(&rooms)
	server := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
