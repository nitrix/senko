package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"os"
)

func main() {
	go webServer()
	discordBot()
}

func webServer() {
	_ = os.RemoveAll("downloads")
	_ = os.MkdirAll("downloads", 0777)
	_ = os.Chdir("downloads")

	fs := http.FileServer(http.Dir("downloads"))
	http.Handle("/downloads", fs)

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatalln("Unable to listen on port 80:", err)
	}
}

func discordBot() {
	token := loadConfig("DISCORD_TOKEN")

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