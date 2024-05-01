package bot

import (
	"ZakuBot/bot/commands"
	"ZakuBot/mongo"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type dropReactionInfo struct {
	MessageID string
	UserID    string
}
type burnReactionInfo struct {
	MessageID string
	UserID    string
	CharID    []string
}

var trackedDropMessages = make(map[string]dropReactionInfo)
var trackedBurnMessages = make(map[string]burnReactionInfo)

func Run(BotToken string) {

	// create a session
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// add an event handler
	discord.AddHandler(receivedMessage)

	// Add a handler for the MessageReactionAdd event
	discord.AddHandler(addedReaction)

	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions

	// open session
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// keep bot running until there is NO os interruption (ctrl + C)
	fmt.Println("Bot running....")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-c

	// Cleanly close down the Discord session.
	err = discord.Close()
	if err != nil {
		return
	}

}

func receivedMessage(session *discordgo.Session, message *discordgo.MessageCreate) {

	// prevent bot responding to its own messages
	if message.Author.ID == session.State.User.ID {
		return
	}

	// respond to user message if it matches a case
	switch message.Content {
	case "zreg", "zregister", "Zreg", "Zregister":
		response := commands.Register(message.Author.ID, message.Author.GlobalName, message.Author.Username)
		session.ChannelMessageSend(message.ChannelID, response)
		return

	case "zd", "zdrop", "zdraw", "Zd", "Zdrop", "Zdraw":
		//Make sure user is registered
		if !commands.IsRegistered(message.Author.ID) {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("<@%s> You must register first. Type `zreg` to register.",
				message.Author.ID))
			return
		}
		// Check if last draw was less than 5 minutes ago
		if !commands.CanUserDrop(message.Author.ID) {
			lastDropTime := commands.GetUserDropTime(message.Author.ID)
			currentTime := time.Now().Unix()
			timeDiff := 300 - (currentTime - lastDropTime)
			fmt.Println(timeDiff)
			if timeDiff < 60 {
				session.ChannelMessageSend(message.ChannelID,
					fmt.Sprintf("<@%s> You can't drop right now, must wait %ds.", message.Author.ID, timeDiff))
			} else {
				amountMinutes := timeDiff / 60
				amountSeconds := timeDiff - amountMinutes*60
				session.ChannelMessageSend(message.ChannelID,
					fmt.Sprintf("<@%s> You can't drop right now, must wait %dmin%ds.",
						message.Author.ID, amountMinutes, amountSeconds))
			}
			return
		}
		cards, err := mongo.DrawCards()
		if err != nil {
			session.ChannelMessageSend(message.ChannelID, "Error drawing cards")
		}
		combinedImgPath, err := commands.CombinedCardsFile(cards)
		if err != nil {
			log.Fatal(err)
		}
		combinedImageFile, err := os.Open(combinedImgPath)
		if err != nil {
			log.Fatal(err)
		}
		messageToSend := discordgo.MessageSend{
			Content: fmt.Sprintf("<@%s>, Here is your drop.", message.Author.ID),
			Files: []*discordgo.File{
				{
					Reader: combinedImageFile,
					Name:   "cards.png",
				},
			},
		}
		sentMessage, _ := session.ChannelMessageSendComplex(message.ChannelID, &messageToSend)

		// Add reactions to the message
		emojis := []string{"1️⃣", "2️⃣", "3️⃣"}
		for _, emoji := range emojis {
			session.MessageReactionAdd(sentMessage.ChannelID, sentMessage.ID, emoji)
		}
		// Track the message
		trackedDropMessages[sentMessage.ID] = dropReactionInfo{MessageID: sentMessage.ID, UserID: message.Author.ID}
		mongo.SetUserDropTimer(message.Author.ID)
		time.AfterFunc(15*time.Second, func() {
			delete(trackedDropMessages, sentMessage.ID)

			// Remove bot reaction before drawing winners
			for _, emoji := range emojis {
				session.MessageReactionRemove(sentMessage.ChannelID, sentMessage.ID, emoji, sentMessage.Author.ID)
			}

			winners := commands.ChooseWinners(session, sentMessage, message.Author.ID, cards)
			commands.AddDropsToInventories(winners)
			commands.NotifyWinners(session, sentMessage.ChannelID, winners)

			// Remove all reactions and change drop message
			commands.MessageCleanup(session, sentMessage)
		})
		return

	case "zm", "zmoney", "Zm", "Zmoney":
		//Make sure user is registered
		if !commands.IsRegistered(message.Author.ID) {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("<@%s> You must register first. Type `zreg` to register.",
				message.Author.ID))
			return
		}
		userMoney := commands.GetUserMoney(message.Author.ID)
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("<@%s> You have %d € in your balance.",
			message.Author.ID, userMoney))
		return

	case "zv", "zview", "Zv", "Zview":
		//Make sure user is registered
		if !commands.IsRegistered(message.Author.ID) {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("<@%s> You must register first. Type `zreg` to register.",
				message.Author.ID))
			return
		}
		lastCardDroppedEmbed := commands.ViewLastCardDropped(message.Author.ID)
		_, _ = session.ChannelMessageSendComplex(message.ChannelID, &lastCardDroppedEmbed)
		return

	case "zb", "zburn", "Zb", "Zburn":
		//Make sure user is registered
		if !commands.IsRegistered(message.Author.ID) {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("<@%s> You must register first. Type `zreg` to register.",
				message.Author.ID))
			return
		}
		confirmationBurnMessage := commands.ConfirmationBurnLastCard(message.Author.ID)
		sentMessage, _ := session.ChannelMessageSendComplex(message.ChannelID, &confirmationBurnMessage)
		card := mongo.GetLastCardDropped(message.Author.ID)
		charId := card["characterId"].(string)
		// Add reactions to the message
		emojis := []string{"👍", "👎"}
		for _, emoji := range emojis {
			session.MessageReactionAdd(sentMessage.ChannelID, sentMessage.ID, emoji)
		}
		// Track the message
		trackedBurnMessages[sentMessage.ID] = burnReactionInfo{
			MessageID: sentMessage.ID,
			UserID:    message.Author.ID,
			CharID:    []string{charId},
		}
		time.AfterFunc(15*time.Second, func() {
			delete(trackedBurnMessages, sentMessage.ID)
		})
		return
	}
}

func addedReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Check if the reaction is made by the bot
	if r.UserID == s.State.User.ID {
		return
	}
	// Check if the reaction is on a tracked message for drop
	_, okDrop := trackedDropMessages[r.MessageID]
	if okDrop {
		// Define the emojis
		emojis := []string{"1️⃣", "2️⃣", "3️⃣"}

		// Remove the other reactions
		for _, emoji := range emojis {
			if emoji != r.Emoji.Name {
				s.MessageReactionRemove(r.ChannelID, r.MessageID, emoji, r.UserID)
			}
		}
	}

	// Check if the reaction is on a tracked message for drop
	_, okBurn := trackedBurnMessages[r.MessageID]
	if okBurn {
		authorId := trackedBurnMessages[r.MessageID].UserID
		// Don't consider emoji if not by author
		if r.UserID != authorId {
			s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
			return
		}
		if r.Emoji.Name == "👍" {
			cardsIds := trackedBurnMessages[r.MessageID].CharID
			successText := commands.SuccessfullyBurnt(authorId, cardsIds)
			s.ChannelMessageEdit(r.ChannelID, r.MessageID, successText)
			delete(trackedBurnMessages, r.MessageID)
		}
		if r.Emoji.Name == "👎" {
			s.ChannelMessageDelete(r.ChannelID, r.MessageID)
			delete(trackedBurnMessages, r.MessageID)
		}
	}
}
