package eggplant

import (
	"fmt"
	"senko/app"
	"strings"
	"sync"
)

type Eggplant struct {
	victims      []string
	victimsMutex sync.Mutex
}

func (e *Eggplant) OnLoad(store *app.Store) {
	victims := store.Read("eggplant.victims")

	switch t := victims.(type) {
	case []string:
		e.victims = t
	default:
		e.victims = make([]string, 0)
	}
}

func (e *Eggplant) OnUnload(store *app.Store) {
	store.Write("eggplant.victims", e.victims)
}

func (e *Eggplant) OnEvent(event *app.Event) error {
	// Handling eggplant commands.
	if event.Kind == app.CommandEvent {
		if !strings.HasPrefix(event.Content, "eggplant ") {
			return nil
		}

		parts := strings.Split(strings.TrimPrefix(event.Content, "eggplant "), " ")

		if len(parts) == 2 && parts[0] == "enable" {
			return e.enable(event, parts[1])
		}

		if len(parts) == 2 && parts[0] == "disable" {
			return e.disable(event, parts[1])
		}
	}

	// Handling the eggplant prank.
	if event.Kind == app.MessageCreatedEvent {
		if strings.Contains(strings.ToLower(event.Content), "o.o") && e.isVictim(event.UserId) {
			_ = event.React("🍆") // Eggplant
			_ = event.React("🙄") // Rolling eyes
		}
	}

	return nil
}

func (e *Eggplant) enable(event *app.Event, nick string) error {
	userId, err := event.ResolveNick(nick)
	if err != nil {
		return event.Reply(fmt.Sprintf("Nick `%s` not found.", nick))
	}

	if e.isVictim(userId) {
		return event.Reply(fmt.Sprintf("Eggplant already enabled for `%s`.", nick))
	}

	e.addVictim(userId)

	return event.Reply(fmt.Sprintf("Eggplant enabled for `%s`.", nick))
}

func (e *Eggplant) disable(event *app.Event, nick string) error {
	userId, err := event.ResolveNick(nick)
	if err != nil {
		return event.Reply(fmt.Sprintf("Nick `%s` not found.", nick))
	}

	if !e.isVictim(userId) {
		return event.Reply(fmt.Sprintf("Eggplant isn't enabled for `%s`.", nick))
	}

	e.removeVictim(userId)

	return event.Reply(fmt.Sprintf("Eggplant disabled for `%s`.", nick))
}

func (e *Eggplant) isVictim(userId string) bool {
	e.victimsMutex.Lock()
	defer e.victimsMutex.Unlock()

	for _, victim := range e.victims {
		if userId == victim {
			return true
		}
	}

	return false
}

func (e *Eggplant) addVictim(userId string) {
	e.victimsMutex.Lock()
	defer e.victimsMutex.Unlock()

	e.victims = append(e.victims, userId)
}

func (e *Eggplant) removeVictim(userId string) {
	e.victimsMutex.Lock()
	defer e.victimsMutex.Unlock()

	for i, victim := range e.victims {
		if victim == userId {
			e.victims[i] = e.victims[len(e.victims)-1]
			e.victims = e.victims[:len(e.victims)-1]
		}
	}
}