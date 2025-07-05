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
)

type CustomerHandler struct {
	db        *sql.DB
	validator *validator.Validate
}

func NewCustomerHandler(db *sql.DB) *CustomerHandler {
	return &CustomerHandler{
		db:        db,
		validator: validator.New(),
	}
}

// @Summary Get all customers
// @Description Get all customers with stats (admin only)
// @Tags customers
// @Security ApiKeyAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param business_type query string false "Filter by business type"
// @Success 200 {array} models.Customer
// @Router /api/customers [get]
func (h *CustomerHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can view all customers
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	statusFilter := r.URL.Query().Get("status")
	businessTypeFilter := r.URL.Query().Get("business_type")
	
	query := `
		SELECT 
			c.id, c.user_id, c.company_name, c.contact_person, c.phone, 
			c.alternate_phone, c.website, c.tax_id, c.business_type, 
			c.status, c.credit_limit, c.payment_terms, c.notes,
			c.created_at, c.updated_at,
			u.name, u.email,
			COALESCE(s.total_shipments, 0) as total_shipments,
			COALESCE(s.total_spent, 0) as total_spent,
			s.last_shipment
		FROM customers c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN (
			SELECT 
				customer_id,
				COUNT(*) as total_shipments,
				SUM(weight * z.price_per_kg) as total_spent,
				MAX(created_at) as last_shipment
			FROM shipments sh
			JOIN zones z ON sh.zone_id = z.id
			GROUP BY customer_id
		) s ON c.user_id = s.customer_id
		WHERE 1=1`

	var args []interface{}
	argIndex := 1

	if statusFilter != "" {
		query += " AND c.status = $" + strconv.Itoa(argIndex)
		args = append(args, statusFilter)
		argIndex++
	}

	if businessTypeFilter != "" {
		query += " AND c.business_type = $" + strconv.Itoa(argIndex)
		args = append(args, businessTypeFilter)
		argIndex++
	}

	query += " ORDER BY c.created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(
			&c.ID, &c.UserID, &c.CompanyName, &c.ContactPerson, &c.Phone,
			&c.AlternatePhone, &c.Website, &c.TaxID, &c.BusinessType,
			&c.Status, &c.CreditLimit, &c.PaymentTerms, &c.Notes,
			&c.CreatedAt, &c.UpdatedAt,
			&c.Name, &c.Email,
			&c.TotalShipments, &c.TotalSpent, &c.LastShipment,
		)
		if err != nil {
			http.Error(w, "Failed to scan customer", http.StatusInternalServerError)
			return
		}
		customers = append(customers, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

// @Summary Get customer stats
// @Description Get customer statistics (admin only)
// @Tags customers
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} models.CustomerStats
// @Router /api/customers/stats [get]
func (h *CustomerHandler) GetCustomerStats(w http.ResponseWriter, r *http.Request) {
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

	var stats models.CustomerStats

	// Get customer counts
	err := h.db.QueryRow(`
		SELECT 
			COUNT(*) as total_customers,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_customers,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_customers
		FROM customers`,
	).Scan(&stats.TotalCustomers, &stats.ActiveCustomers, &stats.InactiveCustomers)

	if err != nil {
		http.Error(w, "Failed to get customer counts", http.StatusInternalServerError)
		return
	}

	// Get revenue stats
	err = h.db.QueryRow(`
		SELECT 
			COALESCE(SUM(weight * z.price_per_kg), 0) as total_revenue,
			COALESCE(AVG(weight * z.price_per_kg), 0) as average_order_value
		FROM shipments s
		JOIN zones z ON s.zone_id = z.id`,
	).Scan(&stats.TotalRevenue, &stats.AverageOrderValue)

	if err != nil {
		http.Error(w, "Failed to get revenue stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Placeholder methods for other customer operations
func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *CustomerHandler) GetCustomerShipments(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *CustomerHandler) AddCustomerAddress(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}


