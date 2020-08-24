package discord

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
	"senko/app"
	"senko/requests"
	"senko/responses"
	"strings"
)

type Discord struct {
	quit chan struct{}
	session *discordgo.Session
}

type discordContext struct {
	channelId string
}

func (d *Discord) Name() string {
	return "discord"
}

func (d *Discord) OnRegister() {
	d.quit = make(chan struct{})
}

func (d *Discord) Run(a *app.App) error {
	token := a.Envs["DISCORD_TOKEN"]

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Unable to create Discord bot:", err)
	}

	// Event handlers
	d.setupDiscordEventHandlers(a, session)

	d.session = session

	// This call is non-blocking and spawns goroutines that then uses the handlers that were registered.
	// Those handlers seems to also be called in their own goroutines (processed asynchronously).
	err = session.Open()
	if err != nil {
		log.Fatalln("Error connecting to Discord:", err)
	}

	d.quit = make(chan struct{})
	<- d.quit

	_ = session.Close()
	return nil
}

func (d *Discord) Stop() {
	close(d.quit)
}

func (d *Discord) setupDiscordEventHandlers(a *app.App, discord *discordgo.Session) {
	discord.AddHandler(func (session *discordgo.Session, message *discordgo.MessageCreate) {
		var request interface{}

		if strings.HasPrefix(message.Content, "!") {
			request = requests.EventCommand{
				Author: "discord/" + message.Author.ID,
				Content: strings.TrimPrefix(message.Content, "!"),
			}
		} else {
			request = requests.EventMessageCreated{
				Author: "discord/" + message.Author.ID,
				Message: message.Content,
			}
		}

		err := a.BroadcastRequest(request, func (response interface{}) error {
			return d.onResponse(response, discordContext {
				channelId: message.ChannelID,
			})
		})

		if err != nil {
			log.Println(err)
		}
	})

	discord.AddHandler(func (session *discordgo.Session, ready *discordgo.Ready) {
		/*
		for _, guild := range discord.State.Guilds {
			event := Event{
				Kind:    ReadyEvent,
				session: discord,
				guildId: guild.ID,
				app: a,
			}

			a.BroadcastRequest(&event)
		}
		*/
	})

	discord.AddHandler(func (session *discordgo.Session, state *discordgo.VoiceStateUpdate) {
		/*
		// Exclude itself.
		if state.UserID == session.State.User.ID {
			return
		}

		a.mutex.Lock()
		previous := a.currentVoiceChannel[state.UserID]
		if state.ChannelID != "" {
			a.currentVoiceChannel[state.UserID] = state.ChannelID
		} else {
			delete(a.currentVoiceChannel, state.UserID)
		}
		a.mutex.Unlock()

		event := Event{
			channelId: state.ChannelID,
			guildId: state.GuildID,
			User: User{
				gatewayName: discordGateway.Discord{}.Name(),
				identifier:  state.UserID,
			},
			session: session,
			app: a,
		}

		// Left voice channel
		if previous != state.ChannelID && previous != "" {
			event.Kind = VoiceLeaveEvent
			event.channelId = previous

			a.BroadcastRequest(&event)
		}

		// Joined voice channel
		if previous != state.ChannelID && state.ChannelID != "" {
			event.Kind = VoiceJoinEvent

			a.BroadcastRequest(&event)
		}
		*/
	})

	discord.AddHandler(func (session *discordgo.Session, create *discordgo.GuildCreate) {
		/*
		for _, voiceState := range create.Guild.VoiceStates {
			if voiceState.UserID != session.State.User.ID {
				a.mutex.Lock()
				a.currentVoiceChannel[voiceState.UserID] = voiceState.ChannelID
				a.mutex.Unlock()

				event := Event{
					Kind: CurrentlyInVoiceEvent,
					User: User{
						gatewayName: discordGateway.Discord{}.Name(),
						identifier: voiceState.UserID,
					},
					guildId: voiceState.GuildID,
					channelId: voiceState.ChannelID,
					session: session,
					app: a,
				}

				a.BroadcastRequest(&event)
			}
		}
		*/
	})
}

func (d *Discord) onResponse(response interface{}, context discordContext) error {
	switch a := response.(type) {
	case responses.Reply:
		_, err := d.session.ChannelMessageSend(context.channelId, a.Content)
		return err
	default:
		return errors.New("action not supported by the Discord gateway")
	}
}