package dto

import (
	"time"

	"github.com/google/uuid"
)

type ResponseModel struct {
	ID        uuid.UUID `json:"_id"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorModel struct {
	Message string    `json:"message"`
	Details []*string `json:"details"`
}
