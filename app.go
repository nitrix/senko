package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type App struct {
	modules []Module
}

func (a *App) RegisterModule(m Module) {
	a.modules = append(a.modules, m)
}

func (a *App) Run() error {
	log.Println("Starting up...")

	token := os.Getenv("DISCORD_TOKEN")

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	// session.LogLevel = discordgo.LogDebug

	log.Println("Loading modules...")

	for _, module := range a.modules {
		err := module.OnLoad()
		if err != nil {
			return err
		}
	}

	log.Println("Registering handlers...")

	ready := sync.WaitGroup{}
	ready.Add(1)

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for _, module := range a.modules {
			module.OnReady(s, r)
		}

		ready.Done()
	})

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		for _, module := range a.modules {
			module.OnInteractionCreate(s, i)
		}
	})

	session.AddHandler(func(s *discordgo.Session, u *discordgo.VoiceStateUpdate) {
		for _, module := range a.modules {
			module.OnVoiceStateUpdate(s, u)
		}
	})

	session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		for _, module := range a.modules {
			module.OnGuildCreate(s, g)
		}
	})

	log.Println("Opening Discord session...")

	err = session.Open()
	if err != nil {
		log.Fatalln("Unable to connect to Discord", err)
	}

	ready.Wait()

	log.Println("Ready!")

	log.Println("Registering global commands...")

	commands := make([]*discordgo.ApplicationCommand, 0)
	for _, module := range a.modules {
		commands = append(commands, module.Commands()...)
	}

	_, err = session.ApplicationCommandBulkOverwrite(session.State.User.ID, "", make([]*discordgo.ApplicationCommand, 0)) // FIXME: Remove global ones.
	if err != nil {
		return err
	}

	_, err = session.ApplicationCommandBulkOverwrite(session.State.User.ID, "628921721680035840", commands) // FIXME: Specific guild.
	if err != nil {
		return err
	}

	log.Println("Running...")

	a.waitForTerminationSignal()

	log.Println("Shutting down...")

	log.Println("Unloading modules...")

	for _, module := range a.modules {
		err = module.OnUnload()
		if err != nil {
			return err
		}
	}

	log.Println("Closing Discord sesion...")

	err = session.Close()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) waitForTerminationSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
