package modules

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Autojoin struct {
	targets map[string]string // Guild ID -> Channel ID
}

func (a *Autojoin) OnLoad() error {
	a.targets = make(map[string]string)
	a.targets["628921721680035840"] = "628921722929807370"
	return nil
}

func (a *Autojoin) OnUnload() error {
	return nil
}

func (a *Autojoin) Commands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "autojoin",
			Description: "Enable or disable the ability to automatically join/leave a channel when occupied.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "on",
					Description: "Enables automatically joining/leaving a voice channel when occupied.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "channel",
							Description: "The name of the voice channel to join.",
							Type:        discordgo.ApplicationCommandOptionChannel,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildVoice,
							},
							Required: true,
						},
					},
				},
				{
					Name:        "off",
					Description: "Disables automatically joining/leaving of a voice channel when occupied.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
}

func (a *Autojoin) OnReady(s *discordgo.Session, r *discordgo.Ready) {}

func (a *Autojoin) OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	target := a.targets[g.ID]

	occupied := false

	for _, voiceState := range g.VoiceStates {
		if voiceState.ChannelID == target {
			occupied = true
			break
		}
	}

	if occupied {
		_, err := s.ChannelVoiceJoin(g.ID, target, false, false)
		if err != nil {
			log.Println("Couldn't autojoin voice channel:", err)
		}
	}
}

func (a *Autojoin) OnVoiceStateUpdate(s *discordgo.Session, i *discordgo.VoiceStateUpdate) {
	if i.BeforeUpdate != nil {
		if voiceConnection, ok := s.VoiceConnections[i.GuildID]; ok {
			if voiceConnection.ChannelID == i.BeforeUpdate.ChannelID {
				if !a.channelIsOccupied(s, i.GuildID, i.BeforeUpdate.ChannelID) {
					voiceConnection.Disconnect()
				}
			}
		}
	}

	if i.ChannelID != "" && a.targets[i.GuildID] == i.ChannelID {
		if voiceConnection, ok := s.VoiceConnections[i.GuildID]; ok {
			if voiceConnection.ChannelID == i.ChannelID {
				// Already in the right channel
				return
			}
		}

		_, err := s.ChannelVoiceJoin(i.GuildID, i.ChannelID, false, false)
		if err != nil {
			log.Println("Couldn't autojoin voice channel:", err)
		}
	}
}

func (a *Autojoin) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (a *Autojoin) channelIsOccupied(s *discordgo.Session, guildID string, channelID string) bool {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return false
	}

	for _, voiceState := range guild.VoiceStates {
		if voiceState.ChannelID == channelID && voiceState.UserID != s.State.User.ID {
			return true
		}
	}

	return false
}
