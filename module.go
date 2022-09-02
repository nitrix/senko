package main

import "github.com/bwmarrin/discordgo"

type Module interface {
	OnLoad() error
	OnUnload() error

	Commands() []discordgo.ApplicationCommand
	OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate)
}
