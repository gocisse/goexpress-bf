package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"goexpress-api/middleware"
	"goexpress-api/models"
	"goexpress-api/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type DriverHandler struct {
	db        *sql.DB
	validator *validator.Validate
}

func NewDriverHandler(db *sql.DB) *DriverHandler {
	return &DriverHandler{
		db:        db,
		validator: validator.New(),
	}
}

// @Summary Get all drivers
// @Description Get all drivers with their details and stats
// @Tags drivers
// @Security ApiKeyAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Success 200 {array} models.Driver
// @Router /api/drivers [get]
func (h *DriverHandler) GetDrivers(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can view all drivers
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	statusFilter := r.URL.Query().Get("status")
	
	query := `
		SELECT 
			u.id, u.name, u.email, u.role, u.created_at, u.updated_at
		FROM users u
		WHERE u.role = 'driver'`

	var args []interface{}

	if statusFilter != "" {
		// For now, we'll just return all drivers since we don't have a drivers table
		// In a real implementation, you'd join with a drivers table
	}

	query += " ORDER BY u.created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var drivers []models.Driver
	for rows.Next() {
		var d models.Driver
		err := rows.Scan(
			&d.ID, &d.Name, &d.Email, &d.Role, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to scan driver", http.StatusInternalServerError)
			return
		}
		// Set default values for driver-specific fields
		d.Status = "available"
		d.Rating = 4.5
		d.TotalDeliveries = 0
		drivers = append(drivers, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drivers)
}

// @Summary Get driver stats
// @Description Get driver statistics (admin only)
// @Tags drivers
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} models.DriverStats
// @Router /api/drivers/stats [get]
func (h *DriverHandler) GetDriverStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can view stats
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var stats models.DriverStats

	// Get driver counts from users table
	err := h.db.QueryRow(`
		SELECT 
			COUNT(*) as total_drivers
		FROM users WHERE role = 'driver'`,
	).Scan(&stats.TotalDrivers)

	if err != nil {
		http.Error(w, "Failed to get driver stats", http.StatusInternalServerError)
		return
	}

	// Set default values for other stats
	stats.AvailableDrivers = stats.TotalDrivers
	stats.BusyDrivers = 0
	stats.OfflineDrivers = 0
	stats.TotalDeliveries = 0
	stats.AverageRating = 4.5

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Placeholder methods for other driver operations
func (h *DriverHandler) GetDriver(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	driverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	var driver models.Driver
	err = h.db.QueryRow(`
		SELECT id, name, email, role, created_at, updated_at
		FROM users WHERE id = $1 AND role = 'driver'`,
		driverID,
	).Scan(&driver.ID, &driver.Name, &driver.Email, &driver.Role, &driver.CreatedAt, &driver.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Driver not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Set default values for driver-specific fields
	driver.Status = "available"
	driver.Rating = 4.5
	driver.TotalDeliveries = 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

func (h *DriverHandler) CreateDriver(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can create drivers
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req models.CreateDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create driver user
	var driver models.Driver
	err = h.db.QueryRow(`
		INSERT INTO users (name, email, password_hash, role) 
		VALUES ($1, $2, $3, 'driver') 
		RETURNING id, name, email, role, created_at, updated_at`,
		req.Name, req.Email, hashedPassword,
	).Scan(&driver.ID, &driver.Name, &driver.Email, &driver.Role, &driver.CreatedAt, &driver.UpdatedAt)
	
	if err != nil {
		http.Error(w, "Failed to create driver", http.StatusInternalServerError)
		return
	}

	// Set driver-specific fields from request
	driver.Phone = req.Phone
	driver.LicenseNumber = req.LicenseNumber
	driver.VehicleType = req.VehicleType
	driver.VehicleNumber = req.VehicleNumber
	driver.CurrentLocation = req.CurrentLocation
	driver.Status = "available"
	driver.Rating = 4.5
	driver.TotalDeliveries = 0

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(driver)
}

func (h *DriverHandler) UpdateDriver(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can update drivers
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	driverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update driver user
	var driver models.Driver
	err = h.db.QueryRow(`
		UPDATE users SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3 AND role = 'driver'
		RETURNING id, name, email, role, created_at, updated_at`,
		req.Name, req.Email, driverID,
	).Scan(&driver.ID, &driver.Name, &driver.Email, &driver.Role, &driver.CreatedAt, &driver.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Driver not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update driver", http.StatusInternalServerError)
		return
	}

	// Set driver-specific fields from request
	driver.Phone = req.Phone
	driver.LicenseNumber = req.LicenseNumber
	driver.VehicleType = req.VehicleType
	driver.VehicleNumber = req.VehicleNumber
	driver.CurrentLocation = req.CurrentLocation
	driver.Status = req.Status
	driver.Rating = 4.5
	driver.TotalDeliveries = 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

func (h *DriverHandler) DeleteDriver(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can delete drivers
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	driverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM users WHERE id = $1 AND role = 'driver'", driverID)
	if err != nil {
		http.Error(w, "Failed to delete driver", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Driver not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DriverHandler) GetDriverShipments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	driverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, tracking_number, origin, destination, weight, zone_id, 
		       status, customer_id, driver_id, created_at, updated_at
		FROM shipments WHERE driver_id = $1 ORDER BY created_at DESC`,
		driverID,
	)
	if err != nil {
		http.Error(w, "Failed to get driver shipments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var shipments []models.Shipment
	for rows.Next() {
		var s models.Shipment
		err := rows.Scan(&s.ID, &s.TrackingNumber, &s.Origin, &s.Destination, &s.Weight,
			&s.ZoneID, &s.Status, &s.CustomerID, &s.DriverID, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to scan shipment", http.StatusInternalServerError)
			return
		}
		shipments = append(shipments, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shipments)
}


