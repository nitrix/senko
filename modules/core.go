package modules

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Core struct{}

func (c *Core) OnLoad() error {
	return nil
}

func (c *Core) Commands() []discordgo.ApplicationCommand {
	return []discordgo.ApplicationCommand{
		{
			Name:        "join",
			Description: "Joins the specified voice channel or the one you're currently in when omitted.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "channel",
					Description: "The name of the voice channel to join.",
					Type:        discordgo.ApplicationCommandOptionChannel,
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildVoice,
					},
					Required: false,
				},
			},
		},
		{
			Name:        "leave",
			Description: "Leaves the current voice channel.",
		},
	}
}

func (c *Core) OnUnload() error {
	return nil
}

func (c *Core) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	if data.Name == "join" {
		var channel *discordgo.Channel

		for _, option := range data.Options {
			if option.Name == "channel" {
				channel = option.ChannelValue(s)
			}
		}

		if len(data.Options) == 0 {
			voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
			if err != nil {
				return
			}
			channel, err = s.Channel(voiceState.ChannelID)
			if err != nil {
				return
			}
		}

		s.ChannelVoiceJoin(i.GuildID, channel.ID, false, false)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6, // "Only you can see this"
				Content: fmt.Sprintf("Joined channel <#%s>.", channel.ID),
			},
		})
	}

	if data.Name == "leave" {
		if voiceConnection, ok := s.VoiceConnections[i.GuildID]; ok {
			channelID := voiceConnection.ChannelID

			voiceConnection.Disconnect()

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6, // "Only you can see this"
					Content: fmt.Sprintf("Left channel <#%s>.", channelID),
				},
			})
		}
	}
}
