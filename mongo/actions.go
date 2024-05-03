package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

var client = NewMongoClient()
var db = client.Database("ZakuBot")
var usersColl = db.Collection("Users")
var charactersColl = db.Collection("Characters")
var artworksColl = db.Collection("Artworks")

func RegisterUser(userID string, viewName string, userName string) string {
	var userDoc bson.M
	err := usersColl.FindOne(ctx, bson.M{"userId": userID}).Decode(&userDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			userDoc = bson.M{
				"userId": userID, "viewName": viewName, "username": userName, "money": 0,
				"inventory": map[string]int{}, "wishlist": []string{},
				"lastDropTime": int64(0), "dropOrder": []string{},
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
	updatedUserDoc := bson.M{"$set": bson.M{"viewName": viewName}}
	_, err = usersColl.UpdateOne(ctx, bson.M{"userId": userID}, updatedUserDoc)
	return "User already exists"
}

func DrawCards() ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"owned", false}}}},
		{{"$group", bson.D{{"_id", "$characterId"}, {"doc", bson.D{{"$first", "$$ROOT"}}}}}},
		{{"$sample", bson.D{{"size", 3}}}},
	}

	cursor, err := charactersColl.Aggregate(ctx, pipeline)
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

func GetArtworkInfos(characterId string, artworkId int) (bson.M, error) {
	var artwork bson.M
	err := artworksColl.FindOne(ctx, bson.M{"characterId": characterId, "artworkId": artworkId}).Decode(&artwork)
	if err != nil {
		return nil, err
	}
	return artwork, nil
}

func GetAllArtworks(characterId string) ([]bson.M, error) {
	var artworks []bson.M
	cursor, err := artworksColl.Find(ctx, bson.M{"characterId": characterId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &artworks)
	if err != nil {
		log.Fatal(err)
	}

	return artworks, nil
}

func AddToInventory(userID string, characterID string, artworkID int) {
	var userDoc bson.M
	err := usersColl.FindOne(ctx, bson.M{"userId": userID}).Decode(&userDoc)
	if err != nil {
		log.Fatal(err)
	}
	// Add card with specific artwork ID to user's inventory
	userFilter := bson.M{"userId": userID}
	updateInventoryDoc := bson.M{"$set": bson.M{"inventory." + characterID: artworkID},
		"$push": bson.M{"dropOrder": characterID}}
	_, err = usersColl.UpdateOne(ctx, userFilter, updateInventoryDoc)
	if err != nil {
		log.Fatal(err)
	}

	// Add user to character's owners
	characterFilter := bson.M{"characterId": characterID}
	characterUpdateDoc := bson.M{"$set": bson.M{"owner": userID, "owned": true}}
	_, err = charactersColl.UpdateOne(ctx, characterFilter, characterUpdateDoc)
	if err != nil {
		log.Fatal(err)
	}
}

func GetUser(userID string) (bson.M, error) {
	var user bson.M
	err := usersColl.FindOne(ctx, bson.M{"userId": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func SetUserDropTimer(userId string) {
	updatedTimerDoc := bson.M{"$set": bson.M{"lastDropTime": time.Now().Unix()}}
	usersColl.UpdateOne(ctx, bson.M{"userId": userId}, updatedTimerDoc)
}

func GetLastCardDropped(userId string) bson.M {
	user, _ := GetUser(userId)
	dropOrderInterface := user["dropOrder"].(primitive.A)
	if len(dropOrderInterface) == 0 {
		return nil
	}
	cardsDropOrder := make([]string, len(dropOrderInterface))
	for i, v := range dropOrderInterface {
		cardsDropOrder[i] = v.(string)
	}
	charIdLastDrop := cardsDropOrder[len(cardsDropOrder)-1]
	cardInfo, _ := GetCharInfos(charIdLastDrop)
	return cardInfo
}

func RemoveCardFromInventory(userId string, characterId string) {
	var userDoc bson.M
	err := usersColl.FindOne(ctx, bson.M{"userId": userId}).Decode(&userDoc)
	if err != nil {
		log.Fatal(err)
	}

	// Remove card from user's inventory
	userFilter := bson.M{"userId": userId}
	updateInventoryDoc := bson.M{"$unset": bson.M{"inventory." + characterId: ""},
		"$pull": bson.M{"dropOrder": characterId}}
	usersColl.UpdateOne(ctx, userFilter, updateInventoryDoc)

	// Remove user from card's owners
	characterFilter := bson.M{"characterId": characterId}
	characterUpdateDoc := bson.M{"$set": bson.M{"owner": "", "owned": false}}
	charactersColl.UpdateOne(ctx, characterFilter, characterUpdateDoc)
}

func ChangeBalance(userId string, change int) {
	user, _ := GetUser(userId)
	currentBalance := user["money"].(int32)
	newBalance := int(currentBalance) + change
	userFilter := bson.M{"userId": userId}
	updateBalanceDoc := bson.M{"$set": bson.M{"money": newBalance}}
	_, err := usersColl.UpdateOne(ctx, userFilter, updateBalanceDoc)
	if err != nil {
		log.Fatal(err)
	}
}

func SearchCardByName(charName string) ([]bson.M, error) {
	filter := bson.M{"name": bson.M{"$regex": charName, "$options": "i"}}
	cursor, err := charactersColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func SearchCardBySeries(seriesName string) ([]bson.M, error) {
	filter := bson.M{"series": bson.M{"$regex": seriesName, "$options": "i"}}
	cursor, err := charactersColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
