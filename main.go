package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/nitrix/senko/helpers"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Unable to create the connection for Discord", err)
	}

	session.LogLevel = discordgo.LogDebug

	session.AddHandler(onReady)
	session.AddHandler(onMessageCreate)
	session.AddHandler(onMessageDelete)
	session.AddHandler(onVoiceStateUpdate)
	session.AddHandler(onInteractionCreate)

	log.Println("Connecting to Discord...")

	err = session.Open()
	if err != nil {
		log.Fatalln("Unable to connect to Discord", err)
	}

	reloadApplicationCommands(session)

	waitForTerminationSignal()

	session.Close()
}

func waitForTerminationSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("Ready")
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {}

func onMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {}

func onVoiceStateUpdate(s *discordgo.Session, u *discordgo.VoiceStateUpdate) {}

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if data.Name == "join" {
		var channel *discordgo.Channel

		for _, option := range data.Options {
			if option.Name == "channel" {
				channel = option.ChannelValue(s)
			}
		}

		if len(data.Options) == 0 {
			voiceState, err := s.State.VoiceState(i.GuildID, i.User.ID)
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

	if data.Name == "test" {
		if voiceConnection, ok := s.VoiceConnections[i.GuildID]; ok {
			filename := "oh-yeah.mp3"

			stopper, err := helpers.PlayFile(voiceConnection, filename, true)
			if err != nil {
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6, // "Only you can see this"
					Content: fmt.Sprintf(":arrow_forward: Playing `%s`...", filename),
				},
			})

			func() {
				<-stopper
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: fmt.Sprintf("Played `%s`.", filename),
				})
			}()
		}
	}
}

func reloadApplicationCommands(s *discordgo.Session) {
	for _, guild := range s.State.Guilds {
		// Delete all existing commands.
		commands, err := s.ApplicationCommands(s.State.User.ID, guild.ID)
		if err != nil {
			log.Fatalln("Unable to get existing application commands:", err)
		}

		for _, command := range commands {
			fmt.Println("Removing old command", command.Name)
			_ = s.ApplicationCommandDelete(s.State.User.ID, guild.ID, command.ID)
		}

		// Replace them with newer ones.
		commands = []*discordgo.ApplicationCommand{
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
			{
				Name:        "test",
				Description: "Plays a test sound.",
			},
		}

		for _, command := range commands {
			fmt.Println("Registering new command", command.Name)
			_, err := s.ApplicationCommandCreate(s.State.User.ID, guild.ID, command)
			if err != nil {
				log.Fatalln("Unable to register application command:", err)
			}
		}
	}
}
