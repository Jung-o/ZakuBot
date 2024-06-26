package commands

import (
	"ZakuBot/mongo"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"os"
)

func ViewLastCardDropped(userId string) discordgo.MessageSend {
	cardInfo := mongo.GetLastCardDropped(userId)
	if cardInfo == nil {
		viewCardMessage := discordgo.MessageSend{
			Content: fmt.Sprintf("<@%s> You have no cards in your inventory, go drop some with `zdrop`.", userId),
		}
		return viewCardMessage
	}
	artworksSaved := cardInfo["viewedArtwork"].(int32)
	response := ViewCardMessage(cardInfo, int(artworksSaved))
	newResponseContent := fmt.Sprintf("<@%s>\n", userId) + response.Content
	response.Content = newResponseContent
	return response
}

func ViewCardMessage(cardInfo bson.M, artworkId int) discordgo.MessageSend {
	charId := cardInfo["characterId"].(string)
	artworkImage, err := openArtworkImage(charId, artworkId)
	if err != nil {
		response := discordgo.MessageSend{
			Content: "Error printing card",
		}
		return response
	}
	cardOwner := cardInfo["owner"].(string)
	owner, _ := mongo.GetUser(cardOwner)
	var ownerUsername string
	if owner == nil {
		ownerUsername = ""
	} else {
		ownerUsername = "@" + owner["username"].(string)
	}
	embedMessage := discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeImage,
		Title: cardInfo["name"].(string),
		Description: fmt.Sprintf("Series: %s", cardInfo["series"]) +
			fmt.Sprintf("\nOwner: %s", ownerUsername) +
			fmt.Sprintf("\nCharacter ID: %s", cardInfo["characterId"]),
		Image: &discordgo.MessageEmbedImage{URL: "attachment://" + artworkImage.Name},
	}
	response := discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{&embedMessage},
		Files:  []*discordgo.File{artworkImage},
	}
	return response
}

func ViewSpecifiedCard(userId string, searchFilter string) ([]discordgo.MessageSend, int) {
	var matchingCardsMessages []discordgo.MessageSend
	// try to find card by characterId
	card, err := mongo.GetCharInfos(searchFilter)
	if err == nil {
		viewedArtworkId := card["viewedArtwork"].(int32)
		response := ViewCardMessage(card, int(viewedArtworkId))
		newResponseContent := fmt.Sprintf("<@%s> 1 card matched your search.\n", userId) + response.Content
		response.Content = newResponseContent
		matchingCardsMessages = append(matchingCardsMessages, response)
		return matchingCardsMessages, 1
	}
	// try to find card by its name
	cards, err := mongo.SearchCardByName(searchFilter)
	amountMatched := len(cards)
	if err == nil && amountMatched > 0 {
		for _, cardInfo := range cards {
			viewedArtworkId := cardInfo["viewedArtwork"].(int32)
			response := ViewCardMessage(cardInfo, int(viewedArtworkId))
			response.Content = fmt.Sprintf("<@%s> Found %d cards matching your criteria.",
				userId, amountMatched)
			matchingCardsMessages = append(matchingCardsMessages, response)
		}
		return matchingCardsMessages, amountMatched
	}
	// try to find card by its series
	cards, err = mongo.SearchCardBySeries(searchFilter)
	amountMatched = len(cards)
	if err == nil && amountMatched > 0 {
		for _, cardInfo := range cards {
			viewedArtworkId := cardInfo["viewedArtwork"].(int32)
			response := ViewCardMessage(cardInfo, int(viewedArtworkId))
			response.Content = fmt.Sprintf("<@%s> Found %d cards matching your criteria.",
				userId, amountMatched)
			matchingCardsMessages = append(matchingCardsMessages, response)
		}
		return matchingCardsMessages, amountMatched
	}
	// if no match then respond a fail
	response := discordgo.MessageSend{
		Content: fmt.Sprintf("<@%s>\n", userId) +
			fmt.Sprintf("Couldn't find card matching `%s` criteria.", searchFilter),
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
