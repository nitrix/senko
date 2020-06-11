package app

import "github.com/bwmarrin/discordgo"

type Plugin interface {
	Save() error
	Restore() error
	OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) error
}
