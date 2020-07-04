package eggplant

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"senko/app"
	"strings"
	"sync"
)

type Eggplant struct {
	victims      []string
	victimsMutex sync.Mutex
}

func (e *Eggplant) Load() error {
	e.victimsMutex.Lock()
	defer e.victimsMutex.Unlock()

	bytes, err := json.Marshal(e.victims)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("config/victims.txt", bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (e *Eggplant) Unload() error {
	e.victimsMutex.Lock()
	defer e.victimsMutex.Unlock()

	file, err := os.Open("config/victims.txt")
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	return decoder.Decode(&e.victims)
}

func (e *Eggplant) OnCommand(event *app.CommandEvent) error {
	if !strings.HasPrefix(event.Content, "!eggplant ") {
		return nil
	}

	parts := strings.Split(strings.TrimPrefix(event.Content, "eggplant "), " ")

	if len(parts) == 2 && parts[0] == "enable" {
		return e.enable(event, parts[1])
	}

	if len(parts) == 2 && parts[0] == "disable" {
		return e.disable(event, parts[1])
	}

	return nil
}

func (e *Eggplant) OnMessageCreated(event *app.MessageCreatedEvent) error {
	if strings.Contains(strings.ToLower(event.Content), "o.o") && e.isVictim(event.AuthorId) {
		_ = event.React("🍆") // Eggplant
		_ = event.React("🙄") // Rolling eyes
	}

	return nil
}

func (e *Eggplant) enable(event *app.CommandEvent, nick string) error {
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

func (e *Eggplant) disable(event *app.CommandEvent, nick string) error {
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