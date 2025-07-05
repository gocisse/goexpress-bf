package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"goexpress-api/models"
	"goexpress-api/utils"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	db        *sql.DB
	validator *validator.Validate
	jwtSecret string
	refreshSecret string
}

func NewAuthHandler(db *sql.DB, jwtSecret, refreshSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		validator: validator.New(),
		jwtSecret: jwtSecret,
		refreshSecret: refreshSecret,
	}
}

// @Summary User registration
// @Description Register a new user with GoExpress
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.UserRegistration true "User registration data"
// @Success 201 {object} models.AuthResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegistration
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

	// Generate tokens
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, user.Role, h.refreshSecret)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// @Summary User login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.UserLogin true "User login credentials"
// @Success 200 {object} models.AuthResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user from database
	var user models.User
	err := h.db.QueryRow(`
		SELECT id, name, email, password_hash, role, created_at, updated_at 
		FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Validate password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, user.Role, h.refreshSecret)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}