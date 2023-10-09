package main

import (
	"log"
	"net/http"
	"pong/internal/api"
	"pong/internal/game"
)

func main() {
  game := game.NewGame() 
	handler := api.NewHandler(game)

	server := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
