package commands

import (
	"ZakuBot/mongo"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
)

func ViewLastCardDropped(userId string) discordgo.MessageSend {
	user, _ := mongo.GetUser(userId)
	dropOrderInterface := user["dropOrder"].(primitive.A)
	if len(dropOrderInterface) == 0 {
		viewCardMessage := discordgo.MessageSend{
			Content: fmt.Sprintf("<@%s> You have no cards in your inventory, go drop some with `zdrop`.", userId),
		}
		return viewCardMessage
	}
	cardsDropOrder := make([]string, len(dropOrderInterface))
	for i, v := range dropOrderInterface {
		cardsDropOrder[i] = v.(string)
	}
	charIdLastDrop := cardsDropOrder[len(cardsDropOrder)-1]
	artworksSaved := user["inventory"].(primitive.M)[charIdLastDrop].(int32)
	cardInfo, _ := mongo.GetCharInfos(charIdLastDrop)
	artworkImage, err := openArtworkImage(charIdLastDrop, int(artworksSaved))
	if err != nil {
		viewCardMessage := discordgo.MessageSend{
			Content: "Error printing card",
		}
		return viewCardMessage
	} else {
		cardOwnersInterface := cardInfo["owners"].(primitive.A)
		cardOwners := make([]string, len(cardOwnersInterface))
		for i, v := range cardOwnersInterface {
			cardOwners[i] = v.(string)
		}
		pingOwnersString := createUsernamesMessage(cardOwners)
		viewCardMessage := discordgo.MessageSend{
			Content: fmt.Sprintf("Name: %s", cardInfo["name"]) +
				fmt.Sprintf("\nSeries: %s", cardInfo["series"]) +
				"\nOwners: " + pingOwnersString +
				fmt.Sprintf("\nCharacter ID: %s", cardInfo["characterId"]),
			Files: []*discordgo.File{artworkImage},
		}
		return viewCardMessage
	}
}

func openArtworkImage(charId string, artworkId int) (*discordgo.File, error) {
	imageName := fmt.Sprintf("%s-%d.jpg", charId, artworkId)
	filepath := fmt.Sprintf("files/cards/%s", imageName)
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return &discordgo.File{
		Reader: file,
		Name:   imageName,
	}, nil
}

func createPingAllUsers(usersList []string) string {
	var pingString string
	for _, owner := range usersList {
		pingString = pingString + fmt.Sprintf("<@%s> ", owner)
	}
	return pingString
}

func createUsernamesMessage(usersList []string) string {
	var usernamesMessage string
	var user bson.M
	for _, owner := range usersList {
		user, _ = mongo.GetUser(owner)
		usernamesMessage = usernamesMessage + fmt.Sprintf("%s ", user["username"])
	}
	return usernamesMessage
}
