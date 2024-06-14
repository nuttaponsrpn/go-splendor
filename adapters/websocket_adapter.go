package adapters

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nuttaponsrpn/go-splendor/core"
	"github.com/nuttaponsrpn/go-splendor/gotype"
)

type WebsocketAdapter struct {
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]bool
	broadcast   chan gotype.GameState
	gameService core.GameService
}

func NewWebsocketAdapter(
	upgrader websocket.Upgrader,
	clients map[*websocket.Conn]bool,
	broadcast chan gotype.GameState,
	gameService core.GameService,
) *WebsocketAdapter {
	return &WebsocketAdapter{upgrader: upgrader, clients: clients, broadcast: broadcast, gameService: gameService}
}

func (wsAdapter *WebsocketAdapter) CreateClientSocket(w http.ResponseWriter, r *http.Request) {
	// Create websocket connection
	ws, err := wsAdapter.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Close connection
	defer ws.Close()

	// Add clients to the websocket pool
	wsAdapter.clients[ws] = true

	// Detect messages from clients and broadcast to the rest of the clients in the websocket pool
	for {
		var msg core.WebsocketPlayerAction

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(wsAdapter.clients, ws)
			wsAdapter.CloseClientConnection(ws)
			break
		}

		gameState := wsAdapter.gameService.GetGameState()

		switch msg.Status {
		case gotype.Waiting:
			wsAdapter.gameService.JoinPlayer(msg.PlayerId)
			gameState = wsAdapter.gameService.GetGameState()

		case gotype.Started:
			wsAdapter.gameService.UpdateGameState(msg)
			gameState = wsAdapter.gameService.GetGameState()

		case gotype.CloseConnection:
			wsAdapter.CloseClientConnection(ws)
			gameState = wsAdapter.gameService.GetGameState()
		}

		wsAdapter.broadcast <- gameState
	}
}

func (wsAdapter *WebsocketAdapter) BroadcastMessage() {
	for {
		msg := <-wsAdapter.broadcast

		for client := range wsAdapter.clients {
			if client != nil {

				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					wsAdapter.CloseClientConnection(client)
				}

			}
		}
	}
}

func (wsAdapter *WebsocketAdapter) CloseClientConnection(ws *websocket.Conn) {
	err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Printf("error: %v", err)
	}
	ws.Close()
	delete(wsAdapter.clients, ws)
}
