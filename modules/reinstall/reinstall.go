package reinstall

import (
	"fmt"
	"github.com/andanhm/go-prettytime"
	"senko/app"
	"sync"
	"time"
)

type reinstallEntry struct{
	Count int
	Last time.Time
}

type Reinstall struct {
	mutex sync.Mutex
	entries map[app.UserID]reinstallEntry
}

func (e *Reinstall) OnRegister(store *app.Store) {
	store.Link("reinstall.entries", &e.entries, make(map[app.UserID]reinstallEntry, 0))
}

func (e *Reinstall) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch v := event.(type) {
	case app.EventCommand:
		if vars, ok := v.Match("reinstall++ <username>"); ok {
			return e.increment(gateway, v.ChannelID, v.GuildID, vars["username"])
		}

		if vars, ok := v.Match("reinstall reset <username>"); ok {
			return e.reset(gateway, v.ChannelID, v.GuildID, vars["username"])
		}
	}

	return nil
}

func (e *Reinstall) increment(gateway *app.Gateway, channelID app.ChannelID, guildID app.GuildID, nick string) error {
	user, err := gateway.ResolveNick(guildID, nick)
	if err != nil {
		return gateway.SendMessage(channelID, fmt.Sprintf("Nick `%s` not found.", nick))
	}

	count := 0
	last := time.Now()

	e.mutex.Lock()
	entry, ok := e.entries[user]
	if ok {
		count = entry.Count
		last = entry.Last
	}

	count++

	e.entries[user] = reinstallEntry{
		Count: count,
		Last: time.Now(),
	}
	e.mutex.Unlock()

	if count == 1 {
		return gateway.SendMessage(channelID, fmt.Sprintf("`%s` reinstalled Windows for the first time.", nick))
	} else {
		ordinal := ordinalSuffix(count)
		ago := prettytime.Format(last)
		return gateway.SendMessage(channelID, fmt.Sprintf("`%s` reinstalled Windows again for the `%d%s` time. (Last time on `%s`, `%s`)", nick, count, ordinal, last.Format("2006-01-02"), ago))
	}
}

func (e *Reinstall) reset(gateway *app.Gateway, channelID app.ChannelID, guildID app.GuildID, nick string) error {
	user, err := gateway.ResolveNick(guildID, nick)
	if err != nil {
		return gateway.SendMessage(channelID, fmt.Sprintf("Nick `%s` not found.", nick))
	}

	e.mutex.Lock()
	delete(e.entries, user)
	e.mutex.Unlock()

	return gateway.SendMessage(channelID, fmt.Sprintf("Reset count for `%s`.", nick))
}

func ordinalSuffix(count int) string {
	switch count {
	case 1, 21, 31:
		return "st"
	case 2, 22:
		return "nd"
	case 3, 23:
		return "rd"
	}
	return "th"
}