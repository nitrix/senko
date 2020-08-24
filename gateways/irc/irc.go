package irc

import "senko/app"

type IRC struct {}

func (g *IRC) Name() string {
	return "irc"
}

func (g *IRC) OnRegister() {}

func (g *IRC) Run(app *app.App) error {
	return nil
}

func (g *IRC) Stop() {}
