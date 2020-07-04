package app

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"os"
	"strings"
)

const DownloadDir = "downloads"

type handlerOnCommand = func (command string) error
type handlerOnMessageCreated = func (session *discordgo.Session, message *discordgo.MessageCreate) error

type App struct {
	modules   []Module
	quitChan  chan bool
	webServer *http.Server

	handlersForOnCommand        []handlerOnCommand
	handlersForOnMessageCreated []handlerOnMessageCreated
}

func (a *App) Run() {
	a.doOnceAtStartup()

	go a.runWebServer()
	go a.runDiscordBot()
	a.waitForQuitSignal()

	a.doOnceAtCleanup()
}

func (a *App) RegisterModule(module Module) error {
	a.modules = append(a.modules, module)

	return nil
}

func (a *App) doOnceAtStartup() {
	// Create the download directory if it's missing.
	_ = os.Mkdir(DownloadDir, 0644)

	// The termination channel.
	a.quitChan = make(chan bool)

	// Handlers.
	a.handlersForOnCommand = make([]handlerOnCommand, 0)
	a.handlersForOnMessageCreated = make([]handlerOnMessageCreated, 0)

	// FIXME: Terminate on some signals, for kubernetes and stuff.
	/*
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc
	*/

	a.loadModules()
}

func (a *App) loadModules() {
	for _, module := range a.modules {
		err := module.Load()
		if err != nil {
			// TODO: Handle module load failures?
		}
	}
}

func (a *App) unloadModules() {
	for _, module := range a.modules {
		err := module.Unload()
		if err != nil {
			// TODO: Handle module unload failures?
		}
	}
}

func (a *App) doOnceAtCleanup() {
	a.unloadModules()
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
	token := GetToken("DISCORD_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Unable to create Discord bot:", err)
	}

	// Event handlers
	discord.AddHandler(func (session *discordgo.Session, message *discordgo.MessageCreate) {
		for _, module := range a.modules {
			err = module.OnMessageCreated(&MessageCreatedEvent{
				Content: message.Content,
				AuthorId: message.Author.ID,
				session: session,
				message: message,
			})

			if strings.HasPrefix(message.Content, "!") {
				err = module.OnCommand(&CommandEvent{
					app: a,
					Content: strings.TrimPrefix(message.Content, "!"),
					session: session,
					channelId: message.ChannelID,
					guildId: message.GuildID,
				})
			}

			if err != nil {
				log.Println(err)
			}
		}
	})

	// This call is non-blocking and spawns goroutines that then uses the handlers that were registered.
	// Those handlers seems to also be called in their own goroutines (processed asynchronously).
	err = discord.Open()
	if err != nil {
		log.Fatalln("Error connecting to Discord:", err)
	}

	a.waitForQuitSignal()

	_ = discord.Close()
}
