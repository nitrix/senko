package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type App struct {
	// TODO: Can that be behind methods?
	Envs     map[string]string

	gateway  Gateway
	store    Store
	modules  []Module
}

func (a *App) Run() error {
	err := a.startup()
	if err != nil {
		return err
	}

	// Blocking
	err = a.gateway.Run(a)
	if err != nil {
		fmt.Println(err)
	}

	err = a.cleanup()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) RegisterModule(module Module) {
	module.OnRegister(&a.store)
	a.modules = append(a.modules, module)
}

func (a *App) startup() error {
	// Create the config directory if it's missing.
	_ = os.Mkdir("config", 0644)

	// Restore the store.
	a.store.filepath = "config/storage.gob"
	err := a.store.restore()
	if err != nil {
		return err
	}

	// Environment variables.
	a.loadEnvironmentVariables()

	// Terminate on some signals, for kubernetes and stuff.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		<-sc
		a.gateway.Stop()
	}()

	return nil
}

func (a *App) cleanup() error {
	err := a.store.save()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) loadEnvironmentVariables() {
	defaultEnvs := map[string]string{
		"EXTERNAL_URL_PREFIX": "http://localhost",
		"DISCORD_TOKEN": "",
	}

	if a.Envs == nil {
		a.Envs = make(map[string]string)
	}

	for k, v := range defaultEnvs {
		// 1. Use the value from the environment when available.
		str := strings.TrimSpace(os.Getenv(k))
		if str != "" {
			a.Envs[k] = str
			continue
		}

		// Use the default hard-coded value otherwise.
		a.Envs[k] = v
	}
}

func (a *App) BroadcastEvent(event interface{}) {
	// TODO: Use goroutines + waitgroup for the dispatching of requests?
	// TODO: Might need a mutex in case the modules become dynamic?
	// TODO: We should prevent being able to call RegisterModule while the app is running.
	// TODO: Do something about the awful error handling?

	for _, module := range a.modules {
		err := module.OnEvent(&a.gateway, event)
		if err != nil {
			log.Println(err)

			// Report errors publicly for commands when they fail.
			switch e := event.(type) {
			case EventCommand:
				_ = a.gateway.SendMessage(e.ChannelID, "Error: " + err.Error())
			}
		}
	}
}