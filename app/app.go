package app

import (
	"log"
	"os"
	"os/signal"
	"senko/responses"
	"strings"
	"syscall"
)

type App struct {
	Envs     map[string]string
	modules  []Module
	gateways []Gateway
	store    *Store
	quit     chan struct{}
}

func (a *App) Run() error {
	err := a.startup()
	if err != nil {
		return err
	}

	for _, gateway := range a.gateways {
		go func(gw Gateway) {
			_ = gw.Run(a)
		}(gateway)
	}

	a.wait()

	err = a.cleanup()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) RegisterModule(module Module) {
	module.OnRegister(a.store)
	a.modules = append(a.modules, module)
}

func (a *App) RegisterGateway(gateway Gateway) {
	gateway.OnRegister()
	a.gateways = append(a.gateways, gateway)
}

func (a *App) startup() error {
	// Create the config directory if it's missing.
	_ = os.Mkdir("config", 0644)

	// The termination channel.
	a.quit = make(chan struct{})

	// Restore the store.
	a.store = NewStore("config/storage.gob")
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
		a.Stop()
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

func (a *App) wait() {
	<- a.quit
}

func (a *App) Stop() {
	for _, gateway := range a.gateways {
		gateway.Stop()
	}

	close(a.quit)
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

func (a *App) BroadcastRequest(request interface{}, reply func(response interface{}) error) error {
	// TODO: Use goroutines + waitgroup for the dispatching of requests?
	// TODO: Might need a mutex in case the modules become dynamic?
	// TODO: We should prevent being able to call RegisterModule while the app is running.
	// TODO: Do something about the awful error handling?

	// Quit is a special case that the application intercepts.
	replyMiddleware := func(response interface{}) error {
		_, ok := response.(responses.Quit)
		if ok {
			a.Stop()
			return nil
		}

		return reply(response)
	}

	// Otherwise, broadcast the request to every module.
	for _, module := range a.modules {
		err := module.OnRequest(request, replyMiddleware)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}