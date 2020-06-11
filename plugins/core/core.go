package core

import (
	"github.com/bwmarrin/discordgo"
	"senko/app"
)

type Plugin struct {}

func (p *Plugin) Save() error { return nil }
func (p *Plugin) Restore() error { return nil }

func (p Plugin) OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) error {
	if message.Content == "!version" {
		_, err := session.ChannelMessageSend(message.ChannelID, app.Version)
		return err
	}

	if message.Content == "!help" {
		msg := "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md"
		_, err := session.ChannelMessageSend(message.ChannelID, msg)
		return err
	}

	if message.Content == "!state save" {
		err := app.SaveState()
		if err != nil {
			return err
		}

		_, err = session.ChannelMessageSend(message.ChannelID, "State saved.")
		return err
	}

	if message.Content == "!state restore" {
		err := app.RestoreState()
		if err != nil {
			return err
		}

		_, err = session.ChannelMessageSend(message.ChannelID, "State restored.")
		return err
	}

	return nil
}