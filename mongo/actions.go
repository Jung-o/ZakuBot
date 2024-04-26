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

func RegisterUser(userID string, userName string) string {
	var userDoc bson.M
	err := usersColl.FindOne(ctx, bson.M{"userId": userID}).Decode(&userDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			userDoc = bson.M{
				"userId": userID, "userName": userName, "money": 0,
				"inventory": []string{}, "wishlist": []string{},
			}
			_, err := usersColl.InsertOne(ctx, userDoc)
			if err != nil {
				log.Fatal(err)
			}
			return "User registered successfully"
		} else {
			log.Fatal(err)
		}
	}
	updatedUserDoc := bson.M{"$set": bson.M{"userName": userName}}
	_, err = usersColl.UpdateOne(ctx, bson.M{"userId": userID}, updatedUserDoc)
	return "User already exists"
}

func DrawCards() ([]bson.M, error) {
	pipeline := mongo.Pipeline{
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

func GetArtworkInfos(artworkId string) (bson.M, error) {
	var artwork bson.M
	err := artworksColl.FindOne(ctx, bson.M{"artworkId": artworkId}).Decode(&artwork)
	if err != nil {
		return nil, err
	}
	return artwork, nil
}

func AddToInventory(userID string, characterID string, artworkID string) {
	var userDoc bson.M
	err := usersColl.FindOne(ctx, bson.M{"userId": userID}).Decode(&userDoc)
	if err != nil {
		log.Fatal(err)
	}
	// Add artwork to user's inventory
	userFilter := bson.M{"userId": userID}
	userUpdateDoc := bson.M{"$push": bson.M{"inventory": artworkID}}
	_, err = usersColl.UpdateOne(ctx, userFilter, userUpdateDoc)
	if err != nil {
		log.Fatal(err)
	}

	// Add user to artwork's owners
	artworkFilter := bson.M{"artworkId": artworkID}
	artworkUpdateDoc := bson.M{"$push": bson.M{"owners": userID}}
	_, err = artworksColl.UpdateOne(ctx, artworkFilter, artworkUpdateDoc)
	if err != nil {
		log.Fatal(err)
	}

	// Add user to character's owners
	characterFilter := bson.M{"characterId": characterID}
	characterUpdateDoc := bson.M{"$push": bson.M{"owners": userID}}
	_, err = charactersColl.UpdateOne(ctx, characterFilter, characterUpdateDoc)
	if err != nil {
		log.Fatal(err)
	}
}
