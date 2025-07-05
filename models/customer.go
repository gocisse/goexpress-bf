package models

import (
	"time"
)

type Customer struct {
	ID              int       `json:"id" db:"id"`
	UserID          int       `json:"user_id" db:"user_id"`
	CompanyName     string    `json:"company_name" db:"company_name"`
	ContactPerson   string    `json:"contact_person" db:"contact_person"`
	Phone           string    `json:"phone" db:"phone"`
	AlternatePhone  string    `json:"alternate_phone" db:"alternate_phone"`
	Website         string    `json:"website" db:"website"`
	TaxID           string    `json:"tax_id" db:"tax_id"`
	BusinessType    string    `json:"business_type" db:"business_type"`
	Status          string    `json:"status" db:"status"` // active, inactive, suspended
	CreditLimit     float64   `json:"credit_limit" db:"credit_limit"`
	PaymentTerms    string    `json:"payment_terms" db:"payment_terms"`
	Notes           string    `json:"notes" db:"notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	
	// Joined fields from users table
	Name            string    `json:"name" db:"name"`
	Email           string    `json:"email" db:"email"`
	
	// Calculated fields
	TotalShipments  int       `json:"total_shipments"`
	TotalSpent      float64   `json:"total_spent"`
	LastShipment    *time.Time `json:"last_shipment"`
}

type CustomerAddress struct {
	ID          int       `json:"id" db:"id"`
	CustomerID  int       `json:"customer_id" db:"customer_id"`
	Type        string    `json:"type" db:"type"` // billing, shipping, both
	Label       string    `json:"label" db:"label"` // home, office, warehouse
	AddressLine1 string   `json:"address_line1" db:"address_line1"`
	AddressLine2 string   `json:"address_line2" db:"address_line2"`
	City        string    `json:"city" db:"city"`
	State       string    `json:"state" db:"state"`
	PostalCode  string    `json:"postal_code" db:"postal_code"`
	Country     string    `json:"country" db:"country"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CustomerStats struct {
	TotalCustomers     int     `json:"total_customers"`
	ActiveCustomers    int     `json:"active_customers"`
	InactiveCustomers  int     `json:"inactive_customers"`
	TotalRevenue       float64 `json:"total_revenue"`
	AverageOrderValue  float64 `json:"average_order_value"`
	TopCustomers       []Customer `json:"top_customers"`
}

// Request/Response models
type CreateCustomerRequest struct {
	UserID          int     `json:"user_id" validate:"required"`
	CompanyName     string  `json:"company_name" validate:"required"`
	ContactPerson   string  `json:"contact_person" validate:"required"`
	Phone           string  `json:"phone" validate:"required"`
	AlternatePhone  string  `json:"alternate_phone"`
	Website         string  `json:"website"`
	TaxID           string  `json:"tax_id"`
	BusinessType    string  `json:"business_type"`
	CreditLimit     float64 `json:"credit_limit"`
	PaymentTerms    string  `json:"payment_terms"`
	Notes           string  `json:"notes"`
}

type UpdateCustomerRequest struct {
	CompanyName     string  `json:"company_name" validate:"required"`
	ContactPerson   string  `json:"contact_person" validate:"required"`
	Phone           string  `json:"phone" validate:"required"`
	AlternatePhone  string  `json:"alternate_phone"`
	Website         string  `json:"website"`
	TaxID           string  `json:"tax_id"`
	BusinessType    string  `json:"business_type"`
	Status          string  `json:"status" validate:"required,oneof=active inactive suspended"`
	CreditLimit     float64 `json:"credit_limit"`
	PaymentTerms    string  `json:"payment_terms"`
	Notes           string  `json:"notes"`
}

type CreateAddressRequest struct {
	Type         string `json:"type" validate:"required,oneof=billing shipping both"`
	Label        string `json:"label" validate:"required"`
	AddressLine1 string `json:"address_line1" validate:"required"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city" validate:"required"`
	State        string `json:"state" validate:"required"`
	PostalCode   string `json:"postal_code" validate:"required"`
	Country      string `json:"country" validate:"required"`
	IsDefault    bool   `json:"is_default"`
}

type CustomerWithAddresses struct {
	Customer
	Addresses []CustomerAddress `json:"addresses"`
}

type CustomerShipmentHistory struct {
	Customer
	Shipments []Shipment `json:"shipments"`
	Stats     struct {
		TotalShipments int     `json:"total_shipments"`
		TotalSpent     float64 `json:"total_spent"`
		LastShipment   *time.Time `json:"last_shipment"`
	} `json:"stats"`
}
