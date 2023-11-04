package api

import (
	"Fortune_Tracker_API/api/ledger"
	"Fortune_Tracker_API/api/transaction"
	"Fortune_Tracker_API/api/user"
	"Fortune_Tracker_API/config"
	"Fortune_Tracker_API/internal/auth"
	"Fortune_Tracker_API/internal/validator"
	"Fortune_Tracker_API/pkg/logger"
	"Fortune_Tracker_API/pkg/mariadb"
	"Fortune_Tracker_API/pkg/mongodb"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func Main() {
	// Init API
	apiInit()
	Quit := make(chan os.Signal, 1)

	// Create gin router
	r := gin.Default()

	// Users (no token validation)
	r.POST("/user", user.Register)
	r.POST("/user/login", user.Login)

	// Auth middleware for all routes below
	r.Use(auth.ValidateToken)

	// Users
	r.GET("/user/:uuid", user.Get)
	r.PUT("/user/", user.Update)

	// Ledger
	r.GET("/ledger", ledger.Get)
	r.POST("/ledger", ledger.Create)

	ledgerRoutes := r.Group("/ledger/:ulid")
	ledgerRoutes.Use(validator.ValidateULIDParam)
	{
		// ledger info
		ledgerRoutes.PATCH("/", ledger.Update)

		// Ledger members
		ledgerRoutes.POST("/member", ledger.AddMember)
		ledgerRoutes.PATCH("/member", ledger.UpdateNickname)
		ledgerRoutes.DELETE("/member", ledger.RemoveMember)

		// Ledger transactions
		ledgerRoutes.POST("/transaction", transaction.Create)
		ledgerRoutes.DELETE("/transaction/:utid", transaction.Delete)
		ledgerRoutes.GET("/transaction/:utid", transaction.Get)
		ledgerRoutes.GET("/transaction/time", transaction.GetByTime)
		ledgerRoutes.PUT("/transaction/:utid", transaction.Update)
	}

	// Start API service
	srv := &http.Server{
		Addr:    ":" + config.Viper.GetString("API_PORT"),
		Handler: r,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("[API] " + err.Error())
		os.Exit(1)
	}

	// Graceful shutdown
	signal.Notify(Quit, syscall.SIGINT, syscall.SIGTERM)
	<-Quit
	logger.Info("[API] Shutting down server...")
	if err := srv.Shutdown(nil); err != nil {
		logger.Error("[API] Error shutting down API server: " + err.Error())
		os.Exit(1)
	}
	if err := mariadb.Disconnect(); err != nil {
		logger.Error("[MARIADB] Error closing DB: " + err.Error())
		os.Exit(1)
	}
	if err := mongodb.Disconnect(); err != nil {
		logger.Error("[MONGODB] Error closing DB: " + err.Error())
		os.Exit(1)
	}
	logger.Info("[API] Server exited properly")
}

func apiInit() {
	config.LoadConfig() // Load config
	auth.SetJWTKey()    // Set JWT key
	logger.InitLogger() // Init logger
	ginInit()           // Init gin

	// Connect to MariaDB
	var err error
	if err = mariadb.Connect(); err != nil {
		logger.Error("[MARIADB] " + err.Error())
		return
	}

	// Connect to MongoDB
	if err = mongodb.Connect(); err != nil {
		logger.Error("[MONGODB] " + err.Error())
		return
	}
}

func ginInit() {
	// Set gin log path
	// gin.SetMode(gin.ReleaseMode)
	f, _ := os.Create(config.Viper.GetString("GIN_LOG_PATH"))
	gin.DefaultWriter = io.MultiWriter(f)
}
