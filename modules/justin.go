package modules

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/nitrix/senko/helpers"
)

const synthesizeEndpoint = "https://justin.nitrix.me/synthesize"

type Justin struct{}

func (j *Justin) OnLoad() error {
	return nil
}

func (j *Justin) OnUnload() error {
	return nil
}

func (j *Justin) Commands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "say",
			Description: "Says the specified text out loud in the voice channel currently in.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "text",
					Description: "The text to say out loud.",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	}
}

func (j *Justin) OnReady(s *discordgo.Session, r *discordgo.Ready) {}

func (j *Justin) OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {}

func (j *Justin) OnVoiceStateUpdate(s *discordgo.Session, i *discordgo.VoiceStateUpdate) {}

func (j *Justin) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	if data.Name == "say" {
		for _, option := range data.Options {
			if option.Name == "text" {
				text := option.StringValue()

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   1 << 6, // "Only you can see this"
						Content: fmt.Sprintf("Synthesizing `%s`...", text),
					},
				})

				body := bytes.NewBufferString(text)

				response, err := http.Post(synthesizeEndpoint, "application/text", body)
				if err != nil {
					return
				}

				voiceConnection, ok := s.VoiceConnections[i.GuildID]
				if ok {
					mixer := helpers.VoiceConnectionToMixer(voiceConnection)

					content := fmt.Sprintf("Saying `%s`.", text)
					s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
						Content: &content,
					})

					stopper, err := mixer.PlayReader(response.Body, true)
					if err != nil {
						return
					}

					stopper.Wait()
				}

				content := fmt.Sprintf("Said `%s`.", text)
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
			}
		}
	}
}
