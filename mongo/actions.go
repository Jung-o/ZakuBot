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
var charactersColl = db.Collection("Characters")
var artworksColl = db.Collection("Artworks")

func RegisterUser(userID string) string {
	var userDoc bson.M
	err := usersColl.FindOne(ctx, bson.M{"userID": userID}).Decode(&userDoc)
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

func DrawCards() ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"owned", false}}}},
		{{"$group", bson.D{{"_id", "$characterId"}, {"doc", bson.D{{"$first", "$$ROOT"}}}}}},
		{{"$sample", bson.D{{"size", 3}}}},
	}

	cursor, err := artworksColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func GetCharInfos(characterId string) (bson.M, error) {
	var character bson.M
	err := charactersColl.FindOne(ctx, bson.M{"characterId": characterId}).Decode(&character)
	if err != nil {
		return nil, err
	}
	return character, nil
}
