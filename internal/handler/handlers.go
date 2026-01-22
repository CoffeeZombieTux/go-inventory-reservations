package handler

import "go-inventory-reservations/internal/logger"

type Handlers struct {
	Logger *logger.Logger
}

func NewHandlers(logger *logger.Logger) *Handlers {
	return &Handlers{Logger: logger}
}
