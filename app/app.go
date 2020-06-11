package app

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const Version = "v0.0.10"

var plugins []Plugin

func Run() {
	go webServer()
	discordBot()
}

func webServer() {
	fs := http.FileServer(http.Dir("downloads"))
	http.Handle("/downloads/", http.StripPrefix("/downloads/", fs))

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatalln("Unable to listen on port 80:", err)
	}
}

func discordBot() {
	token := GetToken("DISCORD_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Unable to create Discord bot:", err)
	}

	for _, plugin := range plugins {
		discord.AddHandler(func (p Plugin) func(session *discordgo.Session, message *discordgo.MessageCreate) {
			return func(session *discordgo.Session, message *discordgo.MessageCreate) {
				err = p.OnMessageCreate(session, message)
				if err != nil {
					log.Println(err)
				}
			}
		}(plugin))
	}

	// This call is non-blocking and spawns goroutines that then uses the handlers that were registered.
	// Those handlers seems to also be called in their own goroutines (processed asynchronously).
	err = discord.Open()
	if err != nil {
		log.Fatalln("Error connecting to Discord:", err)
	}
	defer func() {
		_ = discord.Close()
	}()

	// Wait for exit signal.
	// That also comes from their documentation, as much as I hate it.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func RegisterPlugin(plugin Plugin) {
	plugins = append(plugins, plugin)
}