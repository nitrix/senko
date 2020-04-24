package app

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"strings"
)

const Version = "v0.0.10"

var modules []Module

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

	discord.AddHandler(discordReadyHandler)
	discord.AddHandler(discordMessageHandler)

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

func discordReadyHandler(s *discordgo.Session, event *discordgo.Ready) {
	err := s.UpdateStatus(0, "")
	if err != nil {
		log.Fatalln("unable to update status:", err)
	}
}

func discordMessageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignore your own messages.
	if message.Author.ID == session.State.User.ID {
		return
	}

	if message.Content[0] == '!' {
		args := strings.Split(message.Content, " ")
		args[0] = strings.TrimPrefix(args[0], "!")

		channel, err := session.Channel(message.ChannelID)
		if err != nil {
			log.Println("unable to lookup channel by id:", err)
			return
		}

		req := Request {}
		req.Args = args
		req.NSFW = channel.NSFW

		resp := Response {}
		resp.channelId = message.ChannelID
		resp.session = session

		for _, module := range modules {
			err := module.Dispatch(req, resp)

			if err != nil {
				_, err := session.ChannelMessageSend(message.ChannelID, err.Error())
				if err != nil {
					log.Println("Unable to send channel message:", err)
				}
			}
		}
	}
}

func RegisterModule(module Module) {
	modules = append(modules, module)
}