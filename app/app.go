package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type App struct {
	// TODO: Can that be behind methods?
	Envs     map[string]string

	gateway  Gateway
	store    Store

	mutex    sync.RWMutex
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

	a.mutex.Lock()
	a.modules = append(a.modules, module)
	a.mutex.Unlock()
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
		"GOOGLE_APPLICATION_CREDENTIALS": "",
		"EXTERNAL_URL_PREFIX": "http://localhost",
		"DISCORD_TOKEN": "",
		"WOLFRAM_TOKEN": "",
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
	go func() {
		switch e := event.(type) {
		case EventCommand:
			log.Println("Command:", e.Content)
		}

		a.mutex.RLock()
		defer a.mutex.RUnlock()

		wg := sync.WaitGroup{}

		for _, module := range a.modules {
			wg.Add(1)
			go func(m Module) {
				err := m.OnEvent(&a.gateway, event)
				if err != nil {
					log.Println(err)

					// Report errors publicly for commands when they fail.
					switch e := event.(type) {
					case EventCommand:
						_ = a.gateway.SendMessage(e.ChannelID, "Error: " + err.Error())
					}
				}

				wg.Done()
			}(module)
		}

		wg.Wait()
	}()
}