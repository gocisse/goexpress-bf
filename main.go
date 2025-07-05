package main

import (
	"log"
	"net/http"

	"goexpress-api/config"
	"goexpress-api/database"
	"goexpress-api/handlers"
	"goexpress-api/middleware"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title GoExpress Delivery Management API
// @version 1.0
// @description A comprehensive API for GoExpress delivery operations
// @termsOfService http://swagger.io/terms/
// @contact.name GoExpress API Support
// @contact.email support@goexpress.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("🚀 Starting GoExpress API Server...")
	log.Printf("📊 Environment: %s", cfg.Environment)
	log.Printf("🔧 Port: %s", cfg.Port)

	// Connect to database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	log.Printf("✅ Connected to GoExpress database")

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatal("❌ Failed to run migrations:", err)
	}

	log.Printf("✅ Database migrations completed")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db.DB, cfg.JWTSecret, cfg.JWTRefreshSecret)
	shipmentHandler := handlers.NewShipmentHandler(db.DB)
	zoneHandler := handlers.NewZoneHandler(db.DB)
	userHandler := handlers.NewUserHandler(db.DB, cfg.JWTSecret)
	customerHandler := handlers.NewCustomerHandler(db.DB)
	driverHandler := handlers.NewDriverHandler(db.DB)

	// Setup router
	r := mux.NewRouter()

	// Apply middleware
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware())

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Auth routes (public)
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Public routes
	api.HandleFunc("/shipments/{tracking_number}", shipmentHandler.GetShipmentByTracking).Methods("GET")
	api.HandleFunc("/quote", shipmentHandler.GetQuote).Methods("POST")
	api.HandleFunc("/zones", zoneHandler.GetZones).Methods("GET")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// User routes (protected)
	protected.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	protected.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	protected.HandleFunc("/users/profile", userHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/users/profile", userHandler.UpdateProfile).Methods("PUT")
	protected.HandleFunc("/users/change-password", userHandler.ChangePassword).Methods("POST")
	protected.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	protected.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	protected.HandleFunc("/users/{id}/reset-password", userHandler.ResetPassword).Methods("POST")

	// Customer routes (protected)
	protected.HandleFunc("/customers", customerHandler.GetCustomers).Methods("GET")
	protected.HandleFunc("/customers", customerHandler.CreateCustomer).Methods("POST")
	protected.HandleFunc("/customers/stats", customerHandler.GetCustomerStats).Methods("GET")
	protected.HandleFunc("/customers/{id}", customerHandler.GetCustomer).Methods("GET")
	protected.HandleFunc("/customers/{id}", customerHandler.UpdateCustomer).Methods("PUT")
	protected.HandleFunc("/customers/{id}", customerHandler.DeleteCustomer).Methods("DELETE")
	protected.HandleFunc("/customers/{id}/shipments", customerHandler.GetCustomerShipments).Methods("GET")
	protected.HandleFunc("/customers/{id}/addresses", customerHandler.AddCustomerAddress).Methods("POST")

	// Driver routes (protected)
	protected.HandleFunc("/drivers", driverHandler.GetDrivers).Methods("GET")
	protected.HandleFunc("/drivers", driverHandler.CreateDriver).Methods("POST")
	protected.HandleFunc("/drivers/stats", driverHandler.GetDriverStats).Methods("GET")
	protected.HandleFunc("/drivers/{id}", driverHandler.GetDriver).Methods("GET")
	protected.HandleFunc("/drivers/{id}", driverHandler.UpdateDriver).Methods("PUT")
	protected.HandleFunc("/drivers/{id}", driverHandler.DeleteDriver).Methods("DELETE")
	protected.HandleFunc("/drivers/{id}/shipments", driverHandler.GetDriverShipments).Methods("GET")

	// Shipment routes (protected)
	protected.HandleFunc("/shipments", shipmentHandler.GetShipments).Methods("GET")
	protected.HandleFunc("/shipments", shipmentHandler.CreateShipment).Methods("POST")
	protected.HandleFunc("/shipments/{id}", shipmentHandler.GetShipmentById).Methods("GET")
	protected.HandleFunc("/shipments/{id}/tracking-history", shipmentHandler.GetTrackingHistory).Methods("GET")
	protected.HandleFunc("/shipments/{id}/status", shipmentHandler.UpdateShipmentStatus).Methods("PUT")

	// Admin-only routes
	admin := protected.PathPrefix("").Subrouter()
	admin.Use(middleware.RequireRole("admin"))

	// Zone management (admin only)
	admin.HandleFunc("/zones", zoneHandler.CreateZone).Methods("POST")
	admin.HandleFunc("/zones/{id}", zoneHandler.UpdateZone).Methods("PUT")
	admin.HandleFunc("/zones/{id}", zoneHandler.DeleteZone).Methods("DELETE")

	// Swagger documentation
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"goexpress-api","version":"1.0.0"}`))
	}).Methods("GET")

	// Root endpoint
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Welcome to GoExpress Delivery API","version":"1.0.0","docs":"/swagger/index.html"}`))
	}).Methods("GET")

	log.Printf("🌐 GoExpress API Server starting on port %s", cfg.Port)
	log.Printf("📚 Swagger documentation: http://localhost:%s/swagger/index.html", cfg.Port)
	log.Printf("🏥 Health check: http://localhost:%s/health", cfg.Port)
	
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}


