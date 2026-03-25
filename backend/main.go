package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/handlers"
	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/utils/logger"
)

func main() {
	database, err := db.ConnectResilient()
	if err != nil {
		logger.Warn("Could not connect to database on startup: %v", err)
		logger.Warn("Server starting in degraded mode (some features will be unavailable)")
	} else {
		defer database.Close()
	}

	productHandler := handlers.NewProductHandler(database)
	adminHandler := handlers.NewAdminHandler(database)
	categoriesHandler := handlers.NewCategoriesHandler(database)
	lookupHandler := handlers.NewLookupHandler()
	settingsHandler := handlers.NewSettingsHandler(database)
	refreshHandler := handlers.NewRefreshHandler(database)
	orderHandler := handlers.NewOrderHandler(database)

	// Start nightly price refresh at midnight
	handlers.StartMidnightScheduler(database)

	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.CORS)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Public API
	r.Route("/api", func(r chi.Router) {
		r.Get("/products", productHandler.List)
		r.Get("/products/{id}", productHandler.GetByID)
		r.Get("/tcgs", productHandler.ListTCGs)
		r.Get("/categories", categoriesHandler.List)
		r.Get("/settings", settingsHandler.Get)

		// Public order creation
		r.Post("/orders", orderHandler.Create)

		// Frontend logging
		logHandler := handlers.NewLogHandler()
		r.Post("/logs", logHandler.Receive)

		// Admin routes (protected)
		r.Route("/admin", func(r chi.Router) {
			r.Post("/login", adminHandler.Login)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminAuth)

				// Products CRUD
				r.Get("/products", productHandler.List)
				r.Post("/products", productHandler.Create)
				r.Post("/products/bulk", productHandler.BulkCreate)
				r.Put("/products/{id}", productHandler.Update)
				r.Delete("/products/{id}", productHandler.Delete)

				// Product Storage
				r.Get("/products/{id}/storage", productHandler.GetStorage)
				r.Put("/products/{id}/storage", productHandler.UpdateStorage)

				// Storage Locations CRUD
				storageHandler := handlers.NewStorageHandler(database)
				r.Get("/storage", storageHandler.List)
				r.Post("/storage", storageHandler.Create)
				r.Put("/storage/{id}", storageHandler.Update)
				r.Delete("/storage/{id}", storageHandler.Delete)

				// Custom Categories CRUD
				r.Get("/categories", categoriesHandler.List)
				r.Post("/categories", categoriesHandler.Create)
				r.Put("/categories/{id}", categoriesHandler.Update)
				r.Delete("/categories/{id}", categoriesHandler.Delete)

				// Exchange rate settings
				r.Get("/settings", settingsHandler.Get)
				r.Put("/settings", settingsHandler.Update)

				// External card lookup (Scryfall + Pokémon TCG API)
				r.Get("/lookup/mtg", lookupHandler.MTG)
				r.Post("/lookup/mtg/batch", lookupHandler.BatchMTG)
				r.Get("/lookup/pokemon", lookupHandler.Pokemon)

				// Price refresh (manual trigger + scheduled nightly)
				r.Post("/prices/refresh", refreshHandler.Trigger)

				// Order Management
				r.Get("/orders", orderHandler.List)
				r.Get("/orders/{id}", orderHandler.GetDetail)
				r.Put("/orders/{id}", orderHandler.Update)
				r.Post("/orders/{id}/complete", orderHandler.Complete)
			})
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("🚀 El Bulk API running on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Error("Server failed: %v", err)
		os.Exit(1)
	}
}
