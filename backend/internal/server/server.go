// Package server wires Fiber middleware around the handlers.
package server

import (
	"log/slog"
	"strings"
	"time"

	json "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/toontoonz/logistic-go/backend/internal/config"
	"github.com/toontoonz/logistic-go/backend/internal/handler"
)

// New builds the HTTP application. It does not listen; main does.
// Read and write timeouts are left unset on purpose: a WebSocket is a
// long-lived connection, and a short server timeout would cut it.
func New(cfg config.Config, api handler.API) *fiber.App {
	app := fiber.New(fiber.Config{
		// goccy/go-json is a faster drop-in for encoding/json.
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		AppName:               "lane",
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(cfg.CORSOrigins, ","),
		AllowMethods: "GET,POST,PATCH,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/health" {
			return c.Next()
		}
		reqID := uuid.NewString()
		c.Set("X-Request-ID", reqID)
		start := time.Now()
		err := c.Next()
		slog.Info("http",
			"req_id", reqID,
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"ms", time.Since(start).Milliseconds(),
		)
		return err
	})

	handler.Register(app, api, cfg.CORSOrigins)
	return app
}
