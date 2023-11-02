package api

import (
	"Fortune_Tracker_API/api/ledger"
	"Fortune_Tracker_API/api/transaction"
	"Fortune_Tracker_API/api/user"
	"Fortune_Tracker_API/config"
	"Fortune_Tracker_API/internal/mariadb"
	"Fortune_Tracker_API/internal/mongodb"
	"Fortune_Tracker_API/pkg/auth"
	"Fortune_Tracker_API/pkg/logger"
	"Fortune_Tracker_API/pkg/validator"
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func Main() {
	// Init API
	initAPI()

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

	r.Run()
}

func initAPI() {
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
