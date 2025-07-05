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

type ShipmentHandler struct {
	db        *sql.DB
	validator *validator.Validate
}

func NewShipmentHandler(db *sql.DB) *ShipmentHandler {
	return &ShipmentHandler{
		db:        db,
		validator: validator.New(),
	}
}

// @Summary Get shipment tracking history
// @Description Get tracking history for a shipment
// @Tags shipments
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "Shipment ID"
// @Success 200 {array} models.TrackingUpdate
// @Router /api/shipments/{id}/tracking-history [get]
func (h *ShipmentHandler) GetTrackingHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shipmentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	// Get tracking updates
	rows, err := h.db.Query(`
		SELECT id, shipment_id, status, location, timestamp, created_at 
		FROM tracking_updates WHERE shipment_id = $1 ORDER BY timestamp DESC`,
		shipmentID,
	)
	if err != nil {
		http.Error(w, "Failed to get tracking updates", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var trackingUpdates []models.TrackingUpdate
	for rows.Next() {
		var tu models.TrackingUpdate
		err := rows.Scan(&tu.ID, &tu.ShipmentID, &tu.Status, &tu.Location, &tu.Timestamp, &tu.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan tracking update", http.StatusInternalServerError)
			return
		}
		trackingUpdates = append(trackingUpdates, tu)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trackingUpdates)
}

// @Summary Get shipment by ID
// @Description Get shipment details by ID
// @Tags shipments
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "Shipment ID"
// @Success 200 {object} models.ShipmentResponse
// @Router /api/shipments/{id} [get]
func (h *ShipmentHandler) GetShipmentById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shipmentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	// Get shipment
	var shipment models.Shipment
	err = h.db.QueryRow(`
		SELECT id, tracking_number, origin, destination, weight, zone_id, 
		       status, customer_id, driver_id, created_at, updated_at 
		FROM shipments WHERE id = $1`,
		shipmentID,
	).Scan(&shipment.ID, &shipment.TrackingNumber, &shipment.Origin, &shipment.Destination,
		&shipment.Weight, &shipment.ZoneID, &shipment.Status, &shipment.CustomerID,
		&shipment.DriverID, &shipment.CreatedAt, &shipment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Shipment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get tracking updates
	rows, err := h.db.Query(`
		SELECT id, shipment_id, status, location, timestamp, created_at 
		FROM tracking_updates WHERE shipment_id = $1 ORDER BY timestamp DESC`,
		shipment.ID,
	)
	if err != nil {
		http.Error(w, "Failed to get tracking updates", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var trackingUpdates []models.TrackingUpdate
	for rows.Next() {
		var tu models.TrackingUpdate
		err := rows.Scan(&tu.ID, &tu.ShipmentID, &tu.Status, &tu.Location, &tu.Timestamp, &tu.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan tracking update", http.StatusInternalServerError)
			return
		}
		trackingUpdates = append(trackingUpdates, tu)
	}

	// Get zone info
	var zone models.Zone
	err = h.db.QueryRow(`
		SELECT id, name, price_per_kg, created_at, updated_at 
		FROM zones WHERE id = $1`,
		shipment.ZoneID,
	).Scan(&zone.ID, &zone.Name, &zone.PricePerKg, &zone.CreatedAt, &zone.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to get zone info", http.StatusInternalServerError)
		return
	}

	response := models.ShipmentResponse{
		Shipment:       shipment,
		TrackingUpdate: trackingUpdates,
		Zone:           zone,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
// @Summary Get all shipments
// @Description Get all shipments (filtered by user role)
// @Tags shipments
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} models.Shipment
// @Router /api/shipments [get]
func (h *ShipmentHandler) GetShipments(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var query string
	var args []interface{}

	switch claims.Role {
	case "admin":
		query = `SELECT id, tracking_number, origin, destination, weight, zone_id, 
				 status, customer_id, driver_id, created_at, updated_at FROM shipments ORDER BY created_at DESC`
	case "driver":
		query = `SELECT id, tracking_number, origin, destination, weight, zone_id, 
				 status, customer_id, driver_id, created_at, updated_at FROM shipments 
				 WHERE driver_id = $1 ORDER BY created_at DESC`
		args = append(args, claims.UserID)
	default: // client
		query = `SELECT id, tracking_number, origin, destination, weight, zone_id, 
				 status, customer_id, driver_id, created_at, updated_at FROM shipments 
				 WHERE customer_id = $1 ORDER BY created_at DESC`
		args = append(args, claims.UserID)
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
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

// @Summary Create a new shipment
// @Description Create a new shipment with GoExpress
// @Tags shipments
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param shipment body models.ShipmentRequest true "Shipment data"
// @Success 201 {object} models.Shipment
// @Router /api/shipments [post]
func (h *ShipmentHandler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate tracking number with GoExpress prefix
	trackingNumber, err := utils.GenerateTrackingNumber()
	if err != nil {
		http.Error(w, "Failed to generate tracking number", http.StatusInternalServerError)
		return
	}

	// Create shipment
	var shipment models.Shipment
	err = h.db.QueryRow(`
		INSERT INTO shipments (tracking_number, origin, destination, weight, zone_id, customer_id, status) 
		VALUES ($1, $2, $3, $4, $5, $6, 'pending') 
		RETURNING id, tracking_number, origin, destination, weight, zone_id, status, customer_id, driver_id, created_at, updated_at`,
		trackingNumber, req.Origin, req.Destination, req.Weight, req.ZoneID, claims.UserID,
	).Scan(&shipment.ID, &shipment.TrackingNumber, &shipment.Origin, &shipment.Destination,
		&shipment.Weight, &shipment.ZoneID, &shipment.Status, &shipment.CustomerID,
		&shipment.DriverID, &shipment.CreatedAt, &shipment.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to create shipment", http.StatusInternalServerError)
		return
	}

	// Create initial tracking update
	_, err = h.db.Exec(`
		INSERT INTO tracking_updates (shipment_id, status, location) 
		VALUES ($1, $2, $3)`,
		shipment.ID, "pending", req.Origin,
	)
	if err != nil {
		http.Error(w, "Failed to create tracking update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shipment)
}

// @Summary Get shipment by tracking number
// @Description Get shipment details by tracking number (public endpoint)
// @Tags shipments
// @Produce json
// @Param tracking_number path string true "Tracking number"
// @Success 200 {object} models.ShipmentResponse
// @Router /api/shipments/{tracking_number} [get]
func (h *ShipmentHandler) GetShipmentByTracking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trackingNumber := vars["tracking_number"]

	if !utils.ValidateTrackingNumber(trackingNumber) {
		http.Error(w, "Invalid tracking number format", http.StatusBadRequest)
		return
	}

	// Get shipment
	var shipment models.Shipment
	err := h.db.QueryRow(`
		SELECT id, tracking_number, origin, destination, weight, zone_id, 
		       status, customer_id, driver_id, created_at, updated_at 
		FROM shipments WHERE tracking_number = $1`,
		trackingNumber,
	).Scan(&shipment.ID, &shipment.TrackingNumber, &shipment.Origin, &shipment.Destination,
		&shipment.Weight, &shipment.ZoneID, &shipment.Status, &shipment.CustomerID,
		&shipment.DriverID, &shipment.CreatedAt, &shipment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Shipment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get tracking updates
	rows, err := h.db.Query(`
		SELECT id, shipment_id, status, location, timestamp, created_at 
		FROM tracking_updates WHERE shipment_id = $1 ORDER BY timestamp DESC`,
		shipment.ID,
	)
	if err != nil {
		http.Error(w, "Failed to get tracking updates", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var trackingUpdates []models.TrackingUpdate
	for rows.Next() {
		var tu models.TrackingUpdate
		err := rows.Scan(&tu.ID, &tu.ShipmentID, &tu.Status, &tu.Location, &tu.Timestamp, &tu.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan tracking update", http.StatusInternalServerError)
			return
		}
		trackingUpdates = append(trackingUpdates, tu)
	}

	// Get zone info
	var zone models.Zone
	err = h.db.QueryRow(`
		SELECT id, name, price_per_kg, created_at, updated_at 
		FROM zones WHERE id = $1`,
		shipment.ZoneID,
	).Scan(&zone.ID, &zone.Name, &zone.PricePerKg, &zone.CreatedAt, &zone.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to get zone info", http.StatusInternalServerError)
		return
	}

	response := models.ShipmentResponse{
		Shipment:       shipment,
		TrackingUpdate: trackingUpdates,
		Zone:           zone,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get shipping quote
// @Description Get shipping quote based on weight and zone
// @Tags shipments
// @Accept json
// @Produce json
// @Param quote body models.QuoteRequest true "Quote request data"
// @Success 200 {object} models.QuoteResponse
// @Router /api/quote [post]
func (h *ShipmentHandler) GetQuote(w http.ResponseWriter, r *http.Request) {
	var req models.QuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get zone info
	var zone models.Zone
	err := h.db.QueryRow(`
		SELECT id, name, price_per_kg, created_at, updated_at 
		FROM zones WHERE id = $1`,
		req.ZoneID,
	).Scan(&zone.ID, &zone.Name, &zone.PricePerKg, &zone.CreatedAt, &zone.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Zone not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	totalPrice := req.Weight * zone.PricePerKg

	response := models.QuoteResponse{
		Weight:     req.Weight,
		ZoneID:     req.ZoneID,
		ZoneName:   zone.Name,
		PricePerKg: zone.PricePerKg,
		TotalPrice: totalPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Update shipment status
// @Description Update shipment status (admin/driver only)
// @Tags shipments
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Shipment ID"
// @Param status body map[string]string true "Status update"
// @Success 200 {object} models.Shipment
// @Router /api/shipments/{id}/status [put]
func (h *ShipmentHandler) UpdateShipmentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shipmentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status   string `json:"status" validate:"required"`
		Location string `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update shipment status
	_, err = h.db.Exec(`
		UPDATE shipments SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`,
		req.Status, shipmentID,
	)
	if err != nil {
		http.Error(w, "Failed to update shipment", http.StatusInternalServerError)
		return
	}

	// Add tracking update
	_, err = h.db.Exec(`
		INSERT INTO tracking_updates (shipment_id, status, location) 
		VALUES ($1, $2, $3)`,
		shipmentID, req.Status, req.Location,
	)
	if err != nil {
		http.Error(w, "Failed to add tracking update", http.StatusInternalServerError)
		return
	}

	// Get updated shipment
	var shipment models.Shipment
	err = h.db.QueryRow(`
		SELECT id, tracking_number, origin, destination, weight, zone_id, 
		       status, customer_id, driver_id, created_at, updated_at 
		FROM shipments WHERE id = $1`,
		shipmentID,
	).Scan(&shipment.ID, &shipment.TrackingNumber, &shipment.Origin, &shipment.Destination,
		&shipment.Weight, &shipment.ZoneID, &shipment.Status, &shipment.CustomerID,
		&shipment.DriverID, &shipment.CreatedAt, &shipment.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to get updated shipment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shipment)
}

