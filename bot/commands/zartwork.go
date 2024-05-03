package commands

import (
	"ZakuBot/mongo"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func ViewAllCharArtworks(userId string, searchFilter string) ([]discordgo.MessageSend, int) {
	// try to find card by characterId
	artworkAmount, err := mongo.GetAmountArtworks(searchFilter)
	if err == nil && artworkAmount > 0 {
		var cardArtworksMessages []discordgo.MessageSend
		cardInfo, _ := mongo.GetCharInfos(searchFilter)
		for artworkId := range artworkAmount {
			response := ViewCardMessage(cardInfo, artworkId+1)
			response.Content = fmt.Sprintf("<@%s> Found %d artworks for character with `%s` id.",
				userId, artworkAmount, searchFilter)
			cardArtworksMessages = append(cardArtworksMessages, response)
		}
		return cardArtworksMessages, artworkAmount
	}
	// try to find card by its name
	cards, err := mongo.SearchCardByName(searchFilter)
	amountMatched := len(cards)
	if err == nil && amountMatched == 1 {
		card := cards[0]
		cardId := card["characterId"].(string)
		return ViewAllCharArtworks(userId, cardId) // tip to shorten func length
	} else if err == nil && amountMatched > 1 {
		response := discordgo.MessageSend{
			Content: fmt.Sprintf("Too much cards matched `%s` criteria, try to be more precise.", searchFilter),
		}
		return []discordgo.MessageSend{response}, 0
	}
	// if no match then respond a fail
	response := discordgo.MessageSend{
		Content: fmt.Sprintf("Couldn't find character with `%s` criteria.", searchFilter),
	}
	return []discordgo.MessageSend{response}, 0

}
