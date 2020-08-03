package core

import (
	"fmt"
	"senko/app"
	"strings"
)

type Core struct{
	stops map[string]chan struct{}
}

func (c *Core) OnLoad(store *app.Store) {
	c.stops = make(map[string]chan struct{})
}

func (c *Core) OnUnload(store *app.Store) {}

func (c *Core) OnEvent(event *app.Event) error {
	if event.Content == "help" {
		return event.Reply("For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md")
	}

	if event.Content == "quit" {
		err := event.Reply("Quitting...")
		if err != nil {
			return err
		}

		return event.Quit()
	}

	if strings.HasPrefix(event.Content, "voice join ") {
		channelName := strings.TrimPrefix(event.Content, "voice join ")

		channel, err := event.FindChannelByName(channelName)
		if err != nil {
			return event.Reply(fmt.Sprintf("Unknown channel named `%s`", channelName))
		}

		return event.JoinVoice(channel.ID)
	}

	if event.Content == "voice leave" {
		fmt.Println(event.CurrentVoiceChannelId())
		return event.LeaveVoice(event.CurrentVoiceChannelId())
	}

	return nil
}