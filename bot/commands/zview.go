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
	cardInfo := mongo.GetLastCardDropped(userId)
	if cardInfo == nil {
		viewCardMessage := discordgo.MessageSend{
			Content: fmt.Sprintf("<@%s> You have no cards in your inventory, go drop some with `zdrop`.", userId),
		}
		return viewCardMessage
	}
	charIdLastDrop := cardInfo["characterId"].(string)
	artworksSaved := user["inventory"].(primitive.M)[charIdLastDrop].(int32)
	response := viewCardMessage(cardInfo, int(artworksSaved))
	newResponseContent := fmt.Sprintf("<@%s>\n", userId) + response.Content
	response.Content = newResponseContent
	return response
}

func viewCardMessage(cardInfo bson.M, artworkId int) discordgo.MessageSend {
	charId := cardInfo["characterId"].(string)
	artworkImage, err := openArtworkImage(charId, artworkId)
	if err != nil {
		response := discordgo.MessageSend{
			Content: "Error printing card",
		}
		return response
	}
	cardOwnersInterface := cardInfo["owners"].(primitive.A)
	cardOwners := make([]string, len(cardOwnersInterface))
	for i, v := range cardOwnersInterface {
		cardOwners[i] = v.(string)
	}
	pingOwnersString := createUsernamesMessage(cardOwners)
	response := discordgo.MessageSend{
		Content: fmt.Sprintf("Name: %s", cardInfo["name"]) +
			fmt.Sprintf("\nSeries: %s", cardInfo["series"]) +
			"\nOwners: " + pingOwnersString +
			fmt.Sprintf("\nCharacter ID: %s", cardInfo["characterId"]),
		Files: []*discordgo.File{artworkImage},
	}
	return response
}

func ViewSpecifiedCard(userId string, searchFilter string) ([]discordgo.MessageSend, int) {
	var matchingCardsMessages []discordgo.MessageSend
	// try to find card by characterId
	card, err := mongo.GetCharInfos(searchFilter)
	if err == nil {
		response := viewCardMessage(card, 1)
		newResponseContent := fmt.Sprintf("<@%s>\n", userId) + response.Content
		response.Content = newResponseContent
		matchingCardsMessages = append(matchingCardsMessages, response)
		return matchingCardsMessages, 1
	}
	// try to find card by its name
	cards, err := mongo.SearchCardByName(searchFilter)
	if err == nil {
		for _, cardInfo := range cards {
			response := viewCardMessage(cardInfo, 1)
			newResponseContent := fmt.Sprintf("<@%s>\n", userId) + response.Content
			response.Content = newResponseContent
			matchingCardsMessages = append(matchingCardsMessages, response)
		}
		return matchingCardsMessages, len(cards)
	}
	// try to find card by its series
	cards, err = mongo.SearchCardBySeries(searchFilter)
	if err == nil {
		for _, cardInfo := range cards {
			response := viewCardMessage(cardInfo, 1)
			newResponseContent := fmt.Sprintf("<@%s>\n", userId) + response.Content
			response.Content = newResponseContent
			matchingCardsMessages = append(matchingCardsMessages, response)
		}
		return matchingCardsMessages, len(cards)
	}
	// if no match then respond a fail
	response := discordgo.MessageSend{
		Content: fmt.Sprintf("<@%s>\n", userId) +
			fmt.Sprintf("Couldn't find card matching `%s` criteria", searchFilter),
	}
	return []discordgo.MessageSend{response}, 0
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
