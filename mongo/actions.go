package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

var client = NewMongoClient()
var db = client.Database("ZakuBot")
var usersColl = db.Collection("Users")
var inventoriesColl = db.Collection("Inventories")
var cardsColl = db.Collection("Cards")

func RegisterUser(userID string) string {
	userDoc := bson.M{"userID": userID}
	err := usersColl.FindOne(ctx, userDoc).Decode(&userDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err := usersColl.InsertOne(ctx, userDoc)
			if err != nil {
				log.Fatal(err)
			}
			return "User registered successfully"
		} else {
			log.Fatal(err)
		}
	}
	return "User already exists"
}
