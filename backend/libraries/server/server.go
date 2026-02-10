package server

import (
	routes "backend/libraries/server/handlers"
	"backend/libraries/sessionManager"
	"backend/models"
	"log"
	"net/http"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

// server struct is used to inject dependencies into route handlers
type Server struct {
	Router         *http.ServeMux
	SessionManager *sessionManager.SessionManager
	DB             *gorm.DB
}

// returnes a new server instance. all dependencies should be initialized here
func NewServer() *Server {

	driver := sqlite.Open("./xrpay.db")

	if(os.Getenv("production") == "true") {
		log.Println("Running in production mode")
		dsn := "root:secret@tcp(mysql-container:3306)/xrpay?charset=utf8mb4&parseTime=True&loc=Local"
		driver = mysql.Open(dsn)
	}


	db, err := gorm.Open(driver, &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	return &Server{
		Router:         http.NewServeMux(),
		SessionManager: sessionManager.NewSessionManager(),
		DB:             db,
	}
}

// start the servers and sets up any dependencies that are needed.
func (s *Server) Start(address string) error {
	routes.AddRoutesToServer(s.Router)
	s.AutoMigrateDB()
	return http.ListenAndServe(address, s.Router)
}

func (s *Server) AutoMigrateDB() error {
	return s.DB.AutoMigrate(
		&models.Merchant{},
		&models.MerchantAPIKey{},
		&models.MerchantAPIKey{},
		&models.MerchantAddress{},
		&models.MerchantBusinessProfile{},
		&models.MerchantOwner{},
		&models.MerchantCryptoWallet{},

		&models.Role{},
		&models.User{},
		&models.Customer{},
		&models.Deposit{},
		&models.Invoice{},
	)
}
