package core

import (
	"log"

	"github.com/gofiber/websocket/v2"
	"github.com/nuttaponsrpn/go-splendor/gotype"
)

type Client struct {
	conn *websocket.Conn
	Room *Room `json:"room"`
}

type Room struct {
	clients     map[*Client]string
	register    chan *Client
	unregister  chan *Client
	broadcast   chan gotype.GameState
	close       chan bool
	GameService GameService `json:"gameService"`
}

type GameRoom interface {
	CreateRoom(roomID string, playerID string, c *websocket.Conn)
	DeleteRoom(roomID string) *Room
}

type GameRoomService struct {
	rooms map[string]*Room
}

func NewGameRoomService(rooms *map[string]*Room) GameRoom {
	return &GameRoomService{rooms: *rooms}
}

func (gs *GameRoomService) CreateRoom(roomID string, playerID string, conn *websocket.Conn) {
	room, exists := gs.rooms[roomID]
	if !exists {
		room := &Room{
			clients:     make(map[*Client]string),
			register:    make(chan *Client),
			unregister:  make(chan *Client),
			broadcast:   make(chan gotype.GameState),
			GameService: NewGameService(gotype.GameState{}),
		}
		// Detect message from other client
		go room.run(roomID, playerID, gs.rooms)
		// Add room id to rooms pool
		gs.rooms[roomID] = room
	}

	// Send client to register in room channle
	client := &Client{conn: conn, Room: room}
	gs.rooms[roomID].register <- client

	defer func() {
		room, exists := gs.rooms[roomID]
		if exists {
			_, exists := room.clients[client]
			if exists {
				room.unregister <- client
			}
			conn.Close()
		}
	}()

	for {
		var msg WebsocketPlayerAction
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}

		gameState := gs.rooms[roomID].GameService.GetGameState()

		switch msg.Status {
		case gotype.Waiting:
			gs.rooms[roomID].GameService.JoinPlayer(msg.PlayerId)
			gameState = gs.rooms[roomID].GameService.GetGameState()

		case gotype.Started:
			gs.rooms[roomID].GameService.UpdateGameState(msg)
			gameState = gs.rooms[roomID].GameService.GetGameState()
		case gotype.CloseConnection:
			gs.rooms[roomID].unregister <- client
			gs.rooms[roomID].GameService.RemovePlayer(msg.PlayerId)
			gameState = gs.rooms[roomID].GameService.GetGameState()
		}

		if gs.rooms[roomID] != nil {
			gs.rooms[roomID].broadcast <- gameState
		}
	}
}

func (r *Room) run(roomID string, playerID string, rooms map[string]*Room) {
	for {
		select {
		case client := <-r.register:
			r.clients[client] = playerID

		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(rooms[roomID].clients, client)
				client.conn.Close()

				if len(r.clients) == 0 {
					delete(rooms, roomID)
				}
			}
		case message := <-r.broadcast:
			for client := range r.clients {
				err := client.conn.WriteJSON(message)
				if err != nil {
					log.Printf("error: %v", err)
					client.conn.Close()
					delete(r.clients, client)
				}
			}
		case <-r.close:
			for client := range r.clients {
				client.conn.Close()
				delete(r.clients, client)
			}
			delete(rooms, roomID)
			return
		}
	}
}

func (gs *GameRoomService) DeleteRoom(roomID string) *Room {
	clientsLen := gs.rooms[roomID].clients

	for client := range clientsLen {
		gs.rooms[roomID].unregister <- client
	}

	return gs.rooms[roomID]
}
