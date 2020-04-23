package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
)

const Version = "v0.0.7"

func main() {
	// go webServer()
	discordBot()
}

func webServer() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/assets/downloads/", http.StripPrefix("/assets/downloads/", fs))

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatalln("Unable to listen on port 80:", err)
	}
}

func discordBot() {
	token := getToken("DISCORD_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Unable to create Discord bot:", err)
	}

	discord.AddHandler(readyHandler)
	discord.AddHandler(messageHandler)

	// This call is non-blocking and spawns goroutines that then uses the handlers that were registered.
	// Those handlers seems to also be called in their own goroutines (processed asynchronously).
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