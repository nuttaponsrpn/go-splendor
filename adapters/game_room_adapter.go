package adapters

import (
	"log"

	"github.com/gofiber/websocket/v2"
	"github.com/nuttaponsrpn/go-splendor/core"
)

type GameRoomAdapter struct {
	gr core.GameRoom
}

func NewGameRoomAdapter(gr *core.GameRoom) *GameRoomAdapter {
	return &GameRoomAdapter{
		gr: *gr,
	}
}

func (roomAdapter *GameRoomAdapter) HandleConnections(conn *websocket.Conn) {
	roomID := conn.Query("room_id")
	playerID := conn.Query("player_id")
	if roomID == "" {
		conn.Close()
		return
	}

	roomAdapter.gr.CreateRoom(roomID, playerID, conn)
}

var roomClients = make(map[*websocket.Conn]bool)

func (roomAdapter *GameRoomAdapter) ShowPlayerRooms(conn *websocket.Conn) {
	defer func() {
		conn.Close()
		delete(roomClients, conn)
	}()

	_, exists := roomClients[conn]
	if exists {
		for clients := range roomClients {
			clients.WriteJSON(roomAdapter.gr.GetRoom())
		}
	} else {
		roomClients[conn] = true
		for clients := range roomClients {
			clients.WriteJSON(roomAdapter.gr.GetRoom())
		}
	}

	go func() {
		for {
			room := <-roomAdapter.gr.GetRoomChannel()
			_, exists := roomClients[conn]
			if exists && room != "" {
				for clients := range roomClients {
					clients.WriteJSON(roomAdapter.gr.GetRoom())
				}
			} else {
				roomClients[conn] = true
				for clients := range roomClients {
					clients.WriteJSON(roomAdapter.gr.GetRoom())
				}
			}

		}
	}()

	for {
		var msg []core.DisplayRooms
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		if len(msg) == 1 && msg[0].RoomID == "close" {
			conn.Close()
			return
		}
	}
}
