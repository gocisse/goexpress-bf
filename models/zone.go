package models

import (
	"time"
)

type Zone struct {
	ID         int       `json:"id" db:"id"`
	Name       string    `json:"name" db:"name" validate:"required"`
	PricePerKg float64   `json:"price_per_kg" db:"price_per_kg" validate:"required,gt=0"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}