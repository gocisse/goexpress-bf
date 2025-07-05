package models

import (
	"time"
)

type Driver struct {
	ID                   int       `json:"id" db:"id"`
	UserID               int       `json:"user_id,omitempty" db:"user_id"`
	Name                 string    `json:"name" db:"name"`
	Email                string    `json:"email" db:"email"`
	Role                 string    `json:"role" db:"role"`
	Phone                string    `json:"phone,omitempty" db:"phone"`
	LicenseNumber        string    `json:"license_number,omitempty" db:"license_number"`
	VehicleType          string    `json:"vehicle_type,omitempty" db:"vehicle_type"`
	VehicleNumber        string    `json:"vehicle_number,omitempty" db:"vehicle_number"`
	Status               string    `json:"status" db:"status"` // available, busy, offline
	CurrentLocation      string    `json:"current_location,omitempty" db:"current_location"`
	Rating               float64   `json:"rating" db:"rating"`
	TotalDeliveries      int       `json:"total_deliveries" db:"total_deliveries"`
	SuccessfulDeliveries int       `json:"successful_deliveries,omitempty" db:"successful_deliveries"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

type DriverStats struct {
	TotalDrivers     int     `json:"total_drivers"`
	AvailableDrivers int     `json:"available_drivers"`
	BusyDrivers      int     `json:"busy_drivers"`
	OfflineDrivers   int     `json:"offline_drivers"`
	TotalDeliveries  int     `json:"total_deliveries"`
	AverageRating    float64 `json:"average_rating"`
}

// Request/Response models
type CreateDriverRequest struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=6"`
	Phone           string `json:"phone"`
	LicenseNumber   string `json:"license_number"`
	VehicleType     string `json:"vehicle_type"`
	VehicleNumber   string `json:"vehicle_number"`
	CurrentLocation string `json:"current_location"`
}

type UpdateDriverRequest struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Phone           string `json:"phone"`
	LicenseNumber   string `json:"license_number"`
	VehicleType     string `json:"vehicle_type"`
	VehicleNumber   string `json:"vehicle_number"`
	Status          string `json:"status" validate:"required,oneof=available busy offline"`
	CurrentLocation string `json:"current_location"`
}


