package core

import (
	"github.com/bwmarrin/discordgo"
	"senko/app"
)

type Plugin struct {}

func (p Plugin) OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) error {
	if message.Content == "!version" {
		return p.version(session, message.ChannelID)
	}

	if message.Content == "!help" {
		return p.help(session, message.ChannelID)
	}

	return nil
}

func (p Plugin) version(session *discordgo.Session, channelId string) error {
	_, err := session.ChannelMessageSend(channelId, app.Version)
	return err
}

func (p Plugin) help(session *discordgo.Session, channelId string) error {
	msg := "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md"
	_, err := session.ChannelMessageSend(channelId, msg)
	return err
}