package core

import (
	"senko/app"
)

type Core struct{
	stops map[string]chan struct{}
}

func (c *Core) OnRegister(store *app.Store) {
	c.stops = make(map[string]chan struct{})
}

func (c *Core) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		if _, ok := e.Match("help"); ok {
			return gateway.SendMessage(e.ChannelID, "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md")
		}

		if _, ok := e.Match("quit"); ok {
			gateway.Stop()
			return nil
		}

		if vars, ok := e.Match("voice join <channel>"); ok {
			channelId, err := gateway.FindChannelByName(e.GuildID, vars["channel"])
			if err != nil {
				return err
			}

			return gateway.JoinVoice(e.GuildID, channelId)
		}

		if _, ok := e.Match("voice leave"); ok {
			return gateway.LeaveVoice(e.GuildID)
		}
	}

	return nil
}
