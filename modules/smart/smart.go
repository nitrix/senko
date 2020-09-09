package smart

import (
	"github.com/Krognol/go-wolfram"
	"senko/app"
)

type Smart struct{}

func (s *Smart) OnRegister(store *app.Store) {}

func (s *Smart) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		if vars, ok := e.Match("{who,what,where,when,why,how} <something>"); ok {
			question := vars["who,what,where,when,why,how"] + " " + vars["something"]
			return s.smart(gateway, e.GuildID, question)
		}
	}

	return nil
}

func (s *Smart) smart(gateway *app.Gateway, guildId app.GuildID, something string) error {
	c := &wolfram.Client{
		AppID: gateway.GetEnv("WOLFRAM_TOKEN"),
	}

	//answer, err := c.GetShortAnswerQuery(something, wolfram.Imperial, 10000)
	answer, err := c.GetSpokentAnswerQuery(something, wolfram.Imperial, 10000)
	if err != nil {
		return err
	}

	// TODO: Reply on voice or text, based on the event?
	// TODO: Oh god, need to refactor the event/gateway.
	return gateway.Say(guildId, answer)
}
