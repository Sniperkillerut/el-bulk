package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/handlers"
	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/storage"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found or error loading it: %v. Relying on system environment variables.", err)
	}

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
	tcgHandler := handlers.NewTCGHandler(database)
	bountyHandler := handlers.NewBountyHandler(database)
	healthHandler := handlers.NewHealthHandler(database)
	accountingHandler := handlers.NewAccountingHandler(database)
	translationHandler := handlers.NewTranslationHandler(database)
	
	// Initialize Storage Backend
	var storageDriver storage.StorageDriver
	storageType := os.Getenv("STORAGE_TYPE")
	ctx := context.Background()

	switch storageType {
	case "gcp":
		bucket := os.Getenv("GCP_BUCKET_NAME")
		if bucket == "" {
			logger.Warn("STORAGE_TYPE is gcp but GCP_BUCKET_NAME is not set")
		} else {
			driver, err := storage.NewGCPDriver(ctx, bucket)
			if err != nil {
				logger.Error("Failed to initialize GCP storage: %v", err)
			} else {
				storageDriver = driver
				logger.Info("✅ Cloud Storage: Google Cloud Storage initialized (Bucket: %s)", bucket)
			}
		}
	case "s3":
		bucket := os.Getenv("AWS_S3_BUCKET")
		region := os.Getenv("AWS_REGION")
		if bucket == "" || region == "" {
			logger.Warn("STORAGE_TYPE is s3 but AWS_S3_BUCKET or AWS_REGION is not set")
		} else {
			driver, err := storage.NewS3Driver(ctx, bucket, region)
			if err != nil {
				logger.Error("Failed to initialize S3 storage: %v", err)
			} else {
				storageDriver = driver
				logger.Info("✅ Cloud Storage: AWS S3 initialized (Bucket: %s, Region: %s)", bucket, region)
			}
		}
	default:
		logger.Warn("⚠️ No Cloud Storage configured (STORAGE_TYPE empty or unknown). Image uploads will be disabled.")
	}
	
	uploadHandler := handlers.NewUploadHandler(storageDriver)

	// Start nightly price refresh at midnight
	handlers.StartMidnightScheduler(database)

	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestSize(1 << 20)) // 1MB global body limit
	r.Use(middleware.CORS)
	r.Use(middleware.SecurityHeaders)

	// Health check
	r.Get("/health", healthHandler.Ping)

	// Public API
	r.Route("/api", func(r chi.Router) {
		r.Get("/products", productHandler.List)
		r.Post("/products/search-deck", productHandler.BulkSearch)
		r.Get("/products/{id}", productHandler.GetByID)
		r.Get("/tcgs", productHandler.ListTCGs)
		r.Get("/categories", categoriesHandler.List)
		r.Get("/settings", settingsHandler.Get)
		
		themeHandler := handlers.NewThemeHandler(database)
		r.Get("/themes", themeHandler.List)
		
		r.Get("/api/settings", settingsHandler.PublicGet)
		r.Get("/api/translations", translationHandler.List)
		
		r.Get("/translations", translationHandler.List)

		r.Get("/bounties", bountyHandler.List)
		r.With(middleware.OptionalUserAuth).Post("/bounties/offers", bountyHandler.SubmitOffer)
		r.With(middleware.RequireUserAuth).Get("/bounties/offers/me", bountyHandler.ListMeOffers)
		r.With(middleware.RequireUserAuth).Delete("/bounties/offers/me/{id}", bountyHandler.CancelMeOffer)
		r.With(middleware.OptionalUserAuth, middleware.RateLimit(5, 10*time.Minute)).Post("/client-requests", bountyHandler.CreateRequest)
		r.With(middleware.RequireUserAuth).Get("/client-requests/me", bountyHandler.ListMeRequests)
		r.With(middleware.RequireUserAuth).Delete("/client-requests/me/{id}", bountyHandler.CancelMeRequest)
		
		// Newsletter
		newsletterHandler := &handlers.NewsletterHandler{DB: database}
		r.With(middleware.RateLimit(3, 30*time.Minute)).Post("/newsletter/subscribe", newsletterHandler.Subscribe)

		// Public order creation (with optional user context)
		r.With(middleware.OptionalUserAuth).Post("/orders", orderHandler.Create)
		r.With(middleware.RequireUserAuth).Get("/orders/me", orderHandler.ListMe)
		r.With(middleware.RequireUserAuth).Get("/orders/me/{id}", orderHandler.GetMeDetail)
		r.With(middleware.RequireUserAuth).Post("/orders/me/{id}/cancel", orderHandler.CancelMe)

		// Frontend logging
		logHandler := handlers.NewLogHandler()
		r.Post("/logs", logHandler.Receive)

		// User Auth
		userAuthHandler := handlers.NewUserAuthHandler(database)
		r.Route("/auth", func(r chi.Router) {
			r.With(middleware.RateLimit(10, 5*time.Minute)).Get("/{provider}/login", userAuthHandler.Login)
			r.With(middleware.OptionalUserAuth).Get("/{provider}/callback", userAuthHandler.Callback)
			r.Post("/logout", userAuthHandler.Logout)
			r.With(middleware.RequireUserAuth).Get("/me", userAuthHandler.Me)
			r.With(middleware.RequireUserAuth).Put("/me", userAuthHandler.UpdateMe)
		})

		// Admin routes (protected)
		r.Route("/admin", func(r chi.Router) {
			r.With(middleware.RateLimit(5, 15*time.Minute)).Post("/login", adminHandler.Login)
			r.Post("/logout", adminHandler.Logout)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminAuth)
				r.Get("/stats", healthHandler.GetStats)
				
				// Media Uploads
				r.Post("/upload", uploadHandler.Upload)

				// Themes CRUD
				themeHandler := handlers.NewThemeHandler(database)
				r.Get("/themes", themeHandler.List)
				r.Post("/themes", themeHandler.Create)
				r.Put("/themes/{id}", themeHandler.Update)
				r.Delete("/themes/{id}", themeHandler.Delete)

				// Products CRUD
				r.Get("/products", productHandler.List)
				r.Post("/products", productHandler.Create)
				r.With(chiMiddleware.RequestSize(10 << 20)).Post("/products/bulk", productHandler.BulkCreate)
				r.Put("/products/{id}", productHandler.Update)
				r.Delete("/products/{id}", productHandler.Delete)
				r.Get("/products/low-stock", productHandler.GetLowStock)

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

				// TCG Management
				r.Get("/tcgs", tcgHandler.List)
				r.Post("/tcgs", tcgHandler.Create)
				r.Put("/tcgs/{id}", tcgHandler.Update)
				r.Delete("/tcgs/{id}", tcgHandler.Delete)
				r.Post("/tcgs/sync-sets", tcgHandler.SyncSets)

				// Exchange rate settings
				r.Get("/settings", settingsHandler.Get)
				r.Put("/settings", settingsHandler.Update)

				// Translations
				r.Get("/translations", translationHandler.AdminList)
				r.Put("/translations", translationHandler.Update)
				r.Delete("/translations/{key}", translationHandler.Delete)
				r.Delete("/translations/locales/{locale}", translationHandler.DeleteLocale)

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
				r.Post("/orders/{id}/confirm", orderHandler.Confirm)
				r.Post("/orders/{id}/restore", orderHandler.RestoreStock)

				// Bounties & Client Requests CRUD
				r.Post("/bounties", bountyHandler.Create)
				r.Put("/bounties/{id}", bountyHandler.Update)
				r.Delete("/bounties/{id}", bountyHandler.Delete)
				r.Get("/bounties/offers", bountyHandler.ListOffers)
				r.Put("/bounties/offers/{id}/status", bountyHandler.UpdateOfferStatus)
				r.Get("/client-requests", bountyHandler.ListRequests)
				r.Put("/client-requests/{id}/status", bountyHandler.UpdateRequestStatus)

				// System Health & Stats
				r.Get("/stats", healthHandler.GetStats)

				// Notices (Blog/News) CRUD
				noticeHandler := handlers.NewNoticeHandler(database)
				r.Get("/notices", noticeHandler.AdminList)
				r.Post("/notices", noticeHandler.Create)
				r.Put("/notices/{id}", noticeHandler.Update)
				r.Delete("/notices/{id}", noticeHandler.Delete)

				// CRM - Customers & Subscribers
				customerAdminHandler := &handlers.CustomerAdminHandler{DB: database}
				r.Get("/clients", customerAdminHandler.ListCustomers)
				r.Get("/clients/{id}", customerAdminHandler.GetCustomerDetail)
				r.Post("/clients/{id}/notes", customerAdminHandler.AddNote)

				r.Get("/subscribers", newsletterHandler.AdminGetSubscribers)

				// Accounting
				r.Get("/accounting/export", accountingHandler.ExportCSV)
				r.Get("/accounting/valuation", accountingHandler.GetInventoryValuation)
			})
		})

		// Public Notices
		noticeHandler := handlers.NewNoticeHandler(database)
		r.Get("/notices", noticeHandler.List)
		r.Get("/notices/{slug}", noticeHandler.GetBySlug)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Create context that listens for the interrupt signals from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("🚀 El Bulk API running on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
			os.Exit(1)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signals and notify user of shutdown.
	stop()
	logger.Info("Shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
