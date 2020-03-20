package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func main() {
	token := loadToken()

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Unable to create Discord bot:", err)
	}

	discord.AddHandler(readyHandler)
	discord.AddHandler(messageHandler)

	// This call is non-blocking and spawns goroutines that then uses the handlers that were registered.
	err = discord.Open()
	if err != nil {
		log.Fatalln("Error connecting to Discord:", err)
	}
	defer func() {
		_ = discord.Close()
	}()

	// That also comes from their documentation, as much as I hate it.
	waitForExitSignal()
}
