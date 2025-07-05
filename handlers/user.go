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

type UserHandler struct {
	db        *sql.DB
	validator *validator.Validate
	jwtSecret string
}

func NewUserHandler(db *sql.DB, jwtSecret string) *UserHandler {
	return &UserHandler{
		db:        db,
		validator: validator.New(),
		jwtSecret: jwtSecret,
	}
}

// @Summary Get all users
// @Description Get all users (admin only)
// @Tags users
// @Security ApiKeyAuth
// @Produce json
// @Param role query string false "Filter by role"
// @Success 200 {array} models.User
// @Router /api/users [get]
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can view all users
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	roleFilter := r.URL.Query().Get("role")
	
	var query string
	var args []interface{}

	if roleFilter != "" {
		query = `SELECT id, name, email, role, created_at, updated_at FROM users WHERE role = $1 ORDER BY created_at DESC`
		args = append(args, roleFilter)
	} else {
		query = `SELECT id, name, email, role, created_at, updated_at FROM users ORDER BY created_at DESC`
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to scan user", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// @Summary Get user profile
// @Description Get current user profile
// @Tags users
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} models.User
// @Router /api/users/profile [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := h.db.QueryRow(`
		SELECT id, name, email, role, created_at, updated_at 
		FROM users WHERE id = $1`,
		claims.UserID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Update user profile
// @Description Update current user profile
// @Tags users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param profile body models.UpdateProfileRequest true "Profile data"
// @Success 200 {object} models.User
// @Router /api/users/profile [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if email is already taken by another user
	var existingID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1 AND id != $2", req.Email, claims.UserID).Scan(&existingID)
	if err == nil {
		http.Error(w, "Email already taken", http.StatusConflict)
		return
	}

	// Update user profile
	var user models.User
	err = h.db.QueryRow(`
		UPDATE users SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3 
		RETURNING id, name, email, role, created_at, updated_at`,
		req.Name, req.Email, claims.UserID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Change password
// @Description Change user password
// @Tags users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param password body models.ChangePasswordRequest true "Password data"
// @Success 200 {object} map[string]string
// @Router /api/users/change-password [post]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get current password hash
	var currentPasswordHash string
	err := h.db.QueryRow("SELECT password_hash FROM users WHERE id = $1", claims.UserID).Scan(&currentPasswordHash)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, currentPasswordHash) {
		http.Error(w, "Current password is incorrect", http.StatusBadRequest)
		return
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Update password
	_, err = h.db.Exec(`
		UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`,
		newPasswordHash, claims.UserID,
	)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}

// @Summary Create user (Admin only)
// @Description Create a new user
// @Tags users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User data"
// @Success 201 {object} models.User
// @Router /api/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can create users
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req models.CreateUserRequest
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

	// Create user
	var user models.User
	err = h.db.QueryRow(`
		INSERT INTO users (name, email, password_hash, role) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, email, role, created_at, updated_at`,
		req.Name, req.Email, hashedPassword, req.Role,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// @Summary Update user (Admin only)
// @Description Update a user
// @Tags users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body models.UpdateUserRequest true "User data"
// @Success 200 {object} models.User
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can update users
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if email is already taken by another user
	var existingID int
	err = h.db.QueryRow("SELECT id FROM users WHERE email = $1 AND id != $2", req.Email, userID).Scan(&existingID)
	if err == nil {
		http.Error(w, "Email already taken", http.StatusConflict)
		return
	}

	// Update user
	var user models.User
	err = h.db.QueryRow(`
		UPDATE users SET name = $1, email = $2, role = $3, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $4 
		RETURNING id, name, email, role, created_at, updated_at`,
		req.Name, req.Email, req.Role, userID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Delete user (Admin only)
// @Description Delete a user
// @Tags users
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Success 204
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can delete users
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Prevent admin from deleting themselves
	if userID == claims.UserID {
		http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Reset user password (Admin only)
// @Description Reset a user's password
// @Tags users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param password body models.ResetPasswordRequest true "New password"
// @Success 200 {object} map[string]string
// @Router /api/users/{id}/reset-password [post]
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin can reset passwords
	if claims.Role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Update password
	result, err := h.db.Exec(`
		UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`,
		hashedPassword, userID,
	)
	if err != nil {
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successfully",
	})
}
