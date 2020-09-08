package alias

import "senko/app"

type Alias struct {
	// TODO: mutex
	aliases map[app.GuildID]Mappings
}

type Mappings map[string]string

func (c *Alias) OnRegister(store *app.Store) {
	store.Link("alias.aliases", &c.aliases, make(map[app.GuildID]Mappings))
}

func (c *Alias) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		// Adding an alias.
		if vars, ok := e.Match("alias add <original> <replacement>"); ok {
			if c.aliases[e.GuildID] == nil {
				c.aliases[e.GuildID] = make(Mappings)
			}

			c.aliases[e.GuildID][vars["original"]] = vars["replacement"]
			return gateway.SendMessage(e.ChannelID, "Alias added.")
		}

		// TODO: Removing an alias.
		// TODO: Listing aliases.

		// Processing aliases.
		for original, replacement := range c.aliases[e.GuildID] {
			if vars, ok := e.Match(original); ok {
				newEvent := e
				newEvent.Content = replacement
				newEvent.Replace(vars)
				gateway.BroadcastEvent(newEvent)
			}
		}
	}

	return nil
}