package models

import (
	"time"
)

type TrackingUpdate struct {
	ID         int       `json:"id" db:"id"`
	ShipmentID int       `json:"shipment_id" db:"shipment_id"`
	Status     string    `json:"status" db:"status" validate:"required"`
	Location   string    `json:"location" db:"location"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type TrackingUpdateRequest struct {
	ShipmentID int    `json:"shipment_id" validate:"required"`
	Status     string `json:"status" validate:"required"`
	Location   string `json:"location"`
}