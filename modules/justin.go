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

func (j *Justin) Commands() []discordgo.ApplicationCommand {
	return []discordgo.ApplicationCommand{
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

				before := func() {
					content := fmt.Sprintf("Saying `%s`.", text)
					s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
						Content: &content,
					})
				}

				after := func() {
					content := fmt.Sprintf("Said `%s`.", text)
					s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
						Content: &content,
					})
				}

				say(s, i.GuildID, text, before, after)
			}
		}
	}
}

func say(s *discordgo.Session, guildID string, text string, before, after func()) {
	voiceConnection, ok := s.VoiceConnections[guildID]
	if !ok {
		return
	}

	go func() {
		body := bytes.NewBufferString(text)

		response, err := http.Post(synthesizeEndpoint, "application/text", body)
		if err != nil {
			return
		}

		before()

		mixer := helpers.VoiceConnectionToMixer(voiceConnection)
		stopper, err := mixer.PlayReader(response.Body, true)
		if err != nil {
			return
		}

		func() {
			stopper.Wait()
			after()
		}()
	}()
}
