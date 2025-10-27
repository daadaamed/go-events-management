package utils

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/lpernett/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	DB          *mongo.Database
)

func InitDB() (*mongo.Client, *mongo.Database, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	uri := os.Getenv("MONGODB_URI")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
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
