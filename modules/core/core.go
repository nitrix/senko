package core

import (
	"senko/app"
	"senko/requests"
	"senko/responses"
)

type Core struct{
	stops map[string]chan struct{}
}

func (c *Core) OnRegister(store *app.Store) {
	c.stops = make(map[string]chan struct{})
}

func (c *Core) OnRequest(request interface{}, reply app.ReplyFunc) error {
	switch r := request.(type) {
	case requests.EventCommand:
		if _, ok := r.Match("help"); ok {
			return reply(responses.Reply{
				Content: "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md",
			})
		}

		if _, ok := r.Match("quit"); ok {
			return reply(responses.Quit{})
		}

		if vars, ok := r.Match("voice join <channel>"); ok {
			return reply(responses.JoinVoice{
				Location: vars["channel"],
			})
		}

		if _, ok := r.Match("voice leave"); ok {
			return reply(responses.LeaveVoice{})
		}
	}

	return nil
}