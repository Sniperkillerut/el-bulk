package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/handlers"
	"github.com/el-bulk/backend/middleware"
)

func main() {
	database := db.Connect()
	defer database.Close()

	productHandler := handlers.NewProductHandler(database)
	adminHandler := handlers.NewAdminHandler(database)
	lookupHandler := handlers.NewLookupHandler()
	settingsHandler := handlers.NewSettingsHandler(database)
	refreshHandler := handlers.NewRefreshHandler(database)

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

		// Admin routes (protected)
		r.Route("/admin", func(r chi.Router) {
			r.Post("/login", adminHandler.Login)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminAuth)

				// Products CRUD
				r.Get("/products", productHandler.List)
				r.Post("/products", productHandler.Create)
				r.Put("/products/{id}", productHandler.Update)
				r.Delete("/products/{id}", productHandler.Delete)

				// Exchange rate settings
				r.Get("/settings", settingsHandler.Get)
				r.Put("/settings", settingsHandler.Update)

				// External card lookup (Scryfall + Pokémon TCG API)
				r.Get("/lookup/mtg", lookupHandler.MTG)
				r.Get("/lookup/pokemon", lookupHandler.Pokemon)

				// Price refresh (manual trigger + scheduled nightly)
				r.Post("/prices/refresh", refreshHandler.Trigger)
			})
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 El Bulk API running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
