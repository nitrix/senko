package alias

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"senko/app"
	"strings"
	"sync"
)

type Alias struct {
	mutex sync.RWMutex
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
			c.mutex.Lock()
			defer c.mutex.Unlock()

			if c.aliases[e.GuildID] == nil {
				c.aliases[e.GuildID] = make(Mappings)
			}

			c.aliases[e.GuildID][vars["original"]] = vars["replacement"]
			return gateway.SendMessage(e.ChannelID, "Alias added.")
		}

		// Removing an alias.
		if vars, ok := e.Match("alias remove <alias>"); ok {
			c.mutex.Lock()
			defer c.mutex.Unlock()

			alias := vars["alias"]

			aliases := c.aliases[e.GuildID]
			if aliases == nil {
				return fmt.Errorf("alias not found")
			}

			_, ok := aliases[alias]
			if !ok {
				return fmt.Errorf("alias not found")
			}

			delete(aliases, alias)

			return gateway.SendMessage(e.ChannelID, fmt.Sprintf("Alias `%s` removed.", alias))
		}

		// Listing aliases.
		if _, ok := e.Match("alias list"); ok {
			c.mutex.RLock()
			defer c.mutex.RUnlock()

			aliases := c.aliases[e.GuildID]
			if aliases == nil {
				return gateway.SendMessage(e.ChannelID, "No aliases.")
			}

			builder := strings.Builder{}
			table := tablewriter.NewWriter(&builder)
			table.SetHeader([]string{"Alias", "Replacement"})
			for k, v := range aliases {
				table.Append([]string{k, v})
			}
			table.Render()

			return gateway.SendMessage(e.ChannelID, fmt.Sprintf("```\n%s\n```", builder.String()))
		}

		// Processing aliases.
		c.mutex.RLock()
		defer c.mutex.RUnlock()

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
