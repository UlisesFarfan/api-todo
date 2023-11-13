package database

import (
	"api-todo/config"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database connection
func GetDatabase() *mongo.Database {
	var database *mongo.Database
	config, _ := config.LoadConfig(".")
	fmt.Println(config.DbUrl)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.DbUrl))
	if err != nil {
		panic(err)
	}
	database = client.Database(config.DbName)

	database.Collection("user").Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)

	return database
}
