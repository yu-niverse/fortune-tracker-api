package mariadb

import (
	"Fortune_Tracker_API/config"
	"Fortune_Tracker_API/pkg/logger"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() error {
	// Get config values
	dbHost := config.Viper.GetString("MARIADB_HOST")
	dbPort := config.Viper.GetInt("MARIADB_PORT")
	dbUser := config.Viper.GetString("MARIADB_USER")
	dbPass := config.Viper.GetString("MARIADB_PASSWORD")
	dbName := config.Viper.GetString("MARIADB_DATABASE")

	// connect to database
	var err error
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	// check connection
	err = DB.Ping()
	if err != nil {
		return err
	}
	
	logger.Info("[MARIADB] Successfully connected to MariaDB!")
	return nil
}

func Disconnect() error {
	err := DB.Close()
	if err != nil {
		return err
	}
	logger.Info("[MARIADB] Successfully disconnected from MariaDB!")
	return nil
}