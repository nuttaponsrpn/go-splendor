package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nuttaponsrpn/go-splendor/adapters"
	"github.com/nuttaponsrpn/go-splendor/core"
	"github.com/nuttaponsrpn/go-splendor/gotype"
)

func main() {
	var clients = make(map[*websocket.Conn]bool)
	var broadcast = make(chan gotype.GameState)
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	gameService := core.NewGameService(gotype.GameState{})
	websocketAdapters := adapters.NewWebsocketAdapter(upgrader, clients, broadcast, gameService)

	http.HandleFunc("/ws", websocketAdapters.CreateClientSocket)

	go websocketAdapters.BroadcastMessage()

	log.Println("HTTP server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
