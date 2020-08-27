package eggplant

import (
	"fmt"
	"senko/app"
	"strings"
	"sync"
)

type Eggplant struct {
	victims      []app.UserID
	victimsMutex sync.Mutex
}

func (e *Eggplant) OnRegister(store *app.Store) {
	store.Link("eggplant.victims", &e.victims, make([]app.UserID, 0))
}

func (e *Eggplant) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch v := event.(type) {
	// Handling eggplant commands.
	case app.EventCommand:
		if vars, ok := v.Match("eggplant enable <username>"); ok {
			return e.enable(gateway, v.ChannelID, v.GuildID, vars["username"])
		}

		if vars, ok := v.Match("eggplant disable <username>"); ok {
			return e.disable(gateway, v.ChannelID, v.GuildID, vars["username"])
		}

	// Handling the eggplant prank.
	case app.EventMessageCreated:
		if strings.Contains(strings.ToLower(v.Content), "o.o") && e.isVictim(v.UserID) {
			_ = gateway.React(v.ChannelID, v.MessageID, "🍆") // Eggplant
			_ = gateway.React(v.ChannelID, v.MessageID, "🙄") // Rolling eyes
		}
	}

	return nil
}

func (e *Eggplant) enable(gateway *app.Gateway, channelID app.ChannelID, guildID app.GuildID, nick string) error {
	user, err := gateway.ResolveNick(guildID, nick)
	if err != nil {
		return gateway.SendMessage(channelID, fmt.Sprintf("Nick `%s` not found.", nick))
	}

	if e.isVictim(user) {
		return gateway.SendMessage(channelID, fmt.Sprintf("Eggplant already enabled for `%s`.", nick))
	}

	e.addVictim(user)

	return gateway.SendMessage(channelID, fmt.Sprintf("Eggplant enabled for `%s`.", nick))
}

func (e *Eggplant) disable(gateway *app.Gateway, channelID app.ChannelID, guildID app.GuildID, nick string) error {
	userId, err := gateway.ResolveNick(guildID, nick)
	if err != nil {
		return gateway.SendMessage(channelID, fmt.Sprintf("Nick `%s` not found.", nick))
	}

	if !e.isVictim(userId) {
		return gateway.SendMessage(channelID, fmt.Sprintf("Eggplant isn't enabled for `%s`.", nick))
	}

	e.removeVictim(userId)

	return gateway.SendMessage(channelID, fmt.Sprintf("Eggplant disabled for `%s`.", nick))
}

func (e *Eggplant) isVictim(userID app.UserID) bool {
	e.victimsMutex.Lock()
	defer e.victimsMutex.Unlock()

	for _, victim := range e.victims {
		if userID == victim {
			return true
		}
	}

	return false
}

func (e *Eggplant) addVictim(userID app.UserID) {
	e.victimsMutex.Lock()
	defer e.victimsMutex.Unlock()

	e.victims = append(e.victims, userID)
}

func (e *Eggplant) removeVictim(userID app.UserID) {
	e.victimsMutex.Lock()
	defer e.victimsMutex.Unlock()

	for i, victim := range e.victims {
		if victim == userID {
			e.victims[i] = e.victims[len(e.victims)-1]
			e.victims = e.victims[:len(e.victims)-1]
		}
	}
}