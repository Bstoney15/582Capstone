package routes

// Author: Test scaffolding for handler unit tests.
// Provides an in-memory SQLite database, model migrations, and fixture
// seeding helpers shared by every *_test.go file in this package.

import (
	"backend/libraries/apiauth"
	"backend/libraries/sessionManager"
	"backend/models"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestDB returns an isolated in-memory SQLite database with all application
// models migrated. Each call creates a fresh, independent database.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Unique DSN per call so tests cannot accidentally share state.
	dsn := "file:" + uuid.New().String() + "?mode=memory&cache=shared&_foreign_keys=on"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Merchant{},
		&models.MerchantAPIKey{},
		&models.MerchantWebhookKey{},
		&models.MerchantAddress{},
		&models.MerchantBusinessProfile{},
		&models.MerchantOwner{},
		&models.MerchantCryptoWallet{},
		&models.Role{},
		&models.User{},
		&models.Customer{},
		&models.Deposit{},
		&models.Invoice{},
		&models.XRPLCheckpoint{},
		&models.XRPLPayment{},
		&models.WebhookLog{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}

	return db
}

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	return &Handler{DB: newTestDB(t)}
}

func seedMerchant(t *testing.T, db *gorm.DB) models.Merchant {
	t.Helper()
	m := models.Merchant{MerchantID: uuid.New().String(), MerchantName: "Test Merchant"}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("seed merchant: %v", err)
	}
	return m
}

// seedMerchantAPIKey returns the plaintext API key (the hash is stored).
func seedMerchantAPIKey(t *testing.T, db *gorm.DB, merchantID string) string {
	t.Helper()
	plaintext := "test_" + uuid.New().String()
	digest := sha256.Sum256([]byte(plaintext))
	record := models.MerchantAPIKey{
		MerchantAPIKeyID:         uuid.New().String(),
		MerchantAPIKeyName:       "test key",
		MerchantAPIKeyHashed:     hex.EncodeToString(digest[:]),
		MerchantAPIKeyMerchantID: merchantID,
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("seed api key: %v", err)
	}
	return plaintext
}

func seedCustomer(t *testing.T, db *gorm.DB, merchantID string) models.Customer {
	t.Helper()
	c := models.Customer{
		CustomerID:         uuid.New().String(),
		CustomerMerchantID: merchantID,
		CustomerFirstName:  "Test",
		CustomerLastName:   "Customer",
		CustomerEmail:      "test@example.com",
	}
	if err := db.Create(&c).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	return c
}

// callWithAPIKey wraps a handler in the real RequireMerchantAPIKey middleware
// and executes the request, exercising the production auth path.
func callWithAPIKey(db *gorm.DB, handler http.HandlerFunc, r *http.Request, apiKey string) *httptest.ResponseRecorder {
	if apiKey != "" {
		r.Header.Set("X-API-Key", apiKey)
	}
	rec := httptest.NewRecorder()
	apiauth.RequireMerchantAPIKey(db, handler).ServeHTTP(rec, r)
	return rec
}

// callDirect executes a handler without middleware (no merchant id in context).
func callDirect(handler http.HandlerFunc, r *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
	return rec
}

// seedUser inserts a User row with a bcrypt-hashed password and returns it
// alongside the plaintext password (for login flows).
func seedUser(t *testing.T, db *gorm.DB, username, password string) models.User {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	u := models.User{
		UserID:           uuid.New().String(),
		UserUsername:     username,
		UserFirstName:    "First",
		UserLastName:     "Last",
		UserPasswordHash: string(hashed),
		UserStatus:       "active",
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

// seedRole attaches a user to a merchant with the given role name.
func seedRole(t *testing.T, db *gorm.DB, userID, merchantID, name string) models.Role {
	t.Helper()
	role := models.Role{
		RoleID:         uuid.New().String(),
		RoleMerchantID: merchantID,
		RoleUserID:     userID,
		RoleName:       name,
	}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	return role
}

// loginAs returns a session cookie attached to a request, simulating an
// authenticated user. The caller passes the cookie via r.AddCookie.
func loginAs(userID string) *http.Cookie {
	token := sessionManager.CreateSession(userID)
	return &http.Cookie{Name: sessionManager.SessionCookieName, Value: token, Path: "/"}
}
