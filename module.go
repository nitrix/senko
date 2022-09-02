package main

import "github.com/bwmarrin/discordgo"

type Module interface {
	OnLoad() error
	OnUnload() error

	Commands() []*discordgo.ApplicationCommand

	OnReady(s *discordgo.Session, r *discordgo.Ready)
	OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate)
	OnVoiceStateUpdate(s *discordgo.Session, u *discordgo.VoiceStateUpdate)
	OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate)
}
