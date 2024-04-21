package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Run(BotToken string) {

	// create a session
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// add a event handler
	discord.AddHandler(receivedMessage)

	// In this example, we only care about receiving message events.
	discord.Identify.Intents = discordgo.IntentsGuildMessages

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

func receivedMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {

	// prevent bot responding to its own messages
	if message.Author.ID == discord.State.User.ID {
		return
	}

	// respond to user message if it contains `!help` or `!bye`
	switch {
	case strings.Contains(message.Content, "zreg"):
		response := "You have been registered successfully!"
		discord.ChannelMessageSend(message.ChannelID, response)
	}
}
