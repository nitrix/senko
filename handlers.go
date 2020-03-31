package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func readyHandler(s *discordgo.Session, event *discordgo.Ready) {
	err := s.UpdateStatus(0, "")
	if err != nil {
		log.Fatalln("unable to update status:", err)
	}
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore your own messages.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content[0] == '!' {
		parts := strings.Split(m.Content, " ")

		command := strings.TrimPrefix(parts[0], "!")
		args := parts[1:]

		var err error

		switch command {
		case "help": err = commandHelp(s, m.ChannelID)
		case "anime": err = commandAnime(s, m.ChannelID, args)
		case "gif": err = commandGif(s, m.ChannelID, args)
		}

		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, err.Error())
			if err != nil {
				log.Println("Unable to send channel message:", err)
			}
		}
	}
}
