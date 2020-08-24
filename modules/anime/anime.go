package anime

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"senko/app"
	"senko/requests"
	"senko/modules/anime/mal"
	"strings"
	"time"
)

type Anime struct{
	accounts map[app.User]mal.Username
	monitoring map[app.User]bool
}

func (a *Anime) OnRegister(store *app.Store) {
	store.Link("anime.accounts", &a.accounts, make(map[app.User]mal.Username))
	store.Link("anime.monitoring", &a.monitoring, make(map[app.User]bool))
}

func (a *Anime) OnEvent(gateway app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case requests.EventCommand:
		e.Match("anime search <name>", func(vars map[string]string) {
			return a.search(name)
		})
	}

	/*
	if !strings.HasPrefix(event.Content, "anime ") {
				return nil
			}

			command := strings.TrimPrefix(event.Content, "anime ")
			parts := strings.Split(command, " ")

			if len(parts) > 1 && parts[0] == "search" {
				name := strings.Join(parts[1:], " ")
				return a.search(event, name)
			}

			if len(parts) == 2 && parts[0] == "link" {
				return a.setMapping(event, parts[1])
			}

			if len(parts) == 2 && parts[0] == "notify" && parts[1] == "enable" {
				return a.setMonitoring(event, true)
			}

			if len(parts) == 2 && parts[0] == "notify" && parts[1] == "disable" {
				return a.setMonitoring(event, false)
			}
	 */

	return nil
}

func (a *Anime) setMapping(event *app.Event, username string) error {
	// TODO: Verify username.
	a.accounts[event.User] = mal.Username(username)
	return event.Reply("Account linked.")
}

func (a *Anime) setMonitoring(event *app.Event, enable bool) error {
	if enable {
		if a.accounts[event.User] != "" {
			a.monitoring[event.User] = true
			return event.Reply("Enabled notifications.")
		} else {
			return event.Reply("You must first link your account with `!anime link <username>` to be able to enable notifications.")
		}
	} else {
		delete(a.monitoring, event.User)
		return event.Reply("Disable notifications.")
	}
}
