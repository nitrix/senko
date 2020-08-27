package anime

import (
	"senko/app"
	"senko/modules/anime/mal"
)

// TODO
//- `!anime search <name>` Search an anime by name.
//- `!anime link <username>` Link your MyAnimeList account to your Discord account.
//- `!anime notify {enable,disable}` Enable or disable notification when new episodes of anime you watch gets aired.

type Anime struct{
	accounts map[app.UserID]mal.Username
	monitoring map[app.UserID]bool
}

func (a *Anime) OnRegister(store *app.Store) {
	store.Link("anime.accounts", &a.accounts, make(map[app.UserID]mal.Username))
	store.Link("anime.monitoring", &a.monitoring, make(map[app.UserID]bool))
}

func (a *Anime) OnEvent(gateway *app.Gateway, event interface{}) error {
	/*
	switch e := event.(type) {
	case app.EventCommand:
		if vars, ok := e.Match("anime link <username>"); ok {
			return a.account(gateway, vars["username"])
		}

		if vars, ok := e.Match("anime notify enable"); ok {
			return a.monitor(gateway, vars["username"])
		}

		if vars, ok := e.Match("anime notify disable"); ok {
			return a.unmonitor(gateway, vars["username"])
		}
	}
	*/

	return nil
}
