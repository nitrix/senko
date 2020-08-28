package app

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Gateway struct {
	quit chan struct{}
	session *discordgo.Session
	app *App

	// Below are all protected by the mutex.

	mutex               sync.RWMutex
	mixers              map[GuildID]*Mixer
	currentVoiceChannel map[UserID]ChannelID
}

func (g *Gateway) Run(a *App) error {
	g.quit = make(chan struct{})

	token := a.Envs["DISCORD_TOKEN"]

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Unable to create Gateway bot:", err)
	}

	g.app = a
	g.session = session
	g.mixers = make(map[GuildID]*Mixer)
	g.currentVoiceChannel = make(map[UserID]ChannelID)

	// Event handlers
	g.setupReadyEventHandler(a)
	g.setupMessageCreateEventHandler(a)
	g.setupVoiceEventHandler(a)

	// This call is non-blocking and spawns goroutines that then uses the handlers that were registered.
	// Those handlers seems to also be called in their own goroutines (processed asynchronously).
	err = session.Open()
	if err != nil {
		log.Fatalln("Error connecting to Gateway:", err)
	}

	g.quit = make(chan struct{})
	<- g.quit

	_ = g.session.Close()

	return nil
}

func (g *Gateway) Stop() {
	/*
	g.session.Lock()
	for _, voiceConnection := range g.session.VoiceConnections {
		g.session.Unlock()
		_ = voiceConnection.Disconnect()
		g.session.Lock()
	}
	g.session.Unlock()
	*/

	close(g.quit)
}

func (g *Gateway) setupMessageCreateEventHandler(a *App) {
	g.session.AddHandler(func (s *discordgo.Session, message *discordgo.MessageCreate) {
		var event interface{}

		if strings.HasPrefix(message.Content, "!") {
			event = EventCommand{
				UserID:  UserID(message.Author.ID),
				ChannelID: ChannelID(message.ChannelID),
				GuildID: GuildID(message.GuildID),
				Content: strings.TrimPrefix(message.Content, "!"),
			}
		} else {
			event = EventMessageCreated{
				MessageID: MessageID(message.ID),
				UserID: UserID(message.Author.ID),
				ChannelID: ChannelID(message.ChannelID),
				Content: message.Content,
			}
		}

		a.BroadcastEvent(event)
	})
}

func (g *Gateway) setupReadyEventHandler(a *App) {
	g.session.AddHandler(func(s *discordgo.Session, ready *discordgo.Ready) {
		event := EventReady{}

		a.BroadcastEvent(event)
	})
}

func (g *Gateway) setupVoiceEventHandler(a *App) {
	g.session.AddHandler(func (s *discordgo.Session, state *discordgo.VoiceStateUpdate) {
		// Exclude itself.
		if state.UserID == s.State.User.ID {
			return
		}

		// Keep track of which voice channel users are in.
		g.mutex.Lock()
		previous := g.currentVoiceChannel[UserID(state.UserID)]
		if state.ChannelID != "" {
			g.currentVoiceChannel[UserID(state.UserID)] = ChannelID(state.ChannelID)
		} else {
			delete(g.currentVoiceChannel, UserID(state.UserID))
		}
		g.mutex.Unlock()

		// Left voice channel
		if string(previous) != state.ChannelID && previous != "" {
			event := EventVoiceLeave{
				UserID: UserID(state.UserID),
				ChannelID: previous,
				GuildID: GuildID(state.GuildID),
			}

			a.BroadcastEvent(event)
		}

		// Joined voice channel
		if string(previous) != state.ChannelID && state.ChannelID != "" {
			event := EventVoiceJoin{
				UserID: UserID(state.UserID),
				ChannelID: ChannelID(state.ChannelID),
				GuildID: GuildID(state.GuildID),
			}

			a.BroadcastEvent(event)
		}
	})

	g.session.AddHandler(func (s *discordgo.Session, create *discordgo.GuildCreate) {
		for _, voiceState := range create.Guild.VoiceStates {
			if voiceState.UserID != s.State.User.ID {
				// Keep track of which channel users are in.
				g.mutex.RLock()
				g.currentVoiceChannel[UserID(voiceState.UserID)] = ChannelID(voiceState.ChannelID)
				g.mutex.RUnlock()

				event := EventVoiceAlready{
					UserID: UserID(voiceState.UserID),
					ChannelID: ChannelID(voiceState.ChannelID),
					GuildID: GuildID(voiceState.GuildID),
				}

				a.BroadcastEvent(event)
			}
		}
	})
}

func (g *Gateway) SendMessage(channelID ChannelID, message string) error {
	_, err := g.session.ChannelMessageSend(string(channelID), message)
	return err
}

func (g *Gateway) FindChannelByName(guildID GuildID, name string) (ChannelID, error) {
	channels, err := g.session.GuildChannels(string(guildID))
	if err != nil {
		return "", fmt.Errorf("error while looking up channel: %w", err)
	}

	for _, channel := range channels {
		if channel.Name == name {
			return ChannelID(channel.ID), nil
		}
	}

	return "", fmt.Errorf("unable to find channel `%s`", name)
}

func (g *Gateway) JoinVoice(guildId GuildID, channelId ChannelID) error {
	// Leave previous voice channel, whatever it was.
	err := g.LeaveVoiceAny(guildId)
	if err != nil {
		return err
	}

	// Join the new channel.
	voiceConnection, err := g.session.ChannelVoiceJoin(string(guildId), string(channelId), false, false)
	if err != nil {
		return err
	}

	// Do the audio processing in a goroutine.
	mixer := &Mixer{}

	g.mutex.Lock()
	g.mixers[guildId] = mixer
	g.mutex.Unlock()

	go mixer.handleRealtime(g, voiceConnection)

	return nil
}

func (g *Gateway) LeaveVoiceAny(guildId GuildID) error {
	g.session.Lock()
	voiceConnection := g.session.VoiceConnections[string(guildId)]
	g.session.Unlock()

	if voiceConnection != nil {
		return voiceConnection.Disconnect()
	}

	return nil
}

func (g *Gateway) LeaveVoice(guildId GuildID, channelID ChannelID) error {
	g.session.Lock()
	voiceConnection := g.session.VoiceConnections[string(guildId)]
	g.session.Unlock()

	if voiceConnection != nil && voiceConnection.ChannelID == string(channelID) {
		return voiceConnection.Disconnect()
	}

	return nil
}

func (g *Gateway) IsChannelInUse(channelId ChannelID) bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	for _, currentChannelId := range g.currentVoiceChannel {
		if currentChannelId == channelId {
			return true
		}
	}

	return false
}

func (g *Gateway) BroadcastEvent(event interface{}) {
	g.app.BroadcastEvent(event)
}

func (g *Gateway) PlayAudioFile(guildID GuildID, filepath string) (chan struct{}, error) {
	mixer := g.mixers[guildID]
	if mixer == nil {
		return nil, errors.New("no active mixer available")
	}

	stopper := mixer.Play(filepath)

	return stopper, nil
}

func (g *Gateway) SendFile(channelID ChannelID, path string) error {
	file, err := os.Open(path)
	defer func() {
		_ = file.Close()
	}()

	if err != nil {
		return err
	}

	message := discordgo.MessageSend{
		Files: []*discordgo.File{
			{
				Name:        filepath.Base(path),
				ContentType: mime.TypeByExtension(filepath.Ext(path)),
				Reader:      file,
			},
		},
	}

	_, err = g.session.ChannelMessageSendComplex(string(channelID), &message)

	return err
}

func (g *Gateway) SendEmbed(channelID ChannelID, embed discordgo.MessageEmbed) error {
	_, err := g.session.ChannelMessageSendEmbed(string(channelID), &embed)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gateway) GetEnv(name string) string {
	return g.app.Envs[name]
}

func (g *Gateway) ResolveNick(guildID GuildID, nick string) (UserID, error) {
	members, err := g.session.GuildMembers(string(guildID), "", 1000)
	if err != nil {
		return UserID(""), err
	}

	for _, member := range members {
		if member.User.Username == nick || member.Nick == nick {
			return UserID(member.User.ID), nil
		}
	}

	return UserID(""), errors.New("nick not found")
}

func (g *Gateway) React(channelID ChannelID, messageID MessageID, emoji string) error {
	return g.session.MessageReactionAdd(string(channelID), string(messageID), emoji)
}