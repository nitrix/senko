package app

import "github.com/bwmarrin/discordgo"

type Plugin interface {
	OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) error
}
