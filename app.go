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

	// token := os.Getenv("DISCORD_TOKEN")
	token := "NjkwNDk3ODY3NjE3NDAyOTgw.GfJwfR.7aOd0fv61l1Fcbml4POm_XGAZ_JZ_8033q3DHU"

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

	session.AddHandler(func(s *discordgo.Session, m *discordgo.Ready) {
		ready.Done()
	})

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		for _, module := range a.modules {
			module.OnInteractionCreate(s, i)
		}
	})

	// session.AddHandler(func (s *discordgo.Session, u *discordgo.VoiceStateUpdate) {})

	log.Println("Opening Discord session...")

	err = session.Open()
	if err != nil {
		log.Fatalln("Unable to connect to Discord", err)
	}

	log.Println("Registering commands...")

	for _, module := range a.modules {
		commands := module.Commands()

		for _, command := range commands {
			log.Println("Registering command", command.Name)

			_, err := session.ApplicationCommandCreate(session.State.User.ID, "", &command)
			if err != nil {
				return err
			}
		}
	}

	ready.Wait()
	log.Println("Ready!")

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
