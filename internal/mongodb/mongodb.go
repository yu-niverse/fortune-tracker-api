package mongodb

import (
	"Fortune_Tracker_API/config"
	"Fortune_Tracker_API/pkg/logger"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var LedgerCollection *mongo.Collection
var TransactionCollection *mongo.Collection

func Connect() error {
	// Get config values
	dbHost := config.Viper.GetString("MONGODB_HOST")
	dbPort := config.Viper.GetInt("MONGODB_PORT")
	dbUser := config.Viper.GetString("MONGODB_USER")
	dbPass := config.Viper.GetString("MONGODB_PASSWORD")
	options := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d", dbUser, dbPass, dbHost, dbPort))

	// connect to database
	DB, err := mongo.Connect(context.TODO(), options)
	if err != nil {
		return err
	}

	// check connection
	err = DB.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return err
	}

	// Set collection
	LedgerCollection = DB.Database("Fortune_Tracker").Collection("Ledger")
	TransactionCollection = DB.Database("Fortune_Tracker").Collection("Transaction")

	logger.Info("[MONGODB] Successfully connected to MongoDB!")
	return nil
}
