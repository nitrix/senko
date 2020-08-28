package autojoin

import (
	"senko/app"
	"sync"
)

type Autojoin struct {
	mutex     sync.RWMutex
	channels  map[app.GuildID]app.ChannelID
}

func (a *Autojoin) OnRegister(store *app.Store) {
	store.Link("autojoin.channels", &a.channels, make(map[app.GuildID]app.ChannelID))
}

func (a *Autojoin) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		if vars, ok := e.Match("autojoin enable <channel>"); ok {
			channelID, err := gateway.FindChannelByName(e.GuildID, vars["channel"])
			if err != nil {
				return err
			}

			a.mutex.Lock()
			a.channels[e.GuildID] = channelID
			a.mutex.Unlock()

			// Pretend we were in the channel to trigger an autojoin.
			return a.autojoin(gateway, e.GuildID, channelID)
		}

		if _, ok := e.Match("autojoin disable"); ok {
			a.mutex.Lock()
			channelID := a.channels[e.GuildID]
			delete(a.channels, e.GuildID)
			a.mutex.Unlock()

			// Pretend we were in the channel to trigger an autojoin.
			return a.autojoin(gateway, e.GuildID, channelID)
		}
	case app.EventVoiceJoin:
		return a.autojoin(gateway, e.GuildID, e.ChannelID)
	case app.EventVoiceLeave:
		return a.autojoin(gateway, e.GuildID, e.ChannelID)
	case app.EventVoiceAlready:
		return a.autojoin(gateway, e.GuildID, e.ChannelID)
	}

	return nil
}

func (a *Autojoin) autojoin(gateway *app.Gateway, guildID app.GuildID, channelID app.ChannelID) error {
	// Making the entire scope critical ensures that join and leave events aren't processed
	// concurrently, causing strange reconnects while in a room due to the order of messages
	// arriving in.
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	autojoinChannelID := a.channels[guildID]

	// Check if the channel is in use.
	inUse := gateway.IsChannelInUse(autojoinChannelID)

	// If autojoin is enabled, then automatically join the channel when it's in use.
	if autojoinChannelID == channelID && inUse {
		_ = gateway.JoinVoice(guildID, channelID)
		return nil  // Special case, ignore errors here on failure.
	}

	// If there's an autojoin channel configured, then leave that channel if it's no longer in use.
	if autojoinChannelID != "" && !inUse {
		return gateway.LeaveVoice(guildID, autojoinChannelID)
	}

	return nil
}