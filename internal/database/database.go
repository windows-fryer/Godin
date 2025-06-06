package database

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var MongoSession *mongo.Client
var MongoContext = context.TODO()

func connectToMongoDB() (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI")).SetServerAPIOptions(serverAPI)

	return mongo.Connect(clientOptions)
}

func Start() {
	log.Info("Database started")

	go func() {
		mongoSession, err := connectToMongoDB()

		if err != nil {
			panic("Failed to connect to MongoDB: " + err.Error())
		}

		MongoSession = mongoSession
	}()
}

func Stop() {
	log.Info("Database stopped")

	if err := MongoSession.Disconnect(MongoContext); err != nil {
		log.Error("Error disconnecting from database: %v", err)
	}
}
