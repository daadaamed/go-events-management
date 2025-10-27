package utils

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	DB          *mongo.Database
)

func InitDB() (*mongo.Client, *mongo.Database, error) {
	MONGODB_URI := os.Getenv("MONGODB_URI")
	if MONGODB_URI == "" {
		MONGODB_URI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "eventsdb"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MONGODB_URI))
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping error: %v", err)
	}

	MongoClient = client
	DB = client.Database("eventsdb")
	return MongoClient, DB, nil
}
