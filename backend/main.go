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
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/storage"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found or error loading it: %v. Relying on system environment variables.", err)
	}

	// Initialize logger level and format
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		logger.SetLevel(logger.ParseLevel(logLevel))
	}
	if os.Getenv("LOG_FORMAT") == "json" {
		logger.SetJSON(true)
	}
	
	// Auto-detect GCP context (overrides manual format if on GCP)
	logger.AutoDetectGCP()

	if os.Getenv("APP_ENV") == "" {
		logger.Warn("⚠️ APP_ENV is not set. CORS and security middleware will run in development mode.")
	}
	
	logger.Info("Logger initialized | Level: %s | Format: %s", 
		logger.Default.GetLevel().String(), 
		func() string { if os.Getenv("LOG_FORMAT") == "json" || os.Getenv("K_SERVICE") != "" { return "JSON" }; return "TEXT" }())

	database, err := db.ConnectResilient()
	if err != nil {
		logger.Warn("Could not connect to database on startup: %v", err)
		logger.Warn("Server starting in degraded mode (some features will be unavailable)")
	} else {
		defer database.Close()
	}

	// Initialize Stores
	categoryStore := store.NewCategoryStore(database)
	themeStore := store.NewThemeStore(database)
	tcgStore := store.NewTCGStore(database)
	noticeStore := store.NewNoticeStore(database)
	productStore := store.NewProductStore(database)
	settingsStore := store.NewSettingsStore(database)
	orderStore := store.NewOrderStore(database)
	customerStore := store.NewCustomerStore(database)
	bountyStore := store.NewBountyStore(database)
	adminStore := store.NewAdminStore(database)
	newsletterStore := store.NewNewsletterStore(database)
	storageLocationStore := store.NewStorageLocationStore(database)
	authStore := store.NewAuthStore(database)
	healthStore := store.NewHealthStore(database)
	refreshStore := store.NewRefreshStore(database)
	accountingStore := store.NewAccountingStore(database)
	translationStore := store.NewTranslationStore(database)
	auditStore := store.NewAuditStore(database)

	translationService := service.NewTranslationService(translationStore)
	auditService := service.NewAuditService(auditStore, adminStore)

	// Initialize Services
	settingsService := service.NewSettingsService(settingsStore)
	productService := service.NewProductService(productStore, tcgStore, settingsService, auditService)
	orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, auditService)
	categoryService := service.NewCategoryService(categoryStore)
	refreshService := service.NewRefreshService(refreshStore, settingsService)
	tcgService := service.NewTCGService(tcgStore, refreshService)
	noticeService := service.NewNoticeService(noticeStore)
	themeService := service.NewThemeService(themeStore)
	bountyService := service.NewBountyService(bountyStore)
	adminService := service.NewAdminService(adminStore)
	newsletterService := service.NewNewsletterService(newsletterStore)
	storageLocationService := service.NewStorageLocationService(storageLocationStore)
	authService := service.NewAuthService(authStore)
	healthService := service.NewHealthService(healthStore)
	// refreshService already initialized above
	accountingService := service.NewAccountingService(accountingStore, settingsService)

	// Initialize Handlers
	productHandler := handlers.NewProductHandler(productService, database)
	adminHandler := handlers.NewAdminHandler(adminService, auditService)
	categoriesHandler := handlers.NewCategoriesHandler(categoryService)
	lookupHandler := handlers.NewLookupHandler(productService)
	settingsHandler := handlers.NewSettingsHandler(settingsService)
	refreshHandler := handlers.NewRefreshHandler(refreshService)
	orderHandler := handlers.NewOrderHandler(orderService)
	tcgHandler := handlers.NewTCGHandler(tcgService)
	bountyHandler := handlers.NewBountyHandler(bountyService)
	healthHandler := handlers.NewHealthHandler(healthService)
	accountingHandler := handlers.NewAccountingHandler(database, accountingService)
	translationHandler := handlers.NewTranslationHandler(translationService)
	themeHandler := handlers.NewThemeHandler(themeService)
	noticeHandler := handlers.NewNoticeHandler(noticeService)
	userAuthHandler := handlers.NewUserAuthHandler(authService)
	newsletterHandler := handlers.NewNewsletterHandler(newsletterService)
	storageLocationHandler := handlers.NewStorageHandler(storageLocationService)

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
				
				// Ensure driver is closed on shutdown
				if closer, ok := storageDriver.(interface{ Close() error }); ok {
					defer func() {
						if err := closer.Close(); err != nil {
							logger.Error("Failed to close storage driver: %v", err)
						} else {
							logger.Info("Storage driver closed successfully")
						}
					}()
				}
			}
		}
	default:
		logger.Warn("⚠️ No Cloud Storage configured (STORAGE_TYPE empty or unknown). Image uploads will be disabled.")
	}
	
	uploadHandler := handlers.NewUploadHandler(storageDriver)

	// Start nightly price refresh at midnight
	handlers.StartMidnightScheduler(refreshService)

	r := chi.NewRouter()

	// Global middleware
	r.Use(logger.RequestLogger)
	r.Use(logger.Recoverer)
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
		
		r.Get("/themes", themeHandler.List)
		
		r.Get("/settings/public", settingsHandler.PublicGet)
		r.Get("/translations", translationHandler.List)

		r.Get("/bounties", bountyHandler.List)
		r.With(middleware.RequireUserAuth).Post("/bounties/offers", bountyHandler.SubmitOffer)
		r.With(middleware.RequireUserAuth).Get("/bounties/offers/me", bountyHandler.ListMeOffers)
		r.With(middleware.RequireUserAuth).Delete("/bounties/offers/me/{id}", bountyHandler.CancelMeOffer)
		r.With(middleware.RequireUserAuth, middleware.RateLimit(5, 10*time.Minute)).Post("/client-requests", bountyHandler.CreateRequest)
		r.With(middleware.RequireUserAuth).Get("/client-requests/me", bountyHandler.ListMeRequests)
		r.With(middleware.RequireUserAuth).Delete("/client-requests/me/{id}", bountyHandler.CancelMeRequest)
		
		// Newsletter
		r.With(middleware.RateLimit(3, 30*time.Minute)).Post("/newsletter/subscribe", newsletterHandler.Subscribe)

		// Public order creation (with optional user context)
		r.With(middleware.RequireUserAuth).Post("/orders", orderHandler.Create)
		r.With(middleware.RequireUserAuth).Get("/orders/me", orderHandler.ListMe)
		r.With(middleware.RequireUserAuth).Get("/orders/me/{id}", orderHandler.GetMeDetail)
		r.With(middleware.RequireUserAuth).Post("/orders/me/{id}/cancel", orderHandler.CancelMe)

		// Frontend logging
		logHandler := handlers.NewLogHandler()
		r.Post("/logs", logHandler.Receive)

		// User Auth
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
			r.Get("/auth/google", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, r.URL.Path+"/login", http.StatusMovedPermanently)
			})
			r.Get("/auth/google/login", adminHandler.GoogleLogin)
			r.Get("/auth/google/callback", adminHandler.GoogleCallback)
			r.Post("/logout", adminHandler.Logout)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminAuth)
				r.Get("/stats", healthHandler.GetStats)
				
				// Media Uploads
				r.Post("/upload", uploadHandler.Upload)

				// Themes CRUD
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
				r.Put("/products/bulk-source", productHandler.BulkUpdateSource)

				// Product Storage
				r.Get("/products/{id}/storage", productHandler.GetStorage)
				r.Put("/products/{id}/storage", productHandler.UpdateStorage)

				// Storage Locations CRUD
				r.Get("/storage", storageLocationHandler.List)
				r.Post("/storage", storageLocationHandler.Create)
				r.Put("/storage/{id}", storageLocationHandler.Update)
				r.Delete("/storage/{id}", storageLocationHandler.Delete)

				// Custom Categories CRUD
				r.Get("/categories", categoriesHandler.List)
				r.Post("/categories", categoriesHandler.Create)
				r.Put("/categories/{id}", categoriesHandler.Update)
				r.Delete("/categories/{id}", categoriesHandler.Delete)

				// TCG Management
				r.Get("/tcgs", tcgHandler.List)
				r.Post("/tcgs", tcgHandler.Create)
				r.Post("/tcgs/sync-sets", tcgHandler.SyncSets)
				r.Put("/tcgs/{id}", tcgHandler.Update)
				r.Delete("/tcgs/{id}", tcgHandler.Delete)
				r.Post("/tcgs/{id}/sync-sets", tcgHandler.SyncSets)
				r.Post("/tcgs/{id}/sync-prices", tcgHandler.SyncPrices)

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
				
				// External Price Lookups
				r.Get("/lookup/external/prices", tcgHandler.GetExternalPrice)

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

				// Notices (Blog/News) CRUD
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

				// Dynamic Logging
				r.Get("/logs/level", adminHandler.GetLogLevel)
				r.Put("/logs/level", adminHandler.UpdateLogLevel)

				// Audit Logs
				r.Get("/audit-logs", adminHandler.ListAuditLogs)
			})
		})

		// Public Notices
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
