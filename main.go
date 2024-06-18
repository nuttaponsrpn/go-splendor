package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/nuttaponsrpn/go-splendor/adapters"
	"github.com/nuttaponsrpn/go-splendor/core"
)

func main() {
	var rooms = make(map[string]*core.Room)
	gameRoomService := core.NewGameRoomService(&rooms)
	gameRoomAdapter := adapters.NewGameRoomAdapter(&gameRoomService)

	app := fiber.New()
	// Middleware to upgrade the HTTP connection to WebSocket
	app.Use("/ws", handleConnections)
	// WebSocket route
	app.Get("/ws", websocket.New(gameRoomAdapter.HandleConnections))

	// HTTP GET all rooms
	app.Get("/rooms", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(rooms)
	})

	// HTTP GET route
	app.Delete("/rooms", func(c *fiber.Ctx) error {
		m := c.Queries()
		room := gameRoomService.DeleteRoom(m["room_id"])
		return c.Status(fiber.StatusOK).JSON(room)
	})

	app.Listen(":8080")
}

func handleConnections(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}
