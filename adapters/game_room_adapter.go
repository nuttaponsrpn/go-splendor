package adapters

import (
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

func (roomAdapter *GameRoomAdapter) HandleConnections(c *websocket.Conn) {
	roomID := c.Query("room_id")
	playerID := c.Query("player_id")
	if roomID == "" {
		c.Close()
		return
	}

	roomAdapter.gr.CreateRoom(roomID, playerID, c)
}
