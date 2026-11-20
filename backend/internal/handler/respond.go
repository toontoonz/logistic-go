// Package handler translates HTTP into service calls.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/toontoonz/logistic-go/backend/internal/model"
	"github.com/toontoonz/logistic-go/backend/internal/service"
)

type errorBody struct {
	Error string `json:"error"`
}

type dataBody struct {
	Data any `json:"data"`
}

func writeData(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(dataBody{Data: data})
}

func writeErr(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(errorBody{Error: message})
}

// writeServiceErr maps the sentinel-style service errors onto HTTP status codes.
func writeServiceErr(c *fiber.Ctx, err error) error {
	var input *service.InputError
	var missing *service.NotFoundError
	var conflict *service.ConflictError
	switch {
	case errors.As(err, &input):
		return writeErr(c, fiber.StatusBadRequest, input.Error())
	case errors.As(err, &missing):
		return writeErr(c, fiber.StatusNotFound, missing.Error())
	case errors.As(err, &conflict):
		return writeErr(c, fiber.StatusConflict, conflict.Error())
	case errors.Is(err, model.ErrNotFound):
		return writeErr(c, fiber.StatusNotFound, "not found")
	default:
		slog.Error("request failed", "err", err, "path", c.Path())
		return writeErr(c, fiber.StatusInternalServerError, "internal error")
	}
}

// requestCtx is the standard-library context for this HTTP call.
// Fiber's own c.Context() is a fasthttp value, not a context.Context.
// A timeout means a stuck query cannot hold a database connection open.
func requestCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	parent := c.UserContext()
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, 5*time.Second)
}
