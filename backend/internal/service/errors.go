// Package service holds the use cases: validation, status rules, and
// the orchestration of Postgres, Redis, and the live map.
package service

import (
	"github.com/toontoonz/logistic-go/backend/internal/model"
)

// InputError is a bad request: missing field, impossible coordinate, and so on.
type InputError struct {
	Message string
}

func (e *InputError) Error() string { return e.Message }

// ConflictError is a legal request that the current state refuses,
// such as delivering an order that was never picked up.
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string { return e.Message }

// NotFoundError names the missing thing ("order", "driver").
// Unwrap lets callers use errors.Is(err, model.ErrNotFound).
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string { return e.Message }

func (e *NotFoundError) Unwrap() error { return model.ErrNotFound }

func notFound(what string) error {
	return &NotFoundError{Message: what + " not found"}
}
