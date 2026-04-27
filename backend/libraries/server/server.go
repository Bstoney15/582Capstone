// Authors: Ben Stonestreet
// Created: 02/02/26
// Description: main server struct

// Package server wires together the HTTP router, database connection, GORM
// migrations, development seed data, and the XRPL reconciler into a single
// Server type that the main entry-point uses to start the application.
package server

import (
	routes "backend/libraries/server/handlers"
	"backend/models"
	"log"
	"net/http"
	"strings"

	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Server holds the shared dependencies that are injected into every route
// handler. Keeping them here (rather than as package-level globals) makes the
// application easier to test and avoids hidden coupling between packages.
type Server struct {
	Router *http.ServeMux
	DB     *gorm.DB
}

// NewServer initialises the database connection and returns a Server ready to
// be started. The database driver is chosen at runtime:
//   - SQLite (file ./xrpay.db) is used by default for local development.
//   - MySQL is used when the "production" environment variable is set to "true".
func NewServer() *Server {

	driver := sqlite.Open("./xrpay.db")

	dsn := os.Getenv("DATABASE_URL")

	if dsn != "" {
		log.Println("Using MySQL (DATABASE_URL detected)")
		driver = mysql.Open(dsn)
	} else {
		log.Println("Using SQLite (fallback)")
		driver = sqlite.Open("./xrpay.db")
	}

	db, err := gorm.Open(driver, &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	return &Server{
		Router: http.NewServeMux(),
		DB:     db,
	}
}

// Start registers all HTTP routes, runs database migrations, optionally seeds
// development data, launches the XRPL payment reconciler, and begins listening
// on the given address (e.g. ":8080"). It blocks until the server exits.
func (s *Server) Start(address string) error {
	h := routes.NewHandler(s.DB)
	h.RegisterRoutes(s.Router)

	if err := s.AutoMigrateDB(); err != nil {
		return err
	}

	if os.Getenv("PRODUCTION") == "true" {
		if err := s.EnsureProductionDemoMerchantSeeded(); err != nil {
			return err
		}
	} else {
		if err := s.SeedDevData(); err != nil {
			return err
		}
	}

	// Start the background goroutine that polls the XRPL ledger and reconciles
	// incoming payments against open invoices.
	xrplReconciler := NewXRPLReconciler(s.DB)
	xrplReconciler.Start()

	return http.ListenAndServe(address, s.withWidgetCORS(s.Router))
}

// withWidgetCORS is middleware that adds permissive CORS headers specifically
// for the widget-facing API endpoints (/api/invoices and /api/verify). These
// routes are called directly from the embedded checkout widget on third-party
// merchant storefronts, so they must accept cross-origin requests from any
// origin. All other routes are unaffected and rely on the handler-level CORS
// policy set in handlers/cors.go.
func (s *Server) withWidgetCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Identify requests that originate from the embedded checkout widget.
		isWidgetAPI := r.URL.Path == "/api/invoices" ||
			strings.HasPrefix(r.URL.Path, "/api/invoices/") ||
			r.URL.Path == "/api/verify"

		if isWidgetAPI {
			origin := r.Header.Get("Origin")
			if origin != "" {
				// Reflect the request origin rather than using "*" so that
				// browsers accept responses that include credentials.
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-API-Key")
				w.Header().Set("Access-Control-Max-Age", "600")
			}

			// Respond to CORS pre-flight OPTIONS requests immediately without
			// forwarding them to the underlying handler.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// AutoMigrateDB runs GORM's auto-migration for every model in the application,
// creating or updating tables to match the current struct definitions. Safe to
// call on every startup; GORM will not drop existing columns or data.
func (s *Server) AutoMigrateDB() error {
	return s.DB.AutoMigrate(
		// Merchant and its child entities.
		&models.Merchant{},
		&models.MerchantAPIKey{},
		&models.MerchantWebhookKey{},
		&models.MerchantAddress{},
		&models.MerchantBusinessProfile{},
		&models.MerchantOwner{},
		&models.MerchantCryptoWallet{},

		// Identity and access models.
		&models.Role{},
		&models.User{},

		// Transaction / payment models.
		&models.Customer{},
		&models.Deposit{},
		&models.Invoice{},

		// XRPL ledger reconciliation state.
		&models.XRPLCheckpoint{},
		&models.XRPLPayment{},
	)
}
