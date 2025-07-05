package models

import (
	"time"
)

type Shipment struct {
	ID             int       `json:"id" db:"id"`
	TrackingNumber string    `json:"tracking_number" db:"tracking_number"`
	Origin         string    `json:"origin" db:"origin" validate:"required"`
	Destination    string    `json:"destination" db:"destination" validate:"required"`
	Weight         float64   `json:"weight" db:"weight" validate:"required,gt=0"`
	ZoneID         int       `json:"zone_id" db:"zone_id" validate:"required"`
	Status         string    `json:"status" db:"status"`
	CustomerID     int       `json:"customer_id" db:"customer_id"`
	DriverID       *int      `json:"driver_id" db:"driver_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type ShipmentRequest struct {
	Origin      string  `json:"origin" validate:"required"`
	Destination string  `json:"destination" validate:"required"`
	Weight      float64 `json:"weight" validate:"required,gt=0"`
	ZoneID      int     `json:"zone_id" validate:"required"`
}

type ShipmentResponse struct {
	Shipment       Shipment          `json:"shipment"`
	TrackingUpdate []TrackingUpdate  `json:"tracking_updates"`
	Zone           Zone             `json:"zone"`
}

type QuoteRequest struct {
	Weight float64 `json:"weight" validate:"required,gt=0"`
	ZoneID int     `json:"zone_id" validate:"required"`
}

type QuoteResponse struct {
	Weight    float64 `json:"weight"`
	ZoneID    int     `json:"zone_id"`
	ZoneName  string  `json:"zone_name"`
	PricePerKg float64 `json:"price_per_kg"`
	TotalPrice float64 `json:"total_price"`
}