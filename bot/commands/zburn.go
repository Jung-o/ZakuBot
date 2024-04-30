package commands

import (
	"ZakuBot/mongo"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func ConfirmationBurnLastCard(userId string) discordgo.MessageSend {
	lastCardDropped := mongo.GetLastCardDropped(userId)
	if lastCardDropped == nil {
		noCardsMessage := discordgo.MessageSend{
			Content: fmt.Sprintf("<@%s> You have no cards in your inventory, go drop some with `zdrop`.", userId),
		}
		return noCardsMessage
	}
	confirmationMessage := ViewLastCardDropped(userId)
	confirmationText := fmt.Sprintf("<@%s> Are you sure you want to burn %s from %s ?",
		userId, lastCardDropped["name"], lastCardDropped["series"])
	confirmationMessage.Content = confirmationText
	return confirmationMessage
}

func SuccessfullyBurnt(userId string, cards []string) string {
	for _, charId := range cards {
		mongo.RemoveCardFromInventory(userId, charId)
	}
	mongo.ChangeBalance(userId, 10*len(cards))
	successBurnText := fmt.Sprintf("<@%s> Successfully burnt %d cards.", userId, len(cards))
	return successBurnText
}
