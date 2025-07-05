package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"goexpress-api/models"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type ZoneHandler struct {
	db        *sql.DB
	validator *validator.Validate
}

func NewZoneHandler(db *sql.DB) *ZoneHandler {
	return &ZoneHandler{
		db:        db,
		validator: validator.New(),
	}
}

// @Summary Get all zones
// @Description Get all GoExpress shipping zones
// @Tags zones
// @Produce json
// @Success 200 {array} models.Zone
// @Router /api/zones [get]
func (h *ZoneHandler) GetZones(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, name, price_per_kg, created_at, updated_at 
		FROM zones ORDER BY name`,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var zones []models.Zone
	for rows.Next() {
		var z models.Zone
		err := rows.Scan(&z.ID, &z.Name, &z.PricePerKg, &z.CreatedAt, &z.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to scan zone", http.StatusInternalServerError)
			return
		}
		zones = append(zones, z)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(zones)
}

// @Summary Create a new zone
// @Description Create a new GoExpress shipping zone (admin only)
// @Tags zones
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param zone body models.Zone true "Zone data"
// @Success 201 {object} models.Zone
// @Router /api/zones [post]
func (h *ZoneHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	var req models.Zone
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var zone models.Zone
	err := h.db.QueryRow(`
		INSERT INTO zones (name, price_per_kg) 
		VALUES ($1, $2) 
		RETURNING id, name, price_per_kg, created_at, updated_at`,
		req.Name, req.PricePerKg,
	).Scan(&zone.ID, &zone.Name, &zone.PricePerKg, &zone.CreatedAt, &zone.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to create zone", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(zone)
}

// @Summary Update a zone
// @Description Update a GoExpress shipping zone (admin only)
// @Tags zones
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Zone ID"
// @Param zone body models.Zone true "Zone data"
// @Success 200 {object} models.Zone
// @Router /api/zones/{id} [put]
func (h *ZoneHandler) UpdateZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid zone ID", http.StatusBadRequest)
		return
	}

	var req models.Zone
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var zone models.Zone
	err = h.db.QueryRow(`
		UPDATE zones SET name = $1, price_per_kg = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3 
		RETURNING id, name, price_per_kg, created_at, updated_at`,
		req.Name, req.PricePerKg, zoneID,
	).Scan(&zone.ID, &zone.Name, &zone.PricePerKg, &zone.CreatedAt, &zone.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Zone not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update zone", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(zone)
}

// @Summary Delete a zone
// @Description Delete a GoExpress shipping zone (admin only)
// @Tags zones
// @Security ApiKeyAuth
// @Param id path int true "Zone ID"
// @Success 204
// @Router /api/zones/{id} [delete]
func (h *ZoneHandler) DeleteZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid zone ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM zones WHERE id = $1", zoneID)
	if err != nil {
		http.Error(w, "Failed to delete zone", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Zone not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}