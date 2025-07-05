package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"goexpress-api/handlers"
	"goexpress-api/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Register(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handlers.NewAuthHandler(db.DB, "test-secret", "test-refresh-secret")

	// Test successful registration
	t.Run("successful registration", func(t *testing.T) {
		user := models.UserRegistration{
			Name:     "Test User",
			Email:    "test@goexpress.com",
			Password: "password123",
			Role:     "client",
		}

		jsonData, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.Register(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response models.AuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, user.Email, response.User.Email)
		assert.Equal(t, user.Role, response.User.Role)
	})

	// Test invalid email
	t.Run("invalid email", func(t *testing.T) {
		user := models.UserRegistration{
			Name:     "Test User",
			Email:    "invalid-email",
			Password: "password123",
			Role:     "client",
		}

		jsonData, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	// Test duplicate email
	t.Run("duplicate email", func(t *testing.T) {
		user := models.UserRegistration{
			Name:     "Test User 2",
			Email:    "test@goexpress.com", // Same email as first test
			Password: "password123",
			Role:     "client",
		}

		jsonData, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.Register(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := handlers.NewAuthHandler(db.DB, "test-secret", "test-refresh-secret")

	// First, register a user
	user := models.UserRegistration{
		Name:     "Test User",
		Email:    "login@goexpress.com",
		Password: "password123",
		Role:     "client",
	}

	jsonData, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Register(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Test successful login
	t.Run("successful login", func(t *testing.T) {
		loginData := models.UserLogin{
			Email:    "login@goexpress.com",
			Password: "password123",
		}

		jsonData, _ := json.Marshal(loginData)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.AuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, loginData.Email, response.User.Email)
	})

	// Test invalid credentials
	t.Run("invalid credentials", func(t *testing.T) {
		loginData := models.UserLogin{
			Email:    "login@goexpress.com",
			Password: "wrongpassword",
		}

		jsonData, _ := json.Marshal(loginData)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	// Test non-existent user
	t.Run("non-existent user", func(t *testing.T) {
		loginData := models.UserLogin{
			Email:    "nonexistent@goexpress.com",
			Password: "password123",
		}

		jsonData, _ := json.Marshal(loginData)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}