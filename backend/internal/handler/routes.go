package handler

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// Register attaches the HTTP and WebSocket routes.
func Register(app *fiber.App, api API, origins []string) {
	app.Get("/health", api.health)

	app.Post("/api/drivers", api.createDriver)
	app.Get("/api/drivers", api.listDrivers)
	app.Get("/api/drivers/:id", api.getDriver)
	app.Post("/api/drivers/:id/location", api.postLocation)
	app.Get("/api/drivers/:id/history", api.locationHistory)

	app.Post("/api/orders", api.createOrder)
	app.Get("/api/orders", api.listOrders)
	app.Get("/api/orders/:id", api.getOrder)
	app.Post("/api/orders/:id/assign", api.assignOrder)
	app.Patch("/api/orders/:id/status", api.updateOrderStatus)

	app.Get("/api/demo", api.demoStatus)
	app.Post("/api/demo/start", api.startDemo)
	app.Post("/api/demo/stop", api.stopDemo)

	// Fiber only upgrades the connection when this middleware runs first.
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/positions", websocket.New(api.positionsWS, websocket.Config{
		Origins:         origins,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}))
}
