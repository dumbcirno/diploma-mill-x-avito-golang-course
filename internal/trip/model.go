package trip

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive    = "active"
	StatusCompleted = "completed"
)

type Trip struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	DriverID   uuid.UUID
	StartLat   float64
	StartLon   float64
	EndLat     float64
	EndLon     float64
	Price      int64
	Status     string
	StartedAt  time.Time
	FinishedAt *time.Time
}

type CreateInput struct {
	UserID   uuid.UUID
	DriverID uuid.UUID
	StartLat float64
	StartLon float64
	EndLat   float64
	EndLon   float64
	Price    int64
}

func (in CreateInput) validate() error {
	if in.UserID == uuid.Nil || in.DriverID == uuid.Nil {
		return ErrInvalidRequest
	}
	if in.StartLat < -90 || in.StartLat > 90 || in.EndLat < -90 || in.EndLat > 90 {
		return ErrInvalidRequest
	}
	if in.StartLon < -180 || in.StartLon > 180 || in.EndLon < -180 || in.EndLon > 180 {
		return ErrInvalidRequest
	}
	if in.Price < 0 {
		return ErrInvalidRequest
	}
	return nil
}

type IdempotencyRecord struct {
	Key         uuid.UUID
	RequestHash string
	TripID      uuid.UUID
	ExpiresAt   time.Time
}
