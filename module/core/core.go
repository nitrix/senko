package core

import (
	"senko/app"
)

type Core struct{}

func (c *Core) Load() error { return nil }

func (c *Core) Unload() error { return nil }

func (c *Core) OnCommand(event *app.CommandEvent) error {
	if event.Content == "help" {
		return event.Reply("For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md")
	}

	if event.Content == "quit" {
		event.Quit()
	}

	return nil
}

func (c *Core) OnMessageCreated(event *app.MessageCreatedEvent) error { return nil }
