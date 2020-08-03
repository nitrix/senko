package app

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const DownloadDir = "downloads"

type App struct {
	modules   []Module
	quitChan  chan bool
	webServer *http.Server

	// Store for various persistent configs.
	store *Store

	// Mutex protected
	mutex               sync.Mutex
	currentVoiceChannel map[string]string
	voices              map[string]*Voice
}

func (a *App) Run() {
	err := a.atStartup()
	if err != nil {
		log.Fatalln(err)
	}

	a.setupSignalHandlers()
	go a.runWebServer()
	go a.runDiscordBot()
	a.waitForQuitSignal()

	err = a.atCleanup()
	if err != nil {
		log.Fatalln(err)
	}
}

func (a *App) RegisterModule(module Module) {
	a.modules = append(a.modules, module)
}

func (a *App) atStartup() error {
	// Create the download directory if it's missing.
	_ = os.Mkdir(DownloadDir, 0644)

	// The termination channel.
	a.quitChan = make(chan bool)

	a.store = &Store{}
	err := a.store.restore()
	if err != nil {
		return err
	}

	// Voice stuff
	a.voices = make(map[string]*Voice)
	a.currentVoiceChannel = make(map[string]string)

	for _, module := range a.modules {
		module.OnLoad(a.store)
	}

	return nil
}

func (a *App) atCleanup() error {
	for _, module := range a.modules {
		module.OnUnload(a.store)
	}

	err := a.store.save()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) waitForQuitSignal() {
	<- a.quitChan
}

func (a *App) Stop() {
	if err := a.webServer.Shutdown(context.Background()); err != nil {
		log.Println("Error while shutting down web server:", err)
	}

	close(a.quitChan)
}

func (a *App) runWebServer() {
	fs := http.FileServer(http.Dir(DownloadDir))
	http.Handle(fmt.Sprintf("/%s/", DownloadDir), http.StripPrefix(fmt.Sprintf("/%s/", DownloadDir), fs))

	a.webServer = &http.Server{Addr: ":80"}

	err := a.webServer.ListenAndServe()
	if err != nil {
		log.Fatalln("Unable to listen on port 80:", err)
	}
}

func (a *App) runDiscordBot() {
	token := GetEnvironmentVariable("DISCORD_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Unable to create Discord bot:", err)
	}

	// Event handlers
	a.setupDiscordEventHandlers(discord)

	// This call is non-blocking and spawns goroutines that then uses the handlers that were registered.
	// Those handlers seems to also be called in their own goroutines (processed asynchronously).
	err = discord.Open()
	if err != nil {
		log.Fatalln("Error connecting to Discord:", err)
	}

	a.waitForQuitSignal()

	_ = discord.Close()
}

func (a *App) setupSignalHandlers() {
	// Terminate on some signals, for kubernetes and stuff.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		<-sc
		a.Stop()
	}()
}

func (a *App) setupDiscordEventHandlers(discord *discordgo.Session) {
	// Event handlers
	discord.AddHandler(func (session *discordgo.Session, message *discordgo.MessageCreate) {
		event := Event{
			Kind: MessageCreatedEvent,
			Content: message.Content,
			channelId: message.ChannelID,
			UserId: message.Author.ID,
			guildId: message.GuildID,
			session: session,
			message: message,
			app: a,
		}

		if strings.HasPrefix(message.Content, "!") {
			event.Kind = CommandEvent
			event.Content = strings.TrimPrefix(event.Content, "!")
		}

		a.BroadcastEvent(&event)
	})

	discord.AddHandler(func (session *discordgo.Session, ready *discordgo.Ready) {
		for _, guild := range discord.State.Guilds {
			event := Event{
				Kind:    ReadyEvent,
				session: discord,
				guildId: guild.ID,
				app: a,
			}

			a.BroadcastEvent(&event)
		}
	})

	discord.AddHandler(func (session *discordgo.Session, state *discordgo.VoiceStateUpdate) {
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
			UserId: state.UserID,
			session: session,
			app: a,
		}

		// Left voice channel
		if previous != state.ChannelID && previous != "" {
			event.Kind = VoiceLeaveEvent
			event.channelId = previous

			a.BroadcastEvent(&event)
		}

		// Joined voice channel
		if previous != state.ChannelID && state.ChannelID != "" {
			event.Kind = VoiceJoinEvent

			a.BroadcastEvent(&event)
		}
	})

	discord.AddHandler(func (session *discordgo.Session, create *discordgo.GuildCreate) {
		for _, voiceState := range create.Guild.VoiceStates {
			if voiceState.UserID != session.State.User.ID {
				a.mutex.Lock()
				a.currentVoiceChannel[voiceState.UserID] = voiceState.ChannelID
				a.mutex.Unlock()

				event := Event{
					Kind: CurrentlyInVoiceEvent,
					UserId: voiceState.UserID,
					guildId: voiceState.GuildID,
					channelId: voiceState.ChannelID,
					session: session,
					app: a,
				}

				a.BroadcastEvent(&event)
			}
		}
	})
}

func (a *App) BroadcastEvent(event *Event) {
	// TODO: Use goroutines for this?
	// TODO: Might need mutex when modules becomes dynamic?
	for _, module := range a.modules {
		err := module.OnEvent(event)
		if err != nil {
			log.Println(err)
		}
	}
}